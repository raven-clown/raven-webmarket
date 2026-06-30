package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/raven-clown/raven-webmarket/backend/internal/models"
	redisstore "github.com/raven-clown/raven-webmarket/backend/internal/redisstore"
)

const cacheTTL = 5 * time.Minute

type Service struct {
	db    *sql.DB
	redis *redisstore.Store
}

func New(db *sql.DB, redis *redisstore.Store) *Service {
	return &Service{db: db, redis: redis}
}

func (s *Service) GetBanners(ctx context.Context) ([]models.Banner, error) {
	key := "cache:banners"
	if raw, err := s.redis.Session.Get(ctx, key).Bytes(); err == nil {
		var items []models.Banner
		if json.Unmarshal(raw, &items) == nil {
			return items, nil
		}
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, image_url, COALESCE(link_url,''), sort_order, is_active
		FROM shop_banners WHERE is_active = 1 ORDER BY sort_order ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.Banner
	for rows.Next() {
		var b models.Banner
		var active int
		if err := rows.Scan(&b.ID, &b.Title, &b.ImageURL, &b.LinkURL, &b.SortOrder, &active); err != nil {
			return nil, err
		}
		b.IsActive = active == 1
		items = append(items, b)
	}
	if data, err := json.Marshal(items); err == nil {
		s.redis.Session.Set(ctx, key, data, cacheTTL)
	}
	return items, nil
}

func (s *Service) GetCategories(ctx context.Context) ([]models.Category, error) {
	key := "cache:categories"
	if raw, err := s.redis.Session.Get(ctx, key).Bytes(); err == nil {
		var items []models.Category
		if json.Unmarshal(raw, &items) == nil {
			return items, nil
		}
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, slug, name, sort_order, is_active
		FROM shop_categories WHERE is_active = 1 ORDER BY sort_order ASC`)
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
	if data, err := json.Marshal(items); err == nil {
		s.redis.Session.Set(ctx, key, data, cacheTTL)
	}
	return items, nil
}

func (s *Service) GetProducts(ctx context.Context, categoryID uint, search string, featuredOnly bool) ([]models.Product, error) {
	cacheKey := fmt.Sprintf("cache:products:%d:%s:%v", categoryID, search, featuredOnly)
	if raw, err := s.redis.Session.Get(ctx, cacheKey).Bytes(); err == nil {
		var items []models.Product
		if json.Unmarshal(raw, &items) == nil {
			return items, nil
		}
	}
	query := `
		SELECT id, category_id, sku, name, COALESCE(description,''), COALESCE(image_url,''),
			regular_price, sale_price, esx_item_name, esx_item_count, stock_limit, stock_sold,
			max_limit_per_id, expiry_date, is_featured, is_active
		FROM shop_products
		WHERE is_active = 1
			AND (expiry_date IS NULL OR expiry_date > UTC_TIMESTAMP())
			AND (sale_start_date IS NULL OR sale_start_date <= UTC_TIMESTAMP())`
	args := []interface{}{}
	if categoryID > 0 {
		query += " AND category_id = ?"
		args = append(args, categoryID)
	}
	if search != "" {
		query += " AND (name LIKE ? OR sku LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like)
	}
	if featuredOnly {
		query += " AND is_featured = 1"
	}
	query += " ORDER BY sort_order ASC, id DESC"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if data, err := json.Marshal(items); err == nil {
		s.redis.Session.Set(ctx, cacheKey, data, cacheTTL)
	}
	return items, nil
}

func (s *Service) GetPackages(ctx context.Context, featuredOnly bool) ([]models.Package, error) {
	query := `
		SELECT id, sku, name, COALESCE(description,''), COALESCE(image_url,''),
			regular_price, sale_price, stock_limit, stock_sold, max_limit_per_id,
			expiry_date, is_featured, is_active
		FROM shop_packages
		WHERE is_active = 1
			AND (expiry_date IS NULL OR expiry_date > UTC_TIMESTAMP())
			AND (sale_start_date IS NULL OR sale_start_date <= UTC_TIMESTAMP())`
	if featuredOnly {
		query += " AND is_featured = 1"
	}
	query += " ORDER BY id DESC"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.Package
	for rows.Next() {
		pkg, err := scanPackage(rows)
		if err != nil {
			return nil, err
		}
		pkg.Items, _ = s.getPackageItems(ctx, pkg.ID)
		items = append(items, pkg)
	}
	return items, nil
}

func (s *Service) getPackageItems(ctx context.Context, packageID uint) ([]models.PackageItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT esx_item_name, esx_item_count FROM shop_package_items WHERE package_id = ?`, packageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.PackageItem
	for rows.Next() {
		var it models.PackageItem
		if err := rows.Scan(&it.ESXItemName, &it.ESXItemCount); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *Service) GetActivePromotions(ctx context.Context) ([]models.Promotion, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, COALESCE(description,''), target_type, target_id, COALESCE(banner_image_url,''),
			regular_price, sale_price, max_limit_per_id, start_date, end_date, is_active, sort_order
		FROM shop_promotions
		WHERE is_active = 1 AND start_date <= UTC_TIMESTAMP() AND end_date > UTC_TIMESTAMP()
		ORDER BY sort_order ASC, id DESC`)
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

func (s *Service) InvalidateCache(ctx context.Context) {
	iter := s.redis.Session.Scan(ctx, 0, "cache:*", 100).Iterator()
	for iter.Next(ctx) {
		s.redis.Session.Del(ctx, iter.Val())
	}
}

func scanProduct(scanner interface{ Scan(...interface{}) error }) (models.Product, error) {
	var p models.Product
	var active, featured int
	var expiry sql.NullTime
	err := scanner.Scan(&p.ID, &p.CategoryID, &p.SKU, &p.Name, &p.Description, &p.ImageURL,
		&p.RegularPrice, &p.SalePrice, &p.ESXItemName, &p.ESXItemCount, &p.StockLimit, &p.StockSold,
		&p.MaxLimitPerID, &expiry, &featured, &active)
	if err != nil {
		return p, err
	}
	if expiry.Valid {
		t := expiry.Time
		p.ExpiryDate = &t
	}
	p.IsActive = active == 1
	p.IsFeatured = featured == 1
	if p.RegularPrice > 0 && p.SalePrice < p.RegularPrice {
		p.DiscountPct = ((p.RegularPrice - p.SalePrice) / p.RegularPrice) * 100
	}
	if p.StockLimit > 0 {
		p.StockRemaining = p.StockLimit - p.StockSold
	}
	return p, nil
}

func scanPackage(scanner interface{ Scan(...interface{}) error }) (models.Package, error) {
	var p models.Package
	var active, featured int
	var expiry sql.NullTime
	err := scanner.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.ImageURL,
		&p.RegularPrice, &p.SalePrice, &p.StockLimit, &p.StockSold, &p.MaxLimitPerID,
		&expiry, &featured, &active)
	if err != nil {
		return p, err
	}
	if expiry.Valid {
		t := expiry.Time
		p.ExpiryDate = &t
	}
	p.IsActive = active == 1
	p.IsFeatured = featured == 1
	if p.RegularPrice > 0 && p.SalePrice < p.RegularPrice {
		p.DiscountPct = ((p.RegularPrice - p.SalePrice) / p.RegularPrice) * 100
	}
	return p, nil
}
