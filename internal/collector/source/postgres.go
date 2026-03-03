package source

import (
	"context"
	"strings"
	"time"

	"github/com/jackc/pgx/v5"
	"regressionDetector/internal/collector/types"
)

type PostgresSource struct{
	connStr string
}

func NewPostgresSource(connStr string) *PostgresSource  {
	return  &PostgresSource{connStr: connStr}
}