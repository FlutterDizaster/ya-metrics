package application

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/FlutterDizaster/ya-metrics/pkg/errgo"
)

type Service interface {
	Start(context.Context) error
}

type Application struct {
	services []Service
}

// TODO: Добавыть выброс ошибки, если инициализация уже пройдена
func (a *Application) RegisterService(service Service) error {
	a.services = append(a.services, service)
	return nil
}

func (a *Application) Start(ctx context.Context) error {
	slog.Debug("Starting services")
	// Если сервисов нет, то и запускать нечего
	if a.services == nil {
		return errors.New("no registered services")
	}

	eg := errgo.ErrGo{}
	// Слайс функция закрытия контекстов
	stops := make([]func(), len(a.services))
	// Спавним сервисы
	for i := range a.services {
		// Создание контекста для остановки сервиса
		shutdownCtx, shutdownStopCtx := context.WithCancel(context.Background())
		stops[i] = shutdownStopCtx

		// Запуск сервиса
		func(index int) {
			eg.Go(func() error {
				return a.services[index].Start(shutdownCtx)
			})
		}(i)
	}
	// Ждем завершения контекста
	// TODO: Запустить в отдельной горутине. Мешает распространению ошибки во время запуска
	var err error
loop:
	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutdown...")
			defer slog.Info("All services stopped")
			// Запускаем gracefull keeper
			// Завершает выполнение программы через 30 секунд, если программа не завершится сама
			forceCtx, forceStopCtx := context.WithTimeout(
				context.Background(),
				30*time.Second, // TODO: Вынести в конфиг
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
				// TODO: Ожидать закрытия каждого сервиса
				stops[i]()
			}
		case err = <-eg.Wait():
			break loop
		}
	}
	// Ожидание остановки сервисов
	return err
}
