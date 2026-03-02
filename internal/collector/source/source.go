package source

import (
	"context"
	"time"

	"regressiondetector/internal/collector/types"
)

type Source interface {
	Collect(ctx context.Context) ([]types.Record, error)
}

type StaticSource struct {
	records []types.Record
}

func NewStaticSource(records []types.Record) StaticSource {
	return StaticSource{records: records}
}

func (s StaticSource) Collect(_ context.Context) ([]types.Record, error) {
	if len(s.records) == 0 {
		return []types.Record{
			{
				Metric:    "collector.heartbeat",
				Value:     1,
				Timestamp: time.Now().UTC(),
				Labels: map[string]string{
					"module": "collector",
				},
			},
		}, nil
	}

	output := make([]types.Record, len(s.records))
	copy(output, s.records)
	return output, nil
}
