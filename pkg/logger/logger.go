package logger

import (
	"log/slog"
	"os"

	"github.com/rs/zerolog"
)

func New() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	slog.SetDefault(logger)
}

// For passing autotests.
func Zero() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}
