package main

import (
	"flag"
	"log"
	"strings"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	Addr           string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

func parseFlag(opt *Options) {
	opt.Addr = "http://localhost:8080"
	flag.Func("a", "server endpoint", func(s string) error {
		parts := strings.Split(s, ":")
		if len(parts) == 2 {
			opt.Addr = "http://" + s
			return nil
		}
		opt.Addr = s
		return nil
	})

	flag.IntVar(&opt.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&opt.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&opt.Key, "k", "", "hash key")
	flag.IntVar(&opt.RateLimit, "l", 1, "rate limit")
	flag.Parse()

	err := env.Parse(opt)
	if err != nil {
		log.Fatalf("failed to parse config env: %s", err.Error())
	}

	parts := strings.Split(opt.Addr, ":")
	if len(parts) == 2 {
		opt.Addr = "http://" + opt.Addr
	}
}
