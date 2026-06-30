INSERT INTO admin_settings (setting_key, setting_value, updated_by) VALUES
('autoscale_api', JSON_OBJECT(
    'min_replicas', 3,
    'max_replicas', 16,
    'cpu_target_percent', 55,
    'memory_target_percent', 65,
    'enabled', true
), 'system'),
('autoscale_frontend', JSON_OBJECT(
    'min_replicas', 2,
    'max_replicas', 12,
    'cpu_target_percent', 55,
    'memory_target_percent', 65,
    'enabled', true
), 'system')
ON DUPLICATE KEY UPDATE
    setting_value = VALUES(setting_value),
    updated_by = VALUES(updated_by);
