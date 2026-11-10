package notebook

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func TestFindCursorPersistenceAndIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fieldnotes.sqlite")
	s, err := Open(path)
	requireNoError(t, err)
	ctx := context.Background()
	for i := 0; i < 13; i++ {
		id := fmt.Sprintf("f%02d", i)
		if _, err := s.PutFact(ctx, factInput(id, id)); err != nil {
			t.Fatal(err)
		}
	}
	first, err := s.Find(ctx, FindInput{Kind: "fact", Query: "local"})
	requireNoError(t, err)
	requireEqual(t, first.HasMore, true)
	requireEqual(t, len(first.Items), 10)
	requireNoError(t, s.Close())
	reopened, err := Open(path)
	requireNoError(t, err)
	defer reopened.Close()
	second, err := reopened.Find(ctx, FindInput{Cursor: first.NextCursor})
	requireNoError(t, err)
	requireEqual(t, second.HasMore, false)
	requireEqual(t, len(second.Items), 3)
	requireEqual(t, first.Items[9].ID, "f03")
	requireEqual(t, second.Items[0].ID, "f02")
	assertCursorRejections(t, reopened, first.NextCursor)
}

func assertCursorRejections(t *testing.T, s *Store, cursor string) {
	t.Helper()
	ctx := context.Background()
	_, err := s.Find(ctx, FindInput{Cursor: cursor + "x"})
	faultCode(t, err, "invalid_input")
	_, err = s.Find(ctx, FindInput{Cursor: cursor, Query: "override"})
	faultCode(t, err, "invalid_input")
	_, err = s.Read(ctx, ReadInput{Cursor: cursor})
	faultCode(t, err, "invalid_input")
	other := openTest(t)
	_, err = other.Find(ctx, FindInput{Cursor: cursor})
	faultCode(t, err, "invalid_input")
}

func TestReadCursorPersistence(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	in := factInput("changing", "create")
	created, err := s.PutFact(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	cursor, err := s.encodeCursor(s.project, "read", []Reference{{"fact", "changing"}})
	if err != nil {
		t.Fatal(err)
	}
	in.OperationID = "replace"
	in.ExpectedRevision = created.Revision
	in.Fact.Body = "The new complete body."
	if _, err := s.PutFact(ctx, in); err != nil {
		t.Fatal(err)
	}
	page, err := s.Read(ctx, ReadInput{Cursor: cursor})
	if err != nil || page.Items[0].Fact.Body != in.Fact.Body {
		t.Fatalf("current full read: %+v %v", page, err)
	}
}
