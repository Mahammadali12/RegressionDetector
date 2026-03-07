package engine

import (
	"context"
	"errors"
	"fmt"
	"regressiondetector/internal/collector/types"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Detector struct{
	pool* pgxpool.Pool
}

func NewDetector(p* pgxpool.Pool) *Detector {
	return &Detector{pool: p}
}

func(d* Detector) Analyze(ctx context.Context, row types.PgStatRow) error{

	queryId := row.QueryID

	r := d.pool.QueryRow(ctx,"SELECT query_id, mean, stddev, sample_count, last_seen FROM baseline_stats WHERE query_id = $1;",queryId)
	
	var baseline Baseline
	err := r.Scan(
		&baseline.QueryID,
		&baseline.Mean,
		&baseline.Stddev,
		&baseline.SampleCount,
		&baseline.LastSeen,
	)

	if errors.Is(err, pgx.ErrNoRows){ //! new baseline record
		    _, err := d.pool.Exec(ctx,
        `INSERT INTO baseline_stats (query_id, mean, stddev, sample_count, last_seen)
        VALUES ($1, $2, $3, $4, $5)`,
        row.QueryID, row.MeanExecTime, 0, 1, row.SnapshotTime)
    	if err != nil {
    	    return fmt.Errorf("failed to seed baseline: %w", err)
    	}
    	return nil
	}

	if err != nil {
		return fmt.Errorf("failed to load baseline: %w", err)
	}

	if baseline.Stddev == 0 && baseline.SampleCount < 2 {
		// not enough data to compute z-score yet, update baseline and return
		newMean := (baseline.Mean + row.MeanExecTime) / 2
		baseline.Mean = newMean
		baseline.SampleCount += 1
		baseline.LastSeen = row.SnapshotTime

		err = d.updateBaseLine(ctx, &baseline)
		if err != nil {
			return fmt.Errorf("failed to update baseline: %w", err)
		}		

		return nil
	}

	Z := (row.MeanExecTime - baseline.Mean) / baseline.Stddev
	absChange := row.MeanExecTime - baseline.Mean
	percChange := absChange/baseline.Mean * 100

	if( Z >= 3 && absChange > 50 && percChange > 30){
		_, err := d.pool.Exec(ctx,
		`INSERT INTO anomaly_records 
		(query_id, window_start, window_end, metric, z_score, absolute_change, baseline_mean)
		VALUES($1,$2,$3,$4,$5,$6,$7)`,
		row.QueryID,row.SnapshotTime,row.SnapshotTime,"mean_exec_time",Z,absChange,baseline.Mean)
		if err != nil {
			return  fmt.Errorf("error inserting anomaly record: %w",err)
		}
		fmt.Printf("Anomaly was inserted")


	}

	newMean := (baseline.Mean*float64(baseline.SampleCount) + row.MeanExecTime) / (float64(baseline.SampleCount)+1)
	newStdDev := 0.0

	if baseline.SampleCount > 1 {
		newStdDev = ((float64(baseline.SampleCount-1)*baseline.Stddev*baseline.Stddev + (row.MeanExecTime-baseline.Mean)*(row.MeanExecTime-newMean)) / float64(baseline.SampleCount))
	}
	
	baseline.Mean = newMean
	baseline.Stddev = newStdDev
	baseline.SampleCount += 1
	baseline.LastSeen = row.SnapshotTime

	err = d.updateBaseLine(ctx, &baseline)
	if err != nil {
		return fmt.Errorf("failed to update baseline after anomaly: %w", err)
	}

	return nil
}

func(d* Detector) updateBaseLine(ctx context.Context, baseline *Baseline) error {
	_, err := d.pool.Exec(ctx,
	`UPDATE baseline_stats SET mean = $1, stddev = $2, sample_count = $3, last_seen = $4 WHERE query_id = $5`,
	baseline.Mean, baseline.Stddev, baseline.SampleCount, baseline.LastSeen, baseline.QueryID)
	if err != nil {
		return fmt.Errorf("failed to update baseline: %w", err)
	}
	return nil

}