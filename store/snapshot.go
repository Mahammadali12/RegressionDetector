package store

import (
	"context"
	"fmt"
	"regressiondetector/internal/collector/types"

	"github.com/jackc/pgx/v5/pgxpool"
)



type Store struct{
	pool* pgxpool.Pool
}

func NewStore( pool* pgxpool.Pool) *Store {
	return &Store{pool: pool}

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