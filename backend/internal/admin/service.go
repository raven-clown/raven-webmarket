package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/raven-clown/raven-webmarket/backend/internal/models"
)

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Log(ctx context.Context, adminDiscordID, action, targetType, targetID, detail, ip string) {
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO admin_audit_logs (admin_discord_id, action, target_type, target_id, detail, ip_address)
		VALUES (?, ?, ?, ?, ?, ?)`, adminDiscordID, action, targetType, targetID, detail, ip)
}

func (s *Service) ResetAccumulations(ctx context.Context, adminDiscordID, ip string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		UPDATE user_shop_profiles SET monthly_accumulation = 0, redeem_points = 0`)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.Log(ctx, adminDiscordID, "reset_accumulations", "all", "all", "{}", ip)
	return nil
}

func (s *Service) SearchUser(ctx context.Context, query string) ([]models.UserProfile, error) {
	like := "%" + query + "%"
	rows, err := s.db.QueryContext(ctx, `
		SELECT discord_id, identifier, COALESCE(display_name,''), monthly_accumulation,
			redeem_points, total_topup_amount, topup_count
		FROM user_shop_profiles
		WHERE discord_id LIKE ? OR identifier LIKE ? OR display_name LIKE ?
		LIMIT 50`, like, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []models.UserProfile
	for rows.Next() {
		var u models.UserProfile
		if err := rows.Scan(&u.DiscordID, &u.Identifier, &u.DisplayName, &u.MonthlyAccumulation,
			&u.RedeemPoints, &u.TotalTopupAmount, &u.TopupCount); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *Service) UserTopups(ctx context.Context, discordID string) ([]models.TopupTransaction, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tx_ref, discord_id, identifier, amount, points_earned, payment_method,
			COALESCE(gateway_ref,''), COALESCE(slip_url,''), status, created_at
		FROM topup_transactions WHERE discord_id = ? ORDER BY created_at DESC`, discordID)
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

func (s *Service) AuditLogs(ctx context.Context, limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, admin_discord_id, action, COALESCE(target_type,''), COALESCE(target_id,''),
			COALESCE(detail,'{}'), COALESCE(ip_address,''), created_at
		FROM admin_audit_logs ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(&l.ID, &l.AdminDiscordID, &l.Action, &l.TargetType, &l.TargetID,
			&l.Detail, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (s *Service) RevenueOverview(ctx context.Context, period string) ([]models.KPIRevenue, error) {
	format := "%Y-%m-%d"
	switch period {
	case "weekly":
		format = "%Y-%u"
	case "monthly":
		format = "%Y-%m"
	case "yearly":
		format = "%Y"
	}
	query := fmt.Sprintf(`
		SELECT DATE_FORMAT(created_at, '%s') AS period, SUM(amount) AS total
		FROM topup_transactions WHERE status = 'completed'
		GROUP BY period ORDER BY period DESC LIMIT 30`, format)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.KPIRevenue
	for rows.Next() {
		var r models.KPIRevenue
		if err := rows.Scan(&r.Period, &r.Amount); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	return items, nil
}

func (s *Service) PeakTopup(ctx context.Context) (float64, time.Time, error) {
	var amount float64
	var peakTime time.Time
	err := s.db.QueryRowContext(ctx, `
		SELECT amount, created_at FROM topup_transactions
		WHERE status = 'completed' ORDER BY amount DESC LIMIT 1`).Scan(&amount, &peakTime)
	return amount, peakTime, err
}

func (s *Service) TransactionFrequency(ctx context.Context) ([]models.KPIPaymentMethod, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT payment_method, COUNT(*) FROM topup_transactions
		WHERE status = 'completed' GROUP BY payment_method`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.KPIPaymentMethod
	for rows.Next() {
		var m models.KPIPaymentMethod
		if err := rows.Scan(&m.Method, &m.Count); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, nil
}

func (s *Service) TopSpenders(ctx context.Context, limit int) ([]models.KPITopSpender, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT discord_id, COALESCE(display_name,''), total_topup_amount, topup_count
		FROM user_shop_profiles ORDER BY total_topup_amount DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.KPITopSpender
	for rows.Next() {
		var t models.KPITopSpender
		if err := rows.Scan(&t.DiscordID, &t.DisplayName, &t.TotalAmount, &t.TopupCount); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, nil
}

func (s *Service) UpsertProduct(ctx context.Context, adminID, ip string, data map[string]interface{}) error {
	detail, _ := json.Marshal(data)
	s.Log(ctx, adminID, "upsert_product", "product", fmt.Sprint(data["sku"]), string(detail), ip)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO shop_products (category_id, sku, name, description, image_url, regular_price, sale_price,
			esx_item_name, esx_item_count, stock_limit, max_limit_per_id, expiry_date, is_featured, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name), description = VALUES(description), image_url = VALUES(image_url),
			regular_price = VALUES(regular_price), sale_price = VALUES(sale_price),
			stock_limit = VALUES(stock_limit), max_limit_per_id = VALUES(max_limit_per_id),
			expiry_date = VALUES(expiry_date), is_featured = VALUES(is_featured), is_active = VALUES(is_active)`,
		data["category_id"], data["sku"], data["name"], data["description"], data["image_url"],
		data["regular_price"], data["sale_price"], data["esx_item_name"], data["esx_item_count"],
		data["stock_limit"], data["max_limit_per_id"], data["expiry_date"], data["is_featured"], data["is_active"])
	return err
}

func (s *Service) UpsertBanner(ctx context.Context, adminID, ip string, data map[string]interface{}) error {
	detail, _ := json.Marshal(data)
	s.Log(ctx, adminID, "upsert_banner", "banner", fmt.Sprint(data["title"]), string(detail), ip)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO shop_banners (title, image_url, link_url, sort_order, is_active)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			title = VALUES(title), image_url = VALUES(image_url), link_url = VALUES(link_url),
			sort_order = VALUES(sort_order), is_active = VALUES(is_active)`,
		data["title"], data["image_url"], data["link_url"], data["sort_order"], data["is_active"])
	return err
}
