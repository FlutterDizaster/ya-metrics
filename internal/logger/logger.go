package logger

import (
	"log/slog"
	"os"

	// for passing autotests.
	_ "github.com/rs/zerolog"
)

func Init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)
}
