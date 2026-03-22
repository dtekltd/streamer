package config

import "strings"

type AppConfig struct {
	ServerMode    string
	ServerPort    string
	EnableLogging bool
}

func New(mode, port string, enableLogging bool) *AppConfig {
	return &AppConfig{
		ServerMode:    normalizeAppMode(mode),
		ServerPort:    normalizePort(port),
		EnableLogging: enableLogging,
	}
}

func (cfg *AppConfig) IsDevMode() bool {
	return cfg.ServerMode == "dev"
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
