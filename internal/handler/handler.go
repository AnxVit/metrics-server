package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/AnxVit/metrics-server/internal/handler/middleware"
	models "github.com/AnxVit/metrics-server/internal/model"
)

type iService interface {
	SaveMetric(metric *models.Metrics) error
	GetMetric(metricType, name string) (*models.Metrics, error)
	GetAll() []models.Metrics
}

type Handler struct {
	chi.Router

	service iService
}

func NewHandler(service iService) *Handler {
	h := &Handler{
		service: service,
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/", func(r chi.Router) {
		r.Get("/", h.handleGetAll)
		r.Post("/value", h.handleGetMetric)
		r.Post("/update", h.handlePostMetric)
		r.Get("/value/{type}/{name}", h.handleGetMetricParameters)
		r.Post("/update/{type}/{name}/{value}", h.handlePostMetricParameters)
	})

	h.Router = r
	return h
}

func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	metrics := h.service.GetAll()

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

	metric, err := h.service.GetMetric(req.MType, req.ID)
	if err != nil || metric == nil {
		http.Error(w, "", http.StatusNotFound)
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

	metric, err := h.service.GetMetric(metricType, metricName)
	if err != nil || metric == nil {
		http.Error(w, "", http.StatusNotFound)
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
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}

func (h *Handler) handlePostMetric(w http.ResponseWriter, r *http.Request) {
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

	if req.Delta == nil && req.Value == nil {
		http.Error(w, "value or delta should set", 400)
		return
	}

	if err := h.service.SaveMetric(&req); err != nil {
		http.Error(w, "", http.StatusBadRequest)
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

	if err := h.service.SaveMetric(metric); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
