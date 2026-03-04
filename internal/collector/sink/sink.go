package sink

import (
	"context"
	"log"

	"regressiondetector/internal/collector/types"
)

type Sink interface {
	Write(ctx context.Context, records []types.PgStatRow) error
}

type LoggingSink struct{}

func (LoggingSink) Write(_ context.Context, records []types.PgStatRow) error {

	limit := 3
	if len(records) < limit {
	    limit = len(records)
	}

	

	for i := 0; i < limit; i++ {
		
		queryText := "[redacted]"
		if records[i].Query != nil {
		    queryText = *records[i].Query;
		}
		log.Printf("MeanExecTime: %f QueryID: %d Query: %s  Time: %s",records[i].MeanExecTime, records[i].QueryID, queryText, records[i].SnapshotTime)
	}

	log.Printf("collector sink received %d records", len(records))
	return nil
}
