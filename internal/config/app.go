package config

import "strings"

type AppConfig struct {
	Name                   string
	Env                    string
	Port                   string
	CORSAllowedOrigins     []string
	RateLimitRequests      int
	RateLimitWindowSeconds int
	RequestTimeoutSeconds  int
	TrustProxy             bool
	TrustedProxies         []string

	JWTSecret                string
	JWTAccessTokenTTLMinutes int
}

func loadAppConfig() AppConfig {
	origins := getEnv(
		"CORS_ALLOWED_ORIGINS",
		"http://localhost:5173,http://127.0.0.1:5173",
	)

	return AppConfig{
		Name:                     getEnv("APP_NAME", "AviCosystem"),
		Env:                      getEnv("APP_ENV", "development"),
		Port:                     getEnv("APP_PORT", "8080"),
		CORSAllowedOrigins:       splitEnvList(origins),
		RateLimitRequests:        getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowSeconds:   getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60),
		RequestTimeoutSeconds:    getEnvInt("REQUEST_TIMEOUT_SECONDS", 30),
		TrustProxy:               getEnvBool("TRUST_PROXY", false),
		TrustedProxies:           splitEnvList(getEnv("TRUSTED_PROXIES", "")),
		JWTSecret:                getEnv("JWT_SECRET", ""),
		JWTAccessTokenTTLMinutes: getEnvInt("JWT_ACCESS_TOKEN_TTL_MINUTES", 15),
	}
}

func splitEnvList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}
