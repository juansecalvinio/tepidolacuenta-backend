package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI             string
	JWTSecret            string
	Port                 string
	GinMode              string
	CORSAllowedOrigins   []string
	FrontendBaseURL      string
	SMTPHost             string
	SMTPPort             string
	SMTPUsername         string
	SMTPPassword         string
	SMTPFrom             string
	GoogleClientID              string
	GoogleClientSecret          string
	GoogleRedirectURL           string
	SentryDSN                   string
	MercadoPagoAccessToken      string
	MercadoPagoWebhookSecret    string
	MercadoPagoNotificationURL  string
}

func Load() (*Config, error) {
	// Load .env file
	_ = godotenv.Load()

	// Parse CORS origins (can be comma-separated)
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	originsSlice := strings.Split(corsOrigins, ",")
	for i := range originsSlice {
		originsSlice[i] = strings.TrimSpace(originsSlice[i])
	}

	return &Config{
		MongoURI:           getEnv("MONGODB_URI", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		Port:               getEnv("PORT", "8080"),
		GinMode:            getEnv("GIN_MODE", "debug"),
		CORSAllowedOrigins: originsSlice,
		FrontendBaseURL:    getEnv("FRONTEND_BASE_URL", "http://localhost:5173"),
		SMTPHost:           getEnv("SMTP_HOST", ""),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		SMTPUsername:       getEnv("SMTP_USERNAME", ""),
		SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:           getEnv("SMTP_FROM", ""),
		GoogleClientID:             getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:         getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:          getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
		SentryDSN:                  getEnv("SENTRY_DSN", ""),
		MercadoPagoAccessToken:     getEnv("MERCADOPAGO_ACCESS_TOKEN", ""),
		MercadoPagoWebhookSecret:   getEnv("MERCADOPAGO_WEBHOOK_SECRET", ""),
		MercadoPagoNotificationURL: getEnv("MERCADOPAGO_NOTIFICATION_URL", ""),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
