package main

import (
	"context"
	"log"

	"regressiondetector/internal/collector/agent"
	"regressiondetector/internal/collector/config"
	"regressiondetector/internal/collector/processor"
	"regressiondetector/internal/collector/sink"
	"regressiondetector/internal/collector/source"
)

func main() {
	cfg := config.Default()
	// connStr := "postgres://postgres@localhost:5432/postgres?sslmode=disable"
	connStr := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

	collectorAgent := agent.New(
		cfg,
		source.NewPostgresSource(connStr),
		processor.PassThroughProcessor{},
		sink.LoggingSink{})

	if err := collectorAgent.RunOnce(context.Background()); err != nil {
		log.Fatalf("collector run failed: %v", err)
	}
}
