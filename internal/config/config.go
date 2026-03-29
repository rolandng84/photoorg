package config

import (
	"os"
	"path/filepath"
	"runtime"
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
	DataDir        string
}

// Load builds Config from environment variables. dataDir overrides the
// OS-default data location; pass "" to use the platform default.
func Load(dataDir string) *Config {
	if dataDir == "" {
		dataDir = defaultDataDir()
	}

	dbPath := getEnv("DATABASE_PATH", filepath.Join(dataDir, "photoorg.db"))
	thumbDir := getEnv("THUMBNAIL_DIR", filepath.Join(dataDir, "thumbnails"))

	return &Config{
		Env:            getEnv("ENV", "development"),
		APIHost:        getEnv("API_HOST", "0.0.0.0"),
		APIPort:        getEnvInt("API_PORT", 8012),
		APICORSOrigins: getEnvList("API_CORS_ORIGINS", []string{"http://localhost:3011"}),
		DatabasePath:   dbPath,
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		ThumbnailDir:   thumbDir,
		DataDir:        dataDir,
	}
}

// defaultDataDir returns the OS-standard application data directory for photoorg.
//
//   - Linux/BSD:  $XDG_DATA_HOME/photoorg   (~/.local/share/photoorg)
//   - macOS:      ~/Library/Application Support/photoorg
//   - Windows:    %APPDATA%\photoorg
func defaultDataDir() string {
	switch runtime.GOOS {
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "photoorg")
		}
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support", "photoorg")
		}
	default: // linux, freebsd, etc.
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "photoorg")
		}
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".local", "share", "photoorg")
		}
	}
	// Final fallback — should be unreachable on normal systems
	return filepath.Join("data")
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
