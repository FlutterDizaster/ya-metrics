package server

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/server/api"
	"github.com/FlutterDizaster/ya-metrics/internal/server/api/middleware"
	"github.com/FlutterDizaster/ya-metrics/internal/server/repository/memory"
	"github.com/FlutterDizaster/ya-metrics/pkg/logger"
	"github.com/FlutterDizaster/ya-metrics/pkg/utils"
)

// const (
// 	gracefullPeriodSec = 30
// )

// type Service interface {
// 	Start(ctx context.Context) error
// 	Shutdown(ctx context.Context) error
// }

// type Settings struct {
// 	URL                string
// 	StoreInterval      int
// 	FileStoragePath    string
// 	Restore            bool
// 	PGConnectionString string
// }

// type Server struct {
// 	services []Service
// }

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
	routerSettings := &api.Settings{
		Storage: storage,
		Middlewares: []func(http.Handler) http.Handler{
			middleware.Logger,
			middleware.GzipCompressor,
			middleware.GzipUncompressor,
		},
	}

	// Создание сервера
	server := http.Server{Addr: settings.URL, Handler: api.NewRouter(routerSettings)}

	// Контекст бекапов
	backupCtx, backupStopCtx := context.WithCancel(context.Background())

	// Контекст работы сервера
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Создание WaitGroup
	wg := CustomWG{}

	// Прослушивание сигналов системы для старта Gracefull Shutdown
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	wg.Add(1, "server graceefull keeper")
	go func() {
		defer wg.Done("server graceefull keeper goroutine")
		// Ожидание сигнала Gracefull Shutdown
		<-ctx.Done()
		slog.Info("Stopping server...")

		// Сигнал завершения работы с таймером
		shutdownCtx, shutdownStopCtx := context.WithTimeout(
			serverCtx,
			gracefullPeriodSec*time.Second,
		)
		defer shutdownStopCtx()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				slog.Error("graceful shutdown timed out.. forcing exit.")
				os.Exit(1)
			}
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

	// TODO: remove later
	// deadlock avoid
	go shutdown(ctx)

	// Запуск создания бекапов
	wg.Add(1, "backups goroutine")
	go func() {
		defer wg.Done("backups goroutine")
		storage.Start(backupCtx)
	}()
	// go storage.StartBackups(context.Background())

	// Запуск сервера
	wg.Add(1, "server listener goroutine")
	go func() {
		defer wg.Done("server listener goroutine")
		slog.Info("Listening...")
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", slog.String("error", err.Error()))
			panic(err)
		}
	}()

	// Ожидание завершения работы сервера
	<-serverCtx.Done()
	// time.Sleep(droptime * time.Second)

	// Завершение работы бекапов
	backupStopCtx()
	// os.Exit(dropcode)

	wg.Wait()
}

func shutdown(ctx context.Context) {
	<-ctx.Done()
	forceCTX, forceStopCtx := context.WithTimeout(
		context.Background(),
		gracefullPeriodSec*time.Second,
	)

	<-forceCTX.Done()
	if forceCTX.Err() == context.DeadlineExceeded {
		slog.Error("graceful shutdown timed out.. forcing exit.")
		forceStopCtx()
		os.Exit(1)
	}
	forceStopCtx()
}
