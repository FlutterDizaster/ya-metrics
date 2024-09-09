package sender

import (
	"context"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

// Интерфейс для буфера метрик.
type Buffer interface {
	// Метод для вытягивания всех метрик из буфера.
	// Подразумевается, что после вызова буфер будет очищен.
	Pull() ([]view.Metric, error)
}

// Sender - сервис отправки метрик.
type ISender interface {
	Start(ctx context.Context) error
}
