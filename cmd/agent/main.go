package main

import "github.com/AnxVit/metrics-server/internal/agent"

func main() {
	client := agent.NewAgent("http://localhost:8080")

	client.Work()
}
