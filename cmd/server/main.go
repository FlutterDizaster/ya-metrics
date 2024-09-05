package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	_ "github.com/FlutterDizaster/ya-metrics/swagger"

	"github.com/FlutterDizaster/ya-metrics/internal/server"
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

// @title Ya-Metrics API
// @version 0.3
// @description API for getting and setting metrics
// @host localhost:8080
// @BasePath /
// @contact.name Dmitriy Loginov
// @contact.email dmitriy@loginoff.space

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

	// Создание структуры с настройками сервера
	settings := server.Settings{}
	err = configloader.LoadConfig(&settings)
	if err != nil {
		slog.Error("Loading config error", slog.String("error", err.Error()))
		return
	}

	// Создание сервера
	srv, err := server.New(settings)
	if err != nil {
		slog.Error("Creating server error", slog.String("error", err.Error()))
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

	// Запуск сервера
	if err = srv.Start(ctx); err != nil {
		slog.Error("Server startup error", slog.String("error", err.Error()))
		return
	}
}
