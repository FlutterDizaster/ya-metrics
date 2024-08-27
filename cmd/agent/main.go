package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/FlutterDizaster/ya-metrics/internal/agent"
	"github.com/FlutterDizaster/ya-metrics/pkg/appinfoprinter"
	configloader "github.com/FlutterDizaster/ya-metrics/pkg/config-loader"
	"github.com/FlutterDizaster/ya-metrics/pkg/logger"
)

//nolint:gochecknoglobals // build info
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	// initialize logger
	logger.New(slog.LevelDebug)

	// Print AppInfo
	appInfo := appinfoprinter.AppInfo{
		Version: buildVersion,
		Date:    buildDate,
		Commit:  buildCommit,
	}

	err := appinfoprinter.PrintAppInfo(appInfo)
	if err != nil {
		slog.Error("PrintAppInfo error", slog.String("error", err.Error()))
	}

	// Создание сруктуры с настройкаами агента
	settings := agent.Settings{}
	err = configloader.LoadConfig(&settings)
	if err != nil {
		slog.Error("Loading config error", slog.String("error", err.Error()))
		return
	}

	// Создание агента
	agt, err := agent.New(settings)
	if err != nil {
		slog.Error("Creating agent error", slog.String("error", err.Error()))
		return
	}

	// Создание контекста отмены
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	// Запуск агента
	if err = agt.Start(ctx); err != nil {
		slog.Error("Agent startup error", slog.String("error", err.Error()))
		return
	}
}
