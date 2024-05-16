package postgres

const (
	// Запрос для получения всех метрик в базе данных.
	gueryGetAll = `SELECT id, mtype, value, delta FROM metrics`
	// Запрос для получения метрики по её имени и типу.
	queryGetOne = `SELECT value, delta
	FROM metrics
	WHERE id = $1 AND mtype = $2
	LIMIT 1`
	// Запрос для добавления метрики и возврата её обновленного значения.
	queryAdd = `INSERT INTO metrics (id, mtype, value, delta)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE
	SET 
		value = EXCLUDED.value,
		delta = CASE WHEN EXCLUDED.mtype = 'counter' THEN metrics.delta + EXCLUDED.delta ELSE EXCLUDED.delta END
	RETURNING value, delta;`
	// Запрос для проверки существования таблицы и её создания при необходимости.
	queryCheckAndCreateDB = `DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'metrics') THEN
			CREATE TABLE metrics (
				id VARCHAR(255) UNIQUE,
				mtype VARCHAR(255),
				value DOUBLE PRECISION,
				delta BIGINT
			);
		END IF;
	END $$;`
)
