package admin

import (
	"context"
	"database/sql"
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
	return s.AuditLogsFiltered(ctx, "", "", limit)
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
