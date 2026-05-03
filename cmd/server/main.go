package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

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

	conn, err := pgx.Connect(context.Background(), opt.DATABASE_DSN)
	if err != nil {
		logger.Log.Warn(fmt.Sprintf("Couldn't connect to database"), zap.Error(err))
	}
	defer conn.Close(context.Background())

	handler := handler.NewHandler(service, conn)

	logger.Log.Info(fmt.Sprintf("Listen %s", opt.Addr))

	err = http.ListenAndServe(opt.Addr, handler)
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("Listen address: %v", opt.Addr), zap.Error(err))
	}
}
