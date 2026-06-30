package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env              string
	APIHost          string
	APIPort          string
	APIBaseURL       string
	FrontendURL      string
	CORSOrigins      []string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBMaxOpen        int
	DBMaxIdle        int
	ESXDBHost        string
	ESXDBPort        string
	ESXDBUser        string
	ESXDBPassword    string
	ESXDBName        string
	RedisHost        string
	RedisPort        string
	RedisPassword    string
	RedisSessionDB   int
	RedisCartDB      int
	RedisRateLimitDB int
	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURI  string
	DiscordAdminIDs     []string
	DiscordWebhookURL   string
	SessionSecret    string
	JWTSecret        string
	MinioEndpoint    string
	MinioAccessKey   string
	MinioSecretKey   string
	MinioBucket      string
	MinioUseSSL      bool
	MinioPublicURL   string
	PaymentGatewayURL    string
	PaymentGatewayKey    string
	PaymentGatewaySecret string
	PaymentWebhookSecret string
	FiveMWebhookURL   string
	FiveMWebhookSecret string
	FiveMAPIKey       string
	RedeemPointsPerBaht float64
	MinTopupAmount      float64
	RateLimitRequests   int
	RateLimitWindowSec  int
	TrustedProxies      []string
	TrustCloudflare     bool
	MongoEnabled        bool
	MongoURI            string
	MongoHost           string
	MongoPort           string
	MongoUser           string
	MongoPassword       string
	MongoDBName         string
}

func Load() *Config {
	return &Config{
		Env:              getEnv("APP_ENV", "development"),
		APIHost:          getEnv("API_HOST", "0.0.0.0"),
		APIPort:          getEnv("API_PORT", "8080"),
		APIBaseURL:       getEnv("API_BASE_URL", "http://localhost:8080"),
		FrontendURL:      getEnv("FRONTEND_URL", "http://localhost:3000"),
		CORSOrigins:      splitCSV(getEnv("CORS_ORIGINS", "http://localhost:3000")),
		DBHost:           getEnv("DB_HOST", "127.0.0.1"),
		DBPort:           getEnv("DB_PORT", "3306"),
		DBUser:           getEnv("DB_USER", "root"),
		DBPassword:       getEnv("DB_PASSWORD", ""),
		DBName:           getEnv("DB_NAME", "raven_webmarket"),
		DBMaxOpen:        getEnvInt("DB_MAX_OPEN", 25),
		DBMaxIdle:        getEnvInt("DB_MAX_IDLE", 10),
		ESXDBHost:        getEnv("ESX_DB_HOST", "127.0.0.1"),
		ESXDBPort:        getEnv("ESX_DB_PORT", "3306"),
		ESXDBUser:        getEnv("ESX_DB_USER", "root"),
		ESXDBPassword:    getEnv("ESX_DB_PASSWORD", ""),
		ESXDBName:        getEnv("ESX_DB_NAME", "es_extended"),
		RedisHost:        getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort:        getEnv("REDIS_PORT", "6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisSessionDB:   getEnvInt("REDIS_SESSION_DB", 0),
		RedisCartDB:      getEnvInt("REDIS_CART_DB", 1),
		RedisRateLimitDB: getEnvInt("REDIS_RATELIMIT_DB", 2),
		DiscordClientID:     getEnv("DISCORD_CLIENT_ID", ""),
		DiscordClientSecret: getEnv("DISCORD_CLIENT_SECRET", ""),
		DiscordRedirectURI:  getEnv("DISCORD_REDIRECT_URI", ""),
		DiscordAdminIDs:     splitCSV(getEnv("DISCORD_ADMIN_IDS", "")),
		DiscordWebhookURL:   getEnv("DISCORD_WEBHOOK_URL", ""),
		SessionSecret:    getEnv("SESSION_SECRET", "dev-secret"),
		JWTSecret:        getEnv("JWT_SECRET", "dev-jwt-secret"),
		MinioEndpoint:    getEnv("MINIO_ENDPOINT", "127.0.0.1:9000"),
		MinioAccessKey:   getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:   getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:      getEnv("MINIO_BUCKET", "slips"),
		MinioUseSSL:      getEnv("MINIO_USE_SSL", "false") == "true",
		MinioPublicURL:   getEnv("MINIO_PUBLIC_URL", "http://localhost:9000/slips"),
		PaymentGatewayURL:    getEnv("PAYMENT_GATEWAY_URL", ""),
		PaymentGatewayKey:    getEnv("PAYMENT_GATEWAY_KEY", ""),
		PaymentGatewaySecret: getEnv("PAYMENT_GATEWAY_SECRET", ""),
		PaymentWebhookSecret: getEnv("PAYMENT_WEBHOOK_SECRET", ""),
		FiveMWebhookURL:   getEnv("FIVEM_WEBHOOK_URL", ""),
		FiveMWebhookSecret: getEnv("FIVEM_WEBHOOK_SECRET", ""),
		FiveMAPIKey:       getEnv("FIVEM_API_KEY", ""),
		RedeemPointsPerBaht: getEnvFloat("REDEEM_POINTS_PER_BAHT", 1),
		MinTopupAmount:      getEnvFloat("MIN_TOPUP_AMOUNT", 50),
		RateLimitRequests:   getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowSec:  getEnvInt("RATE_LIMIT_WINDOW_SEC", 60),
		TrustedProxies:      splitCSV(getEnv("TRUSTED_PROXIES", "127.0.0.1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16")),
		TrustCloudflare:     getEnv("TRUST_CLOUDFLARE", "false") == "true",
		MongoEnabled:        getEnv("MONGO_ENABLED", "false") == "true",
		MongoURI:            getEnv("MONGO_URI", ""),
		MongoHost:           getEnv("MONGO_HOST", "127.0.0.1"),
		MongoPort:           getEnv("MONGO_PORT", "27017"),
		MongoUser:           getEnv("MONGO_USER", ""),
		MongoPassword:       getEnv("MONGO_PASSWORD", ""),
		MongoDBName:         getEnv("MONGO_DB_NAME", "raven_webmarket"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return fallback
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
