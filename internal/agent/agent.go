package agent

import (
	"context"
	"crypto/rsa"
	"errors"
	"log/slog"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/agent/buffer"
	"github.com/FlutterDizaster/ya-metrics/internal/agent/sender"
	grpcsender "github.com/FlutterDizaster/ya-metrics/internal/agent/sender/grpc-sender"
	httpsender "github.com/FlutterDizaster/ya-metrics/internal/agent/sender/http-sender"
	"github.com/FlutterDizaster/ya-metrics/internal/agent/telemetry"
	"github.com/FlutterDizaster/ya-metrics/internal/application"
	pemreader "github.com/FlutterDizaster/ya-metrics/pkg/pem-reader"
)

// Интерфейс IService описывает объекты, которые могут быть запущены как отдельные потоки приложения.
type IService interface {
	Start(ctx context.Context) error
}

// Settings - настройки агента.
type Settings struct {
	// Адрес сервера агрегатора метрик
	ServerAddr string `name:"address" short:"a" default:"localhost:8080" usage:"server addres" env:"ADDRESS"`

	// Использование gRPC сервера вместо HTTP
	UseGRPC bool `name:"grpc" short:"g" default:"false" usage:"use grpc" env:"USE_GRPC"`

	// Ключ для вычисления Hash суммы
	HashKey string `name:"key" short:"k" default:"" usage:"hash key" env:"KEY"`

	// Количество повторных попыток запроса к серверу
	RetryCount int `default:"3"`

	// Интервал между повторными попытками
	RetryInterval int `default:"1"`

	// Максимальное время ожидания между повторными попытками
	RetryMaxWaitTime int `default:"9"`

	// Интервал между отправками метрик
	ReportInterval int `name:"report" short:"r" default:"10" usage:"report interval" env:"REPORT_INTERVAL"`

	// Интервал между получением метрик
	PollInterval int `name:"poll" short:"p" default:"2" usage:"poll interval" env:"POLL_INTERVAL"`

	// Ограничение на количество запросов в секунду
	RateLimit int `name:"rate-limit" short:"l" default:"1" usage:"rate limit" env:"RATE_LIMIT"`

	// Ключ шифрования
	CryptoKey string `name:"crypto-key" short:"s" default:"" usage:"public RSA key file" env:"CRYPTO_KEY"`
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

	rsaKey, err := pemreader.ReadPublicKey(settings.CryptoKey)
	if err != nil {
		if errors.Is(err, pemreader.ErrReadFile) {
			return nil, err
		}
	}

	// Создание экземпляра Sender
	s := setupSender(settings, buf, rsaKey)

	// Создание агента и регистрация сервисов
	agent := &Agent{}
	err = agent.RegisterService(tlm)
	if err != nil {
		return nil, err
	}
	err = agent.RegisterService(s)
	if err != nil {
		return nil, err
	}

	slog.Debug("Agent instance created")
	return agent, nil
}

func setupSender(
	settings Settings,
	buf sender.Buffer,
	rsaKey *rsa.PublicKey,
) sender.ISender {
	var s sender.ISender

	if settings.UseGRPC {
		senderSettings := grpcsender.Settings{
			Addr:           settings.ServerAddr,
			ReportInterval: time.Duration(settings.ReportInterval) * time.Second,
			Buf:            buf,
			RateLimit:      settings.RateLimit,
		}
		s = grpcsender.New(senderSettings)
	} else {
		senderSettings := httpsender.Settings{
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
		s = httpsender.New(senderSettings)
	}

	return s
}
