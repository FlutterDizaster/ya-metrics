package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

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
	// Проверяем есть ли таблица в бд и при необходимости создаем её
	err := ms.checkAndCreateTable()
	if err != nil {
		return err
	}

	// TODO: Компиляция запросов

	// Ожидание завершения контекста
	<-ctx.Done()
	ms.db.Close()
	return nil
}

func (ms *MetricStorage) AddBatchMetrics(metrics []view.Metric) error {
	ctx, cancle := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancle()
	// Начало транзакции
	tx, err := ms.db.Begin(ctx)
	if err != nil {
		return err
	}

	// записываем каждую метрику
	for _, metric := range metrics {
		_, err = tx.Exec(ctx, queryAdd, metric.ID, metric.MType, metric.Value, metric.Delta)
		if err != nil {
			return errors.Join(err, tx.Rollback(ctx))
		}
	}

	// Коммитим транзакцию
	return tx.Commit(ctx)
}

func (ms *MetricStorage) AddMetric(metric view.Metric) (view.Metric, error) {
	// Подготовка переменных
	var value sql.NullFloat64
	var delta sql.NullInt64
	// Выполнение запроса
	ctx, cancle := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancle()
	err := ms.db.QueryRow(ctx, queryAdd, metric.ID, metric.MType, metric.Value, metric.Delta).
		Scan(&value, &delta)
	if err != nil {
		return view.Metric{}, err
	}
	// Сохранение ответа
	if value.Valid {
		metric.Value = &value.Float64
	} else if delta.Valid {
		metric.Delta = &delta.Int64
	}

	return metric, nil
}

func (ms *MetricStorage) GetMetric(kind string, name string) (view.Metric, error) {
	var metric view.Metric
	metric.ID = name
	metric.MType = kind
	// Подготовка переменных
	var value sql.NullFloat64
	var delta sql.NullInt64
	// Выполнение запроса
	ctx, cancle := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancle()
	err := ms.db.QueryRow(ctx, queryGetOne, name, kind).Scan(&value, &delta)
	// Проверка переменных на валидность
	if value.Valid {
		metric.Value = &value.Float64
	} else if delta.Valid {
		metric.Delta = &delta.Int64
	}
	return metric, err
}

func (ms *MetricStorage) ReadAllMetrics() ([]view.Metric, error) {
	metrics := make([]view.Metric, 0)
	// Выполнение запроса
	ctx, cancle := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancle()
	rows, err := ms.db.Query(ctx, gueryGetAll)
	if err != nil {
		return nil, err
	}
	// Проход по всем полученным строкам
	for rows.Next() {
		// Подготовка переменных
		var (
			metric view.Metric
			id     string
			mtype  string
			value  sql.NullFloat64
			delta  sql.NullInt64
		)
		// Сканирование строки
		err = rows.Scan(&id, &mtype, &value, &delta)
		if err != nil {
			return nil, err
		}
		// Заполнение полей метрики
		metric.ID = id
		metric.MType = mtype
		if value.Valid {
			metric.Value = &value.Float64
		} else if delta.Valid {
			metric.Delta = &delta.Int64
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (ms *MetricStorage) checkAndCreateTable() error {
	ctx, cancle := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancle()
	_, err := ms.db.Exec(ctx, queryCheckAndCreateDB)
	return err
}
