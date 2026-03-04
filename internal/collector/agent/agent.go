package agent

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"
	"time"

	"regressiondetector/internal/collector/config"
	"regressiondetector/internal/collector/processor"
	"regressiondetector/internal/collector/sink"
	"regressiondetector/internal/collector/source"
)

type Agent struct {
	config    config.Config
	source    source.Source
	processor processor.Processor
	sink      sink.Sink
}

func New(
	cfg config.Config,
	src source.Source,
	proc processor.Processor,
	output sink.Sink,
) *Agent {
	return &Agent{
		config:    cfg,
		source:    src,
		processor: proc,
		sink:      output,
	}
}

func (a *Agent) RunOnce(ctx context.Context) error {
	records, err := a.source.Collect(ctx)
	if err != nil {
		return err
	}

	processed, err := a.processor.Process(ctx, records)
	if err != nil {
		return err
	}

	if err := a.sink.Write(ctx, processed); err != nil {
		return err
	}

	_ = a.config
	return nil
}

func (a *Agent) Run(ctx context.Context) error {
	runCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(a.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-runCtx.Done():
			if errors.Is(runCtx.Err(), context.Canceled) {
				return nil
			}
			return runCtx.Err()
		case <-ticker.C:
			if err := a.RunOnce(runCtx); err != nil {
					if errors.Is(err, context.Canceled) {
						continue
					}
				log.Printf("collector cycle failed: %v", err)
			}
		}
	}
}
