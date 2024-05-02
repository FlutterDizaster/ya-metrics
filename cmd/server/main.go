package main

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/FlutterDizaster/ya-metrics/internal/server"
	flag "github.com/spf13/pflag"
)

func main() {
	// Парсинг флагов
	endpoint := flag.String("a", "localhost:8080", "Server endpoint addres. Default localhost:8080")
	storeInterval := flag.Int("i", 300, "Time between backups in seconds. Default 300")
	fileStoragePath := flag.String(
		"f",
		"/tmp/metrics-db.json",
		"Backup file path. Default /tmp/metrics-db.json",
	)
	restoref := flag.String(
		"r",
		"true",
		"the flag indicates whether a backup should be loaded from a file",
	)
	flag.Parse()

	restore, err := strconv.ParseBool(*restoref)
	if err != nil {
		slog.Error("r should be integer", "error", err)
		os.Exit(1)
	}

	// Парсинг переменных окружения
	envEndpoint, ok := os.LookupEnv("ADDRESS")
	if ok {
		endpoint = &envEndpoint
	}
	envStoreInterval, ok := os.LookupEnv("STORE_INTERVAL")
	if ok {
		var pStoreInterval int
		pStoreInterval, err = strconv.Atoi(envStoreInterval)
		if err != nil {
			slog.Error("STORE_INTERVAL should be integer", "error", err)
			os.Exit(1)
		}
		storeInterval = &pStoreInterval
	}
	envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		fileStoragePath = &envFileStoragePath
	}
	envRestore, ok := os.LookupEnv("RESTORE")
	if ok {
		var pRestore bool
		pRestore, err = strconv.ParseBool(envRestore)
		if err != nil {
			slog.Error(
				"RESTORE should be 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.",
				"error",
				err,
			)
			os.Exit(1)
		}
		restore = pRestore
	}

	// Создание структуры с настройками сервера
	settings := &server.Settings{
		URL:             *endpoint,
		StoreInterval:   *storeInterval,
		FileStoragePath: *fileStoragePath,
		Restore:         restore,
	}

	server.Setup(settings)
}
