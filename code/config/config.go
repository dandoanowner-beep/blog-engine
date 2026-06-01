package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL     string
	JWTSecret       string
	JWTRefreshSecret string
	SMTPHost        string
	SMTPPort        string
	SMTPUser        string
	SMTPPass        string
	SMTPFrom        string
	GoogleClientID  string
	GoogleClientSecret string
	GoogleRedirectURL  string
	R2AccountID     string
	R2AccessKey     string
	R2SecretKey     string
	R2BucketName    string
	R2PublicURL     string
	AppURL          string
	Port            string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:        requireEnv("DATABASE_URL"),
		JWTSecret:          requireEnv("JWT_SECRET"),
		JWTRefreshSecret:   requireEnv("JWT_REFRESH_SECRET"),
		SMTPHost:           getEnv("SMTP_HOST", "localhost"),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		SMTPUser:           getEnv("SMTP_USER", ""),
		SMTPPass:           getEnv("SMTP_PASS", ""),
		SMTPFrom:           getEnv("SMTP_FROM", "noreply@blog-engine.com"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
		R2AccountID:        requireEnv("R2_ACCOUNT_ID"),
		R2AccessKey:        requireEnv("R2_ACCESS_KEY_ID"),
		R2SecretKey:        requireEnv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:       requireEnv("R2_BUCKET_NAME"),
		R2PublicURL:        requireEnv("R2_PUBLIC_URL"),
		AppURL:             getEnv("APP_URL", "http://localhost:3000"),
		Port:               getEnv("PORT", "8080"),
	}
	return cfg, nil
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
