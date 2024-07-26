package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Реализация хранилища метрик в таблицах PostgreSQL.
// Экземпляр должен создаваться с помощью New.
type MetricStorage struct {
	db *pgxpool.Pool
}

// Функция фабрика для создания нового экземпляра MetricStorage.
// Принимает строку подключения к БД.
// В случае ошибки возвращает nil и ошибку.
// В случае успеха возвращает новый экземпляр MetricStorage и nil.
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

// Метод запускающий сервис БД.
// Блокирует поток исполнения до закрытия контекста ctx.
// В случае невозможности подключения к бд возвращает ошибку.
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

// Метод добавляющий метрики в БД.
// Принимает слайс метрик, которые необходимо добавить в БД.
// Возвращает список обновленных метрик и nil.
// В случае ошибки возвращает nil и ошибку.
func (ms *MetricStorage) AddMetrics(metrics ...view.Metric) ([]view.Metric, error) {
	ctx, cancle := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancle()
	// Создание слайса возвращаемых метрик
	resutl := make([]view.Metric, 0, len(metrics))
	// Начало транзакции
	tx, err := ms.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// записываем каждую метрику
	for i := range metrics {
		// подготовка переменных
		var value sql.NullFloat64
		var delta sql.NullInt64
		metric := metrics[i]
		// выполнение запроса
		err = tx.QueryRow(ctx, queryAdd, metric.ID, metric.MType, metric.Value, metric.Delta).
			Scan(&value, &delta)
		if err != nil {
			return nil, errors.Join(err, tx.Rollback(ctx))
		}
		// Сохранение ответа
		if value.Valid {
			metric.Value = &value.Float64
		} else if delta.Valid {
			metric.Delta = &delta.Int64
		}
		resutl = append(resutl, metric)
	}

	// Коммитим транзакцию
	return resutl, tx.Commit(ctx)
}

// Метод получения метрики из хранилища.
// Принимает тип метрики и ID метрики.
// Возвращает ошибку в случае если метрика не найдена или у метрики с ID = name другой тип.
// Так же ошибка вернется, если не удалось установить соединение с БД.
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

// Метод получения всех метрик из хранилища.
// Возврашает слайс метрик и ошибку.
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
	slog.Debug("Creating table")
	ctx, cancle := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancle()
	_, err := ms.db.Exec(ctx, queryCheckAndCreateDB)
	return err
}
