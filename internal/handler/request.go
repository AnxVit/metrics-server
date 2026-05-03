package handler

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AnxVit/metrics-server/internal/logger"
	models "github.com/AnxVit/metrics-server/internal/model"
)

type UpdateRequest struct {
	Metrics []*models.Metrics
}

func (r *UpdateRequest) validate() error {
	var err error
	for _, metric := range r.Metrics {
		if metric.MType == "" || metric.ID == "" {
			return errors.New("bad request values")
		}

		if metric.Delta == nil && metric.Value == nil {
			return errors.New("value or delta should set")
		}
	}
	return err
}

func (r *UpdateRequest) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return errors.New("empty request body")
	}

	logger.Log.Info(string(data))

	var metrics []*models.Metrics
	if err := json.Unmarshal(data, &metrics); err == nil {
		r.Metrics = metrics
		return nil
	}

	var metric *models.Metrics
	if err := json.Unmarshal(data, &metric); err == nil {
		r.Metrics = []*models.Metrics{metric}
		return nil
	}

	return fmt.Errorf("invalid request: expected array or single metric object")
}
