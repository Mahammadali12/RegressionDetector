package source

import (
	"context"
	"regressiondetector/internal/collector/types"
)

type Source interface {
	Collect(ctx context.Context) ([]types.PgStatRow, error)
}
