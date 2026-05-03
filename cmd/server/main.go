package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/AnxVit/metrics-server/internal/handler"
	"github.com/AnxVit/metrics-server/internal/logger"
	"github.com/AnxVit/metrics-server/internal/repository"
	"github.com/AnxVit/metrics-server/internal/service"
	"github.com/AnxVit/metrics-server/migrations"
)

const (
	commandUP = "up"
)

func main() {
	var opt Options
	parseFlag(&opt)

	logger.Initialize("INFO") // tmp: to cfg

	ctx := context.Background()

	migrations.Migrate(opt.DatabaseDSN, commandUP, []string{})

	pool, err := pgxpool.New(ctx, opt.DatabaseDSN)
	if err != nil {
		logger.Log.Warn("Couldn't connect to database", zap.Error(err))
	} else {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			logger.Log.Warn("Couldn't connect to database", zap.Error(err))
			pool.Close()
			pool = nil
		} else {
			conn.Release()
		}
	}

	defer func() {
		if pool != nil {
			pool.Close()
		}
	}()

	repoCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	repo := repository.NewRepository(
		repoCtx, pool, opt.FileStoragePath, time.Duration(opt.StoreInterval)*time.Second, opt.Restore,
	)

	service := service.NewService(repo)

	handler := handler.NewHandler(service, pool)

	logger.Log.Info(fmt.Sprintf("Listen %s", opt.Addr))

	err = http.ListenAndServe(opt.Addr, handler)
	if err != nil {
		logger.Log.Fatal(fmt.Sprintf("Listen address: %v", opt.Addr), zap.Error(err))
	}
}
