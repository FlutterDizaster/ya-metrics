package server

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/server/repository/memory"
	"github.com/FlutterDizaster/ya-metrics/pkg/logger"
	"github.com/FlutterDizaster/ya-metrics/pkg/utils"
)

const (
	gracefullPeriodSec = 30
)

type Service interface {
	Start(ctx context.Context)
	// Shutdown(ctx context.Context) error
}

type Settings struct {
	URL                string
	StoreInterval      int
	FileStoragePath    string
	Restore            bool
	PGConnectionString string
}

type Server struct {
	services []Service
}

func New(settings Settings) *Server {
	// initialize logger
	logger.Init()

	// validate url
	if err := utils.ValidateURL(settings.URL); err != nil {
		log.Fatalf("url error: %s", err)
	}

	//TODO: Реализовать создание pg сторейджа, если pgConnectionString != ""

	// creating new storage settings
	storageSettings := memory.Settings{
		StoreInterval:   settings.StoreInterval,
		FileStoragePath: settings.FileStoragePath,
		Restore:         settings.Restore,
	}

	// create new metric storage
	storage := memory.NewMetricStorage(&storageSettings)

	// configure router settings
	// routerSettings := &api.Settings{
	// 	Storage: storage,
	// 	Middlewares: []func(http.Handler) http.Handler{
	// 		middleware.Logger,
	// 		middleware.GzipCompressor,
	// 		middleware.GzipUncompressor,
	// 	},
	// }

	// Создание api сервера
	// apiServer := api.New(routerSettings)

	server := &Server{}

	server.services = append(server.services, storage)
	// server.services = append(server.services, apiServer)

	return server
}

func (s *Server) Start(ctx context.Context) error {
	// Если сервисов нет, то и запускать нечего
	if s.services == nil {
		return errors.New("no registered services")
	}

	wg := sync.WaitGroup{}

	// Слайс функция закрытия контекстов
	stops := make([]func(), len(s.services))

	// Спавним сервисы
	for i := range s.services {
		// Создание контекста для остановки сервиса
		shutdownCtx, shutdownStopCtx := context.WithCancel(context.Background())
		stops[i] = shutdownStopCtx

		// Запуск сервиса
		wg.Add(1)
		go func(index int) {
			s.services[index].Start(shutdownCtx)
			wg.Done()
		}(i)
	}

	// Ждем завершения контекста
	<-ctx.Done()

	// Запускаем gracefull keeper
	forceCtx, forceStopCtx := context.WithTimeout(
		context.Background(),
		gracefullPeriodSec*time.Second,
	)
	defer forceStopCtx()
	go func() {
		<-forceCtx.Done()
		if forceCtx.Err() == context.DeadlineExceeded {
			slog.Error("shutdown timed out... forcing exit.")
			os.Exit(1)
		}
	}()

	// Закрытие контекстов сервисов в порядке создания
	for i := range stops {
		stops[i]()
	}

	// Ожидание остановки сервисов
	wg.Wait()
	return nil
}
