package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	models "github.com/AnxVit/metrics-server/internal/model"
	"github.com/AnxVit/metrics-server/internal/repository"
	"github.com/AnxVit/metrics-server/internal/service"
)

func Test_PostMetric(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		req    *models.Metrics
		code   int
	}{
		{
			name:   "bad method",
			method: http.MethodGet,
			path:   "/update",
			req: &models.Metrics{
				ID:    "someMetrics",
				MType: "gauge",
				Value: toPointer(70.0),
			},
			code: 405,
		},
		{
			name:   "bad path",
			method: http.MethodPost,
			path:   "/gauge",
			req: &models.Metrics{
				ID:    "someMetrics",
				MType: "gauge",
				Value: toPointer(70.0),
			},
			code: 404,
		},
		{
			name:   "success",
			method: http.MethodPost,
			path:   "/update",
			req: &models.Metrics{
				ID:    "someMetrics",
				MType: "gauge",
				Value: toPointer(70.0),
			},
			code: 200,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := repository.NewMemStorage("", time.Hour, false) // later mock
			serv := service.NewService(repo)

			h := NewHandler(serv, nil)

			data, _ := json.Marshal(test.req)

			reader := bytes.NewReader(data)

			req := httptest.NewRequest(test.method, test.path, reader)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != test.code {
				t.Errorf("got %d, want %d", rr.Code, test.code)
			}
		})
	}
}

func Test_GetMetric(t *testing.T) {
	repo := repository.NewMemStorage("", time.Hour, false) // later mock
	serv := service.NewService(repo)
	serv.SaveMetric(&models.Metrics{
		ID:    "someMetric",
		MType: models.Gauge,
		Value: toPointer(37.0),
	})
	tests := []struct {
		name   string
		method string
		path   string
		req    *models.Metrics
		code   int
	}{
		{
			name:   "bad method",
			method: http.MethodGet,
			path:   "/value",
			req: &models.Metrics{
				ID:    "someMetric",
				MType: "gauge",
			},
			code: 405,
		},
		{
			name:   "bad path",
			method: http.MethodPost,
			path:   "/value/gauge",
			req: &models.Metrics{
				ID:    "someMetric",
				MType: "gauge",
			},
			code: 404,
		},
		{
			name:   "success",
			method: http.MethodPost,
			path:   "/value",
			req: &models.Metrics{
				ID:    "someMetric",
				MType: "gauge",
			},
			code: 200,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			h := NewHandler(serv, nil)

			data, _ := json.Marshal(test.req)

			reader := bytes.NewReader(data)

			req := httptest.NewRequest(test.method, test.path, reader)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != test.code {
				t.Errorf("got %d, want %d", rr.Code, test.code)
			}
		})
	}
}

func Test_PostMetricParams(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		code   int
	}{
		{
			name:   "bad method",
			method: http.MethodGet,
			path:   "/update/gauge/someMetrics/70.0",
			code:   405,
		},
		{
			name:   "bad path (without update)",
			method: http.MethodPost,
			path:   "/gauge/someMetrics/70.0",
			code:   404,
		},
		{
			name:   "bad path (length)",
			method: http.MethodPost,
			path:   "/update/gauge/someMetrics",
			code:   404,
		},
		{
			name:   "bad value",
			method: http.MethodPost,
			path:   "/update/gauge/someMetrics/value",
			code:   400,
		},
		{
			name:   "success",
			method: http.MethodPost,
			path:   "/update/gauge/someMetrics/70.0",
			code:   200,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := repository.NewMemStorage("", time.Hour, false) // later mock
			serv := service.NewService(repo)

			h := NewHandler(serv, nil)

			req := httptest.NewRequest(test.method, test.path, nil)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != test.code {
				t.Errorf("got %d, want %d", rr.Code, test.code)
			}
		})
	}
}

func Test_GetParameters(t *testing.T) {
	repo := repository.NewMemStorage("", time.Hour, false) // later mock
	serv := service.NewService(repo)
	serv.SaveMetric(&models.Metrics{
		ID:    "someMetric",
		MType: models.Gauge,
		Value: toPointer(37.0),
	})
	tests := []struct {
		name   string
		method string
		path   string
		code   int
	}{
		{
			name:   "bad method",
			method: http.MethodPost,
			path:   "/value/gauge/someMetric",
			code:   405,
		},
		{
			name:   "bad path",
			method: http.MethodGet,
			path:   "/value/gauge/someMetric/45.0",
			code:   404,
		},
		{
			name:   "success",
			method: http.MethodGet,
			path:   "/value/gauge/someMetric",
			code:   200,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			h := NewHandler(serv, nil)

			req := httptest.NewRequest(test.method, test.path, nil)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != test.code {
				t.Errorf("got %d, want %d", rr.Code, test.code)
			}
		})
	}
}

func toPointer[T any](val T) *T {
	return &val
}
