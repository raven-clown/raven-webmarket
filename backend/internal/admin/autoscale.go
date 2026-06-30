package admin

import "fmt"

const k8sNamespace = "raven-webmarket"

func defaultAutoscaleAPI() AutoscaleConfig {
	return AutoscaleConfig{
		MinReplicas: 3, MaxReplicas: 16,
		CPUTargetPercent: 55, MemoryTargetPercent: 65, Enabled: true,
	}
}

func defaultAutoscaleFrontend() AutoscaleConfig {
	return AutoscaleConfig{
		MinReplicas: 2, MaxReplicas: 12,
		CPUTargetPercent: 55, MemoryTargetPercent: 65, Enabled: true,
	}
}

func normalizeAutoscale(cfg AutoscaleConfig, fallback AutoscaleConfig) AutoscaleConfig {
	if cfg.MinReplicas <= 0 {
		cfg.MinReplicas = fallback.MinReplicas
	}
	if cfg.MaxReplicas <= 0 {
		cfg.MaxReplicas = fallback.MaxReplicas
	}
	if cfg.CPUTargetPercent <= 0 {
		cfg.CPUTargetPercent = fallback.CPUTargetPercent
	}
	if cfg.MemoryTargetPercent <= 0 {
		cfg.MemoryTargetPercent = fallback.MemoryTargetPercent
	}
	if cfg.MaxReplicas < cfg.MinReplicas {
		cfg.MaxReplicas = cfg.MinReplicas
	}
	return cfg
}

func BuildHPAYAML(name, deployment string, cfg AutoscaleConfig) string {
	if !cfg.Enabled {
		return fmt.Sprintf("# HPA disabled for %s in admin settings\n", deployment)
	}
	return fmt.Sprintf(`apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: %s
  namespace: %s
  labels:
    app: %s
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: %s
  minReplicas: %d
  maxReplicas: %d
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: %d
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: %d
`, name, k8sNamespace, deployment, deployment, cfg.MinReplicas, cfg.MaxReplicas, cfg.CPUTargetPercent, cfg.MemoryTargetPercent)
}
