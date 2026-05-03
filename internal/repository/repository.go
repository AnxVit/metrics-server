package repository

import (
	"context"
	"reflect"
	"time"

	"github.com/AnxVit/metrics-server/internal/repository/database"
	"github.com/AnxVit/metrics-server/internal/repository/memstorage"
	"github.com/jackc/pgx/v5"
)

const (
	typeGauge   = "gauge"
	typeCounter = "counter"
)

type Storage interface {
	SaveGauge(ctx context.Context, name string, value float64) error
	SaveCounter(ctx context.Context, name string, value int64) error
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAll(ctx context.Context) (map[string]map[string]interface{}, error)
}

type Repository struct {
	storage Storage
}

func NewRepository(ctx context.Context, conn *pgx.Conn, filePath string, updateDuration time.Duration, restore bool) *Repository {
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

func (m *Repository) SaveGauge(ctx context.Context, name string, value float64) error {
	return m.storage.SaveGauge(ctx, name, value)
}

func (m *Repository) SaveCounter(ctx context.Context, name string, value int64) error {
	return m.storage.SaveCounter(ctx, name, value)
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
