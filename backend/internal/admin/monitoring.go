package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/raven-clown/raven-webmarket/backend/internal/config"
	"github.com/raven-clown/raven-webmarket/backend/internal/database"
	"github.com/raven-clown/raven-webmarket/backend/internal/middleware"
	"github.com/raven-clown/raven-webmarket/backend/internal/rbac"
	redisstore "github.com/raven-clown/raven-webmarket/backend/internal/redisstore"
)

type MonitoringConfig struct {
	Enabled              bool    `json:"enabled"`
	CheckIntervalSec     int     `json:"check_interval_sec"`
	CPUAlertThreshold    float64 `json:"cpu_alert_threshold"`
	MemoryAlertThreshold float64 `json:"memory_alert_threshold"`
	AlertWebhookURL      string  `json:"alert_webhook_url"`
	PrometheusURL        string  `json:"prometheus_url"`
}

type AutoscaleConfig struct {
	MinReplicas           int  `json:"min_replicas"`
	MaxReplicas           int  `json:"max_replicas"`
	CPUTargetPercent      int  `json:"cpu_target_percent"`
	MemoryTargetPercent   int  `json:"memory_target_percent"`
	Enabled               bool `json:"enabled"`
}

type HealthReport struct {
	Status    string                       `json:"status"`
	CheckedAt time.Time                    `json:"checked_at"`
	Services  map[string]ServiceHealth     `json:"services"`
	Config    MonitoringConfig             `json:"config"`
}

type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

func (s *Service) GetSetting(ctx context.Context, key string, dest interface{}) error {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT setting_value FROM system_settings WHERE setting_key = ?`, key).Scan(&raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dest)
}

func (s *Service) SaveSetting(ctx context.Context, key string, value interface{}, updatedBy string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO system_settings (setting_key, setting_value, updated_by)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE setting_value = VALUES(setting_value), updated_by = VALUES(updated_by)`,
		key, string(data), updatedBy)
	return err
}

func (s *Service) GetMonitoringConfig(ctx context.Context) (MonitoringConfig, error) {
	var cfg MonitoringConfig
	err := s.GetSetting(ctx, "monitoring", &cfg)
	return cfg, err
}

func (s *Service) SaveMonitoringConfig(ctx context.Context, actor middleware.AdminClaims, ip string, cfg MonitoringConfig) error {
	if !rbac.Can(actor.Role, "monitoring_edit") {
		return fmt.Errorf("forbidden")
	}
	if err := s.SaveSetting(ctx, "monitoring", cfg, actor.Username); err != nil {
		return err
	}
	detail, _ := json.Marshal(cfg)
	s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "monitoring", "update_monitoring_config", "setting", "monitoring", string(detail), ip)
	return nil
}

func (s *Service) GetAutoscaleConfig(ctx context.Context) (map[string]AutoscaleConfig, error) {
	var apiCfg AutoscaleConfig
	var feCfg AutoscaleConfig
	if err := s.GetSetting(ctx, "autoscale_api", &apiCfg); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err := s.GetSetting(ctx, "autoscale_frontend", &feCfg); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	apiCfg = normalizeAutoscale(apiCfg, defaultAutoscaleAPI())
	feCfg = normalizeAutoscale(feCfg, defaultAutoscaleFrontend())
	return map[string]AutoscaleConfig{"api": apiCfg, "frontend": feCfg}, nil
}

func (s *Service) SaveAutoscaleConfig(ctx context.Context, actor middleware.AdminClaims, ip, target string, cfg AutoscaleConfig) error {
	if !rbac.Can(actor.Role, "autoscale") {
		return fmt.Errorf("forbidden")
	}
	key := "autoscale_api"
	if target == "frontend" {
		key = "autoscale_frontend"
	}
	if err := s.SaveSetting(ctx, key, cfg, actor.Username); err != nil {
		return err
	}
	detail, _ := json.Marshal(map[string]interface{}{"target": target, "config": cfg})
	s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "system", "update_autoscale_config", "setting", key, string(detail), ip)
	return nil
}

func (s *Service) HealthReport(ctx context.Context, cfg *config.Config, redis *redisstore.Store) HealthReport {
	report := HealthReport{
		Status:    "healthy",
		CheckedAt: time.Now().UTC(),
		Services:  map[string]ServiceHealth{},
	}
	monCfg, _ := s.GetMonitoringConfig(ctx)
	report.Config = monCfg
	start := time.Now()
	if err := s.db.PingContext(ctx); err != nil {
		report.Services["database"] = ServiceHealth{Status: "down", Message: err.Error()}
		report.Status = "degraded"
	} else {
		report.Services["database"] = ServiceHealth{Status: "up", Latency: time.Since(start).String()}
	}
	start = time.Now()
	if err := redis.Session.Ping(ctx).Err(); err != nil {
		report.Services["redis"] = ServiceHealth{Status: "down", Message: err.Error()}
		report.Status = "degraded"
	} else {
		report.Services["redis"] = ServiceHealth{Status: "up", Latency: time.Since(start).String()}
	}
	if cfg.MongoEnabled {
		start = time.Now()
		if err := database.PingMongoOptional(ctx, cfg); err != nil {
			report.Services["mongodb"] = ServiceHealth{Status: "down", Message: err.Error()}
			report.Status = "degraded"
		} else {
			report.Services["mongodb"] = ServiceHealth{Status: "up", Latency: time.Since(start).String(), Message: "optional"}
		}
	} else {
		report.Services["mongodb"] = ServiceHealth{Status: "disabled", Message: "MONGO_ENABLED=false"}
	}
	report.Services["api"] = ServiceHealth{Status: "up"}
	var pendingOrders int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM shop_orders WHERE status IN ('pending','processing')`).Scan(&pendingOrders)
	var failedDelivery int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM delivery_queue WHERE status = 'failed'`).Scan(&failedDelivery)
	report.Services["queue"] = ServiceHealth{
		Status:  "up",
		Message: fmt.Sprintf("pending_orders=%d failed_delivery=%d", pendingOrders, failedDelivery),
	}
	if failedDelivery > 0 && report.Status == "healthy" {
		report.Status = "degraded"
	}
	return report
}

func (s *Service) AutoscaleManifest(ctx context.Context) (map[string]interface{}, error) {
	cfg, err := s.GetAutoscaleConfig(ctx)
	if err != nil {
		return nil, err
	}
	apiYAML := BuildHPAYAML("raven-api-hpa", "raven-api", cfg["api"])
	feYAML := BuildHPAYAML("raven-frontend-hpa", "raven-frontend", cfg["frontend"])
	return map[string]interface{}{
		"namespace": k8sNamespace,
		"config":    cfg,
		"yaml": map[string]string{
			"api_hpa":       apiYAML,
			"frontend_hpa":  feYAML,
		},
		"apply_commands": []string{
			"kubectl apply -k deploy/kubernetes/",
			"kubectl apply -f deploy/kubernetes/hpa-api.yaml",
			"kubectl apply -f deploy/kubernetes/hpa-frontend.yaml",
			"kubectl get hpa -n raven-webmarket",
			"kubectl describe hpa raven-api-hpa -n raven-webmarket",
		},
	}, nil
}
func (s *Service) UpsertBanner(ctx context.Context, actor middleware.AdminClaims, ip string, data map[string]interface{}) error {
	if !rbac.CanWithPermissions(actor.Role, actor.Permissions, "cms") {
		return fmt.Errorf("forbidden")
	}
	detail, _ := json.Marshal(data)
	s.LogDetailed(ctx, actor.DiscordIDValue(), actor.Username, actor.Role, "cms", "upsert_banner", "banner", fmt.Sprint(data["title"]), string(detail), ip)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO shop_banners (title, image_url, link_url, sort_order, is_active)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			title = VALUES(title), image_url = VALUES(image_url), link_url = VALUES(link_url),
			sort_order = VALUES(sort_order), is_active = VALUES(is_active)`,
		data["title"], data["image_url"], data["link_url"], data["sort_order"], data["is_active"])
	return err
}

func (s *Service) ResetAccumulations(ctx context.Context, actor middleware.AdminClaims, ip string, includeRedeem bool) error {
	return s.ResetMonthlyAccumulation(ctx, &actor, ip, includeRedeem)
}
