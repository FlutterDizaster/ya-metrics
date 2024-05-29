package application

import "context"

// serviceControls Предоставляет доступ к каналу ошибок и функции закрытия контекста сервиса.
type serviceControls struct {
	// Используется для распространения ошибки и ожидания завершения работы
	errCh chan error

	// Функция закрытия контекста работы сервиса
	cancleFn context.CancelFunc
}

// ServiceKeeper нужен для запуска и остановки сервисов.
// Для создания эеземпляра необходимо использовать функцию NewServiceKeeper().
type ServiceKeeper struct {
	services map[Service]*serviceControls
}

// Возвращает указатель на новый экземпляр ServiceKeeper.
func NewServiceKeeper() *ServiceKeeper {
	return &ServiceKeeper{
		services: make(map[Service]*serviceControls),
	}
}

// Start(service Service) запускает выполнение сервиса в отдельной горутине и возвращает канал с ошибкой,
// возвращаемой сервисом в конце своей работы.
//
// В конце работы необходимо вызвать функцию Stop(service Service) для освобождения ресурсов.
func (sk *ServiceKeeper) Start(service Service) <-chan error {
	// Создание контекста завершения сервиса
	srvCtx, srvCancleCtx := context.WithCancel(context.Background())

	// Создание канала ошибки
	errCh := make(chan error)

	// Создание структуры управления сервисом
	sk.services[service] = &serviceControls{
		errCh:    errCh,
		cancleFn: srvCancleCtx,
	}

	// Запуск сервиса
	go func() {
		if err := service.Start(srvCtx); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	return errCh
}

// Stop(service Service) завершает работу сервиса и возвращает ошибку сервиса.
// Если сервис уже был завершен.
func (sk *ServiceKeeper) Stop(service Service) error {
	// Закрытие контекста сервиса
	sk.services[service].cancleFn()

	// Ожидание завершения работы сервиса
	return <-sk.services[service].errCh
}
