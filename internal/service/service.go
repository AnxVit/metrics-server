package service

import (
	"context"
	"errors"
	"strings"

	"github.com/AnxVit/metrics-server/internal/logger"
	models "github.com/AnxVit/metrics-server/internal/model"
	repoErrors "github.com/AnxVit/metrics-server/internal/repository/errors"
	"go.uber.org/zap"
)

var (
	ErrBadMetricType  = errors.New("bad metric type")
	ErrBadMetricValue = errors.New("bad metric value")
)

type iRepo interface {
	SaveGauge(ctx context.Context, name string, value float64) error
	SaveCounter(ctx context.Context, name string, value int64) error

	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)

	GetAll(ctx context.Context) (map[string]map[string]interface{}, error)
}

type Service struct {
	repo iRepo
}

func NewService(repo iRepo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) SaveMetric(ctx context.Context, metric *models.Metrics) error {
	metricType := strings.TrimSpace(strings.ToLower(metric.MType))
	var err error

	switch metricType {
	case models.Gauge:
		if metric.Value == nil {
			return ErrBadMetricValue
		}
		err = s.repo.SaveGauge(ctx, metric.ID, *metric.Value)
	case models.Counter:
		if metric.Delta == nil {
			return ErrBadMetricValue
		}
		err = s.repo.SaveCounter(ctx, metric.ID, *metric.Delta)
	default:
		return ErrBadMetricType
	}

	if err != nil {
		logger.Log.Warn("Couldn't save metric", zap.Error(err))
	}

	return nil
}

func (s *Service) GetMetric(ctx context.Context, metricType, name string) (*models.Metrics, error) {
	metricType = strings.TrimSpace(strings.ToLower(metricType))

	switch metricType {
	case models.Gauge:
		value, err := s.repo.GetGauge(ctx, name)
		if err != nil {
			if !errors.Is(err, repoErrors.ErrorEmptyResult) {
				logger.Log.Warn("Couldn't get metric", zap.Error(err))
				return nil, err
			}
			return nil, nil
		}

		return &models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
		}, nil
	case models.Counter:
		value, err := s.repo.GetCounter(ctx, name)
		if err != nil {
			if !errors.Is(err, repoErrors.ErrorEmptyResult) {
				logger.Log.Warn("Couldn't get metric", zap.Error(err))
				return nil, err
			}
			return nil, nil
		}

		return &models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &value,
		}, nil
	}

	return nil, ErrBadMetricType
}

func (s *Service) GetAll(ctx context.Context) ([]models.Metrics, error) {
	res := make([]models.Metrics, 0)

	values, err := s.repo.GetAll(ctx)
	if err != nil {
		if !errors.Is(err, repoErrors.ErrorEmptyResult) {
			logger.Log.Warn("Couldn't get metrics", zap.Error(err))
			return nil, err
		}
		return res, nil
	}

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
	return res, nil
}
