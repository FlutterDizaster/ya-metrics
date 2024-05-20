package server

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/server/api"
	"github.com/FlutterDizaster/ya-metrics/internal/server/api/middleware"
	"github.com/FlutterDizaster/ya-metrics/internal/server/repository/memory"
	"github.com/FlutterDizaster/ya-metrics/internal/server/repository/postgres"
	"github.com/FlutterDizaster/ya-metrics/pkg/utils"
	"golang.org/x/sync/errgroup"
)

const (
	gracefullPeriodSec = 30
)

type Service interface {
	Start(ctx context.Context) error
}

type StorageService interface {
	Service
	api.MetricsStorage
}

type Settings struct {
	URL             string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	PGConnString    string
	Key             string
}

type Server struct {
	services []Service
}

func New(settings Settings) (*Server, error) {
	slog.Debug("Creating application instance")

	// validate url
	if err := utils.ValidateURL(settings.URL); err != nil {
		slog.Error("url error", slog.String("error", err.Error()))
	}
	// Создание экземпляра StorageService
	var storage StorageService
	var storageMode string
	var err error
	// Если строка для поключения к бд не указана
	if settings.PGConnString == "" {
		// Создание локального хранилища метрик
		storageSettings := memory.Settings{
			StoreInterval:   settings.StoreInterval,
			FileStoragePath: settings.FileStoragePath,
			Restore:         settings.Restore,
		}
		storage, err = memory.New(&storageSettings)
		storageMode = "In Memory"
	} else {
		// Создание хранилища с подключением к базе
		storage, err = postgres.New(settings.PGConnString)
		storageMode = "DB"
	}
	if err != nil {
		slog.Error("error creating storage. forcing exit.", slog.String("error", err.Error()))
		return nil, err
	}
	// Создание списка Middlewares
	middlewares := []middleware.Middleware{
		&middleware.Logger{},
		&middleware.Decompressor{},
		&middleware.Compressor{
			MinDataLength: 1,
		},
	}
	if settings.Key != "" {
		middlewares = append(middlewares, &middleware.Validator{
			Key: []byte(settings.Key),
		})
	}
	// configure router settings
	routerSettings := &api.Settings{
		Addr:        settings.URL,
		Storage:     storage,
		Middlewares: middlewares,
	}
	// Создание api сервера
	apiServer := api.New(routerSettings)

	server := &Server{}

	server.services = append(server.services, storage)
	server.services = append(server.services, apiServer)

	slog.Debug("Application instance created", slog.String("storage mode", storageMode))
	return server, nil
}

func (s *Server) Start(ctx context.Context) error {
	slog.Debug("Starting services")
	// Если сервисов нет, то и запускать нечего
	if s.services == nil {
		return errors.New("no registered services")
	}

	eg := errgroup.Group{}
	// Слайс функция закрытия контекстов
	stops := make([]func(), len(s.services))
	// Спавним сервисы
	for i := range s.services {
		// Создание контекста для остановки сервиса
		shutdownCtx, shutdownStopCtx := context.WithCancel(context.Background())
		stops[i] = shutdownStopCtx

		// Запуск сервиса
		func(index int) {
			eg.Go(func() error {
				return s.services[index].Start(shutdownCtx)
			})
		}(i)
	}
	// Ждем завершения контекста
	// TODO: Запустить в отдельной горутине. Мешает распространению ошибки во время запуска
	<-ctx.Done()
	slog.Info("Shutdown...")
	defer slog.Info("All services stopped")
	// Запускаем gracefull keeper
	// Завершает выполнение программы через gracefullPeriodSec секунд, если программа не завершится сама
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
	return eg.Wait()
}
