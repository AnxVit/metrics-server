package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/AnxVit/metrics-server/internal/logger"
	"github.com/avast/retry-go/v5"
	"go.uber.org/zap"
)

func LinerDelay(n uint, err error, config retry.DelayContext) time.Duration {
	return config.Delay() + time.Duration(2*int(n))*time.Second
}

func NewRetryer(
	attempt uint, delay, maxDelay time.Duration, retryIf func(err error) bool) *retry.Retrier {
	options := []retry.Option{
		retry.Attempts(attempt),
		retry.Delay(delay),
		retry.MaxDelay(maxDelay),
		retry.LastErrorOnly(true),
		retry.DelayType(LinerDelay),
		retry.OnRetry(func(n uint, err error) {
			logger.Log.Warn("Retrying",
				zap.Uint("attempt", n),
				zap.Error(err),
			)
		}),
		retry.RetryIf(retryIf),
	}

	return retry.New(options...)
}

func VerifyHash(jsonBody []byte, hash, key string) bool {
	h := hmac.New(sha256.New, []byte(key))
	if _, err := h.Write(jsonBody); err != nil {
		return false
	}

	hashBody := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(hashBody), []byte(hash))
}

func HashByKey(jsonBody []byte, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))
	if _, err := h.Write(jsonBody); err != nil {
		return "", fmt.Errorf("failed to compute HMAC: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
