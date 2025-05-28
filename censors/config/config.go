// config/config.go
package config

import "os"

type Config struct {
	Censor Censor
}

type Censor struct {
	AdrPort string
}

func New() *Config {
	return &Config{
		Censor: Censor{
			AdrPort: getEnv("CENSOR_PORT", ":8083"),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
