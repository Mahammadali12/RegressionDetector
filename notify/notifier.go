package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)


type Notifier interface {
	Notify(ctx context.Context, queryID int64, zScore float64, absChange float64, baselineMean float64) error
}

type SlackNotifier struct {
	// Add fields for Slack configuration, e.g., webhook URL
	WebhookURL string
	Client    *http.Client
}

func NewSlackNotifier(webhookURL string) *SlackNotifier {
	if webhookURL == "" {
		// If no webhook URL is provided, return warn and return a no-op notifier
		log.Println("Warning: No Slack webhook URL provided. Notifications will be disabled.")
		return &SlackNotifier{WebhookURL: "", Client: &http.Client{}}	
		
	}
	return &SlackNotifier{WebhookURL: webhookURL, Client: &http.Client{}}
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

	req, err := http.NewRequestWithContext(ctx, "POST", s.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response from Slack: %s", resp.Status)
	}

	return nil
}
