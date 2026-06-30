CREATE TABLE IF NOT EXISTS shop_categories (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    slug VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS shop_products (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    category_id INT UNSIGNED NOT NULL,
    sku VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(256) NOT NULL,
    description TEXT,
    image_url VARCHAR(512),
    regular_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    sale_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    esx_item_name VARCHAR(128) NOT NULL,
    esx_item_count INT NOT NULL DEFAULT 1,
    stock_limit INT NOT NULL DEFAULT 0,
    stock_sold INT NOT NULL DEFAULT 0,
    max_limit_per_id INT NOT NULL DEFAULT 0,
    expiry_date DATETIME NULL,
    is_featured TINYINT(1) NOT NULL DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category_id),
    INDEX idx_featured (is_featured, is_active),
    INDEX idx_expiry (expiry_date),
    FOREIGN KEY (category_id) REFERENCES shop_categories(id)
);

CREATE TABLE IF NOT EXISTS shop_banners (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(256) NOT NULL,
    image_url VARCHAR(512) NOT NULL,
    link_url VARCHAR(512),
    sort_order INT NOT NULL DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS shop_packages (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(256) NOT NULL,
    description TEXT,
    image_url VARCHAR(512),
    regular_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    sale_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    stock_limit INT NOT NULL DEFAULT 0,
    stock_sold INT NOT NULL DEFAULT 0,
    max_limit_per_id INT NOT NULL DEFAULT 0,
    expiry_date DATETIME NULL,
    is_featured TINYINT(1) NOT NULL DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS shop_package_items (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    package_id INT UNSIGNED NOT NULL,
    esx_item_name VARCHAR(128) NOT NULL,
    esx_item_count INT NOT NULL DEFAULT 1,
    FOREIGN KEY (package_id) REFERENCES shop_packages(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS shop_orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_ref VARCHAR(64) NOT NULL UNIQUE,
    discord_id VARCHAR(32) NOT NULL,
    identifier VARCHAR(128) NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    status ENUM('pending','processing','completed','failed','refunded') NOT NULL DEFAULT 'pending',
    delivery_status ENUM('pending','delivered','queued','failed') NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_discord (discord_id),
    INDEX idx_status (status)
);

CREATE TABLE IF NOT EXISTS shop_order_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT UNSIGNED NOT NULL,
    product_id INT UNSIGNED NULL,
    package_id INT UNSIGNED NULL,
    item_name VARCHAR(256) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    unit_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    esx_payload JSON NOT NULL,
    FOREIGN KEY (order_id) REFERENCES shop_orders(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS product_purchase_counts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    discord_id VARCHAR(32) NOT NULL,
    product_id INT UNSIGNED NULL,
    package_id INT UNSIGNED NULL,
    purchase_count INT NOT NULL DEFAULT 0,
    UNIQUE KEY uq_discord_product (discord_id, product_id),
    UNIQUE KEY uq_discord_package (discord_id, package_id)
);

CREATE TABLE IF NOT EXISTS user_shop_profiles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    discord_id VARCHAR(32) NOT NULL UNIQUE,
    identifier VARCHAR(128) NOT NULL,
    display_name VARCHAR(256),
    monthly_accumulation DECIMAL(12,2) NOT NULL DEFAULT 0,
    redeem_points DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_topup_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    topup_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_identifier (identifier)
);

CREATE TABLE IF NOT EXISTS milestone_events (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    month_year CHAR(7) NOT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_month (month_year)
);

CREATE TABLE IF NOT EXISTS milestone_tiers (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_id INT UNSIGNED NOT NULL,
    tier_level INT NOT NULL,
    threshold_amount DECIMAL(12,2) NOT NULL,
    reward_name VARCHAR(256) NOT NULL,
    esx_item_name VARCHAR(128) NOT NULL,
    esx_item_count INT NOT NULL DEFAULT 1,
    FOREIGN KEY (event_id) REFERENCES milestone_events(id) ON DELETE CASCADE,
    UNIQUE KEY uq_event_tier (event_id, tier_level)
);

CREATE TABLE IF NOT EXISTS milestone_claims (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_id INT UNSIGNED NOT NULL,
    tier_id INT UNSIGNED NOT NULL,
    discord_id VARCHAR(32) NOT NULL,
    identifier VARCHAR(128) NOT NULL,
    claimed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_claim (event_id, tier_id, discord_id),
    FOREIGN KEY (event_id) REFERENCES milestone_events(id),
    FOREIGN KEY (tier_id) REFERENCES milestone_tiers(id)
);

CREATE TABLE IF NOT EXISTS redeem_catalog (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    description TEXT,
    image_url VARCHAR(512),
    point_cost DECIMAL(12,2) NOT NULL,
    esx_item_name VARCHAR(128) NOT NULL,
    esx_item_count INT NOT NULL DEFAULT 1,
    stock_limit INT NOT NULL DEFAULT 0,
    stock_redeemed INT NOT NULL DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS redeem_transactions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    discord_id VARCHAR(32) NOT NULL,
    identifier VARCHAR(128) NOT NULL,
    catalog_id INT UNSIGNED NOT NULL,
    points_spent DECIMAL(12,2) NOT NULL,
    status ENUM('pending','completed','failed','refunded') NOT NULL DEFAULT 'pending',
    delivery_status ENUM('pending','delivered','queued','failed') NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_discord (discord_id),
    FOREIGN KEY (catalog_id) REFERENCES redeem_catalog(id)
);

CREATE TABLE IF NOT EXISTS topup_transactions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tx_ref VARCHAR(64) NOT NULL UNIQUE,
    discord_id VARCHAR(32) NOT NULL,
    identifier VARCHAR(128) NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    points_earned DECIMAL(12,2) NOT NULL DEFAULT 0,
    payment_method ENUM('promptpay','truewallet','creditcard','manual') NOT NULL,
    gateway_ref VARCHAR(128),
    slip_url VARCHAR(512),
    status ENUM('pending','completed','failed','refunded') NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_discord (discord_id),
    INDEX idx_status (status),
    INDEX idx_created (created_at)
);

CREATE TABLE IF NOT EXISTS web_mailbox (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    identifier VARCHAR(128) NOT NULL,
    discord_id VARCHAR(32) NOT NULL,
    payload JSON NOT NULL,
    source_type ENUM('order','redeem','milestone','topup') NOT NULL,
    source_ref VARCHAR(64) NOT NULL,
    status ENUM('pending','claimed','expired') NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    claimed_at TIMESTAMP NULL,
    INDEX idx_identifier_status (identifier, status)
);

CREATE TABLE IF NOT EXISTS admin_users (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    discord_id VARCHAR(32) NOT NULL UNIQUE,
    display_name VARCHAR(256),
    role ENUM('superadmin','admin','viewer') NOT NULL DEFAULT 'admin',
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    admin_discord_id VARCHAR(32) NOT NULL,
    action VARCHAR(128) NOT NULL,
    target_type VARCHAR(64),
    target_id VARCHAR(64),
    detail JSON,
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_admin (admin_discord_id),
    INDEX idx_action (action),
    INDEX idx_created (created_at)
);

CREATE TABLE IF NOT EXISTS delivery_queue (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    identifier VARCHAR(128) NOT NULL,
    discord_id VARCHAR(32) NOT NULL,
    payload JSON NOT NULL,
    source_type ENUM('order','redeem','milestone') NOT NULL,
    source_ref VARCHAR(64) NOT NULL,
    status ENUM('pending','sent','failed','rolled_back') NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status)
);

INSERT INTO shop_categories (slug, name, sort_order) VALUES
('weapons', 'Weapons', 1),
('gacha', 'Gacha', 2),
('vehicles', 'Vehicles', 3),
('points', 'Points', 4),
('packs', 'Packs', 5);
