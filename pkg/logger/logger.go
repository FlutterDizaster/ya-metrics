package logger

import (
	"log/slog"
	"os"

	"github.com/rs/zerolog"
)

// Функция инициализации логгера.
// Параметр level принимает уровень логгирования.
func New(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	slog.SetDefault(logger)
}

// For passing autotests.
func Zero() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}
