package notebook

import (
	"context"
	"testing"
	"time"
)

func TestEventRetryAndTimestamps(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	in := AppendInput{OperationID: "event", OccurredAt: "2020-01-01T00:00:00Z", Body: "Completed checks.\n", Evidence: []Evidence{{Source: "tests/events"}}}
	first, err := s.AppendEvent(ctx, in)
	requireNoError(t, err)
	retry, err := s.AppendEvent(ctx, in)
	requireNoError(t, err)
	requireEqual(t, retry, first)
	item, err := readOne(ctx, s.db, s.project, first.Reference)
	if err != nil {
		t.Fatal(err)
	}
	if item.Event.Body != in.Body || item.Event.OccurredAt != in.OccurredAt {
		t.Fatalf("altered event: %+v", item)
	}
	recorded, err := time.Parse(time.RFC3339Nano, item.Event.RecordedAt)
	if err != nil || time.Since(recorded) > time.Minute {
		t.Fatalf("recorded_at: %s", item.Event.RecordedAt)
	}
	in.Body = "Changed"
	_, err = s.AppendEvent(ctx, in)
	faultCode(t, err, "operation_conflict")
}

func TestRecentEventsCoverage(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	var last Receipt
	for _, operation := range []string{"a", "b", "c", "d"} {
		var err error
		last, err = s.AppendEvent(ctx, AppendInput{OperationID: operation, Body: operation, Evidence: []Evidence{}})
		if err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.Context(ctx, ContextInput{})
	requireNoError(t, err)
	requireEqual(t, len(got.Recent), 4)
	requireEqual(t, got.HasMore, false)
	requireEqual(t, got.Recent[0].ID, last.ID)
	requireEqual(t, got.Recent[0].Event.Body, "d")
}
