ALTER TABLE admin_users MODIFY COLUMN role ENUM('admin','dev_admin') NOT NULL DEFAULT 'admin';

CREATE TABLE IF NOT EXISTS admin_accounts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('admin','dev_admin') NOT NULL DEFAULT 'admin',
    discord_id VARCHAR(32) NULL,
    display_name VARCHAR(256),
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    last_login_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_role (role),
    INDEX idx_active (is_active)
);

CREATE TABLE IF NOT EXISTS system_settings (
    setting_key VARCHAR(128) NOT NULL PRIMARY KEY,
    setting_value JSON NOT NULL,
    updated_by VARCHAR(64) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS activity_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    category ENUM('purchase','topup','redeem','milestone','delivery','security','system','admin') NOT NULL,
    actor_type ENUM('user','admin','system') NOT NULL,
    actor_id VARCHAR(128) NOT NULL,
    action VARCHAR(128) NOT NULL,
    target_type VARCHAR(64),
    target_id VARCHAR(64),
    detail JSON,
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_actor (actor_type, actor_id),
    INDEX idx_action (action),
    INDEX idx_created (created_at)
);

ALTER TABLE admin_audit_logs
    ADD COLUMN admin_username VARCHAR(64) NULL AFTER admin_discord_id,
    ADD COLUMN admin_role ENUM('admin','dev_admin') NULL AFTER admin_username,
    ADD COLUMN category ENUM('security','cms','purchase','system','monitoring','user') NOT NULL DEFAULT 'system' AFTER action;

INSERT INTO system_settings (setting_key, setting_value, updated_by) VALUES
('monitoring', JSON_OBJECT(
    'enabled', true,
    'check_interval_sec', 60,
    'cpu_alert_threshold', 80,
    'memory_alert_threshold', 80,
    'alert_webhook_url', '',
    'prometheus_url', '/metrics'
), 'system'),
('autoscale_api', JSON_OBJECT(
    'min_replicas', 2,
    'max_replicas', 10,
    'cpu_target_percent', 60,
    'memory_target_percent', 60,
    'enabled', true
), 'system'),
('autoscale_frontend', JSON_OBJECT(
    'min_replicas', 2,
    'max_replicas', 8,
    'cpu_target_percent', 60,
    'memory_target_percent', 60,
    'enabled', true
), 'system')
ON DUPLICATE KEY UPDATE setting_key = setting_key;

INSERT INTO admin_accounts (username, password_hash, role, display_name) VALUES
('devadmin', '$2a$12$kqegytbyt.F7CAs.99kajuDqgdAUjzRxneto3c5nPTDmUlnnbUbba', 'dev_admin', 'Dev Administrator'),
('admin', '$2a$12$JbVe7ptT84H/Y2t///Eyf.baqSNMtc6rWAbr0WfLWTR.Kfraz6Hka', 'admin', 'Shop Administrator')
ON DUPLICATE KEY UPDATE username = username;
