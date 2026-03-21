package service

import (
	"testing"

	models "github.com/AnxVit/metrics-server/internal/model"
	"github.com/AnxVit/metrics-server/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SaveMetric(t *testing.T) {
	tests := []struct {
		name     string
		metric   models.Metrics
		errorMsg string
	}{
		{
			name: "success gauge",
			metric: models.Metrics{
				ID:    "id",
				MType: models.Gauge,
				Value: toPointer(0.0),
			},
		},
		{
			name: "success counter",
			metric: models.Metrics{
				ID:    "id",
				MType: models.Counter,
				Delta: toPointer(int64(0)),
			},
		},
		{
			name: "bad type",
			metric: models.Metrics{
				ID:    "id",
				MType: "vector",
				Delta: toPointer(int64(0)),
			},
			errorMsg: "bad metric type",
		},
		{
			name: "bad gauge value",
			metric: models.Metrics{
				ID:    "id",
				MType: models.Gauge,
				Delta: toPointer(int64(0)),
			},
			errorMsg: "bad gauge type",
		},
		{
			name: "bad counter value",
			metric: models.Metrics{
				ID:    "id",
				MType: models.Counter,
			},
			errorMsg: "bad counter type",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := repository.NewMemStorage() // later mock
			serv := NewService(repo)

			err := serv.SaveMetric(&test.metric)
			if test.errorMsg == "" {
				require.NoError(t, err)
			} else {
				require.Equal(t, test.errorMsg, err.Error())
			}
		})
	}
}

func Test_GetMetric(t *testing.T) {
	repo := repository.NewMemStorage() // later mock
	repo.SaveCounter("counter1", 0)
	repo.SaveGauge("gauge1", 0.0)
	tests := []struct {
		name     string
		input    models.Metrics
		expected *models.Metrics
		errorMsg string
	}{
		{
			name: "success gauge",
			input: models.Metrics{
				ID:    "gauge1",
				MType: models.Gauge,
			},
			expected: &models.Metrics{
				ID:    "gauge1",
				MType: models.Gauge,
				Value: toPointer(0.0),
			},
		},
		{
			name: "success counter",
			input: models.Metrics{
				ID:    "counter1",
				MType: models.Counter,
			},
			expected: &models.Metrics{
				ID:    "counter1",
				MType: models.Counter,
				Delta: toPointer(int64(0)),
			},
		},
		{
			name: "bad type",
			input: models.Metrics{
				ID:    "id",
				MType: "vector",
				Delta: toPointer(int64(0)),
			},
			errorMsg: "bad metric type",
		},
		{
			name: "don't exists",
			input: models.Metrics{
				ID:    "counter2",
				MType: models.Counter,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serv := NewService(repo)

			metric, err := serv.GetMetric(test.input.MType, test.input.ID)
			if test.errorMsg == "" {
				assert.NoError(t, err)
				if test.expected != nil {
					require.Equal(t, test.expected, metric)
				} else {
					require.Nil(t, metric)
				}
			} else {
				require.Equal(t, test.errorMsg, err.Error())
			}
		})
	}
}

func toPointer[T any](val T) *T {
	return &val
}
