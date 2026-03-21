package main

import (
	"flag"
	"strings"
)

type options struct {
	addr           string
	reportInterval int
	pollInterval   int
}

func parseFlag(opt *options) {
	opt.addr = "http://localhost:8080"
	flag.Func("a", "server endpoint", func(s string) error {
		parts := strings.Split(s, ":")
		if len(parts) == 2 {
			opt.addr = "http://" + s
			return nil
		}
		opt.addr = s
		return nil
	})

	flag.IntVar(&opt.reportInterval, "r", 10, "report interval")
	flag.IntVar(&opt.pollInterval, "p", 2, "poll interval")
	flag.Parse()
}
