package config

import "strings"

type AppConfig struct {
	ServerMode     string
	ServerPort     string
	EnableLogging  bool
	FFmpegLogLevel string
	LoginPIN       string
}

func New(mode, port string, enableLogging bool, ffmpegLogLevel, loginPIN string) *AppConfig {
	return &AppConfig{
		ServerMode:     normalizeAppMode(mode),
		ServerPort:     normalizePort(port),
		EnableLogging:  enableLogging,
		FFmpegLogLevel: normalizeFFmpegLogLevel(ffmpegLogLevel),
		LoginPIN:       normalizeLoginPIN(loginPIN),
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

func normalizeFFmpegLogLevel(level string) string {
	trimmed := strings.ToLower(strings.TrimSpace(level))
	switch trimmed {
	case "quiet", "panic", "fatal", "error", "warning", "info", "verbose", "debug":
		return trimmed
	default:
		return "error"
	}
}

func normalizeLoginPIN(pin string) string {
	trimmed := strings.TrimSpace(pin)
	if trimmed == "" {
		return "gostream"
	}
	return trimmed
}
