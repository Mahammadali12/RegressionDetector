package config

import (
	"os"
	"time"
	"fmt"
	// "golang.org/x/tools/go/cfg"
)

type Config struct {
	PollInterval time.Duration
	ConnStr string
	APIToken string
	BackendURL string
	SlackWebhookURL string
}

func Load() (Config, error) {

	cfg := Config{
		PollInterval: 10 * time.Second,
		ConnStr: os.Getenv("DRIFT_DETECTOR_CONN_STR"),
		APIToken: os.Getenv("DRIFT_DETECTOR_API_TOKEN"),
		BackendURL: os.Getenv("DRIFT_DETECTOR_BACKEND_URL"),
		SlackWebhookURL: os.Getenv("DRIFT_DETECTOR_SLACK_WEBHOOK_URL"),
	}

	if cfg.ConnStr == "" {
    	return Config{}, fmt.Errorf("DRIFT_DETECTOR_CONN_STR is required")
	}
	if cfg.APIToken == "" {
	    return Config{}, fmt.Errorf("DRIFT_DETECTOR_API_TOKEN is required")
	}
	if cfg.BackendURL == "" {
	    return Config{}, fmt.Errorf("DRIFT_DETECTOR_BACKEND_URL is required")
	}
	if cfg.SlackWebhookURL == "" {
	    return Config{}, fmt.Errorf("DRIFT_DETECTOR_SLACK_WEBHOOK_URL is required")
	}

	return cfg,nil
}
