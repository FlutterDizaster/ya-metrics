package agent

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/agent/buffer"
	"github.com/FlutterDizaster/ya-metrics/internal/agent/sender"
	"github.com/FlutterDizaster/ya-metrics/internal/agent/telemetry"
	"github.com/FlutterDizaster/ya-metrics/internal/application"
	"github.com/FlutterDizaster/ya-metrics/pkg/utils"
)

// Интерфейс IService описывает объекты, которые могут быть запущены как отдельные потоки приложения.
type IService interface {
	Start(ctx context.Context) error
}

// Settings - настройки агента.
type Settings struct {
	ServerAddr       string // Адрес сервера агрегатора метрик
	HashKey          string // Ключ для вычисления Hash суммы
	RetryCount       int    // Количество повторных попыток запроса к серверу
	RetryInterval    int    // Интервал между повторными попытками
	RetryMaxWaitTime int    // Максимальное время ожидания между повторными попытками
	ReportInterval   int    // Интервал между отправками метрик
	PollInterval     int    // Интервал между получением метрик
	RateLimit        int    // Ограничение на количество запросов в секунду
	CryptoKey        string // Ключ шифрования
}

// Agent управляет запуском сервисов по сбору и отправки метрик.
// Должен быть создан методом New.
// Запуск приложения производится методом Start.
type Agent struct {
	application.Application
}

// New создает новый экземпляр агента.
// Принимает настройки агента.
// Возвращает агента и ошибку.
func New(settings Settings) (*Agent, error) {
	slog.Debug("Creating agent instance")
	// Создание экземпляра Buffer
	buf := buffer.New()

	// Создание экземпляра Telemetry
	telemetrySettings := telemetry.Settings{
		PollInterval: time.Duration(settings.PollInterval) * time.Second,
		Buf:          buf,
	}
	tlm := telemetry.New(telemetrySettings)

	rsaKey, err := utils.ReadPublicKey(settings.CryptoKey)
	if err != nil {
		if errors.Is(err, utils.ErrReadFile) {
			return nil, err
		}
	}

	// Создание экземпляра Sender
	senderSettings := sender.Settings{
		Addr:             settings.ServerAddr,
		RetryCount:       settings.RetryCount,
		RetryInterval:    time.Duration(settings.RetryInterval) * time.Second,
		RetryMaxWaitTime: time.Duration(settings.RetryMaxWaitTime) * time.Second,
		ReportInterval:   time.Duration(settings.ReportInterval) * time.Second,
		HashKey:          settings.HashKey,
		Buf:              buf,
		RateLimit:        settings.RateLimit,
		RSAKey:           rsaKey,
	}
	snd := sender.New(senderSettings)

	// Создание агента и регистрация сервисов
	agent := &Agent{}
	err = agent.RegisterService(tlm)
	if err != nil {
		return nil, err
	}
	err = agent.RegisterService(snd)
	if err != nil {
		return nil, err
	}

	slog.Debug("Agent instance created")
	return agent, nil
}
