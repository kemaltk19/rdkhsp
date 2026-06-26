package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	SystemDatabaseURL string
	Env               string
	FrontendURL       string
	JWTSecret         string
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPass          string
	SMTPFrom          string
}

func Load() *Config {
	_ = godotenv.Load()
	env := getEnv("APP_ENV", "development")
	nodeEnv := getEnv("NODE_ENV", "development")
	
	// Rule P1: Email doğrulama prod ve staging'de zorunludur.
	emailConfirm := os.Getenv("email_confirm")
	if (nodeEnv == "production" || env == "production" || nodeEnv == "staging" || env == "staging") && emailConfirm == "false" {
		log.Fatal("[config] NODE_ENV=production/staging ve email_confirm=false iken uygulama baslatilamaz")
	}

	// Rule P2: DATABASE_URL postgres ile baslamalidir.
	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" || !strings.HasPrefix(databaseURL, "postgres") {
		log.Fatal("[config] DATABASE_URL bos veya 'postgres' ile baslamiyor")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	// Rule P2: JWT_SECRET >= 32 karakter olmalıdır.
	if jwtSecret == "" || len(jwtSecret) < 32 || jwtSecret == "degistir-uzun-rastgele" || jwtSecret == "super_secret_key_change_me_in_prod" {
		log.Fatal("[config] JWT_SECRET (>=32 karakter, varsayilan olmayan) zorunludur")
	}

	return &Config{
		Port:              getEnv("PORT", "9001"),
		DatabaseURL:       databaseURL,
		SystemDatabaseURL: getEnv("SYSTEM_DATABASE_URL", ""),
		Env:               env,
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:5173"),
		JWTSecret:         jwtSecret,
		SMTPHost:          getEnv("SMTP_HOST", ""),
		SMTPPort:          getEnv("SMTP_PORT", "587"),
		SMTPUser:          getEnv("SMTP_USER", ""),
		SMTPPass:          getEnv("SMTP_PASS", ""),
		SMTPFrom:          getEnv("SMTP_FROM", "no-reply@radikalhesap.com"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
