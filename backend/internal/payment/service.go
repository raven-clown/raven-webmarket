package payment

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/raven-clown/raven-webmarket/backend/internal/activity"
	"github.com/raven-clown/raven-webmarket/backend/internal/config"
	"github.com/raven-clown/raven-webmarket/backend/internal/lock"
	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	redisstore "github.com/raven-clown/raven-webmarket/backend/internal/redisstore"
	"github.com/raven-clown/raven-webmarket/backend/internal/storage"
)

type Service struct {
	cfg     *config.Config
	db      *sql.DB
	redis   *redisstore.Store
	storage *storage.Service
	log     *activity.Logger
}

func New(cfg *config.Config, db *sql.DB, redis *redisstore.Store, storage *storage.Service, log *activity.Logger) *Service {
	return &Service{cfg: cfg, db: db, redis: redis, storage: storage, log: log}
}

type WebhookPayload struct {
	Ref           string  `json:"ref"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	GatewayRef    string  `json:"gateway_ref"`
	DiscordID     string  `json:"discord_id"`
	SlipBase64    string  `json:"slip_base64"`
}

func (s *Service) ProcessWebhook(ctx context.Context, payload WebhookPayload) error {
	lockKey := "payment:" + payload.Ref
	return lock.WithLock(ctx, s.redis.Session, lockKey, 60*time.Second, func() error {
		var existing int
		_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM topup_transactions WHERE tx_ref = ?`, payload.Ref).Scan(&existing)
		if existing > 0 {
			return nil
		}
		var identifier, displayName sql.NullString
		_ = s.db.QueryRowContext(ctx, `
			SELECT identifier, display_name FROM user_shop_profiles WHERE discord_id = ?`, payload.DiscordID).Scan(&identifier, &displayName)
		if !identifier.Valid {
			return fmt.Errorf("user not found")
		}
		slipURL := ""
		if payload.SlipBase64 != "" {
			url, err := s.storage.UploadSlip(ctx, payload.Ref, payload.SlipBase64)
			if err == nil {
				slipURL = url
			}
		}
		points := payload.Amount * s.cfg.RedeemPointsPerBaht
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		_, err = tx.ExecContext(ctx, `
			INSERT INTO topup_transactions (tx_ref, discord_id, identifier, amount, points_earned, payment_method, gateway_ref, slip_url, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'completed')`,
			payload.Ref, payload.DiscordID, identifier.String, payload.Amount, points,
			payload.PaymentMethod, payload.GatewayRef, slipURL)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			UPDATE user_shop_profiles SET
				monthly_accumulation = monthly_accumulation + ?,
				redeem_points = redeem_points + ?,
				total_topup_amount = total_topup_amount + ?,
				topup_count = topup_count + 1
			WHERE discord_id = ?`,
			payload.Amount, points, payload.Amount, payload.DiscordID)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		if s.log != nil {
			s.log.Write(ctx, "topup", "user", payload.DiscordID, "topup_completed", "transaction", payload.Ref, "", map[string]interface{}{
				"amount": payload.Amount, "points": points, "method": payload.PaymentMethod,
			})
		}
		go s.notifyDiscord(payload, identifier.String, displayName.String, points, slipURL)
		return nil
	})
}

func (s *Service) CreatePending(ctx context.Context, discordID, identifier, method string, amount float64) (string, error) {
	ref := "TX-" + uuid.New().String()[:12]
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO topup_transactions (tx_ref, discord_id, identifier, amount, payment_method, status)
		VALUES (?, ?, ?, ?, ?, 'pending')`, ref, discordID, identifier, amount, method)
	return ref, err
}

func (s *Service) UploadSlip(ctx context.Context, txRef, discordID, base64Data string) (string, error) {
	var owner string
	err := s.db.QueryRowContext(ctx, `SELECT discord_id FROM topup_transactions WHERE tx_ref = ?`, txRef).Scan(&owner)
	if err != nil || owner != discordID {
		return "", fmt.Errorf("transaction not found")
	}
	url, err := s.storage.UploadSlip(ctx, txRef, base64Data)
	if err != nil {
		return "", err
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE topup_transactions SET slip_url = ? WHERE tx_ref = ?`, url, txRef)
	return url, nil
}

func (s *Service) GetHistory(ctx context.Context, discordID string, limit int) ([]models.TopupTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tx_ref, discord_id, identifier, amount, points_earned, payment_method,
			COALESCE(gateway_ref,''), COALESCE(slip_url,''), status, created_at
		FROM topup_transactions WHERE discord_id = ? ORDER BY created_at DESC LIMIT ?`, discordID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.TopupTransaction
	for rows.Next() {
		var t models.TopupTransaction
		if err := rows.Scan(&t.ID, &t.TxRef, &t.DiscordID, &t.Identifier, &t.Amount, &t.PointsEarned,
			&t.PaymentMethod, &t.GatewayRef, &t.SlipURL, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, nil
}

func (s *Service) notifyDiscord(payload WebhookPayload, identifier, displayName string, points float64, slipURL string) {
	if s.cfg.DiscordWebhookURL == "" {
		return
	}
	name := displayName
	if name == "" {
		name = identifier
	}
	embed := map[string]interface{}{
		"title":       "Payment Completed",
		"color":       5763719,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"fields": []map[string]interface{}{
			{"name": "Player", "value": name, "inline": true},
			{"name": "Discord ID", "value": payload.DiscordID, "inline": true},
			{"name": "Amount", "value": fmt.Sprintf("%.2f", payload.Amount), "inline": true},
			{"name": "Points", "value": fmt.Sprintf("%.0f", points), "inline": true},
			{"name": "Reference", "value": payload.Ref, "inline": false},
		},
	}
	if slipURL != "" {
		embed["image"] = map[string]string{"url": slipURL}
	}
	body, _ := json.Marshal(map[string]interface{}{"embeds": []interface{}{embed}})
	req, _ := http.NewRequest(http.MethodPost, s.cfg.DiscordWebhookURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}
