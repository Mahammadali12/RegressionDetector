package types

import "time"

type PgStatRow struct {
    QueryID         int64
    SnapshotTime    time.Time
    Query           *string
    Calls           int64
    TotalExecTime   float64
    MeanExecTime    float64
    StddevExecTime  float64
    Rows            int64
    SharedBlksHit   int64
    SharedBlksRead  int64
    TempBlksRead    int64
}