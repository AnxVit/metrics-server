package main

import (
	"net/http"

	"github.com/AnxVit/metrics-server/internal/handler"
	"github.com/AnxVit/metrics-server/internal/repository"
	"github.com/AnxVit/metrics-server/internal/service"
)

func main() {
	repo := repository.NewMemStorage()

	service := service.NewService(repo)

	handler := handler.NewHandler(service)

	err := http.ListenAndServe(`:8080`, handler)
	if err != nil {
		panic(err)
	}
}
