package retrier

import (
	"context"
	"errors"
	"time"
)

var ErrMaxRetriesExceeded = errors.New("max retryes exceeded")

// Структура предоставляюзая функционал для повторения вызова функции с указанными параметрами.
// INDEV.
type Retrier struct {
	// Кол-во попыток вызова функции
	// По умолчанию 3
	MaxRetryes int

	// Интерфал между попытками
	// По умолчанию 1000 миллисекунд
	RetryIntervalMS int64

	// Время на которое увеличивается RetryIntervalMS после каждой попытки
	// По умолчанию 2000 миллисекунд
	IncreaseIntervalMS int64
}

// Принимает функцию, которую нужно повторить при возврате true и nil,
// если возвращается ошибка, то других попыток не будет и Try вернет эту ошибку.
// Try обернет ошибку в ErrMaxRetriesExceeded, если достигнуто максимальное кол-во попыток.
func (r *Retrier) Try(ctx context.Context, retryfunc func(context.Context) (bool, error)) error {
	r.setDefaults()

	doNext, err := retryfunc(ctx)
	// if err != nil {
	// 	return err
	// }

	backoffTime := time.Duration(r.RetryIntervalMS) * time.Millisecond
	attempt := 0

	if r.MaxRetryes > 1 {
		for doNext {
			if attempt >= r.MaxRetryes {
				return errors.Join(ErrMaxRetriesExceeded, err)
			}
			timer := time.NewTimer(backoffTime)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				attempt++
				doNext, err = retryfunc(ctx)
			}
			backoffTime = time.Duration(
				r.IncreaseIntervalMS*int64(attempt)+r.RetryIntervalMS,
			) * time.Millisecond
		}
	}

	return err
}

func (r *Retrier) setDefaults() {
	// TODO: Сделать потокобезопасной для использования одного жкземпляра из нескольких потоков
	if r.MaxRetryes == 0 {
		r.MaxRetryes = 3
	}
	if r.RetryIntervalMS == 0 {
		r.RetryIntervalMS = 1000
	}
	if r.IncreaseIntervalMS == 0 {
		r.IncreaseIntervalMS = 2000
	}
}
