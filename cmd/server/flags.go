package main

import (
	"flag"
	"log"
	"strings"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	Addr string `env:"ADDRESS"`
}

func parseFlag(opt *Options) {
	opt.Addr = "localhost:8080"
	flag.Func("a", "server endpoint", func(s string) error {
		parts := strings.Split(s, ":")
		if len(parts) == 3 {
			opt.Addr = strings.Trim(parts[1], "/") + ":" + parts[2]
			return nil
		}
		opt.Addr = s
		return nil
	})

	flag.Parse()

	err := env.Parse(opt)
	if err != nil {
		log.Fatalf("failed to parse: %s", err.Error())
	}
}
