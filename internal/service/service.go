package service

import (
	"errors"
	"strings"

	models "github.com/AnxVit/metrics-server/internal/model"
)

type iRepo interface {
	SaveGauge(name string, value float64)
	SaveCounter(name string, value int64)
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
