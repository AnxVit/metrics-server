package main

import (
	"github.com/AnxVit/metrics-server/internal/agent"
)

func main() {
	var opt Options
	parseFlag(&opt)

	client := agent.NewAgent(
		opt.Addr,
		opt.ReportInterval,
		opt.PollInterval,
	)

	client.Work()
}
