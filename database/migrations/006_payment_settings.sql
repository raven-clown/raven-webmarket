INSERT INTO system_settings (setting_key, setting_value, updated_by) VALUES
('payment_settings', JSON_OBJECT(
    'min_topup_amount', 50,
    'redeem_points_per_baht', 1
), 'system')
ON DUPLICATE KEY UPDATE setting_key = setting_key;
