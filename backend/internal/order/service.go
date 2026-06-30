package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raven-clown/raven-webmarket/backend/internal/activity"
	"github.com/raven-clown/raven-webmarket/backend/internal/lock"
	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	redisstore "github.com/raven-clown/raven-webmarket/backend/internal/redisstore"
)

type Service struct {
	db    *sql.DB
	redis *redisstore.Store
	log   *activity.Logger
	deliver DeliveryFunc
}

type DeliveryFunc func(ctx context.Context, payload models.DeliveryPayload) error

func New(db *sql.DB, redis *redisstore.Store, log *activity.Logger, deliver DeliveryFunc) *Service {
	return &Service{db: db, redis: redis, log: log, deliver: deliver}
}

func (s *Service) Checkout(ctx context.Context, discordID, identifier string, items []models.CartItem) (string, error) {
	lockKey := "checkout:" + discordID
	var orderRef string
	err := lock.WithLock(ctx, s.redis.Session, lockKey, 30*time.Second, func() error {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		var total float64
		var deliveryItems []models.DeliveryItem
		var orderItems []orderItemRow
		for _, item := range items {
			if err := s.validateItem(ctx, tx, discordID, item); err != nil {
				return err
			}
			total += item.Price * float64(item.Quantity)
			payload, dItems, err := s.buildPayload(ctx, tx, item)
			if err != nil {
				return err
			}
			for _, d := range dItems {
				deliveryItems = append(deliveryItems, d)
			}
			orderItems = append(orderItems, orderItemRow{
				item:    item,
				payload: payload,
			})
		}
		orderRef = "ORD-" + uuid.New().String()[:8]
		res, err := tx.ExecContext(ctx, `
			INSERT INTO shop_orders (order_ref, discord_id, identifier, total_amount, status, delivery_status)
			VALUES (?, ?, ?, ?, 'processing', 'pending')`,
			orderRef, discordID, identifier, total)
		if err != nil {
			return err
		}
		orderID, _ := res.LastInsertId()
		for _, oi := range orderItems {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO shop_order_items (order_id, product_id, package_id, item_name, quantity, unit_price, esx_payload)
				VALUES (?, ?, ?, ?, ?, ?, ?)`,
				orderID, oi.productID, oi.packageID, oi.item.Name, oi.item.Quantity, oi.item.Price, oi.payload)
			if err != nil {
				return err
			}
			if err := s.incrementPurchaseCount(ctx, tx, discordID, oi); err != nil {
				return err
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		payload := models.DeliveryPayload{
			Identifier: identifier,
			DiscordID:  discordID,
			Items:      deliveryItems,
			SourceType: "order",
			SourceRef:  orderRef,
		}
		if err := s.deliver(ctx, payload); err != nil {
			s.markFailed(ctx, orderRef, err.Error())
			return fmt.Errorf("delivery failed, rollback initiated: %w", err)
		}
		_, _ = s.db.ExecContext(ctx, `
			UPDATE shop_orders SET status = 'completed', delivery_status = 'delivered' WHERE order_ref = ?`, orderRef)
		if s.log != nil {
			s.log.Write(ctx, "purchase", "user", discordID, "checkout_completed", "order", orderRef, "", map[string]interface{}{
				"identifier": identifier, "total": total, "items": len(items),
			})
		}
		return nil
	})
	return orderRef, err
}

type orderItemRow struct {
	item      models.CartItem
	payload   string
	productID sql.NullInt64
	packageID sql.NullInt64
}

func (s *Service) validateItem(ctx context.Context, tx *sql.Tx, discordID string, item models.CartItem) error {
	if item.Type == "product" {
		var stockLimit, stockSold, maxPerID, active int
		var expiry sql.NullTime
		err := tx.QueryRowContext(ctx, `
			SELECT stock_limit, stock_sold, max_limit_per_id, is_active, expiry_date
			FROM shop_products WHERE id = ? FOR UPDATE`, item.ID).Scan(
			&stockLimit, &stockSold, &maxPerID, &active, &expiry)
		if err != nil {
			return fmt.Errorf("product not found")
		}
		if active != 1 {
			return fmt.Errorf("product inactive")
		}
		if expiry.Valid && expiry.Time.Before(time.Now().UTC()) {
			return fmt.Errorf("product expired")
		}
		if stockLimit > 0 && stockSold+item.Quantity > stockLimit {
			return fmt.Errorf("insufficient stock")
		}
		if maxPerID > 0 {
			var count int
			_ = tx.QueryRowContext(ctx, `
				SELECT COALESCE(purchase_count,0) FROM product_purchase_counts
				WHERE discord_id = ? AND product_id = ?`, discordID, item.ID).Scan(&count)
			if count+item.Quantity > maxPerID {
				return fmt.Errorf("purchase limit exceeded")
			}
		}
		_, err = tx.ExecContext(ctx, `UPDATE shop_products SET stock_sold = stock_sold + ? WHERE id = ?`, item.Quantity, item.ID)
		return err
	}
	if item.Type == "package" {
		var stockLimit, stockSold, maxPerID, active int
		var expiry sql.NullTime
		err := tx.QueryRowContext(ctx, `
			SELECT stock_limit, stock_sold, max_limit_per_id, is_active, expiry_date
			FROM shop_packages WHERE id = ? FOR UPDATE`, item.ID).Scan(
			&stockLimit, &stockSold, &maxPerID, &active, &expiry)
		if err != nil {
			return fmt.Errorf("package not found")
		}
		if active != 1 {
			return fmt.Errorf("package inactive")
		}
		if expiry.Valid && expiry.Time.Before(time.Now().UTC()) {
			return fmt.Errorf("package expired")
		}
		if stockLimit > 0 && stockSold+item.Quantity > stockLimit {
			return fmt.Errorf("insufficient stock")
		}
		if maxPerID > 0 {
			var count int
			_ = tx.QueryRowContext(ctx, `
				SELECT COALESCE(purchase_count,0) FROM product_purchase_counts
				WHERE discord_id = ? AND package_id = ?`, discordID, item.ID).Scan(&count)
			if count+item.Quantity > maxPerID {
				return fmt.Errorf("purchase limit exceeded")
			}
		}
		_, err = tx.ExecContext(ctx, `UPDATE shop_packages SET stock_sold = stock_sold + ? WHERE id = ?`, item.Quantity, item.ID)
		return err
	}
	return fmt.Errorf("invalid item type")
}

func (s *Service) buildPayload(ctx context.Context, tx *sql.Tx, item models.CartItem) (string, []models.DeliveryItem, error) {
	var items []map[string]interface{}
	var delivery []models.DeliveryItem
	if item.Type == "product" {
		var name, esxName string
		var esxCount int
		err := tx.QueryRowContext(ctx, `
			SELECT name, esx_item_name, esx_item_count FROM shop_products WHERE id = ?`, item.ID).Scan(&name, &esxName, &esxCount)
		if err != nil {
			return "", nil, err
		}
		count := esxCount * item.Quantity
		items = append(items, map[string]interface{}{"name": esxName, "count": count})
		delivery = append(delivery, models.DeliveryItem{Name: esxName, Count: count})
	} else {
		rows, err := tx.QueryContext(ctx, `
			SELECT esx_item_name, esx_item_count FROM shop_package_items WHERE package_id = ?`, item.ID)
		if err != nil {
			return "", nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var esxName string
			var esxCount int
			if err := rows.Scan(&esxName, &esxCount); err != nil {
				return "", nil, err
			}
			count := esxCount * item.Quantity
			items = append(items, map[string]interface{}{"name": esxName, "count": count})
			delivery = append(delivery, models.DeliveryItem{Name: esxName, Count: count})
		}
	}
	data, _ := json.Marshal(items)
	return string(data), delivery, nil
}

func (s *Service) incrementPurchaseCount(ctx context.Context, tx *sql.Tx, discordID string, oi orderItemRow) error {
	if oi.item.Type == "product" {
		oi.productID = sql.NullInt64{Int64: int64(oi.item.ID), Valid: true}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO product_purchase_counts (discord_id, product_id, purchase_count)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE purchase_count = purchase_count + VALUES(purchase_count)`,
			discordID, oi.item.ID, oi.item.Quantity)
		return err
	}
	oi.packageID = sql.NullInt64{Int64: int64(oi.item.ID), Valid: true}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO product_purchase_counts (discord_id, package_id, purchase_count)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE purchase_count = purchase_count + VALUES(purchase_count)`,
		discordID, oi.item.ID, oi.item.Quantity)
	return err
}

func (s *Service) markFailed(ctx context.Context, orderRef, msg string) {
	_, _ = s.db.ExecContext(ctx, `
		UPDATE shop_orders SET status = 'failed', delivery_status = 'failed' WHERE order_ref = ?`, orderRef)
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO delivery_queue (identifier, discord_id, payload, source_type, source_ref, status, error_message)
		SELECT identifier, discord_id, '{}', 'order', order_ref, 'failed', ? FROM shop_orders WHERE order_ref = ?`,
		msg, orderRef)
}
