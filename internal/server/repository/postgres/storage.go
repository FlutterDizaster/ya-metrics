package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	// PostgreSQL driver.
	_ "github.com/jackc/pgx/v5"
)

// type DataProvider interface {
// 	AddGauge(view.Metric) (view.Metric, error)
// 	AddCounter(view.Metric) (view.Metric, error)
// 	GetMetric(kind string, name string) (view.Metric, error)
// 	ReadAllMetrics() []view.Metric
// }

// var _ DataProvider = &MetricStorage{}

type MetricStorage struct {
	pgConnString string
	db           *sql.DB
}

func New(conn string) (*MetricStorage, error) {
	// Создание экземпляра DB
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return nil, err
	}

	// Проверка подключения
	ctx, cancle := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancle()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &MetricStorage{
		pgConnString: conn,
		db:           db,
	}, nil
}

func (ms *MetricStorage) Start(ctx context.Context) error {
	// Ожидание завершения контекста
	<-ctx.Done()
	return ms.db.Close()
}

func (ms *MetricStorage) AddMetric(_ view.Metric) (view.Metric, error) {
	return view.Metric{}, nil
}

func (ms *MetricStorage) GetMetric(_ string, _ string) (view.Metric, error) {
	return view.Metric{}, nil
}

func (ms *MetricStorage) ReadAllMetrics() []view.Metric {
	return make([]view.Metric, 0)
}
