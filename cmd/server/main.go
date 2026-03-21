package main

import (
	"log"
	"net/http"

	"github.com/AnxVit/metrics-server/internal/handler"
	"github.com/AnxVit/metrics-server/internal/repository"
	"github.com/AnxVit/metrics-server/internal/service"
)

func main() {
	var opt Options
	parseFlag(&opt)

	repo := repository.NewMemStorage()

	service := service.NewService(repo)

	handler := handler.NewHandler(service)

	log.Printf("Listen %s", opt.Addr)

	err := http.ListenAndServe(opt.Addr, handler)
	if err != nil {
		panic(err)
	}
}
