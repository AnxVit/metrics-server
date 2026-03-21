package repository

type MemStorage struct {
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
	}
}

func (m *MemStorage) SaveGauge(name string, value float64) {
	m.gaugeMetrics[name] = value
}

func (m *MemStorage) SaveCounter(name string, value int64) {
	m.counterMetrics[name] += value
}
