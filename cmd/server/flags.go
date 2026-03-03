package main

import (
	"flag"
	"strings"
)

var addr string

func parseFlag() {
	addr = "localhost:8080"
	flag.Func("a", "server endpoint", func(s string) error {
		parts := strings.Split(s, ":")
		if len(parts) == 3 {
			addr = strings.Trim(parts[1], "/") + ":" + parts[2]
			return nil
		}
		addr = s
		return nil
	})

	flag.Parse()
}
