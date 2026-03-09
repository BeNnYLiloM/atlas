package config

import (
	"os"
	"strconv"
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
	Port string
	Mode string // development, production
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
	Secret     string
	ExpireHour int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Mode: getEnv("MODE", "development"),
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
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpireHour: 24,
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
