package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"regressiondetector/notify"
)

func TestSlackNotifierNotifyPostsWebhookPayload(t *testing.T) {
	var gotMethod string
	var gotContentType string
	var gotBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := &notify.SlackNotifier{
		WebhookURL: server.URL,
		Client:     server.Client(),
	}

	err := n.Notify(context.Background(), 12345, 4.2, 190, 10)
	if err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("expected method %q, got %q", http.MethodPost, gotMethod)
	}

	if gotContentType != "application/json" {
		t.Fatalf("expected content-type application/json, got %q", gotContentType)
	}

	text := gotBody["text"]
	if text == "" {
		t.Fatal("expected Slack payload to include non-empty text")
	}

	if !strings.Contains(text, "Query ID: `12345`") {
		t.Fatalf("expected payload to contain query id, got %q", text)
	}

	if !strings.Contains(text, "Z-score: `4.2`") {
		t.Fatalf("expected payload to contain z-score, got %q", text)
	}
}
