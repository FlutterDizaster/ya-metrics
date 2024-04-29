package logger

import (
	"log/slog"
	"os"

	"github.com/rs/zerolog"
)

func Init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)
}

// For passing autotests.
func Zero() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}
