package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	MinIO    MinIOConfig
	LiveKit  LiveKitConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port           string
	Mode           string
	AllowedOrigins []string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	PublicURL string
}

type LiveKitConfig struct {
	Host      string
	URL       string
	APIKey    string
	APISecret string
}

type JWTConfig struct {
	Secret                string
	AccessTokenTTLMinutes int
	RefreshTokenTTLDays   int
	Issuer                string
	Audience              string
	RefreshCookieName     string
	RefreshCookieDomain   string
	RefreshCookieSecure   bool
}

func Load() *Config {
	serverMode := getEnv("MODE", "development")

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Mode: serverMode,
			AllowedOrigins: getCSVEnv(
				"CORS_ALLOWED_ORIGINS",
				[]string{
					"http://localhost:5173",
					"http://127.0.0.1:5173",
					"http://localhost:4173",
					"http://127.0.0.1:4173",
				},
			),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5433"),
			User:     getEnv("DB_USER", "atlas"),
			Password: getEnv("DB_PASSWORD", "atlas"),
			DBName:   getEnv("DB_NAME", "atlas"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("MINIO_BUCKET", "atlas"),
			UseSSL:    getBoolEnv("MINIO_USE_SSL", false),
			PublicURL: getEnv("MINIO_PUBLIC_URL", ""),
		},
		LiveKit: LiveKitConfig{
			Host:      getEnv("LIVEKIT_HOST", "localhost:7880"),
			URL:       getEnv("LIVEKIT_URL", "ws://localhost:7880"),
			APIKey:    getEnv("LIVEKIT_API_KEY", "devkey"),
			APISecret: getEnv("LIVEKIT_API_SECRET", "secret_replace_in_production_min_32_chars!!"),
		},
		JWT: JWTConfig{
			Secret:                getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenTTLMinutes: getIntEnv("JWT_ACCESS_TTL_MINUTES", 15),
			RefreshTokenTTLDays:   getIntEnv("JWT_REFRESH_TTL_DAYS", 14),
			Issuer:                getEnv("JWT_ISSUER", "atlas"),
			Audience:              getEnv("JWT_AUDIENCE", "atlas-web"),
			RefreshCookieName:     getEnv("JWT_REFRESH_COOKIE_NAME", "atlas_refresh_token"),
			RefreshCookieDomain:   getEnv("JWT_REFRESH_COOKIE_DOMAIN", ""),
			RefreshCookieSecure:   getBoolEnv("JWT_REFRESH_COOKIE_SECURE", serverMode == "production"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func getCSVEnv(key string, defaultValue []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	if len(origins) == 0 {
		return defaultValue
	}
	return origins
}
