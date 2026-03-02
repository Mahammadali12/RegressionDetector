package types

import "time"

type Record struct {
	Metric    string
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
}
