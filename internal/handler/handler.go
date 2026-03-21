package handler

import (
	"net/http"
	"strconv"
	"strings"

	models "github.com/AnxVit/metrics-server/internal/model"
)

type iService interface {
	SaveMetric(metric *models.Metrics) error
}

type Handler struct {
	service iService
}

func NewHandler(service iService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	pathsParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathsParts) != 4 || pathsParts[0] != "update" {
		http.NotFound(w, r)
		return
	}

	metricType := pathsParts[1]
	metricName := pathsParts[2]
	metricValue := pathsParts[3]

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
