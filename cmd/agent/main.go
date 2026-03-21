package main

import (
	"github.com/AnxVit/metrics-server/internal/agent"
)

func main() {
	var opt options
	parseFlag(&opt)
	client := agent.NewAgent(
		opt.addr,
		opt.reportInterval,
		opt.pollInterval,
	)

	client.Work()
}
