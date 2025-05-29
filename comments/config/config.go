package config

import "os"

type Config struct {
	AdrPort string
	URLdb   string
}

func New() *Config {
	return &Config{
		AdrPort: getEnv("COMMENTS_PORT", ""),
		URLdb:   getEnv("COMMENTS_DB", ""),
	}
}

// getEnv вспомогательная функция для считывания окружения или возврата значения по умолчанию
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
