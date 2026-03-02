package sink

import (
	"context"
	"log"

	"regressiondetector/internal/collector/types"
)

type Sink interface {
	Write(ctx context.Context, records []types.Record) error
}

type LoggingSink struct{}

func (LoggingSink) Write(_ context.Context, records []types.Record) error {
	log.Printf("collector sink received %d records", len(records))
	return nil
}
