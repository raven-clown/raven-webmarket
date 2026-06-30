package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/raven-clown/raven-webmarket/backend/internal/config"
	"github.com/raven-clown/raven-webmarket/backend/internal/middleware"
	redisstore "github.com/raven-clown/raven-webmarket/backend/internal/redisstore"
	"golang.org/x/oauth2"
)

type Service struct {
	cfg   *config.Config
	db    *sql.DB
	esxDB *sql.DB
	redis *redisstore.Store
	oauth *oauth2.Config
}

type ESXUserRow struct {
	Identifier string
	DiscordID  string
}

func New(cfg *config.Config, db, esxDB *sql.DB, redis *redisstore.Store) *Service {
	return &Service{
		cfg:   cfg,
		db:    db,
		esxDB: esxDB,
		redis: redis,
		oauth: &oauth2.Config{
			ClientID:     cfg.DiscordClientID,
			ClientSecret: cfg.DiscordClientSecret,
			RedirectURL:  cfg.DiscordRedirectURI,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://discord.com/api/oauth2/authorize",
				TokenURL: "https://discord.com/api/oauth2/token",
			},
			Scopes: []string{"identify"},
		},
	}
}

func (s *Service) AuthURL(state string) string {
	return s.oauth.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (s *Service) HandleCallback(ctx context.Context, code string) (string, error) {
	token, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	client := s.oauth.Client(ctx, token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var user struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return "", err
	}
	row, err := s.lookupESXUser(ctx, user.ID)
	if err != nil {
		return "", err
	}
	isAdmin := false
	jwtToken, err := s.issueJWT(user.ID, row.Identifier, isAdmin)
	if err != nil {
		return "", err
	}
	s.ensureProfile(ctx, user.ID, row.Identifier, user.Username)
	sessionKey := "session:" + user.ID
	s.redis.Session.Set(ctx, sessionKey, jwtToken, 24*time.Hour)
	return jwtToken, nil
}

func (s *Service) lookupESXUser(ctx context.Context, discordID string) (*ESXUserRow, error) {
	key := discordKey(discordID)
	queries := []string{
		`SELECT identifier FROM users WHERE discord_id = ? LIMIT 1`,
		`SELECT identifier FROM users WHERE discord = ? LIMIT 1`,
	}
	for _, query := range queries {
		var identifier string
		err := s.esxDB.QueryRowContext(ctx, query, key).Scan(&identifier)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			if isUnknownColumn(err) {
				continue
			}
			return nil, err
		}
		if !isValidSteamIdentifier(identifier) {
			return nil, rejectInvalidIdentifier(identifier)
		}
		return &ESXUserRow{Identifier: identifier, DiscordID: key}, nil
	}
	return nil, fmt.Errorf("login denied: discord account not linked in game database (%s)", key)
}

func (s *Service) isAdmin(discordID string) bool {
	for _, id := range s.cfg.DiscordAdminIDs {
		if id == discordID {
			return true
		}
	}
	var count int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM admin_users WHERE discord_id = ? AND is_active = 1`, discordID).Scan(&count)
	return count > 0
}

func (s *Service) issueJWT(discordID, identifier string, isAdmin bool) (string, error) {
	claims := middleware.UserClaims{
		DiscordID:  discordID,
		Identifier: identifier,
		IsAdmin:    isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) ensureProfile(ctx context.Context, discordID, identifier, displayName string) {
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO user_shop_profiles (discord_id, identifier, display_name)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE display_name = VALUES(display_name), identifier = VALUES(identifier)`,
		discordID, identifier, displayName)
}

func (s *Service) Logout(ctx context.Context, discordID string) {
	s.redis.Session.Del(ctx, "session:" + discordID)
}

func (s *Service) Me(ctx context.Context, discordID string) (map[string]interface{}, error) {
	var profile struct {
		DiscordID           string
		Identifier          string
		DisplayName         sql.NullString
		MonthlyAccumulation float64
		RedeemPoints        float64
		TotalTopupAmount    float64
		TopupCount          int
	}
	err := s.db.QueryRowContext(ctx, `
		SELECT discord_id, identifier, display_name, monthly_accumulation,
			redeem_points, total_topup_amount, topup_count
		FROM user_shop_profiles WHERE discord_id = ?`, discordID).Scan(
		&profile.DiscordID, &profile.Identifier, &profile.DisplayName,
		&profile.MonthlyAccumulation, &profile.RedeemPoints,
		&profile.TotalTopupAmount, &profile.TopupCount)
	if err != nil {
		return nil, err
	}
	if !isValidSteamIdentifier(profile.Identifier) {
		return nil, fmt.Errorf("account not eligible: invalid identifier on file")
	}
	return map[string]interface{}{
		"discord_id":           profile.DiscordID,
		"discord_linked":       discordKey(profile.DiscordID),
		"identifier":           profile.Identifier,
		"display_name":         profile.DisplayName.String,
		"monthly_accumulation": profile.MonthlyAccumulation,
		"redeem_points":        profile.RedeemPoints,
		"total_topup_amount":   profile.TotalTopupAmount,
		"topup_count":          profile.TopupCount,
	}, nil
}

func WriteAuthCookie(w http.ResponseWriter, cfg *config.Config, token string) {
	secure := strings.HasPrefix(cfg.FrontendURL, "https")
	http.SetCookie(w, &http.Cookie{
		Name:     "raven_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}
