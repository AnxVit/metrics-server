package database

import (
	"context"
	"errors"
	"time"

	"github.com/AnxVit/metrics-server/internal/logger"
	models "github.com/AnxVit/metrics-server/internal/model"
	dbErrors "github.com/AnxVit/metrics-server/internal/repository/errors"
	"github.com/AnxVit/metrics-server/internal/util"
	"github.com/avast/retry-go/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	typeGauge   = "gauge"
	typeCounter = "counter"
)

type Database struct {
	conn    *pgxpool.Pool
	retrier *retry.Retrier
}

func NewDatabase(conn *pgxpool.Pool) *Database {
	retrier := util.NewRetryer(
		3,
		time.Duration(1)*time.Second,
		time.Duration(5)*time.Second,
		func(err error) bool {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				return len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08"
			}
			return errors.Is(err, context.DeadlineExceeded)
		},
	)

	return &Database{
		conn:    conn,
		retrier: retrier,
	}
}

func (d *Database) SaveMetrics(ctx context.Context, metrics []*models.Metrics) error {
	return d.retrier.Do(func() error {
		tx, err := d.conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() {
			if err := tx.Rollback(ctx); err != nil {
				logger.Log.Warn("Couldn't rollback", zap.Error(err))
			}
		}()

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
	})
}

func (d *Database) SaveGauge(ctx context.Context, name string, value float64) error {
	return d.retrier.Do(func() error {
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
	})
}

func (d *Database) SaveCounter(ctx context.Context, name string, value int64) error {
	return d.retrier.Do(func() error {
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
	})
}

func (d *Database) GetGauge(ctx context.Context, name string) (float64, error) {
	var value float64
	err := d.retrier.Do(func() error {
		err := d.conn.QueryRow(ctx, "select value from metrics where name=$1", name).Scan(&value)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return dbErrors.ErrorEmptyResult
			}
			return err
		}
		return nil
	})

	return value, err
}

func (d *Database) GetCounter(ctx context.Context, name string) (int64, error) {
	var delta int64
	err := d.retrier.Do(func() error {
		err := d.conn.QueryRow(ctx, "select delta from metrics where name=$1", name).Scan(&delta)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return dbErrors.ErrorEmptyResult
			}
			return err
		}
		return nil
	})

	return delta, err
}

func (d *Database) GetAll(ctx context.Context) (map[string]map[string]interface{}, error) {
	allMetrics := make(map[string]map[string]interface{}, 0)

	allMetrics[typeCounter] = make(map[string]interface{})
	allMetrics[typeGauge] = make(map[string]interface{})

	err := d.retrier.Do(func() error {
		rows, err := d.conn.Query(ctx, "SELECT name, type, delta, value from metrics")
		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			var v models.Metrics
			err = rows.Scan(&v.ID, &v.MType, &v.Delta, &v.Value)
			if err != nil {
				return err
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
			return err
		}
		return nil
	})

	return allMetrics, err
}
