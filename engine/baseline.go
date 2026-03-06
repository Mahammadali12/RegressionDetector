package engine

import "time"


type Baseline struct{
	QueryID int64
	Mean float64
	Stddev float64
	SampleCount int64
	LastSeen time.Time
}