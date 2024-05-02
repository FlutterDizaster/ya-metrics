package server

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/repository/memory"
	"github.com/FlutterDizaster/ya-metrics/internal/router"
	"github.com/FlutterDizaster/ya-metrics/internal/router/middleware"
	"github.com/FlutterDizaster/ya-metrics/pkg/logger"
	"github.com/FlutterDizaster/ya-metrics/pkg/utils"
)

const (
	gracefullPeriodSec = 30
	droptime           = 10
	dropcode           = 2
)

type Settings struct {
	URL             string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

func Setup(settings *Settings) {
	// initialize logger
	logger.Init()

	// validate url
	if err := utils.ValidateURL(settings.URL); err != nil {
		log.Fatalf("url error: %s", err)
	}

	// creating new storage settings
	storageSettings := memory.Settings{
		StoreInterval:   settings.StoreInterval,
		FileStoragePath: settings.FileStoragePath,
		Restore:         settings.Restore,
	}

	// create new metric storage
	storage := memory.NewMetricStorage(&storageSettings)

	// configure router settings
	routerSettings := &router.Settings{
		Storage: storage,
		Middlewares: []func(http.Handler) http.Handler{
			middleware.Logger,
			middleware.GzipCompressor,
			middleware.GzipUncompressor,
		},
	}

	// Создание сервера
	server := http.Server{Addr: settings.URL, Handler: router.NewRouter(routerSettings)}

	// Контекст бекапов
	backupCtx, backupStopCtx := context.WithCancel(context.Background())

	// Контекст работы сервера
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Создание WaitGroup
	var wg sync.WaitGroup

	// Прослушивание сигналов системы для старта Gracefull Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(
		c,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Ожидание сигнала Gracefull Shutdown
		<-c
		slog.Info("Stopping server...")

		// Сигнал завершения работы с таймером
		shutdownCtx, shutdownStopCtx := context.WithTimeout(
			serverCtx,
			gracefullPeriodSec*time.Second,
		)
		defer shutdownStopCtx()

		wg.Add(1)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				slog.Error("graceful shutdown timed out.. forcing exit.")
				os.Exit(1)
			}
			wg.Done()
		}()

		// Запуск Gracefull Shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("server shutdown error", "error", err)
			os.Exit(1)
		}
		serverStopCtx()
		slog.Info("Server successfully stopped")
	}()

	// Запуск создания бекапов
	wg.Add(1)
	go func() {
		storage.StartBackups(backupCtx)
		wg.Done()
	}()

	// Запуск сервера
	wg.Add(1)
	go func() {
		slog.Info("Listening...")
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error: %s", err)
		}
		wg.Done()
	}()

	// Ожидание завершения работы сервера
	// <-serverCtx.Done()
	time.Sleep(droptime * time.Second)

	// Завершение работы бекапов
	backupStopCtx()
	// os.Exit(dropcode)

	wg.Wait()
}
