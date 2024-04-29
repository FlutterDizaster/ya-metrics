package logger

import (
	"log/slog"
	"os"

	"github.com/rs/zerolog"
)

func Init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	slog.SetDefault(logger)
}

// For passing autotests.
func Zero() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}
