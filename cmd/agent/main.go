package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	flag "github.com/spf13/pflag"

	"github.com/FlutterDizaster/ya-metrics/internal/agent"
	"github.com/FlutterDizaster/ya-metrics/pkg/logger"
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	logger.New()

	// Создание сруктуры с настройкаами агента
	settings := parseConfig()

	// Создание агента
	agt, err := agent.New(settings)
	if err != nil {
		slog.Error("Creating agent error", slog.String("error", err.Error()))
		return 1
	}

	// Создание контекста отмены
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	// Запуск агента
	if err = agt.Start(ctx); err != nil {
		slog.Error("Agent startup error", slog.String("error", err.Error()))
		return 1
	}

	return 0
}

func parseConfig() agent.Settings {
	const (
		defaultServerAddr     = "localhost:8080"
		defaultHashKey        = ""
		defaultRetryCount     = 3
		defaultRetryInterval  = 1
		defaultRetryMaxWait   = 9
		defaultReportInterval = 10
		defaultPollInterval   = 2
	)
	var settings agent.Settings

	flag.StringVarP(
		&settings.ServerAddr,
		"address",
		"a",
		defaultServerAddr,
		"HTTP-server addres. Default \"localhost:8080\"",
	)
	flag.StringVarP(
		&settings.HashKey,
		"key",
		"k",
		defaultHashKey,
		"Hash key",
	)
	flag.IntVarP(
		&settings.ReportInterval,
		"report",
		"r",
		defaultReportInterval,
		"Report interval in seconds. Default 10 sec",
	)
	flag.IntVarP(
		&settings.PollInterval,
		"poll",
		"p",
		defaultPollInterval,
		"Metrics poll interval. Default 2 sec",
	)

	flag.Parse()

	settings.RetryCount = defaultRetryCount
	settings.ReportInterval = defaultRetryInterval
	settings.RetryMaxWaitTime = defaultRetryMaxWait

	return lookupEnvs(settings)
}

func lookupEnvs(settings agent.Settings) agent.Settings {
	envServerAddr, ok := os.LookupEnv("ADDRESS")
	if ok {
		settings.ServerAddr = envServerAddr
	}
	envHashKey, ok := os.LookupEnv("KEY")
	if ok {
		settings.HashKey = envHashKey
	}
	envReportInterval, ok := lookupIntEnv("REPORT_INTERVAL")
	if ok {
		settings.ReportInterval = envReportInterval
	}
	envPollInterval, ok := lookupIntEnv("POLL_INTERVAL")
	if ok {
		settings.PollInterval = envPollInterval
	}

	return settings
}

func lookupIntEnv(name string) (int, bool) {
	env, ok := os.LookupEnv(name)
	if !ok {
		return 0, false
	}
	val, err := strconv.Atoi(env)
	if err != nil {
		slog.Error(
			"wrong env type",
			slog.String("variable", name),
			slog.String("expected type", "integer"),
		)
		return 0, false
	}
	return val, true
}
