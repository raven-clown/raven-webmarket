package redeem

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raven-clown/raven-webmarket/backend/internal/lock"
	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	redisstore "github.com/raven-clown/raven-webmarket/backend/internal/redisstore"
)

type Service struct {
	db      *sql.DB
	redis   *redisstore.Store
	deliver func(ctx context.Context, payload models.DeliveryPayload) error
}

func New(db *sql.DB, redis *redisstore.Store, deliver func(ctx context.Context, payload models.DeliveryPayload) error) *Service {
	return &Service{db: db, redis: redis, deliver: deliver}
}

func (s *Service) Catalog(ctx context.Context) ([]models.RedeemItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, COALESCE(description,''), COALESCE(image_url,''), point_cost,
			esx_item_name, esx_item_count, stock_limit, stock_redeemed, is_active
		FROM redeem_catalog WHERE is_active = 1 ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.RedeemItem
	for rows.Next() {
		var r models.RedeemItem
		var active int
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.ImageURL, &r.PointCost,
			&r.ESXItemName, &r.ESXItemCount, &r.StockLimit, &r.StockRedeemed, &active); err != nil {
			return nil, err
		}
		r.IsActive = active == 1
		items = append(items, r)
	}
	return items, nil
}

func (s *Service) Redeem(ctx context.Context, discordID, identifier string, catalogID uint) error {
	lockKey := "redeem:" + discordID
	return lock.WithLock(ctx, s.redis.Session, lockKey, 30*time.Second, func() error {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		var item models.RedeemItem
		var active int
		err = tx.QueryRowContext(ctx, `
			SELECT id, name, point_cost, esx_item_name, esx_item_count, stock_limit, stock_redeemed, is_active
			FROM redeem_catalog WHERE id = ? FOR UPDATE`, catalogID).Scan(
			&item.ID, &item.Name, &item.PointCost, &item.ESXItemName, &item.ESXItemCount,
			&item.StockLimit, &item.StockRedeemed, &active)
		if err != nil {
			return fmt.Errorf("item not found")
		}
		if active != 1 {
			return fmt.Errorf("item inactive")
		}
		if item.StockLimit > 0 && item.StockRedeemed >= item.StockLimit {
			return fmt.Errorf("out of stock")
		}
		var points float64
		err = tx.QueryRowContext(ctx, `
			SELECT redeem_points FROM user_shop_profiles WHERE discord_id = ? FOR UPDATE`, discordID).Scan(&points)
		if err != nil {
			return err
		}
		if points < item.PointCost {
			return fmt.Errorf("insufficient points")
		}
		_, err = tx.ExecContext(ctx, `
			UPDATE user_shop_profiles SET redeem_points = redeem_points - ? WHERE discord_id = ?`,
			item.PointCost, discordID)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			UPDATE redeem_catalog SET stock_redeemed = stock_redeemed + 1 WHERE id = ?`, catalogID)
		if err != nil {
			return err
		}
		ref := "RD-" + uuid.New().String()[:8]
		_, err = tx.ExecContext(ctx, `
			INSERT INTO redeem_transactions (discord_id, identifier, catalog_id, points_spent, status, delivery_status)
			VALUES (?, ?, ?, ?, 'processing', 'pending')`, discordID, identifier, catalogID, item.PointCost)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		payload := models.DeliveryPayload{
			Identifier: identifier,
			DiscordID:  discordID,
			Items:      []models.DeliveryItem{{Name: item.ESXItemName, Count: item.ESXItemCount}},
			SourceType: "redeem",
			SourceRef:  ref,
		}
		if err := s.deliver(ctx, payload); err != nil {
			s.rollback(ctx, discordID, catalogID, item.PointCost)
			return err
		}
		_, _ = s.db.ExecContext(ctx, `
			UPDATE redeem_transactions SET status = 'completed', delivery_status = 'delivered'
			WHERE discord_id = ? AND catalog_id = ? ORDER BY id DESC LIMIT 1`, discordID, catalogID)
		return nil
	})
}

func (s *Service) rollback(ctx context.Context, discordID string, catalogID uint, points float64) {
	_, _ = s.db.ExecContext(ctx, `
		UPDATE user_shop_profiles SET redeem_points = redeem_points + ? WHERE discord_id = ?`, points, discordID)
	_, _ = s.db.ExecContext(ctx, `
		UPDATE redeem_catalog SET stock_redeemed = stock_redeemed - 1 WHERE id = ?`, catalogID)
	_, _ = s.db.ExecContext(ctx, `
		UPDATE redeem_transactions SET status = 'failed', delivery_status = 'failed'
		WHERE discord_id = ? AND catalog_id = ? ORDER BY id DESC LIMIT 1`, discordID, catalogID)
}
