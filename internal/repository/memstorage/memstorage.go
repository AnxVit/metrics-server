package memstorage

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/AnxVit/metrics-server/internal/logger"
	models "github.com/AnxVit/metrics-server/internal/model"
	"github.com/AnxVit/metrics-server/internal/repository/errors"
)

const (
	typeGauge   = "gauge"
	typeCounter = "counter"
)

type MemStorage struct {
	dbConn *pgx.Conn

	storePath      string
	updateDuration time.Duration

	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
}

func NewMemStorage(ctx context.Context, filePath string, updateDuration time.Duration, restore bool) *MemStorage {
	storage := &MemStorage{
		storePath:      filePath,
		updateDuration: updateDuration,
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
	}

	if restore {
		storage.restoreData()
	}

	if storage.updateDuration > 0 {
		go storage.worker(ctx)
	}

	return storage
}

func (m *MemStorage) SaveGauge(_ context.Context, name string, value float64) error {
	m.gaugeMetrics[name] = value
	if m.updateDuration == 0 {
		return m.saveMetricsToFile(m.getAllMetrics())
	}
	return nil
}

func (m *MemStorage) SaveCounter(_ context.Context, name string, value int64) error {
	m.counterMetrics[name] += value
	if m.updateDuration == 0 {
		return m.saveMetricsToFile(m.getAllMetrics())
	}
	return nil
}

func (m *MemStorage) GetGauge(_ context.Context, name string) (float64, error) {
	val, ok := m.gaugeMetrics[name]
	if !ok {
		return 0, errors.ErrorEmptyResult
	}
	return val, nil
}

func (m *MemStorage) GetCounter(_ context.Context, name string) (int64, error) {
	val, ok := m.counterMetrics[name]
	if !ok {
		return 0, errors.ErrorEmptyResult
	}
	return val, nil
}

func (m *MemStorage) GetAll(_ context.Context) (map[string]map[string]interface{}, error) {
	values := make(map[string]map[string]interface{})

	values[typeGauge] = make(map[string]interface{})
	for name, value := range m.gaugeMetrics {
		values[typeGauge][name] = value
	}

	values[typeCounter] = make(map[string]interface{})
	for name, value := range m.counterMetrics {
		values[typeCounter][name] = value
	}

	return values, nil
}

func (m *MemStorage) getAllMetrics() []models.Metrics {
	var allMetrics []models.Metrics

	for name, value := range m.gaugeMetrics {
		allMetrics = append(allMetrics, models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &value,
		})
	}

	for name, value := range m.counterMetrics {
		allMetrics = append(allMetrics, models.Metrics{
			ID:    name,
			MType: "counter",
			Delta: &value,
		})
	}

	return allMetrics
}

func (m *MemStorage) restoreData() {
	var allMetrics []models.Metrics

	file, err := os.OpenFile(m.storePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Log.Warn("Couldn't open file", zap.String("file", m.storePath), zap.Error(err))
		return
	}

	defer file.Close()

	err = json.NewDecoder(file).Decode(&allMetrics)
	if err != nil {
		logger.Log.Warn("Couldn't encode from file", zap.String("file", m.storePath), zap.Error(err))
		return
	}

	for _, metric := range allMetrics {
		switch metric.MType {
		case typeGauge:
			m.gaugeMetrics[metric.ID] = *metric.Value
		case typeCounter:
			m.counterMetrics[metric.ID] = *metric.Delta
		default:
			logger.Log.Warn("Invalid type in file", zap.String("file", m.storePath), zap.String("type", metric.MType))
		}
	}

	logger.Log.Info("Successfuly load data from file", zap.String("file", m.storePath))
}

func (m *MemStorage) saveMetricsToFile(metrics []models.Metrics) error {
	file, err := os.OpenFile(m.storePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	bytesMetrics, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	_, err = file.Write(bytesMetrics)
	if err != nil {
		return err
	}
	file.WriteString("\n")

	file.Close()
	return nil
}

func (m *MemStorage) worker(ctx context.Context) {
	ticker := time.NewTicker(m.updateDuration)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if err := m.saveMetricsToFile(m.getAllMetrics()); err != nil {
			logger.Log.Warn("Couldn't save metrics to file", zap.Error(err))
		}

		logger.Log.Info("Successfuly wrtie data to file", zap.String("file", m.storePath))
	}
}
