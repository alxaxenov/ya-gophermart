// Package logger предоставляет логгер
package logger

import (
	"log/slog"
	"os"
)

// Logger используется для хранения инстанса логгера.
var Logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

// Initialize инициализация логгера.
func Initialize() error {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewJSONHandler(os.Stderr, opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
	return nil
}
