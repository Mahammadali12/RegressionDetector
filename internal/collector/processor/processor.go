package processor

import (
	"context"

	"regressiondetector/internal/collector/types"
)

type Processor interface {
	Process(ctx context.Context, records []types.PgStatRow) ([]types.PgStatRow, error)
}

type PassThroughProcessor struct{}

func (PassThroughProcessor) Process(_ context.Context, records []types.PgStatRow) ([]types.PgStatRow, error) {
	output := make([]types.PgStatRow, len(records))
	copy(output, records)
	return output, nil
}
