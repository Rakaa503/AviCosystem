package config

type DatabaseConfig struct {
	Host                   string
	Port                   string
	User                   string
	Password               string
	Name                   string
	SSLMode                string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeSeconds int
	ConnMaxIdleTimeSeconds int
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:                   getEnv("DATABASE_HOST", "localhost"),
		Port:                   getEnv("DATABASE_PORT", "5432"),
		User:                   getEnv("DATABASE_USER", "postgres"),
		Password:               getEnv("DATABASE_PASSWORD", ""),
		Name:                   getEnv("DATABASE_NAME", "avicosystem"),
		SSLMode:                getEnv("DATABASE_SSLMODE", "disable"),
		MaxOpenConns:           getEnvInt("DATABASE_MAX_OPEN_CONNS", 25),
		MaxIdleConns:           getEnvInt("DATABASE_MAX_IDLE_CONNS", 10),
		ConnMaxLifetimeSeconds: getEnvInt("DATABASE_CONN_MAX_LIFETIME_SECONDS", 1800),
		ConnMaxIdleTimeSeconds: getEnvInt("DATABASE_CONN_MAX_IDLE_TIME_SECONDS", 300),
	}
}
