package service

import (
	"testing"

	models "github.com/AnxVit/metrics-server/internal/model"
	"github.com/AnxVit/metrics-server/internal/repository"
	"github.com/stretchr/testify/require"
)

func Test_Service(t *testing.T) {
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

func toPointer[T any](val T) *T {
	return &val
}
