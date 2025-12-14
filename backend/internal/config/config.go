package config

import (
	"os"
	"time"
)

type Config struct {
	ServerAddr      string
	ServerEnv       string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiry       time.Duration
	CardTokenSecret string
	TLSCert         string
	TLSKey          string
	CORSOrigins     string
	LogLevel        string
	AdminPassword   string
	BaseURL         string
	ServiceUserID   string // 小安服務帳號 ID，新用戶自動加為好友
}

func Load() *Config {
	secret := getEnv("JWT_SECRET", "")
	if len(secret) < 32 {
		panic("JWT_SECRET must be at least 32 characters")
	}

	cardSecret := getEnv("CARD_TOKEN_SECRET", "")
	if cardSecret == "" {
		panic("CARD_TOKEN_SECRET is required")
	}

	adminPassword := getEnv("ADMIN_PASSWORD", "")
	if adminPassword == "" {
		panic("ADMIN_PASSWORD is required")
	}

	expiry, _ := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	return &Config{
		ServerAddr:      getEnv("SERVER_ADDR", ":8443"),
		ServerEnv:       getEnv("SERVER_ENV", "development"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       secret,
		JWTExpiry:       expiry,
		CardTokenSecret: cardSecret,
		TLSCert:         getEnv("TLS_CERT_FILE", "../certs/localhost+2.pem"),
		TLSKey:          getEnv("TLS_KEY_FILE", "../certs/localhost+2-key.pem"),
		CORSOrigins:     getEnv("CORS_ORIGINS", "https://localhost:5173"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		AdminPassword:   adminPassword,
		BaseURL:         getEnv("BASE_URL", "https://localhost:5173"),
		ServiceUserID:   getEnv("SERVICE_USER_ID", ""), // 可選，設定後新用戶自動加好友
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
