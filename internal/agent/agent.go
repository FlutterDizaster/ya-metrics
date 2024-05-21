package agent

import (
	"context"
	"log/slog"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/agent/buffer"
	"github.com/FlutterDizaster/ya-metrics/internal/agent/sender"
	"github.com/FlutterDizaster/ya-metrics/internal/agent/telemetry"
	"github.com/FlutterDizaster/ya-metrics/internal/application"
)

type Service interface {
	Start(ctx context.Context) error
}

type Settings struct {
	ServerAddr       string
	HashKey          string
	RetryCount       int
	RetryInterval    int
	RetryMaxWaitTime int
	ReportInterval   int
	PollInterval     int
	// GracefullPeriod  time.Duration
}

type Agent struct {
	application.Application
}

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

	// Создание экземпляра Sender
	senderSettings := sender.Settings{
		Addr:             settings.ServerAddr,
		RetryCount:       settings.RetryCount,
		RetryInterval:    time.Duration(settings.RetryInterval) * time.Second,
		RetryMaxWaitTime: time.Duration(settings.RetryMaxWaitTime) * time.Second,
		ReportInterval:   time.Duration(settings.ReportInterval) * time.Second,
		Key:              settings.HashKey,
		Buf:              buf,
	}
	snd := sender.New(senderSettings)

	// Создание агента и регистрация сервисов
	agent := &Agent{}
	err := agent.RegisterService(tlm)
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
