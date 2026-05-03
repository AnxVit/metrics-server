package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/AnxVit/metrics-server/internal/logger"
	models "github.com/AnxVit/metrics-server/internal/model"
	"github.com/AnxVit/metrics-server/internal/util"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type Request struct {
	Metrics []models.Metrics
}

type Agent struct {
	client *http.Client

	addr           string
	reportInterval time.Duration
	pollInterval   time.Duration
}

func NewAgent(addr string, reportInter, pollInter int) *Agent {
	return &Agent{
		client: &http.Client{
			Timeout: 100 * time.Millisecond,
		},
		addr:           addr,
		reportInterval: time.Duration(reportInter) * time.Second,
		pollInterval:   time.Duration(pollInter) * time.Second,
	}
}

func (a *Agent) Work() {
	mu := &sync.Mutex{}
	var info map[string]float64
	var pollCount int

	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			time.Sleep(a.pollInterval)
			mu.Lock()
			info = a.getRuntimeInfo()
			pollCount += 1
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		for {
			time.Sleep(a.reportInterval)
			mu.Lock()
			err := a.sendAllInfo(info, int64(pollCount))
			if err != nil {
				log.Printf("ERR: %v\n", err)
			}
			pollCount = 0
			mu.Unlock()
		}
	}()

	wg.Wait()
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

func (a *Agent) sendAllInfo(info map[string]float64, poolCount int64) error {
	client := resty.New()

	metrics := make([]models.Metrics, 0)

	for name, value := range info {
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &value,
		})
	}

	if err := a.sendInfo(client, metrics); err != nil {
		logger.Log.Warn("Couldn't send info to main service", zap.Error(err))
	}

	metrics = []models.Metrics{
		{
			ID:    "PollCount",
			MType: "counter",
			Delta: &poolCount,
		},
	}

	if err := a.sendInfo(client, metrics); err != nil {
		logger.Log.Warn("Couldn't send info to main service", zap.Error(err))
		return err
	}

	log.Println("Successfully send")
	return nil
}

func (a *Agent) sendInfo(client *resty.Client, models []models.Metrics) error {
	jsonBody, err := json.Marshal(models)
	if err != nil {
		return err
	}

	jsonBody, err = compress(jsonBody)
	if err != nil {
		return err
	}

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

	return retrier.Do(func() error {
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
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
