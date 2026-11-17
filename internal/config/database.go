package config

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DATABASE_HOST", "localhost"),
		Port:     getEnv("DATABASE_PORT", "5432"),
		User:     getEnv("DATABASE_USER", "postgres"),
		Password: getEnv("DATABASE_PASSWORD", ""),
		Name:     getEnv("DATABASE_NAME", "avicosystem"),
		SSLMode:  getEnv("DATABASE_SSLMODE", "disable"),
	}
}
