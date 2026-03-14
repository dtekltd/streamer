package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type AppConfig struct {
	ServerMode        string
	ServerPort        string
	StreamURLTemplate string
	EnableLogging     bool
}

func Load() AppConfig {
	loadDotEnv(".env")

	cfg := AppConfig{
		ServerMode:        normalizeAppMode(getEnv("SERVER_MODE", "prod")),
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		StreamURLTemplate: getEnv("STREAM_URL_TEMPLATE", "rtmp://10.16.0.165:1935/live/%s"),
		EnableLogging:     getEnvBool("ENABLE_LOGGING", true),
	}

	cfg.ServerPort = normalizePort(cfg.ServerPort)
	return cfg
}

func normalizePort(port string) string {
	trimmed := strings.TrimSpace(port)
	if trimmed == "" {
		return "8080"
	}
	return strings.TrimPrefix(trimmed, ":")
}

func normalizeAppMode(mode string) string {
	trimmed := strings.ToLower(strings.TrimSpace(mode))
	if trimmed == "dev" {
		return "dev"
	}
	return "prod"
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(value) == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func loadDotEnv(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}

	file, err := os.Open(absPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"")

		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}
