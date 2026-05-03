package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AnxVit/metrics-server/internal/handler/middleware"
	models "github.com/AnxVit/metrics-server/internal/model"
	"github.com/AnxVit/metrics-server/internal/service"
)

type iService interface {
	SaveMetrics(ctx context.Context, metric []*models.Metrics) error
	GetMetric(ctx context.Context, metricType, name string) (*models.Metrics, error)
	GetAll(ctx context.Context) ([]models.Metrics, error)
}

type Handler struct {
	chi.Router

	service      iService
	postgresConn *pgxpool.Pool
}

func NewHandler(service iService, conn *pgxpool.Pool) *Handler {
	h := &Handler{
		service:      service,
		postgresConn: conn,
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.GZipMiddleware)
	r.Use(chimiddleware.StripSlashes)
	r.Route("/", func(r chi.Router) {
		r.Get("/", h.handleGetAll)
		r.Post("/value", h.handleGetMetric)
		r.Post("/update", h.handlePostMetric)
		r.Get("/value/{type}/{name}", h.handleGetMetricParameters)
		r.Post("/update/{type}/{name}/{value}", h.handlePostMetricParameters)
		r.Get("/ping", h.handlePing)
	})

	h.Router = r
	return h
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metrics, err := h.service.GetAll(ctx)
	if err != nil {
		http.Error(w, "", 500)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Метрики</title>
    <style>
        table { border-collapse: collapse; width: 50%; margin: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:hover { background-color: #f5f5f5; }
    </style>
</head>
<body>
    <h1>Метрики приложения</h1>
    <table>
        <tr>
            <th>Ключ</th>
            <th>Значение</th>
        </tr>`

	for _, m := range metrics {
		if m.Value != nil {
			html += fmt.Sprintf(`
        <tr>
            <td>%s</td>
            <td>%v</td>
        </tr>`, m.ID, *m.Value)
		}
		if m.Delta != nil {
			html += fmt.Sprintf(`
        <tr>
            <td>%s</td>
            <td>%v</td>
        </tr>`, m.ID, *m.Delta)
		}

	}

	html += `
    </table>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (h *Handler) handleGetMetric(w http.ResponseWriter, r *http.Request) {
	var req models.Metrics
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "bad request format", 400)
		return
	}
	if req.MType == "" || req.ID == "" {
		http.Error(w, "bad request values", 400)
		return
	}

	ctx := r.Context()

	metric, err := h.service.GetMetric(ctx, req.MType, req.ID)
	if metric == nil {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		if errors.Is(err, service.ErrBadMetricType) {
			http.Error(w, "bad metric type", http.StatusBadRequest)
			return
		}
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) handleGetMetricParameters(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	if metricType == "" || metricName == "" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()

	metric, err := h.service.GetMetric(ctx, metricType, metricName)
	if metric == nil {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		if errors.Is(err, service.ErrBadMetricType) {
			http.Error(w, "bad metric type", http.StatusBadRequest)
			return
		}
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var resp []byte
	switch {
	case metric.Delta != nil:
		resp, err = json.Marshal(metric.Delta)
	case metric.Value != nil:
		resp, err = json.Marshal(metric.Value)
	default:
		http.Error(w, "", 500)
		return
	}
	if err != nil {
		http.Error(w, "", 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) handlePostMetric(w http.ResponseWriter, r *http.Request) {
	var req UpdateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "bad request format", 400)
		return
	}

	if err := req.validate(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	ctx := r.Context()

	if err := h.service.SaveMetrics(ctx, req.Metrics); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handlePostMetricParameters(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricType == "" || metricName == "" || metricValue == "" {
		http.NotFound(w, r)
		return
	}

	metric := &models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	delta, err := strconv.ParseInt(metricValue, 0, 64)
	if err == nil {
		metric.Delta = &delta
	}

	value, err := strconv.ParseFloat(metricValue, 64)
	if err == nil {
		metric.Value = &value
	}

	ctx := r.Context()

	if err := h.service.SaveMetrics(ctx, []*models.Metrics{metric}); err != nil {
		if errors.Is(err, service.ErrBadMetricType) {
			http.Error(w, "bad metric type", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrBadMetricValue) {
			http.Error(w, "bad metric value", http.StatusBadRequest)
			return
		}
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handlePing(w http.ResponseWriter, r *http.Request) {
	if h.postgresConn == nil && reflect.ValueOf(h.postgresConn).IsNil() {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err := h.postgresConn.Ping(context.Background()); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
