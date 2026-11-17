package config

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  int
	ReadTimeout  int
	WriteTimeout int
}

func loadRedisConfig() RedisConfig {
	return RedisConfig{
		Host:         getEnv("REDIS_HOST", "localhost"),
		Port:         getEnv("REDIS_PORT", "6379"),
		Password:     getEnv("REDIS_PASSWORD", ""),
		DB:           getEnvInt("REDIS_DB", 0),
		PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
		MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 2),
		MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
		DialTimeout:  getEnvInt("REDIS_DIAL_TIMEOUT_SECONDS", 5),
		ReadTimeout:  getEnvInt("REDIS_READ_TIMEOUT_SECONDS", 3),
		WriteTimeout: getEnvInt("REDIS_WRITE_TIMEOUT_SECONDS", 3),
	}
}
