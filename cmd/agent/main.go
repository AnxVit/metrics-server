package main

import (
	"context"

	"github.com/AnxVit/metrics-server/internal/agent"
	"github.com/AnxVit/metrics-server/internal/logger"
)

func main() {
	var opt Options
	parseFlag(&opt)

	logger.Initialize("INFO")

	ctx := context.Background()

	client := agent.NewAgent(
		ctx,
		opt.Addr,
		opt.Key,
		opt.ReportInterval,
		opt.PollInterval,
		opt.RateLimit,
	)

	if err := client.Work(ctx); err != nil {
		panic(err)
	}
}
