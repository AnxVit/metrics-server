package database

import (
	"context"
	"errors"

	"github.com/AnxVit/metrics-server/internal/logger"
	models "github.com/AnxVit/metrics-server/internal/model"
	dbErrors "github.com/AnxVit/metrics-server/internal/repository/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	typeGauge   = "gauge"
	typeCounter = "counter"
)

type Database struct {
	conn *pgxpool.Pool
}

func NewDatabase(conn *pgxpool.Pool) *Database {
	return &Database{
		conn: conn,
	}
}

func (d *Database) SaveMetrics(ctx context.Context, metrics []*models.Metrics) error {
	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, m := range metrics {
		var query string
		var args []interface{}

		if m.MType == "gauge" {
			query = `INSERT INTO metrics (name, type, value) 
                     VALUES ($1, $2, $3)
                     ON CONFLICT (name) DO UPDATE SET value = $3`
			args = []interface{}{m.ID, m.MType, m.Value}
		} else {
			query = `INSERT INTO metrics (name, type, delta) 
                     VALUES ($1, $2, $3)
                     ON CONFLICT (name) 
                     DO UPDATE SET delta = metrics.delta + $3`
			args = []interface{}{m.ID, m.MType, m.Delta}
		}

		if _, err := tx.Exec(ctx, query, args...); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (d *Database) SaveGauge(ctx context.Context, name string, value float64) error {
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			logger.Log.Warn("Couldn't rollback", zap.Error(err))
		}
	}()

	_, err = tx.Exec(ctx, `INSERT INTO metrics(type, name, value) VALUES ($1, $2, $3) ON CONFLICT (name) 
	DO UPDATE SET 
    value = EXCLUDED.value,
    updated_at = NOW();`, typeGauge, name, value)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (d *Database) SaveCounter(ctx context.Context, name string, value int64) error {
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			logger.Log.Warn("Couldn't rollback", zap.Error(err))
		}
	}()
	_, err = tx.Exec(ctx, `INSERT INTO metrics(type, name, delta) VALUES ($1, $2, $3) ON CONFLICT (name) 
	DO UPDATE SET 
    delta = metrics.delta + EXCLUDED.delta,
    updated_at = NOW();`, typeCounter, name, value)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (d *Database) GetGauge(ctx context.Context, name string) (float64, error) {
	var value float64
	err := d.conn.QueryRow(context.Background(), "select value from metrics where name=$1", name).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, dbErrors.ErrorEmptyResult
		}
		return 0, err
	}
	return value, nil
}

func (d *Database) GetCounter(ctx context.Context, name string) (int64, error) {
	var delta int64
	err := d.conn.QueryRow(context.Background(), "select delta from metrics where name=$1", name).Scan(&delta)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, dbErrors.ErrorEmptyResult
		}
		return 0, err
	}

	return delta, nil
}

func (d *Database) GetAll(ctx context.Context) (map[string]map[string]interface{}, error) {
	allMetrics := make(map[string]map[string]interface{}, 0)

	allMetrics[typeCounter] = make(map[string]interface{})
	allMetrics[typeGauge] = make(map[string]interface{})

	rows, err := d.conn.Query(ctx, "SELECT name, type, delta, value from metrics")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var v models.Metrics
		err = rows.Scan(&v.ID, &v.MType, &v.Delta, &v.Value)
		if err != nil {
			return nil, err
		}

		switch v.MType {
		case typeCounter:
			allMetrics[typeCounter][v.ID] = *v.Delta
		case typeGauge:
			allMetrics[typeGauge][v.ID] = *v.Value
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return allMetrics, nil
}
