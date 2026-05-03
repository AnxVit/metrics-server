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

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, opt.DatabaseDSN)
	if err != nil {
		logger.Log.Warn("Couldn't connect to database", zap.Error(err))
	} else {
		defer conn.Close(ctx)
	}

	repoCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	repo := repository.NewRepository(
		repoCtx, conn, opt.FileStoragePath, time.Duration(opt.StoreInterval)*time.Second, opt.Restore,
	)

	service := service.NewService(repo)

	handler := handler.NewHandler(service, conn)

	logger.Log.Info(fmt.Sprintf("Listen %s", opt.Addr))

	err = http.ListenAndServe(opt.Addr, handler)
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("Listen address: %v", opt.Addr), zap.Error(err))
	}
}
