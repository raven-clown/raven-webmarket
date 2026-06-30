CREATE TABLE IF NOT EXISTS site_posts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    post_type ENUM('announcement','daily_update','ad','notice') NOT NULL DEFAULT 'announcement',
    title_en VARCHAR(512) NOT NULL,
    title_th VARCHAR(512) NULL,
    body_en TEXT NULL,
    body_th TEXT NULL,
    image_url VARCHAR(512) NULL,
    link_url VARCHAR(512) NULL,
    placement ENUM('home','sidebar','news','banner','all') NOT NULL DEFAULT 'home',
    sort_order INT NOT NULL DEFAULT 0,
    is_pinned TINYINT(1) NOT NULL DEFAULT 0,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    publish_date DATE NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_type_active (post_type, is_active),
    INDEX idx_placement (placement, is_active),
    INDEX idx_publish (publish_date)
);

CREATE TABLE IF NOT EXISTS forum_threads (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    discord_id VARCHAR(32) NOT NULL,
    author_name VARCHAR(256) NOT NULL,
    title VARCHAR(512) NOT NULL,
    body TEXT NOT NULL,
    is_pinned TINYINT(1) NOT NULL DEFAULT 0,
    is_locked TINYINT(1) NOT NULL DEFAULT 0,
    reply_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_pinned (is_pinned, created_at),
    INDEX idx_author (discord_id)
);

CREATE TABLE IF NOT EXISTS forum_replies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    thread_id INT UNSIGNED NOT NULL,
    discord_id VARCHAR(32) NOT NULL,
    author_name VARCHAR(256) NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_thread (thread_id, created_at),
    FOREIGN KEY (thread_id) REFERENCES forum_threads(id) ON DELETE CASCADE
);

INSERT INTO site_posts (post_type, title_en, title_th, body_en, body_th, placement, is_pinned, publish_date) VALUES
('announcement', 'Welcome to Raven Webmarket', 'ยินดีต้อนรับสู่ Raven Webmarket',
 'Top up, shop, and redeem rewards for your FiveM ESX character. Login with Discord linked to your in-game account.',
 'เติมเงิน ช้อป และแลกของรางวัลสำหรับตัวละคร FiveM ESX ของคุณ ล็อกอินด้วย Discord ที่ผูกกับบัญชีในเกม',
 'home', 1, CURDATE()),
('daily_update', 'Server Patch Notes', 'บันทึกการอัปเดตเซิร์ฟเวอร์',
 'Daily maintenance completed. Shop catalog synced with in-game items.',
 'บำรุงรักษารายวันเสร็จสิ้น คatalog ร้านค้าซิงค์กับไอเทมในเกมแล้ว',
 'news', 0, CURDATE()),
('ad', 'Limited Time Promo', 'โปรโมชั่นจำกัดเวลา',
 'Check featured items for exclusive discounts this week.',
 'ดูสินค้าแนะนำสำหรับส่วนลดพิเศษสัปดาห์นี้',
 'sidebar', 0, CURDATE()),
('notice', 'Cookie & Privacy', 'คุกกี้และความเป็นส่วนตัว',
 'We use essential cookies for login sessions and preferences. See our cookie banner for details.',
 'เราใช้คุกกี้ที่จำเป็นสำหรับเซสชันล็อกอินและการตั้งค่า ดูรายละเอียดในแบนเนอร์คุกกี้',
 'all', 0, CURDATE())
ON DUPLICATE KEY UPDATE title_en = title_en;
