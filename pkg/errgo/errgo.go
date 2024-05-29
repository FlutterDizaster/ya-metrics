package errgo

import (
	"sync"
)

// TODO: Добавить механизм закрытия запущенных горутин при выбросе ошибки одной из них.

// ErrGo альтернатива errorsgroup, но распространяет ошибку не дожидаясь завершения работы всех горутин.
// При этом, если одна из горутин вернет ошибку, ожидание завершения остальных не происходит.
type ErrGo struct {
	wg    sync.WaitGroup
	errCh chan error
	once  sync.Once
}

// Wait() ожидает закрытия горутин запущенных функцие Go() или возврата первой ошибки.
// Возвращает канал с ошибкой.
func (eg *ErrGo) Wait() <-chan error {
	wgWaiter := make(chan any)
	resultCh := make(chan error, 1)

	defer close(eg.errCh)

	go func() {
		defer close(wgWaiter)
		eg.wg.Wait()
		wgWaiter <- struct{}{}
	}()

	select {
	case <-wgWaiter:
		resultCh <- nil
	case err := <-eg.errCh:
		resultCh <- err
	}

	return resultCh
}

// Go() запускает горутину.
func (eg *ErrGo) Go(fn func() error) {
	eg.once.Do(func() {
		eg.errCh = make(chan error)
	})

	eg.wg.Add(1)
	go func() {
		err := fn()
		if err != nil {
			eg.errCh <- err
		}
		eg.wg.Done()
	}()
}
