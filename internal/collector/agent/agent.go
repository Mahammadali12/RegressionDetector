package agent

import (
	"context"

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
