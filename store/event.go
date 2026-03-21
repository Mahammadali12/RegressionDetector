package store

import (
	"context"
	"fmt"
	"time"
)

type EventRecord struct {
	ReceivedAt    time.Time `json:"received_at"`
	Repo          string    `json:"repo"`
	Branch        string    `json:"branch"`
	CommitSHA     string    `json:"commit_sha"`
	CommitMessage *string   `json:"commit_message"`
	Pusher        string    `json:"pusher"`
}


func (s Store) SaveEvent(ctx context.Context, eventRecord EventRecord) error {

	_, err := s.pool.Exec(ctx, `INSERT INTO event_records (received_at,repo,branch,commit_sha,commit_message,pusher)
					VALUES($1,$2,$3,$4,$5,$6) `,eventRecord.ReceivedAt,eventRecord.Repo,eventRecord.Branch,eventRecord.CommitSHA,eventRecord.CommitMessage,eventRecord.Pusher)
	if err != nil {
		return fmt.Errorf("failed to insert event record: %w", err)
	}
	return nil
}