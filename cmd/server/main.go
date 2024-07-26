package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "net/http/pprof"

	_ "github.com/FlutterDizaster/ya-metrics/swagger"

	flag "github.com/spf13/pflag"

	"github.com/FlutterDizaster/ya-metrics/internal/server"
	"github.com/FlutterDizaster/ya-metrics/pkg/logger"
)

// @title Ya-Metrics API
// @version 0.3
// @description API for getting and setting metrics
// @BasePath /
// @contact.name Dmitriy Loginov
// @contact.email dmitriy@loginoff.space

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	// initialize logger
	logger.New(slog.LevelDebug)

	// Создание структуры с настройками сервера
	settings := parseConfig()
	// Создание сервера
	srv, err := server.New(settings)
	if err != nil {
		slog.Error("Creating server error", slog.String("error", err.Error()))
		return 1
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
		return 1
	}

	return 0
}

func parseConfig() server.Settings {
	const (
		defaultEndpoint        = "localhost:8080"
		defaultStoreInterval   = 300
		defaultFileStoragePath = "/tmp/metrics-db.json"
		defaultRestore         = true
		defaultPGConnString    = ""
	)
	var settings server.Settings
	flag.StringVarP(
		&settings.URL,
		"address",
		"a",
		defaultEndpoint,
		"Server endpoint addres. Default localhost:8080",
	)
	flag.StringVarP(
		&settings.FileStoragePath,
		"file",
		"f",
		defaultFileStoragePath,
		"Backup file path. Default /tmp/metrics-db.json",
	)
	flag.StringVarP(
		&settings.PGConnString,
		"dbconn",
		"d",
		defaultPGConnString,
		"Postgres connection string",
	)
	flag.BoolVarP(
		&settings.Restore,
		"restore",
		"r",
		defaultRestore,
		"the flag indicates whether a backup should be loaded from a file",
	)
	flag.IntVarP(
		&settings.StoreInterval,
		"interval",
		"i",
		defaultStoreInterval,
		"Time between backups in seconds. Default 300",
	)
	flag.StringVarP(
		&settings.Key,
		"key",
		"k",
		"",
		"Hash key",
	)

	flag.Parse()

	return lookupEnvs(settings)
}

func lookupEnvs(settings server.Settings) server.Settings {
	envEndpoint, ok := os.LookupEnv("ADDRESS")
	if ok {
		settings.URL = envEndpoint
	}
	envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		settings.FileStoragePath = envFileStoragePath
	}
	envPGConnString, ok := os.LookupEnv("DATABASE_DSN")
	if ok {
		settings.PGConnString = envPGConnString
	}
	envRestore, ok := lookupBoolEnv("RESTORE")
	if ok {
		settings.Restore = envRestore
	}
	envStoreInterval, ok := lookupIntEnv("STORE_INTERVAL")
	if ok {
		settings.StoreInterval = envStoreInterval
	}
	envHashKey, ok := os.LookupEnv("KEY")
	if ok {
		settings.Key = envHashKey
	}

	return settings
}

func lookupIntEnv(name string) (int, bool) {
	env, ok := os.LookupEnv(name)
	if !ok {
		return 0, false
	}
	val, err := strconv.Atoi(env)
	if err != nil {
		slog.Error(
			"wrong env type",
			slog.String("variable", name),
			slog.String("expected type", "integer"),
		)
		return 0, false
	}
	return val, true
}

func lookupBoolEnv(name string) (bool, bool) {
	env, ok := os.LookupEnv(name)
	if !ok {
		return false, false
	}
	val, err := strconv.ParseBool(env)
	if err != nil {
		slog.Error(
			"wrong env type",
			slog.String("variable", name),
			slog.String("expected type", "boolean"),
		)
		return false, false
	}
	return val, true
}
