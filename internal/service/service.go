package service

import (
	"errors"
	"strings"

	models "github.com/AnxVit/metrics-server/internal/model"
)

type iRepo interface {
	SaveGauge(name string, value float64)
	SaveCounter(name string, value int64)

	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)

	GetAll() map[string]map[string]interface{}
}

type Service struct {
	repo iRepo
}

func NewService(repo iRepo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) SaveMetric(metric *models.Metrics) error {
	metricType := strings.TrimSpace(strings.ToLower(metric.MType))

	switch metricType {
	case models.Gauge:
		if metric.Value == nil {
			return errors.New("bad gauge type")
		}
		s.repo.SaveGauge(metric.ID, *metric.Value)
	case models.Counter:
		if metric.Delta == nil {
			return errors.New("bad counter type")
		}
		s.repo.SaveCounter(metric.ID, *metric.Delta)
	default:
		return errors.New("bad metric type")
	}

	return nil
}

func (s *Service) GetMetric(metricType, name string) (*models.Metrics, error) {
	metricType = strings.TrimSpace(strings.ToLower(metricType))

	switch metricType {
	case models.Gauge:
		value, ok := s.repo.GetGauge(name)
		if ok {
			return &models.Metrics{
				ID:    name,
				MType: models.Gauge,
				Value: &value,
			}, nil
		}
		return nil, nil
	case models.Counter:
		value, ok := s.repo.GetCounter(name)
		if ok {
			return &models.Metrics{
				ID:    name,
				MType: models.Counter,
				Delta: &value,
			}, nil
		}
		return nil, nil
	}

	return nil, errors.New("bad metric type")
}

func (s *Service) GetAll() []models.Metrics {
	res := make([]models.Metrics, 0)

	values := s.repo.GetAll()

	for name, iValue := range values[models.Gauge] {
		val, _ := iValue.(float64)
		res = append(res, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &val,
		})
	}
	for name, iValue := range values[models.Counter] {
		val, _ := iValue.(int64)
		res = append(res, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Delta: &val,
		})
	}
	return res
}
