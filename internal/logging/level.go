package logging

import (
	"log/slog"
	"strings"
)

func SetLevel(level string) {
	slogLevel := getLevel(level)
	slog.SetLogLoggerLevel(slogLevel)
}

func getLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}
	return slog.LevelInfo
}
