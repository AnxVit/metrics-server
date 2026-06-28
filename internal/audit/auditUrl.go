package audit

import (
	"encoding/json"

	"github.com/AnxVit/metrics-server/internal/logger"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var _ Audit = &AuditURL{}

const (
	auditURLName = "audit-url"
)

type AuditURL struct {
	url string

	client *resty.Client
}

func NewAuditURL(url string) *AuditURL {
	client := resty.New()

	return &AuditURL{
		url:    url,
		client: client,
	}
}

func (a *AuditURL) GetID() string {
	return auditURLName
}

func (a *AuditURL) Update(event Event) {
	jsonBody, err := json.Marshal(event)
	if err != nil {
		logger.Log.Warn("Failed to marshal event", zap.Error(err))
		return
	}
	_, err = a.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonBody).
		Post(a.url)

	if err != nil {
		logger.Log.Warn("Failed to send post", zap.Error(err))
		return
	}
}
