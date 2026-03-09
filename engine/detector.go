package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"regressiondetector/internal/collector/types"
	"regressiondetector/notify"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Detector struct {
	Pool *pgxpool.Pool
	Notifier notify.Notifier
}

func NewDetector(p *pgxpool.Pool, notifier notify.Notifier) *Detector {
	return &Detector{Pool: p, Notifier: notifier}
}

func (d *Detector) Analyze(ctx context.Context, row types.PgStatRow) error {
	// fmt.Printf("Analyzing query %d with mean exec time %f\n", row.QueryID, row.MeanExecTime)

	queryId := row.QueryID

	r := d.Pool.QueryRow(ctx, "SELECT query_id, mean, stddev, sample_count, last_seen FROM baseline_stats WHERE query_id = $1;", queryId)

	var baseline Baseline
	err := r.Scan(
		&baseline.QueryID,
		&baseline.Mean,
		&baseline.Stddev,
		&baseline.SampleCount,
		&baseline.LastSeen,
	)

	if errors.Is(err, pgx.ErrNoRows) { //! new baseline record
		_, err := d.Pool.Exec(ctx,
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

	oldMean := baseline.Mean
	oldStddev := baseline.Stddev
	oldCount := baseline.SampleCount

	

	newMean := (baseline.Mean*float64(baseline.SampleCount) + row.MeanExecTime) / (float64(baseline.SampleCount) + 1)
	newVariance := 0.0
	newStdDev := 0.0
	if baseline.SampleCount >= 1 {
		newVariance = (float64(baseline.SampleCount-1)*baseline.Stddev*baseline.Stddev +
			(row.MeanExecTime-baseline.Mean)*(row.MeanExecTime-newMean)) / float64(baseline.SampleCount)
	}
	newStdDev = math.Sqrt(newVariance)

	baseline.Mean = newMean
	baseline.Stddev = newStdDev
	baseline.SampleCount += 1
	baseline.LastSeen = row.SnapshotTime


	if err := d.updateBaseLine(ctx, &baseline); err != nil { //! update baseline with new stats
    	return err
	}

	effectiveStddev := oldStddev
	if effectiveStddev < 0.001 {
		effectiveStddev = oldMean * 0.05 // 5% of mean as a minimum stddev to avoid false positives in low variance scenarios
	}



	if oldCount < 3 {
		return nil
	}

	Z := (row.MeanExecTime - oldMean) / effectiveStddev
	absChange := row.MeanExecTime - oldMean
	percChange := absChange / oldMean * 100

	
	
	if Z >= 3 && absChange > 50 && percChange > 30{
		
		hasAnomaly, err := d.hasOpenAnomaly(ctx, row)
		if err != nil {
			return fmt.Errorf("failed to check for open anomaly: %w", err)
		}
		if hasAnomaly {
			fmt.Printf("Anomaly already exists for query %d in the current window, skipping insertion\n", row.QueryID)
			return nil
		}

		_, err = d.Pool.Exec(ctx,
			`INSERT INTO anomaly_records 
		(query_id, window_start, window_end, metric, z_score, absolute_change, baseline_mean)
		VALUES($1,$2,$3,$4,$5,$6,$7)`,
			row.QueryID, row.SnapshotTime, row.SnapshotTime, "mean_exec_time", Z, absChange, oldMean)
		if err != nil {
			return fmt.Errorf("error inserting anomaly record: %w", err)
		}
		fmt.Printf("Anomaly was inserted\n")

		err = d.Notifier.Notify(ctx, row.QueryID, Z, absChange, oldMean)
		if err != nil {
			// return fmt.Errorf("failed to send notification: %w", err)
			log.Printf("Failed to send notification: %v\n", err)
		}

	}
	return nil
}

func (d *Detector) updateBaseLine(ctx context.Context, baseline *Baseline) error {
	_, err := d.Pool.Exec(ctx,
		`UPDATE baseline_stats SET mean = $1, stddev = $2, sample_count = $3, last_seen = $4 WHERE query_id = $5`,
		baseline.Mean, baseline.Stddev, baseline.SampleCount, baseline.LastSeen, baseline.QueryID)
	if err != nil {
		return fmt.Errorf("failed to update baseline: %w", err)
	}
	return nil

}

func (d* Detector) hasOpenAnomaly(ctx context.Context, row types.PgStatRow) (bool, error) {
	//if this query returns a record, it means that there is already an anomaly recorded for this query in the current window, so we should skip it to avoid duplicates
	//SELECT query_id FROM anomaly_records
    //WHERE window_start = window_end AND query_id = $1;
    var existing int64
    err := d.Pool.QueryRow(ctx, `
        SELECT query_id FROM anomaly_records
        WHERE window_start = window_end AND query_id = $1
        LIMIT 1`, row.QueryID).Scan(&existing)
    
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil  // no open anomaly
    }
    if err != nil {
        return false, fmt.Errorf("failed to check open anomaly: %w", err)
    }
    return true, nil  // record found, anomaly is open


}
