package repository

const (
	typeGauge   = "gauge"
	typeCounter = "counter"
)

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
