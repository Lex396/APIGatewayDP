package config

import "os"

type Config struct {
	Censor   Censor
	Comments Comments
	News     News
	Gateway  Gateway
}

type Censor struct {
	AdrPort string
	URLdb   string
}

type Comments struct {
	AdrPort string
	URLdb   string
}

type News struct {
	AdrPort string
	URLdb   string
}

type Gateway struct {
	AdrPort string
}

func New() *Config {
	return &Config{
		Censor: Censor{
			AdrPort: getEnv("CENSOR_PORT", ""),
			URLdb:   getEnv("CENSOR_DB", ""),
		},
		Comments: Comments{
			AdrPort: getEnv("COMMENTS_PORT", ""),
			URLdb:   getEnv("COMMENTS_DB", ""),
		},
		News: News{
			AdrPort: getEnv("NEWS_PORT", ""),
			URLdb:   getEnv("NEWS_DB", ""),
		},
		Gateway: Gateway{
			AdrPort: getEnv("GATEWAY_PORT", ""),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
