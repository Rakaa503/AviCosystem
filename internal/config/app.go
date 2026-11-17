package config

type AppConfig struct {
	Name string
	Env  string
	Port string
}

func loadAppConfig() AppConfig {
	return AppConfig{
		Name: getEnv("APP_NAME", "AviCosystem"),
		Env:  getEnv("APP_ENV", "development"),
		Port: getEnv("APP_PORT", "8080"),
	}
}
