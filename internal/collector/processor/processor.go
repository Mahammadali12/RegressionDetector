package processor

import (
	"context"

	"regressiondetector/internal/collector/types"
)

type Processor interface {
	Process(ctx context.Context, records []types.Record) ([]types.Record, error)
}

type PassThroughProcessor struct{}

func (PassThroughProcessor) Process(_ context.Context, records []types.Record) ([]types.Record, error) {
	output := make([]types.Record, len(records))
	copy(output, records)
	return output, nil
}
