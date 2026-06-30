package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/raven-clown/raven-webmarket/backend/internal/middleware"
	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	"github.com/raven-clown/raven-webmarket/backend/internal/rbac"
)

type MonthlyResetConfig struct {
	Enabled           bool `json:"enabled"`
	ResetDay          int  `json:"reset_day"`
	ResetHourUTC      int  `json:"reset_hour_utc"`
	ResetRedeemPoints bool `json:"reset_redeem_points"`
}

type Promotion = models.Promotion

func (s *Service) checkPerm(actor middleware.AdminClaims, permission string) error {
	if !rbac.CanWithPermissions(actor.Role, actor.Permissions, permission) {
		return fmt.Errorf("forbidden")
	}
	return nil
}

func (s *Service) ListProducts(ctx context.Context, actor middleware.AdminClaims) ([]models.Product, error) {
	if err := s.checkPerm(actor, "products"); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, category_id, sku, name, COALESCE(description,''), COALESCE(image_url,''),
			regular_price, sale_price, esx_item_name, esx_item_count, stock_limit, stock_sold,
			max_limit_per_id, expiry_date, sale_start_date, is_featured, is_active
		FROM shop_products ORDER BY sort_order ASC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

func scanProducts(rows *sql.Rows) ([]models.Product, error) {
	var items []models.Product
	for rows.Next() {
		var p models.Product
		var active, featured int
		var expiry, saleStart sql.NullTime
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.SKU, &p.Name, &p.Description, &p.ImageURL,
			&p.RegularPrice, &p.SalePrice, &p.ESXItemName, &p.ESXItemCount, &p.StockLimit, &p.StockSold,
			&p.MaxLimitPerID, &expiry, &saleStart, &featured, &active); err != nil {
			return nil, err
		}
		if expiry.Valid {
			p.ExpiryDate = &expiry.Time
		}
		if saleStart.Valid {
			p.SaleStartDate = &saleStart.Time
		}
		p.IsFeatured = featured == 1
		p.IsActive = active == 1
		items = append(items, p)
	}
	return items, nil
}

func (s *Service) ListPackages(ctx context.Context, actor middleware.AdminClaims) ([]models.Package, error) {
	if err := s.checkPerm(actor, "packages"); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, sku, name, COALESCE(description,''), COALESCE(image_url,''),
			regular_price, sale_price, stock_limit, stock_sold, max_limit_per_id,
			expiry_date, sale_start_date, is_featured, is_active
		FROM shop_packages ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.Package
	for rows.Next() {
		var p models.Package
		var active, featured int
		var expiry, saleStart sql.NullTime
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.ImageURL,
			&p.RegularPrice, &p.SalePrice, &p.StockLimit, &p.StockSold, &p.MaxLimitPerID,
			&expiry, &saleStart, &featured, &active); err != nil {
			return nil, err
		}
		if expiry.Valid {
			p.ExpiryDate = &expiry.Time
		}
		if saleStart.Valid {
			p.SaleStartDate = &saleStart.Time
		}
		p.IsFeatured = featured == 1
		p.IsActive = active == 1
		itemRows, err := s.db.QueryContext(ctx, `
			SELECT id, package_id, esx_item_name, esx_item_count FROM shop_package_items WHERE package_id = ?`, p.ID)
		if err == nil {
			for itemRows.Next() {
				var it models.PackageItem
				if err := itemRows.Scan(&it.ID, &it.PackageID, &it.ESXItemName, &it.ESXItemCount); err == nil {
					p.Items = append(p.Items, it)
				}
			}
			itemRows.Close()
		}
		items = append(items, p)
	}
	return items, nil
}

func (s *Service) ListCategories(ctx context.Context, actor middleware.AdminClaims) ([]models.Category, error) {
	if err := s.checkPerm(actor, "products"); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, slug, name, sort_order, is_active FROM shop_categories ORDER BY sort_order ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.Category
	for rows.Next() {
		var c models.Category
		var active int
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.SortOrder, &active); err != nil {
			return nil, err
		}
		c.IsActive = active == 1
		items = append(items, c)
	}
	return items, nil
}

func (s *Service) UpsertProduct(ctx context.Context, actor middleware.AdminClaims, ip string, data map[string]interface{}) error {
	if err := s.checkPerm(actor, "products"); err != nil {
		return err
	}
	detail, _ := json.Marshal(data)
	s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "cms", "upsert_product", "product", fmt.Sprint(data["sku"]), string(detail), ip)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO shop_products (category_id, sku, name, description, image_url, regular_price, sale_price,
			esx_item_name, esx_item_count, stock_limit, max_limit_per_id, expiry_date, sale_start_date, is_featured, is_active, sort_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			category_id = VALUES(category_id), name = VALUES(name), description = VALUES(description),
			image_url = VALUES(image_url), regular_price = VALUES(regular_price), sale_price = VALUES(sale_price),
			esx_item_name = VALUES(esx_item_name), esx_item_count = VALUES(esx_item_count),
			stock_limit = VALUES(stock_limit), max_limit_per_id = VALUES(max_limit_per_id),
			expiry_date = VALUES(expiry_date), sale_start_date = VALUES(sale_start_date),
			is_featured = VALUES(is_featured), is_active = VALUES(is_active), sort_order = VALUES(sort_order)`,
		data["category_id"], data["sku"], data["name"], data["description"], data["image_url"],
		data["regular_price"], data["sale_price"], data["esx_item_name"], data["esx_item_count"],
		data["stock_limit"], data["max_limit_per_id"], nullIfEmpty(data["expiry_date"]),
		nullIfEmpty(data["sale_start_date"]), data["is_featured"], data["is_active"], data["sort_order"])
	return err
}

func (s *Service) UpsertPackage(ctx context.Context, actor middleware.AdminClaims, ip string, data map[string]interface{}, items []map[string]interface{}) error {
	if err := s.checkPerm(actor, "packages"); err != nil {
		return err
	}
	detail, _ := json.Marshal(map[string]interface{}{"package": data, "items": items})
	s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "cms", "upsert_package", "package", fmt.Sprint(data["sku"]), string(detail), ip)
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO shop_packages (sku, name, description, image_url, regular_price, sale_price,
			stock_limit, max_limit_per_id, expiry_date, sale_start_date, is_featured, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name), description = VALUES(description), image_url = VALUES(image_url),
			regular_price = VALUES(regular_price), sale_price = VALUES(sale_price),
			stock_limit = VALUES(stock_limit), max_limit_per_id = VALUES(max_limit_per_id),
			expiry_date = VALUES(expiry_date), sale_start_date = VALUES(sale_start_date),
			is_featured = VALUES(is_featured), is_active = VALUES(is_active)`,
		data["sku"], data["name"], data["description"], data["image_url"],
		data["regular_price"], data["sale_price"], data["stock_limit"], data["max_limit_per_id"],
		nullIfEmpty(data["expiry_date"]), nullIfEmpty(data["sale_start_date"]),
		data["is_featured"], data["is_active"])
	if err != nil {
		return err
	}
	var packageID int64
	if id, ok := data["id"]; ok && id != nil && fmt.Sprint(id) != "0" {
		packageID, _ = parseInt64(id)
	} else {
		packageID, _ = res.LastInsertId()
		if packageID == 0 {
			_ = s.db.QueryRowContext(ctx, `SELECT id FROM shop_packages WHERE sku = ?`, data["sku"]).Scan(&packageID)
		}
	}
	if packageID > 0 {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM shop_package_items WHERE package_id = ?`, packageID)
		for _, item := range items {
			_, _ = s.db.ExecContext(ctx, `
				INSERT INTO shop_package_items (package_id, esx_item_name, esx_item_count) VALUES (?, ?, ?)`,
				packageID, item["esx_item_name"], item["esx_item_count"])
		}
	}
	return nil
}

func (s *Service) ListPromotions(ctx context.Context, actor middleware.AdminClaims) ([]models.Promotion, error) {
	if err := s.checkPerm(actor, "promotions"); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, COALESCE(description,''), target_type, target_id, COALESCE(banner_image_url,''),
			regular_price, sale_price, max_limit_per_id, start_date, end_date, is_active, sort_order
		FROM shop_promotions ORDER BY sort_order ASC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.Promotion
	for rows.Next() {
		var p models.Promotion
		var active int
		var start, end time.Time
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.TargetType, &p.TargetID, &p.BannerImageURL,
			&p.RegularPrice, &p.SalePrice, &p.MaxLimitPerID, &start, &end, &active, &p.SortOrder); err != nil {
			return nil, err
		}
		p.StartDate = start.Format(time.RFC3339)
		p.EndDate = end.Format(time.RFC3339)
		p.IsActive = active == 1
		items = append(items, p)
	}
	return items, nil
}

func (s *Service) UpsertPromotion(ctx context.Context, actor middleware.AdminClaims, ip string, p models.Promotion) error {
	if err := s.checkPerm(actor, "promotions"); err != nil {
		return err
	}
	detail, _ := json.Marshal(p)
	s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "cms", "upsert_promotion", "promotion", p.Name, string(detail), ip)
	if p.ID > 0 {
		_, err := s.db.ExecContext(ctx, `
			UPDATE shop_promotions SET name=?, description=?, target_type=?, target_id=?, banner_image_url=?,
				regular_price=?, sale_price=?, max_limit_per_id=?, start_date=?, end_date=?, is_active=?, sort_order=?
			WHERE id=?`, p.Name, p.Description, p.TargetType, p.TargetID, p.BannerImageURL,
			p.RegularPrice, p.SalePrice, p.MaxLimitPerID, p.StartDate, p.EndDate, boolToInt(p.IsActive), p.SortOrder, p.ID)
		return err
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO shop_promotions (name, description, target_type, target_id, banner_image_url,
			regular_price, sale_price, max_limit_per_id, start_date, end_date, is_active, sort_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.Description, p.TargetType, p.TargetID, p.BannerImageURL,
		p.RegularPrice, p.SalePrice, p.MaxLimitPerID, p.StartDate, p.EndDate, boolToInt(p.IsActive), p.SortOrder)
	return err
}

func (s *Service) DeletePromotion(ctx context.Context, actor middleware.AdminClaims, ip string, id uint) error {
	if err := s.checkPerm(actor, "promotions"); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM shop_promotions WHERE id = ?`, id)
	if err == nil {
		s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "cms", "delete_promotion", "promotion", fmt.Sprint(id), "{}", ip)
	}
	return err
}

func (s *Service) ListMilestoneEvents(ctx context.Context, actor middleware.AdminClaims) ([]map[string]interface{}, error) {
	if err := s.checkPerm(actor, "milestones"); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, month_year, is_active FROM milestone_events ORDER BY month_year DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []map[string]interface{}
	for rows.Next() {
		var id uint
		var name, monthYear string
		var active int
		if err := rows.Scan(&id, &name, &monthYear, &active); err != nil {
			return nil, err
		}
		tierRows, _ := s.db.QueryContext(ctx, `
			SELECT id, tier_level, threshold_amount, reward_name, esx_item_name, esx_item_count
			FROM milestone_tiers WHERE event_id = ? ORDER BY tier_level ASC`, id)
		tiers := []map[string]interface{}{}
		if tierRows != nil {
			for tierRows.Next() {
				var tid, level, count int
				var threshold float64
				var reward, item string
				if err := tierRows.Scan(&tid, &level, &threshold, &reward, &item, &count); err == nil {
					tiers = append(tiers, map[string]interface{}{
						"id": tid, "tier_level": level, "threshold_amount": threshold,
						"reward_name": reward, "esx_item_name": item, "esx_item_count": count,
					})
				}
			}
			tierRows.Close()
		}
		events = append(events, map[string]interface{}{
			"id": id, "name": name, "month_year": monthYear, "is_active": active == 1, "tiers": tiers,
		})
	}
	return events, nil
}

func (s *Service) UpsertMilestoneEvent(ctx context.Context, actor middleware.AdminClaims, ip string, name, monthYear string, active bool, tiers []map[string]interface{}) error {
	if err := s.checkPerm(actor, "milestones"); err != nil {
		return err
	}
	detail, _ := json.Marshal(map[string]interface{}{"name": name, "month_year": monthYear, "tiers": tiers})
	s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "cms", "upsert_milestone", "milestone", monthYear, string(detail), ip)
	var eventID uint
	err := s.db.QueryRowContext(ctx, `SELECT id FROM milestone_events WHERE month_year = ?`, monthYear).Scan(&eventID)
	if err == sql.ErrNoRows {
		res, err := s.db.ExecContext(ctx, `
			INSERT INTO milestone_events (name, month_year, is_active) VALUES (?, ?, ?)`,
			name, monthYear, boolToInt(active))
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		eventID = uint(id)
	} else if err != nil {
		return err
	} else {
		_, _ = s.db.ExecContext(ctx, `UPDATE milestone_events SET name=?, is_active=? WHERE id=?`, name, boolToInt(active), eventID)
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM milestone_tiers WHERE event_id = ?`, eventID)
	for _, t := range tiers {
		_, _ = s.db.ExecContext(ctx, `
			INSERT INTO milestone_tiers (event_id, tier_level, threshold_amount, reward_name, esx_item_name, esx_item_count)
			VALUES (?, ?, ?, ?, ?, ?)`,
			eventID, t["tier_level"], t["threshold_amount"], t["reward_name"], t["esx_item_name"], t["esx_item_count"])
	}
	return nil
}

func (s *Service) ListRedeemCatalog(ctx context.Context, actor middleware.AdminClaims) ([]models.RedeemItem, error) {
	if err := s.checkPerm(actor, "redeem"); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, COALESCE(description,''), COALESCE(image_url,''), point_cost,
			esx_item_name, esx_item_count, stock_limit, stock_redeemed, is_active
		FROM redeem_catalog ORDER BY id DESC`)
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

func (s *Service) UpsertRedeemItem(ctx context.Context, actor middleware.AdminClaims, ip string, data map[string]interface{}) error {
	if err := s.checkPerm(actor, "redeem"); err != nil {
		return err
	}
	detail, _ := json.Marshal(data)
	s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "cms", "upsert_redeem", "redeem", fmt.Sprint(data["name"]), string(detail), ip)
	if id, ok := data["id"]; ok && fmt.Sprint(id) != "0" && fmt.Sprint(id) != "" {
		_, err := s.db.ExecContext(ctx, `
			UPDATE redeem_catalog SET name=?, description=?, image_url=?, point_cost=?,
				esx_item_name=?, esx_item_count=?, stock_limit=?, is_active=? WHERE id=?`,
			data["name"], data["description"], data["image_url"], data["point_cost"],
			data["esx_item_name"], data["esx_item_count"], data["stock_limit"], data["is_active"], id)
		return err
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO redeem_catalog (name, description, image_url, point_cost, esx_item_name, esx_item_count, stock_limit, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		data["name"], data["description"], data["image_url"], data["point_cost"],
		data["esx_item_name"], data["esx_item_count"], data["stock_limit"], data["is_active"])
	return err
}

func (s *Service) GetMonthlyResetConfig(ctx context.Context) (MonthlyResetConfig, error) {
	var cfg MonthlyResetConfig
	err := s.GetSetting(ctx, "monthly_reset", &cfg)
	if err == sql.ErrNoRows {
		return MonthlyResetConfig{Enabled: true, ResetDay: 1, ResetHourUTC: 17}, nil
	}
	return cfg, err
}

func (s *Service) SaveMonthlyResetConfig(ctx context.Context, actor middleware.AdminClaims, ip string, cfg MonthlyResetConfig) error {
	if !rbac.CanWithPermissions(actor.Role, actor.Permissions, "reset_monthly") {
		return fmt.Errorf("forbidden")
	}
	if err := s.SaveSetting(ctx, "monthly_reset", cfg, actor.Username); err != nil {
		return err
	}
	detail, _ := json.Marshal(cfg)
	s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "system", "update_monthly_reset", "setting", "monthly_reset", string(detail), ip)
	return nil
}

func (s *Service) ResetMonthlyAccumulation(ctx context.Context, actor *middleware.AdminClaims, ip string, includeRedeem bool) error {
	if actor != nil {
		perm := "reset_monthly"
		if includeRedeem {
			perm = "reset"
		}
		if !rbac.CanWithPermissions(actor.Role, actor.Permissions, perm) {
			return fmt.Errorf("forbidden")
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if includeRedeem {
		_, err = tx.ExecContext(ctx, `UPDATE user_shop_profiles SET monthly_accumulation = 0, redeem_points = 0`)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE user_shop_profiles SET monthly_accumulation = 0`)
	}
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if actor != nil {
		detail := fmt.Sprintf(`{"include_redeem":%v}`, includeRedeem)
		s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "user", "reset_accumulations", "all", "all", detail, ip)
	}
	return nil
}

func (s *Service) RunScheduledMonthlyReset(ctx context.Context) error {
	cfg, err := s.GetMonthlyResetConfig(ctx)
	if err != nil || !cfg.Enabled {
		return nil
	}
	now := time.Now().UTC()
	if now.Day() != cfg.ResetDay || now.Hour() != cfg.ResetHourUTC {
		return nil
	}
	key := fmt.Sprintf("monthly_reset_done:%s", now.Format("2006-01"))
	var existing string
	_ = s.db.QueryRowContext(ctx, `SELECT setting_key FROM system_settings WHERE setting_key = ?`, key).Scan(&existing)
	if existing != "" {
		return nil
	}
	if err := s.ResetMonthlyAccumulation(ctx, nil, "", cfg.ResetRedeemPoints); err != nil {
		return err
	}
	_ = s.SaveSetting(ctx, key, map[string]string{"done": now.Format(time.RFC3339)}, "system")
	return nil
}

func nullIfEmpty(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	s := fmt.Sprint(v)
	if s == "" || s == "null" {
		return nil
	}
	return v
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func parseInt64(v interface{}) (int64, error) {
	switch n := v.(type) {
	case float64:
		return int64(n), nil
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	case json.Number:
		return n.Int64()
	default:
		return 0, fmt.Errorf("invalid id")
	}
}
