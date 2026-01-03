package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI            string
	JWTSecret           string
	Port                string
	GinMode             string
	CORSAllowedOrigins  string
	FrontendBaseURL     string
}

func Load() (*Config, error) {
	// Load .env file
	_ = godotenv.Load()

	return &Config{
		MongoURI:           getEnv("MONGODB_URI", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		Port:               getEnv("PORT", "8080"),
		GinMode:            getEnv("GIN_MODE", "debug"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
		FrontendBaseURL:    getEnv("FRONTEND_BASE_URL", "http://localhost:5173"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
