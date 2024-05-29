package errgo

import (
	"sync"
)

// TODO: Добавить механизм закрытия запущенных горутин при выбросе ошибки одной из них.
// TODO: Где-то есть дедлок. Нужно фиксить. Хотя, возможно, проблема и не здесь.

// ErrGo альтернатива errorsgroup, но распространяет ошибку не дожидаясь завершения работы всех горутин.
// При этом, если одна из горутин вернет ошибку, ожидание завершения остальных не происходит.
type ErrGo struct {
	wg    sync.WaitGroup
	errCh chan error
	once  sync.Once
}

// Wait() ожидает закрытия горутин запущенных функцие Go() или возврата первой ошибки.
func (eg *ErrGo) Wait() error {
	wgWaiter := make(chan any)

	defer close(wgWaiter)
	defer close(eg.errCh)

	go func() {
		eg.wg.Wait()
		wgWaiter <- struct{}{}
	}()

	select {
	case <-wgWaiter:
		return nil
	case err := <-eg.errCh:
		return err
	}
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
