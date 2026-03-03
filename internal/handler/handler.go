package handler

import (
	"net/http"
	"strconv"

	models "github.com/AnxVit/metrics-server/internal/model"
	"github.com/go-chi/chi/v5"
)

type iService interface {
	SaveMetric(metric *models.Metrics) error
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
	r.Post("/update/{type}/{name}/{value}", h.handlePostMetric)

	h.Router = r
	return h
}

func (h *Handler) handlePostMetric(w http.ResponseWriter, r *http.Request) {
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
