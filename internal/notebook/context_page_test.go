package notebook

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func TestContextCursorPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fieldnotes.sqlite")
	s, err := Open(path)
	requireNoError(t, err)
	t.Cleanup(func() { s.Close() })
	ctx := context.Background()
	for i := 0; i < 13; i++ {
		_, err := s.AppendEvent(ctx, AppendInput{OperationID: fmt.Sprint(i), Body: fmt.Sprint(i), Evidence: []Evidence{}})
		requireNoError(t, err)
	}
	first, err := s.Context(ctx, ContextInput{})
	requireNoError(t, err)
	requireEqual(t, len(first.Recent), 10)
	requireEqual(t, first.HasMore, true)
	requireNoError(t, s.Close())
	reopened, err := Open(path)
	requireNoError(t, err)
	defer reopened.Close()
	second, err := reopened.Context(ctx, ContextInput{Cursor: first.NextCursor})
	requireNoError(t, err)
	requireEqual(t, len(second.Recent), 3)
	requireEqual(t, second.HasMore, false)
	requireEqual(t, second.NextCursor, "")
	requireEqual(t, second.Recent[0].Event.Body, "2")
	requireEqual(t, second.Recent[2].Event.Body, "0")
	assertContextCursorRejections(t, reopened, first.NextCursor)
}

func assertContextCursorRejections(t *testing.T, s *Store, cursor string) {
	t.Helper()
	ctx := context.Background()
	_, err := s.Context(ctx, ContextInput{Cursor: cursor + "x"})
	faultCode(t, err, "invalid_input")
	_, err = s.Read(ctx, ReadInput{Cursor: cursor})
	faultCode(t, err, "invalid_input")
	_, err = s.Find(ctx, FindInput{Cursor: cursor})
	faultCode(t, err, "invalid_input")
	wrongOperation, err := s.encodeCursor(s.project, "find", findState{Kind: "event", Before: 1})
	requireNoError(t, err)
	_, err = s.Context(ctx, ContextInput{Cursor: wrongOperation})
	faultCode(t, err, "invalid_input")
	wrongProject, err := s.encodeCursor("other-project", "context", contextState{Before: 1})
	requireNoError(t, err)
	_, err = s.Context(ctx, ContextInput{Cursor: wrongProject})
	faultCode(t, err, "invalid_input")
	invalidPosition, err := s.encodeCursor(s.project, "context", contextState{})
	requireNoError(t, err)
	_, err = s.Context(ctx, ContextInput{Cursor: invalidPosition})
	faultCode(t, err, "invalid_input")
}
