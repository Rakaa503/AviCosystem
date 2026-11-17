package config

import "strings"

type AppConfig struct {
	Name                   string
	Env                    string
	Port                   string
	CORSAllowedOrigins     []string
	RateLimitRequests      int
	RateLimitWindowSeconds int
}

func loadAppConfig() AppConfig {
	origins := getEnv(
		"CORS_ALLOWED_ORIGINS",
		"http://localhost:5173,http://127.0.0.1:5173",
	)

	return AppConfig{
		Name:                   getEnv("APP_NAME", "AviCosystem"),
		Env:                    getEnv("APP_ENV", "development"),
		Port:                   getEnv("APP_PORT", "8080"),
		CORSAllowedOrigins:     strings.Split(origins, ","),
		RateLimitRequests:      getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowSeconds: getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60),
	}
}
