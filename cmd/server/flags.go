package main

import (
	"flag"
	"log"
	"strings"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	Addr            string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
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

	flag.IntVar(&opt.StoreInterval, "i", 300, "store interval")
	flag.StringVar(&opt.FileStoragePath, "f", "~/tmp_store.txt", "file path for store")
	flag.BoolVar(&opt.Restore, "r", false, "should load previously saved values from the specified file when starting the server")

	flag.Parse()

	err := env.Parse(opt)
	if err != nil {
		log.Fatalf("failed to parse: %s", err.Error())
	}
}
