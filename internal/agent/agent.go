package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type Request struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
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
			err := a.sendInfo(info, int64(pollCount))
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

func (a *Agent) sendInfo(info map[string]float64, poolCount int64) error {
	client := resty.New()

	for name, value := range info {
		req := &Request{
			ID:    name,
			MType: "gauge",
			Value: &value,
		}
		jsonBody, err := json.Marshal(req)
		if err != nil {
			return err
		}
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(jsonBody).
			Post(a.addr + "/update")
		if err != nil || resp.StatusCode() != 200 {
			return fmt.Errorf("bad answer: %d", resp.StatusCode())
		}
	}

	req := &Request{
		ID:    "PollCount",
		MType: "counter",
		Delta: &poolCount,
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonBody).
		Post(a.addr + "/update")
	if err != nil || resp.StatusCode() != 200 {
		return fmt.Errorf("bad answer: %d", resp.StatusCode())
	}

	log.Println("Successfully send")
	return nil
}
