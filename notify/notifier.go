package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)


type Notifier interface {
	Notify(ctx context.Context, queryID int64, zScore float64, absChange float64, baselineMean float64) error
}

type SlackNotifier struct {
	// Add fields for Slack configuration, e.g., webhook URL
	webhookURL string
	client    *http.Client
}

func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{webhookURL: webhookURL, client: &http.Client{}}
}

func (s *SlackNotifier) Notify(ctx context.Context, queryID int64, zScore float64, absChange float64, baselineMean float64) error {
	// Implement the logic to send a notification to Slack using the webhook URL
	// You can use an HTTP client to post a message to the Slack channel
	// The message can include details about the anomaly, such as the query ID, z-score, absolute change, and baseline mean

		text := fmt.Sprintf(
		"🚨 *Query regression detected*\n"+
			"Query ID: `%d`\n"+
			"Z-score: `%.1f`\n"+
			"Change: `+%.2f ms` above baseline\n"+
			"Baseline mean: `%.3f ms`\n"+
			"_Investigate the most recent deploy or migration as a likely correlating event._",
		queryID, zScore, absChange, baselineMean,
	)

	payload, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response from Slack: %s", resp.Status)
	}

	return nil
}
