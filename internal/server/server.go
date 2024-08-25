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

// Интерфейс IService описывает объекты, которые могут быть запущены как отдельные потоки приложения.
type IService interface {
	Start(ctx context.Context) error
}

// Интерфейс IStorageService объединяет интерфейсы IService и api.MetricsStorage.
type IStorageService interface {
	IService
	api.MetricsStorage
}

// Settings хранит параметры необходимые для создания экземпляра Server.
type Settings struct {
	URL             string // URL адрес сервера
	StoreInterval   int    // Интервал сохранения данных в хранилище
	FileStoragePath string // Путь к файлу бекапа данных
	Restore         bool   // Флаг восстановления данных. Если true, то данные будут восстановлены из бекапа
	PGConnString    string // Строка подключения к базе данных
	Key             string // Ключ хеширования данных
	CryptoKey       string // Ключ шифрования
}

// Server - структура, которая представляет собоей сервер метрик.
// Занимается логикой работы сервера метрик.
// Объединяет API сервер и хранилище метрик.
// Должен создаваться методом New().
// Для запуска используется метод Start().
type Server struct {
	application.Application // Application управляет жизненным циклом приложения
}

// New - создание нового экземпляра Server.
// В качестве параметров принимает настройки сервера.
// Возвращает экземпляр Server и ошибку инициализации.
func New(settings Settings) (*Server, error) {
	slog.Debug("Creating server instance")

	// validate url
	if err := utils.ValidateURL(settings.URL); err != nil {
		slog.Error("url error", slog.String("error", err.Error()))
	}
	// Создание экземпляра StorageService
	var storage IStorageService
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
	}

	// Получение RSA ключа и добавление в список Middlewares декодера
	if settings.CryptoKey != "" {
		key, crErr := utils.ReadPrivateKey(settings.CryptoKey)
		if crErr != nil {
			return nil, crErr
		}
		middlewares = append(middlewares, &middleware.RSADecoder{Key: key})
	}

	// Добавление в список Middlewares валидатора
	if settings.Key != "" {
		middlewares = append(middlewares, &middleware.Validator{
			Key: []byte(settings.Key),
		})
	}

	// Добавление в список Middlewares прочих Middleware
	middlewares = append(middlewares,
		&middleware.Decompressor{},
		&middleware.Compressor{
			MinDataLength: 1,
		})

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
