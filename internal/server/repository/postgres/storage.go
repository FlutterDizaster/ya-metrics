package postgres

import (
	"context"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	// PostgreSQL driver.

	"github.com/jackc/pgx/v5/pgxpool"
)

// type DataProvider interface {
// 	AddGauge(view.Metric) (view.Metric, error)
// 	AddCounter(view.Metric) (view.Metric, error)
// 	GetMetric(kind string, name string) (view.Metric, error)
// 	ReadAllMetrics() []view.Metric
// }

// var _ DataProvider = &MetricStorage{}

type MetricStorage struct {
	db *pgxpool.Pool
}

func New(conn string) (*MetricStorage, error) {
	ms := &MetricStorage{}
	// Создание экземпляра DB
	poolConfig, err := pgxpool.ParseConfig(conn)
	if err != nil {
		return nil, err
	}

	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	ms.db = db

	// Проверка подключения
	err = ms.Ping()
	if err != nil {
		return nil, err
	}

	return ms, nil
}

func (ms *MetricStorage) Start(ctx context.Context) error {
	// Ожидание завершения контекста
	<-ctx.Done()
	ms.db.Close()
	return nil
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
