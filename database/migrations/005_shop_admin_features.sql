ALTER TABLE shop_products ADD COLUMN sale_start_date DATETIME NULL AFTER expiry_date;
ALTER TABLE shop_packages ADD COLUMN sale_start_date DATETIME NULL AFTER expiry_date;
ALTER TABLE admin_accounts ADD COLUMN permissions JSON NULL AFTER role;

CREATE TABLE IF NOT EXISTS shop_promotions (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    description TEXT,
    target_type ENUM('product','package') NOT NULL DEFAULT 'product',
    target_id INT UNSIGNED NOT NULL,
    banner_image_url VARCHAR(512),
    regular_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    sale_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    max_limit_per_id INT NOT NULL DEFAULT 0,
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_active_dates (is_active, start_date, end_date),
    INDEX idx_target (target_type, target_id)
);

INSERT INTO system_settings (setting_key, setting_value, updated_by) VALUES
('monthly_reset', JSON_OBJECT(
    'enabled', true,
    'reset_day', 1,
    'reset_hour_utc', 17,
    'reset_redeem_points', false
), 'system')
ON DUPLICATE KEY UPDATE setting_key = setting_key;

INSERT INTO milestone_events (name, month_year, is_active) VALUES
('Monthly Top-up Rewards', DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m'), 1)
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO shop_categories (slug, name, sort_order) VALUES
('packs', 'Packs & Bundles', 5)
ON DUPLICATE KEY UPDATE name = VALUES(name);
