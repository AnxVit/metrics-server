package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

const (
	reportInterval = 10 * time.Second
	pollInterval   = 2 * time.Second
)

type Agent struct {
	client *http.Client
	addr   string
}

func NewAgent(addr string) *Agent {
	return &Agent{
		client: &http.Client{
			Timeout: 100 * time.Millisecond,
		},
		addr: addr,
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
			time.Sleep(pollInterval)
			mu.Lock()
			info = a.getRuntimeInfo()
			pollCount += 1
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		for {
			time.Sleep(reportInterval)
			mu.Lock()
			err := a.sendInfo(info, pollCount)
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
	}
	return runtimeInfo
}

func (a *Agent) sendInfo(info map[string]float64, poolCount int) error {
	for name, value := range info {
		url := a.addr + fmt.Sprintf("/update/gauge/%s/%f", name, value)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "text/plain")
		response, err := a.client.Do(req)
		if err != nil {
			return err
		}
		response.Body.Close()
	}

	url := a.addr + fmt.Sprintf("/update/counter/PollCount/%d", poolCount)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")
	response, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	log.Println("Successfully send")
	return nil
}
