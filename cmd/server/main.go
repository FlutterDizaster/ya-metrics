package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	flag "github.com/spf13/pflag"

	"github.com/FlutterDizaster/ya-metrics/internal/server"
	"github.com/FlutterDizaster/ya-metrics/pkg/logger"
)

const (
	defaultEndpoint        = "localhost:8080"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "/tmp/metrics-db.json"
	defaultRestore         = true
	defaultPGConnString    = ""
)

func main() {
	// initialize logger
	logger.Init()

	// Создание структуры с настройками сервера
	settings := parseConfig()
	// Создание сервера
	srv, err := server.New(settings)
	if err != nil {
		panic(err)
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
		panic(err)
	}
}

func parseConfig() server.Settings {
	var settings server.Settings
	flag.StringVar(
		&settings.URL,
		"a",
		defaultEndpoint,
		"Server endpoint addres. Default localhost:8080",
	)
	flag.StringVar(
		&settings.FileStoragePath,
		"f",
		defaultFileStoragePath,
		"Backup file path. Default /tmp/metrics-db.json",
	)
	flag.StringVar(
		&settings.PGConnString,
		"d",
		defaultPGConnString,
		"Postgres connection string",
	)
	flag.BoolVar(
		&settings.Restore,
		"r",
		defaultRestore,
		"the flag indicates whether a backup should be loaded from a file",
	)
	flag.IntVar(
		&settings.StoreInterval,
		"i",
		defaultStoreInterval,
		"Time between backups in seconds. Default 300",
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
