package config

import "os"

type Config struct {
	AdrPort string
	URLdb   string
}

// New возвращает новую Config структуру
func New() *Config {
	return &Config{
		AdrPort: getEnv("NEWS_PORT", ""),
		URLdb:   getEnv("NEWS_DB", ""),
	}
}

// Простая вспомогательная функция для считывания окружения или возврата значения по умолчанию
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
