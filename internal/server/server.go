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

	// URL адрес сервера
	URL string `name:"address" short:"a" default:"localhost:8080" env:"ADDRESS" usage:"Server endpoint with port"`

	// Интервал сохранения данных в хранилище
	StoreInterval int `name:"interval" short:"i" default:"300" env:"STORE_INTERVAL" usage:"Interval to backup metrics"`

	// Путь к файлу бекапа данных
	//nolint:lll // tags too long. idk how to fix that
	FileStoragePath string `name:"file" short:"f" default:"/tmp/metrics-db.json" env:"FILE_STORAGE_PATH" usage:"File path to backup"`

	// Флаг восстановления данных. Если true, то данные будут восстановлены из бекапа
	Restore bool `name:"restore" short:"r" default:"false" env:"RESTORE" usage:"Restore metrics from backup"`

	// Строка подключения к базе данных
	PGConnString string `name:"dbconn" short:"d" default:"" env:"DATABASE_DSN" usage:"Postgres connection string"`

	// Ключ хеширования данных
	Key string `name:"key" short:"k" default:"" env:"KEY" usage:"Hash key"`

	// Ключ шифрования
	CryptoKey string `name:"crypto-key" short:"c" default:"" env:"CRYPTO_KEY" usage:"Private RSA key file"`
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
