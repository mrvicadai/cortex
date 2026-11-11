package notebook

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type Event struct {
	ID         string     `json:"id"`
	RecordedAt string     `json:"recorded_at"`
	OccurredAt string     `json:"occurred_at,omitempty"`
	Body       string     `json:"body"`
	Evidence   []Evidence `json:"evidence"`
}

type AppendInput struct {
	OperationID string     `json:"operation_id" jsonschema:"Stable retry key; reuse with the identical request after response loss."`
	OccurredAt  string     `json:"occurred_at,omitempty" jsonschema:"Optional caller-declared RFC3339 event time, separate from server insertion time."`
	Body        string     `json:"body" jsonschema:"Chronological outcome or milestone description (e.g. 'Merged PR #42', 'Investigated memory leak', 'Passed test suite')."`
	Evidence    []Evidence `json:"evidence" jsonschema:"Structured evidence pointers (e.g. git commit hash, test logs, PR link)."`
}

func (s *Store) AppendEvent(ctx context.Context, in AppendInput) (Receipt, error) {
	if err := validateBody(in.Body); err != nil {
		return Receipt{}, err
	}
	if err := validateEvidence(in.Evidence); err != nil {
		return Receipt{}, err
	}
	if in.OccurredAt != "" {
		if _, err := time.Parse(time.RFC3339Nano, in.OccurredAt); err != nil {
			return Receipt{}, invalid("occurred_at: require RFC3339")
		}
	}
	return s.mutate(ctx, s.project, in.OperationID, "event_append", in, func(tx *sql.Tx) (Receipt, error) {
		event := Event{ID: token(), RecordedAt: time.Now().UTC().Format(time.RFC3339Nano), OccurredAt: in.OccurredAt, Body: in.Body, Evidence: in.Evidence}
		document, err := encodeDocument(event)
		if err != nil {
			return Receipt{}, err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO records(project_id,kind,id,revision,document,search_text) VALUES (?,'event',?,'',?,?)", s.project, event.ID, document, eventSearch(event))
		return Receipt{Reference: Reference{Kind: "event", ID: event.ID}, RecordedAt: event.RecordedAt}, err
	})
}

func eventSearch(event Event) string {
	parts := []string{event.Body, event.RecordedAt, event.OccurredAt}
	for _, item := range event.Evidence {
		parts = append(parts, item.Source)
	}
	return strings.ToLower(strings.Join(parts, "\n"))
}
