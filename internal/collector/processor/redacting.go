package processor

import (
	"context"
	"regressiondetector/internal/collector/types"
)


type RedactingProcessor struct{}

func(RedactingProcessor) Process(_ context.Context, records []types.PgStatRow) ([]types.PgStatRow,error){
	length := len(records)
	for i := 0; i < length; i++ { //! makes a copy first
		records[i].Query = nil
	}
	return  records, nil;
}