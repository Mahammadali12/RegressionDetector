package store

import (
	"context"
	"fmt"
	"regressiondetector/internal/collector/types"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)



type Store struct{
	pool* pgxpool.Pool
}

func NewStore( pool* pgxpool.Pool) *Store {
	return &Store{pool: pool}

}

type Anomaly struct{
	ID int64 `json:"id"`
	QueryID int64 `json:"query_id"`
	WindowStart time.Time `json:"window_start"`
	WindowEnd time.Time `json:"window_end"`
	Metric string `json:"metric"`
	ZScore float64 `json:"z_score"`
	AbsoluteChange float64 `json:"absolute_change"`
	BaselineMean float64 `json:"baseline_mean"`

}

func(s Store) Save(ctx context.Context, records []types.PgStatRow) error{
	// query := fmt.Sprint("")
	length := len(records)
	for i := 0; i < length; i++ {
		_, err := s.pool.Exec(ctx,"INSERT INTO snapshot_records (query_id, query, snapshot_time, calls,total_exec_time, mean_exec_time, stddev_exec_time,rows, shared_blks_read, shared_blks_hit, temp_blks_read) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)ON CONFLICT (query_id, snapshot_time) DO NOTHING",
				records[i].QueryID,records[i].Query,records[i].SnapshotTime,records[i].Calls,records[i].TotalExecTime,records[i].MeanExecTime,records[i].StddevExecTime,records[i].Rows,records[i].SharedBlksRead,records[i].SharedBlksHit,records[i].TempBlksRead)
		if err != nil {
			return  fmt.Errorf("failed to insert snapshot: %w", err)
		}	
	}
	return nil
}

func(s Store) GetAll(ctx context.Context) ([]Anomaly, error){
	rows, err := s.pool.Query(ctx,"SELECT id, query_id, window_start, window_end, metric, z_score, absolute_change, baseline_mean FROM anomaly_records")
	if err != nil {
		return nil, fmt.Errorf("failed to query anomalies: %w", err)
	}
	defer rows.Close()

	var anomalies []Anomaly
	for rows.Next() {
		var anomaly Anomaly
		err := rows.Scan(&anomaly.ID, &anomaly.QueryID, &anomaly.WindowStart, &anomaly.WindowEnd, &anomaly.Metric, &anomaly.ZScore, &anomaly.AbsoluteChange, &anomaly.BaselineMean)
		if err != nil {
			return nil, fmt.Errorf("failed to scan anomaly record: %w", err)
		}
		anomalies = append(anomalies, anomaly)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over anomaly records: %w", err)
	}

	return anomalies, nil
}