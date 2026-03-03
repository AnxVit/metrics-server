package main

import (
	"log"
	"net/http"

	"github.com/AnxVit/metrics-server/internal/handler"
	"github.com/AnxVit/metrics-server/internal/repository"
	"github.com/AnxVit/metrics-server/internal/service"
)

func main() {
	parseFlag()

	repo := repository.NewMemStorage()

	service := service.NewService(repo)

	handler := handler.NewHandler(service)

	log.Printf("Listen %s", addr)

	err := http.ListenAndServe(addr, handler)
	if err != nil {
		panic(err)
	}
}
