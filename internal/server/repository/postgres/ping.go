package postgres

import (
	"context"
	"time"
)

func (ms *MetricStorage) Ping() error {
	pingCtx, pingCancleCtx := context.WithTimeout(context.Background(), 1*time.Second)
	defer pingCancleCtx()
	return ms.db.Ping(pingCtx)
}
