package server

import (
	"context"
	"log/slog"

	"github.com/FlutterDizaster/ya-metrics/internal/application"
	"github.com/FlutterDizaster/ya-metrics/internal/server/api"
	"github.com/FlutterDizaster/ya-metrics/internal/server/api/middleware"
	"github.com/FlutterDizaster/ya-metrics/internal/server/repository/memory"
	"github.com/FlutterDizaster/ya-metrics/internal/server/repository/postgres"
	"github.com/FlutterDizaster/ya-metrics/pkg/utils"
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
	application.Application
}

func New(settings Settings) (*Server, error) {
	slog.Debug("Creating server instance")

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

	err = server.RegisterService(storage)
	if err != nil {
		return nil, err
	}
	err = server.RegisterService(apiServer)
	if err != nil {
		return nil, err
	}

	slog.Debug("Application instance created", slog.String("storage mode", storageMode))
	return server, nil
}
