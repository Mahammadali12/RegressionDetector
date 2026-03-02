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

	collectorAgent := agent.New(
		cfg,
		source.NewStaticSource(nil),
		processor.PassThroughProcessor{},
		sink.LoggingSink{},
	)

	if err := collectorAgent.RunOnce(context.Background()); err != nil {
		log.Fatalf("collector run failed: %v", err)
	}
}
