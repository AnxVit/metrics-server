package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AnxVit/metrics-server/internal/handler"
	"github.com/AnxVit/metrics-server/internal/logger"
	"github.com/AnxVit/metrics-server/internal/repository"
	"github.com/AnxVit/metrics-server/internal/service"
)

func main() {
	var opt Options
	parseFlag(&opt)

	logger.Initialize("INFO") // tmp: to cfg

	repo := repository.NewMemStorage(
		opt.FileStoragePath, time.Duration(opt.StoreInterval)*time.Second, opt.Restore,
	)

	service := service.NewService(repo)

	handler := handler.NewHandler(service)

	logger.Log.Info(fmt.Sprintf("Listen %s", opt.Addr))

	err := http.ListenAndServe(opt.Addr, handler)
	if err != nil {
		panic(err)
	}
}
