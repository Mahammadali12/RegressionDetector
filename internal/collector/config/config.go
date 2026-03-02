package config

import "time"

type Config struct {
	PollInterval time.Duration
}

func Default() Config {
	return Config{
		PollInterval: 15 * time.Second,
	}
}
