package logging

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

var key = ctxKey{}

func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, key, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	logger := ctx.Value(key)
	if logger != nil {
		switch logger := logger.(type) {
		case *slog.Logger:
			return logger
		}
	}

	return slog.Default()
}
