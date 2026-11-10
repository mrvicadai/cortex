package notebook

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
)

func TestNotebookInitializesOnOpen(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	page, err := s.Find(ctx, FindInput{})
	requireNoError(t, err)
	requireEqual(t, len(page.Items), 0)
	read, err := s.Read(ctx, ReadInput{References: []Reference{{Kind: "fact", ID: "absent"}}})
	requireNoError(t, err)
	requireEqual(t, read.Items[0].Status, "missing")
	_, err = s.PutFact(ctx, factInput("fact", "first-write"))
	requireNoError(t, err)
	_, err = s.AppendEvent(ctx, AppendInput{OperationID: "first-append", Body: "Outcome", Evidence: []Evidence{}})
	requireNoError(t, err)
	var count int
	requireNoError(t, s.db.QueryRow("SELECT count(*) FROM projects").Scan(&count))
	requireEqual(t, count, 1)
}

func TestConcurrentProjectInitialization(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fieldnotes.sqlite")
	stores := make([]*Store, 2)
	errors := make([]error, 2)
	var wg sync.WaitGroup
	for i := range stores {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			stores[i], errors[i] = Open(path)
		}(i)
	}
	wg.Wait()
	for i, err := range errors {
		requireNoError(t, err)
		defer stores[i].Close()
	}
	requireEqual(t, stores[0].project, stores[1].project)
	var count int
	requireNoError(t, stores[0].db.QueryRow("SELECT count(*) FROM projects").Scan(&count))
	requireEqual(t, count, 1)
}

func TestContextDirectoryPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fieldnotes.sqlite")
	s, err := Open(path)
	requireNoError(t, err)
	first, err := s.Context(context.Background(), ContextInput{})
	requireNoError(t, err)
	if first.Directory == "" || first.Context != nil {
		t.Fatalf("new context: %+v", first)
	}
	requireNoError(t, s.Close())
	reopened, err := Open(path)
	requireNoError(t, err)
	defer reopened.Close()
	second, err := reopened.Context(context.Background(), ContextInput{})
	requireNoError(t, err)
	requireEqual(t, second.Directory, first.Directory)
}
