package main

import (
	"context"
	"errors"
	"log"

	"regressiondetector/internal/collector/agent"
	"regressiondetector/internal/collector/config"
	"regressiondetector/internal/collector/processor"
	"regressiondetector/internal/collector/sink"
	"regressiondetector/internal/collector/source"
)

func main() {
	cfg,err := config.Load()

	if err != nil {
		log.Fatalf("Failed to load config: %v",err)
	}
	// connStr := "postgres://postgres@localhost:5432/postgres?sslmode=disable"
	// connStr := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

	collectorAgent := agent.New(
		cfg,
		source.NewPostgresSource(cfg.ConnStr),
		processor.RedactingProcessor{},
		sink.LoggingSink{})

	if err := collectorAgent.Run(context.Background()); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("collector run failed: %v", err)
	}
}
