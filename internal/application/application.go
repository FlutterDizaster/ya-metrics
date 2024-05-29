package application

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

// Service определяет интерфейс для сущностей сервисов, работающих конкурентно во время
// во время исполнения программы.
type Service interface {
	Start(context.Context) error
}

// Application используется для управления дизненным циклом сервисов.
type Application struct {
	services []Service
	sk       *ServiceKeeper
}

// RegisterService(service Service) добавляет сервисы в список к запуску.
func (a *Application) RegisterService(service Service) {
	a.services = append(a.services, service)
}

// Start(ctx context.Context) Запускает все зарегистрированные сервисы и завершает их при закрытии контекста.
func (a *Application) Start(ctx context.Context) error {
	slog.Debug("Starting services...")
	if len(a.services) == 0 {
		return errors.New("no services registered")
	}

	// Создание ServiceKeeper
	a.sk = NewServiceKeeper()

	// Создание fanIn контекста
	fanCtx, fanClancleCtx := context.WithCancel(context.Background())
	// Запуск сервисов
	errCh := fanIn(fanCtx, a.startServices()...)

	// Ожидание закрытия контекста или ошибки из сервисов
	var err error

	select {
	case <-ctx.Done():
		err = a.stopServices() // Сохраняем ошибку для возврата

	case err = <-errCh: // Сохраняем ошибку для возврата
		slog.Error("service error found. exiting...")
		_ = a.stopServices() // Пропускаем сохранение ошибки
	}

	fanClancleCtx()

	return err
}

// Хелпер функция для запуска сервисов.
func (a *Application) startServices() []<-chan error {
	// Слайс каналов с ошибками
	errChans := make([]<-chan error, 0, len(a.services))

	// Запуск сервисов
	for i := range a.services {
		errChans = append(errChans, a.sk.Start(a.services[i]))
	}

	return errChans
}

// Хелпер функция для остановки сервисов.
func (a *Application) stopServices() error {
	var err error

	for i := range a.services {
		err = a.sk.Stop(a.services[i])
	}

	return err
}

// fanIn объединяет несколько каналов resultChs в один.
func fanIn(ctx context.Context, resultChs ...<-chan error) <-chan error {
	// конечный выходной канал в который отправляем данные из всех каналов из слайса, назовём его результирующим
	finalCh := make(chan error)

	// понадобится для ожидания всех горутин
	wg := sync.WaitGroup{}

	// перебираем все входящие каналы
	for _, ch := range resultChs {
		// в горутину передавать переменную цикла нельзя, поэтому делаем так
		chClosure := ch

		// инкрементируем счётчик горутин, которые нужно подождать
		wg.Add(1)

		go func() {
			// откладываем сообщение о том, что горутина завершилась
			defer wg.Done()

			// получаем данные из канала
			for data := range chClosure {
				select {
				// выходим из горутины, если канал закрылся
				case <-ctx.Done():
					return
				// если не закрылся, отправляем данные в конечный выходной канал
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		// ждём завершения всех горутин
		wg.Wait()
		// когда все горутины завершились, закрываем результирующий канал
		close(finalCh)
	}()

	// возвращаем результирующий канал
	return finalCh
}
