package repository

import (
	"encoding/json"
	"os"
	"time"

	"github.com/AnxVit/metrics-server/internal/logger"
	models "github.com/AnxVit/metrics-server/internal/model"
	"go.uber.org/zap"
)

const (
	typeGauge   = "gauge"
	typeCounter = "counter"
)

type MemStorage struct {
	storePath      string
	updateDuration time.Duration

	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
}

func NewMemStorage(filePath string, updateDuration time.Duration, restore bool) *MemStorage {
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
		go storage.worker()
	}

	return storage
}

func (m *MemStorage) SaveGauge(name string, value float64) {
	m.gaugeMetrics[name] = value
	if m.updateDuration == 0 {
		m.saveMetricsToFile(m.getAllMetrics())
	}
}

func (m *MemStorage) SaveCounter(name string, value int64) {
	m.counterMetrics[name] += value
	if m.updateDuration == 0 {
		m.saveMetricsToFile(m.getAllMetrics())
	}
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	val, ok := m.gaugeMetrics[name]
	return val, ok
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	val, ok := m.counterMetrics[name]
	return val, ok
}

func (m *MemStorage) GetAll() map[string]map[string]interface{} {
	values := make(map[string]map[string]interface{})

	values[typeGauge] = make(map[string]interface{})
	for name, value := range m.gaugeMetrics {
		values[typeGauge][name] = value
	}

	values[typeCounter] = make(map[string]interface{})
	for name, value := range m.counterMetrics {
		values[typeCounter][name] = value
	}

	return values
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

func (m *MemStorage) saveMetricsToFile(metrics []models.Metrics) {
	file, err := os.OpenFile(m.storePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		logger.Log.Warn("Couldn't open file", zap.String("file", m.storePath), zap.Error(err))
		return
	}

	bytesMetrics, err := json.Marshal(metrics)
	if err != nil {
		logger.Log.Warn("Couldn't marshal array metrics")
		return
	}

	_, err = file.Write(bytesMetrics)
	if err != nil {
		logger.Log.Warn("Couldn't write in file", zap.String("file", m.storePath))
		return
	}
	file.WriteString("\n")

	file.Close()
}

func (m *MemStorage) worker() { // add ctx for shutdown
	for {
		time.Sleep(m.updateDuration)

		m.saveMetricsToFile(m.getAllMetrics())

		logger.Log.Info("Successfuly wrtie data to file", zap.String("file", m.storePath))
	}
}
