package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/raven-clown/raven-webmarket/backend/internal/config"
	"github.com/raven-clown/raven-webmarket/backend/internal/middleware"
	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	"github.com/raven-clown/raven-webmarket/backend/internal/rbac"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type Account struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	DiscordID   string   `json:"discord_id,omitempty"`
	DisplayName string   `json:"display_name"`
	IsActive    bool     `json:"is_active"`
	Permissions []string `json:"permissions,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

func (s *Service) Login(ctx context.Context, cfg *config.Config, username, password, ip string) (string, *Account, error) {
	var acc Account
	var hash string
	var active int
	var lastLogin sql.NullTime
	var permRaw []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, role, COALESCE(discord_id,''), COALESCE(display_name,''), is_active, last_login_at,
			COALESCE(permissions, '[]')
		FROM admin_accounts WHERE username = ? LIMIT 1`, username).Scan(
		&acc.ID, &acc.Username, &hash, &acc.Role, &acc.DiscordID, &acc.DisplayName, &active, &lastLogin, &permRaw)
	if err == sql.ErrNoRows {
		return "", nil, ErrInvalidCredentials
	}
	if err != nil {
		return "", nil, err
	}
	if active != 1 {
		return "", nil, ErrInvalidCredentials
	}
	_ = json.Unmarshal(permRaw, &acc.Permissions)
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		s.LogDetailed(ctx, "", username, acc.Role, "security", "login_failed", "account", username, `{"reason":"bad_password"}`, ip)
		return "", nil, ErrInvalidCredentials
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE admin_accounts SET last_login_at = UTC_TIMESTAMP() WHERE id = ?`, acc.ID)
	acc.IsActive = true
	if lastLogin.Valid {
		t := lastLogin.Time
		acc.LastLoginAt = &t
	}
	token, err := s.issueAdminJWT(cfg, acc)
	if err != nil {
		return "", nil, err
	}
	s.LogDetailed(ctx, acc.DiscordID, acc.Username, acc.Role, "security", "login_success", "account", acc.Username, "{}", ip)
	return token, &acc, nil
}

func (s *Service) issueAdminJWT(cfg *config.Config, acc Account) (string, error) {
	claims := middleware.AdminClaims{
		Username:    acc.Username,
		Role:        acc.Role,
		AccountID:   acc.ID,
		Permissions: acc.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "admin",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func (s *Service) ListAccounts(ctx context.Context) ([]Account, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, username, role, COALESCE(discord_id,''), COALESCE(display_name,''), is_active, last_login_at,
			COALESCE(permissions, '[]')
		FROM admin_accounts ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Account
	for rows.Next() {
		var a Account
		var active int
		var lastLogin sql.NullTime
		var permRaw []byte
		if err := rows.Scan(&a.ID, &a.Username, &a.Role, &a.DiscordID, &a.DisplayName, &active, &lastLogin, &permRaw); err != nil {
			return nil, err
		}
		a.IsActive = active == 1
		_ = json.Unmarshal(permRaw, &a.Permissions)
		if lastLogin.Valid {
			t := lastLogin.Time
			a.LastLoginAt = &t
		}
		items = append(items, a)
	}
	return items, nil
}

func (s *Service) CreateAccount(ctx context.Context, actor middleware.AdminClaims, ip string, username, password, role, displayName, discordID string, permissions []string) error {
	if !rbac.CanWithPermissions(actor.Role, actor.Permissions, "admin_accounts") {
		return fmt.Errorf("forbidden")
	}
	if role != rbac.RoleAdmin && role != rbac.RoleDevAdmin {
		return fmt.Errorf("invalid role")
	}
	if actor.Role == rbac.RoleAdmin && role == rbac.RoleDevAdmin {
		return fmt.Errorf("forbidden")
	}
	if role == rbac.RoleAdmin && len(permissions) == 0 {
		permissions = rbac.DefaultAdminPermissions
	}
	permJSON, _ := json.Marshal(permissions)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO admin_accounts (username, password_hash, role, display_name, discord_id, permissions)
		VALUES (?, ?, ?, ?, NULLIF(?, ''), ?)`, username, string(hash), role, displayName, discordID, string(permJSON))
	if err != nil {
		return err
	}
	detail, _ := json.Marshal(map[string]string{"username": username, "role": role})
	s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "security", "create_admin_account", "account", username, string(detail), ip)
	return nil
}

func (s *Service) UpdateAccount(ctx context.Context, actor middleware.AdminClaims, ip string, id uint, password, role string, isActive *bool, permissions []string) error {
	if !rbac.CanWithPermissions(actor.Role, actor.Permissions, "admin_accounts") {
		return fmt.Errorf("forbidden")
	}
	var currentRole string
	err := s.db.QueryRowContext(ctx, `SELECT role FROM admin_accounts WHERE id = ?`, id).Scan(&currentRole)
	if err != nil {
		return err
	}
	if actor.Role == rbac.RoleAdmin && currentRole == rbac.RoleDevAdmin {
		return fmt.Errorf("forbidden")
	}
	if role != "" && role != rbac.RoleAdmin && role != rbac.RoleDevAdmin {
		return fmt.Errorf("invalid role")
	}
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			return err
		}
		_, err = s.db.ExecContext(ctx, `UPDATE admin_accounts SET password_hash = ? WHERE id = ?`, string(hash), id)
		if err != nil {
			return err
		}
	}
	if role != "" {
		_, err = s.db.ExecContext(ctx, `UPDATE admin_accounts SET role = ? WHERE id = ?`, role, id)
		if err != nil {
			return err
		}
	}
	if isActive != nil {
		active := 0
		if *isActive {
			active = 1
		}
		_, err = s.db.ExecContext(ctx, `UPDATE admin_accounts SET is_active = ? WHERE id = ?`, active, id)
		if err != nil {
			return err
		}
	}
	if permissions != nil {
		permJSON, _ := json.Marshal(permissions)
		_, err = s.db.ExecContext(ctx, `UPDATE admin_accounts SET permissions = ? WHERE id = ?`, string(permJSON), id)
		if err != nil {
			return err
		}
	}
	detail, _ := json.Marshal(map[string]interface{}{"id": id, "role": role, "password_changed": password != "", "permissions": permissions})
	s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "security", "update_admin_account", "account", fmt.Sprint(id), string(detail), ip)
	return nil
}

func (s *Service) Me(ctx context.Context, username string) (*Account, error) {
	var acc Account
	var active int
	var lastLogin sql.NullTime
	var permRaw []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT id, username, role, COALESCE(discord_id,''), COALESCE(display_name,''), is_active, last_login_at,
			COALESCE(permissions, '[]')
		FROM admin_accounts WHERE username = ? AND is_active = 1`, username).Scan(
		&acc.ID, &acc.Username, &acc.Role, &acc.DiscordID, &acc.DisplayName, &active, &lastLogin, &permRaw)
	if err != nil {
		return nil, err
	}
	acc.IsActive = active == 1
	_ = json.Unmarshal(permRaw, &acc.Permissions)
	if lastLogin.Valid {
		t := lastLogin.Time
		acc.LastLoginAt = &t
	}
	return &acc, nil
}

func (s *Service) LogDetailed(ctx context.Context, discordID, username, role, category, action, targetType, targetID, detail, ip string) {
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO admin_audit_logs (admin_discord_id, admin_username, admin_role, category, action, target_type, target_id, detail, ip_address)
		VALUES (NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?)`,
		discordID, username, role, category, action, targetType, targetID, detail, ip)
}

func (s *Service) Log(ctx context.Context, adminDiscordID, action, targetType, targetID, detail, ip string) {
	s.LogDetailed(ctx, adminDiscordID, adminDiscordID, rbac.RoleAdmin, "system", action, targetType, targetID, detail, ip)
}

func (s *Service) AuditLogsFiltered(ctx context.Context, category, action string, limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 200
	}
	query := `
		SELECT id, COALESCE(admin_discord_id,''), COALESCE(admin_username,''), action, COALESCE(target_type,''), COALESCE(target_id,''),
			COALESCE(detail,'{}'), COALESCE(ip_address,''), created_at
		FROM admin_audit_logs WHERE 1=1`
	args := []interface{}{}
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	if action != "" {
		query += " AND action LIKE ?"
		args = append(args, "%"+action+"%")
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(&l.ID, &l.AdminDiscordID, &l.AdminUsername, &l.Action, &l.TargetType, &l.TargetID,
			&l.Detail, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (s *Service) ActivityLogs(ctx context.Context, category, actorID string, limit int) ([]models.ActivityLog, error) {
	if limit <= 0 {
		limit = 200
	}
	query := `
		SELECT id, category, actor_type, actor_id, action, COALESCE(target_type,''), COALESCE(target_id,''),
			COALESCE(detail,'{}'), COALESCE(ip_address,''), created_at
		FROM activity_logs WHERE 1=1`
	args := []interface{}{}
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	if actorID != "" {
		query += " AND actor_id LIKE ?"
		args = append(args, "%"+actorID+"%")
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []models.ActivityLog
	for rows.Next() {
		var l models.ActivityLog
		if err := rows.Scan(&l.ID, &l.Category, &l.ActorType, &l.ActorID, &l.Action, &l.TargetType, &l.TargetID,
			&l.Detail, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (s *Service) PurchaseLogs(ctx context.Context, discordID, status string, limit int) ([]models.PurchaseLog, error) {
	if limit <= 0 {
		limit = 200
	}
	query := `
		SELECT o.id, o.order_ref, o.discord_id, o.identifier, o.total_amount, o.status, o.delivery_status, o.created_at,
			(SELECT COUNT(*) FROM shop_order_items WHERE order_id = o.id) AS item_count
		FROM shop_orders o WHERE 1=1`
	args := []interface{}{}
	if discordID != "" {
		query += " AND o.discord_id LIKE ?"
		args = append(args, "%"+discordID+"%")
	}
	if status != "" {
		query += " AND o.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY o.created_at DESC LIMIT ?"
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.PurchaseLog
	for rows.Next() {
		var p models.PurchaseLog
		if err := rows.Scan(&p.ID, &p.OrderRef, &p.DiscordID, &p.Identifier, &p.TotalAmount, &p.Status, &p.DeliveryStatus, &p.CreatedAt, &p.ItemCount); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, nil
}

func (s *Service) PurchaseLogDetail(ctx context.Context, orderRef string) (map[string]interface{}, error) {
	var header models.PurchaseLog
	err := s.db.QueryRowContext(ctx, `
		SELECT id, order_ref, discord_id, identifier, total_amount, status, delivery_status, created_at
		FROM shop_orders WHERE order_ref = ?`, orderRef).Scan(
		&header.ID, &header.OrderRef, &header.DiscordID, &header.Identifier, &header.TotalAmount,
		&header.Status, &header.DeliveryStatus, &header.CreatedAt)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT item_name, quantity, unit_price, esx_payload FROM shop_order_items WHERE order_id = ?`, header.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]interface{}{}
	for rows.Next() {
		var name, payload string
		var qty int
		var price float64
		if err := rows.Scan(&name, &qty, &price, &payload); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"item_name": name, "quantity": qty, "unit_price": price, "esx_payload": payload,
		})
	}
	return map[string]interface{}{
		"order": header,
		"items": items,
	}, nil
}
