package config

import "os"

type Config struct {
	AppPort, DBHost, DBPort, DBUser, DBPassword, DBName, DBSSLMode string
	JWTSecret, MidtransBaseURL, MidtransServerKey                 string
}

func Load() Config {
	return Config{
		AppPort:           getEnv("APP_PORT", "8080"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "mini_ecommerce"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		MidtransBaseURL:   getEnv("MIDTRANS_BASE_URL", "https://api.sandbox.midtrans.com"),
		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
