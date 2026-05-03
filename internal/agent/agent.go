package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/AnxVit/metrics-server/internal/logger"
	models "github.com/AnxVit/metrics-server/internal/model"
	"github.com/AnxVit/metrics-server/internal/util"
)

type Request struct {
	Metrics []models.Metrics
}

type Agent struct {
	client  *http.Client
	retrier *retry.Retrier

	addr           string
	key            string
	reportInterval time.Duration
	pollInterval   time.Duration
	rateLimit      int

	wg    *errgroup.Group
	queue chan []models.Metrics
}

func NewAgent(ctx context.Context, addr, key string, reportInter, pollInter, rateLimit int) *Agent {
	retrier := util.NewRetryer(
		3,
		time.Duration(1)*time.Second,
		time.Duration(5)*time.Second,
		func(err error) bool {
			var netErr net.Error
			if errors.As(err, &netErr) {
				return true
			}

			if httpErr, ok := err.(interface{ StatusCode() int }); ok {
				code := httpErr.StatusCode()
				return code == 429 || code == 408 || (code >= 500 && code < 600)
			}

			return false
		},
	)

	agent := &Agent{
		client: &http.Client{
			Timeout: 100 * time.Millisecond,
		},
		retrier: retrier,

		addr:           addr,
		key:            key,
		reportInterval: time.Duration(reportInter) * time.Second,
		pollInterval:   time.Duration(pollInter) * time.Second,
		rateLimit:      rateLimit,

		wg:    new(errgroup.Group),
		queue: make(chan []models.Metrics, rateLimit*2),
	}

	agent.startWorks(ctx)

	return agent
}

func (a *Agent) Work(ctx context.Context) error {
	chanInfo := make(chan map[string]float64, 1)
	chanSystem := make(chan map[string]float64, 1)

	a.wg.Go(func() error {
		ticker := time.NewTicker(a.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return nil
			}
			info := a.getRuntimeInfo()
			select {
			case chanInfo <- info:
			default:
			}
		}
	})

	a.wg.Go(func() error {
		ticker := time.NewTicker(a.reportInterval)
		defer ticker.Stop()

		var (
			lastInfo   map[string]float64
			lastSystem map[string]float64
			pollCount  int64
		)
		for {
			select {
			case <-ctx.Done():
				return nil
			case systemInfo := <-chanInfo:
				lastInfo = make(map[string]float64)
				maps.Copy(lastInfo, systemInfo)
				pollCount += 1
			case systemInfo := <-chanSystem:
				lastSystem = make(map[string]float64)
				maps.Copy(lastSystem, systemInfo)
			case <-ticker.C:
				if lastInfo != nil {
					toSendMap := make(map[string]float64)
					maps.Copy(toSendMap, lastSystem)
					maps.Copy(toSendMap, lastInfo)
					a.sendBatch(toSendMap, int64(pollCount))
					pollCount = 0
					lastInfo = nil
					lastSystem = nil
				}
			}
		}
	})

	a.wg.Go(func() error {
		ticker := time.NewTicker(a.pollInterval)
		defer ticker.Stop()
		defer close(chanSystem)

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				info := a.getSystemInfo()
				select {
				case chanSystem <- info:
				default:
				}
			}
		}
	})

	return a.wg.Wait()
}

func (a *Agent) startWorks(ctx context.Context) {
	logger.Log.Info("Start workers", zap.Int16("count", int16(a.rateLimit)))
	for i := 0; i < a.rateLimit; i++ {
		a.wg.Go(func() error {
			return a.worker(ctx)
		})
	}
}

func (a *Agent) worker(ctx context.Context) error {
	client := resty.New()

	for {
		select {
		case <-ctx.Done():
			return nil
		case batch := <-a.queue:
			logger.Log.Warn("Send batch")
			if err := a.sendInfo(client, batch); err != nil {
				logger.Log.Warn("Failed to send metrics", zap.Error(err))
			}
		}
	}
}

func (a *Agent) sendBatch(info map[string]float64, poolCount int64) {
	metrics := make([]models.Metrics, 0, len(info))
	for name, value := range info {
		val := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &val,
		})
	}

	metrics = append(metrics, models.Metrics{
		ID:    "PollCount",
		MType: "counter",
		Delta: &poolCount,
	})

	select {
	case a.queue <- metrics:
		logger.Log.Debug("Batch submitted",
			zap.Int("metrics_count", len(metrics)))
	default:
	}
}

func (a *Agent) getRuntimeInfo() map[string]float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	runtimeInfo := map[string]float64{
		"Alloc":         float64(m.Alloc),
		"BuckHashSys":   float64(m.BuckHashSys),
		"Frees":         float64(m.Frees),
		"GCCPUFraction": m.GCCPUFraction,
		"GCSys":         float64(m.GCSys),
		"HeapAlloc":     float64(m.HeapAlloc),
		"HeapIdle":      float64(m.HeapIdle),
		"HeapInuse":     float64(m.HeapInuse),
		"HeapObjects":   float64(m.HeapObjects),
		"HeapReleased":  float64(m.HeapReleased),
		"HeapSys":       float64(m.HeapSys),
		"LastGC":        float64(m.LastGC),
		"Lookups":       float64(m.Lookups),
		"MCacheInuse":   float64(m.MCacheInuse),
		"MCacheSys":     float64(m.MCacheSys),
		"MSpanInuse":    float64(m.MSpanInuse),
		"MSpanSys":      float64(m.MSpanSys),
		"Mallocs":       float64(m.Mallocs),
		"NextGC":        float64(m.NextGC),
		"NumForcedGC":   float64(m.NumForcedGC),
		"NumGC":         float64(m.NumGC),
		"OtherSys":      float64(m.OtherSys),
		"PauseTotalNs":  float64(m.PauseTotalNs),
		"StackInuse":    float64(m.StackInuse),
		"StackSys":      float64(m.StackSys),
		"Sys":           float64(m.Sys),
		"RandomValue":   rand.Float64(),
		"TotalAlloc":    float64(m.TotalAlloc),
	}
	return runtimeInfo
}

func (a *Agent) getSystemInfo() map[string]float64 {
	info := make(map[string]float64)

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Warn("Failed to get memory info", zap.Error(err))
	} else {
		info["TotalMemory"] = float64(memInfo.Total)
		info["FreeMemory"] = float64(memInfo.Free)
	}

	cpuPercent, err := cpu.Percent(500*time.Millisecond, true)
	if err != nil {
		logger.Log.Warn("Failed to get CPU percent", zap.Error(err))
	} else {
		for i, percent := range cpuPercent {
			info[fmt.Sprintf("CPUutilization%d", i+1)] = percent
		}
	}

	return info
}

func (a *Agent) sendInfo(client *resty.Client, models []models.Metrics) error {
	jsonBody, err := json.Marshal(models)
	if err != nil {
		return err
	}

	var hashData string
	if a.key != "" {
		hashData, err = util.HashByKey(jsonBody, a.key)
		if err != nil {
			return err
		}
	}

	jsonBody, err = compress(jsonBody)
	if err != nil {
		return err
	}

	return a.retrier.Do(func() error {
		req := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip")
		if a.key != "" {
			req.SetHeader("HashSHA256", hashData)
		}
		resp, err := req.
			SetBody(jsonBody).
			Post(a.addr + "/updates")

		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			return &CodeError{
				err:  resp.String(),
				code: resp.StatusCode(),
			}
		}
		return nil
	})
}

func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer

	writer := gzip.NewWriter(&b)

	_, err := writer.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	return b.Bytes(), nil
}

type CodeError struct {
	err  string
	code int
}

func (c *CodeError) Error() string {
	return fmt.Sprintf("%s (%d)", c.err, c.code)
}

func (c *CodeError) StatusCode() int {
	return c.code
}
