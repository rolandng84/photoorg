package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env            string
	APIHost        string
	APIPort        int
	APICORSOrigins []string
	DatabasePath   string
	LogLevel       string
	ThumbnailDir   string
}

func Load() *Config {
	return &Config{
		Env:            getEnv("ENV", "development"),
		APIHost:        getEnv("API_HOST", "0.0.0.0"),
		APIPort:        getEnvInt("API_PORT", 8012),
		APICORSOrigins: getEnvList("API_CORS_ORIGINS", []string{"http://localhost:3011"}),
		DatabasePath:   getEnv("DATABASE_PATH", "data/photoorg.db"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		ThumbnailDir:   getEnv("THUMBNAIL_DIR", "data/thumbnails"),
	}
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvList(key string, fallback []string) []string {
	if val := os.Getenv(key); val != "" {
		return strings.Split(val, ",")
	}
	return fallback
}
