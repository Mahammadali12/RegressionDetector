package source

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"regressiondetector/internal/collector/types"
)

type PostgresSource struct {
	connStr string
}

func NewPostgresSource(connStr string) *PostgresSource {
	return &PostgresSource{connStr: connStr}
}

func (s *PostgresSource) Collect(ctx context.Context) ([]types.PgStatRow, error) {
	conn, err := pgx.Connect(ctx, s.connStr)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
		SELECT queryid, query, calls, total_exec_time,
		       mean_exec_time, stddev_exec_time, rows,
		       shared_blks_hit, shared_blks_read, temp_blks_read
		FROM pg_stat_statements
		ORDER BY total_exec_time DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []types.PgStatRow
	for rows.Next() {
		var r types.PgStatRow
		r.SnapshotTime = time.Now().UTC()
		err := rows.Scan(
			&r.QueryID,
			&r.Query,
			&r.Calls,
			&r.TotalExecTime,
			&r.MeanExecTime,
			&r.StddevExecTime,
			&r.Rows,
			&r.SharedBlksHit,
			&r.SharedBlksRead,
			&r.TempBlksRead,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}	