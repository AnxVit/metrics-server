package repository

import (
	"context"
	"reflect"
	"time"

	models "github.com/AnxVit/metrics-server/internal/model"
	"github.com/AnxVit/metrics-server/internal/repository/database"
	"github.com/AnxVit/metrics-server/internal/repository/memstorage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	SaveMetrics(ctx context.Context, metrics []*models.Metrics) error
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAll(ctx context.Context) (map[string]map[string]interface{}, error)
}

type Repository struct {
	storage Storage
}

func NewRepository(ctx context.Context, conn *pgxpool.Pool, filePath string, updateDuration time.Duration, restore bool) *Repository {
	var storage Storage
	if conn == nil && reflect.ValueOf(conn).IsNil() {
		storage = memstorage.NewMemStorage(ctx, filePath, updateDuration, restore)
	} else {
		storage = database.NewDatabase(conn)
	}

	return &Repository{
		storage: storage,
	}
}

func (m *Repository) SaveMetrics(ctx context.Context, metrics []*models.Metrics) error {
	return m.storage.SaveMetrics(ctx, metrics)
}

func (m *Repository) GetGauge(ctx context.Context, name string) (float64, error) {
	return m.storage.GetGauge(ctx, name)
}

func (m *Repository) GetCounter(ctx context.Context, name string) (int64, error) {
	return m.storage.GetCounter(ctx, name)
}

func (m *Repository) GetAll(ctx context.Context) (map[string]map[string]interface{}, error) {
	return m.storage.GetAll(ctx)
}
