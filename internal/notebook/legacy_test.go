package notebook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func legacyStore(t *testing.T, path string) *Store {
	t.Helper()
	s, err := Open(path)
	requireNoError(t, err)
	t.Cleanup(func() { s.Close() })
	_, err = s.db.Exec("INSERT INTO projects(id) VALUES ('legacy')")
	requireNoError(t, err)
	_, err = s.db.Exec("INSERT INTO bindings(directory,project_id) VALUES (?,'legacy')", s.directory)
	requireNoError(t, err)
	s.project = "legacy"
	return s
}

func TestLegacyNotebookKeepsRecordsReceiptsAndCursors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fieldnotes.sqlite")
	s := legacyStore(t, path)
	ctx := context.Background()
	put := factInput(ContextID, "context")
	created, err := s.PutFact(ctx, put)
	requireNoError(t, err)
	old, err := s.PutFact(ctx, factInput("old", "old"))
	requireNoError(t, err)
	retire := RetireInput{"retire", "old", old.Revision}
	retired, err := s.RetireFact(ctx, retire)
	requireNoError(t, err)
	appendInput := AppendInput{OperationID: "event", Body: "Legacy outcome", Evidence: []Evidence{}}
	event, err := s.AppendEvent(ctx, appendInput)
	requireNoError(t, err)
	cursor, err := s.encodeCursor("legacy", "read", []Reference{{"fact", ContextID}})
	requireNoError(t, err)
	_, err = s.db.Exec("UPDATE receipts SET request=json_set(request,'$.Request.project_id','legacy')")
	requireNoError(t, err)
	requireNoError(t, s.Close())
	reopened, err := Open(path)
	requireNoError(t, err)
	defer reopened.Close()
	got, err := reopened.Context(ctx, ContextInput{})
	requireNoError(t, err)
	requireEqual(t, got.Context.Revision, created.Revision)
	requireEqual(t, got.Recent[0].Reference, event.Reference)
	requireEqual(t, got.Recent[0].Event.Body, appendInput.Body)
	retried, err := reopened.PutFact(ctx, put)
	requireNoError(t, err)
	requireEqual(t, retried, created)
	retried, err = reopened.RetireFact(ctx, retire)
	requireNoError(t, err)
	requireEqual(t, retried, retired)
	retried, err = reopened.AppendEvent(ctx, appendInput)
	requireNoError(t, err)
	requireEqual(t, retried, event)
	page, err := reopened.Read(ctx, ReadInput{Cursor: cursor})
	requireNoError(t, err)
	requireEqual(t, page.Items[0].Revision, created.Revision)
	put.Fact.Body = "Different input"
	_, err = reopened.PutFact(ctx, put)
	faultCode(t, err, "operation_conflict")
}

func TestLaunchDirectoryDoesNotInheritLegacyAncestor(t *testing.T) {
	root := t.TempDir()
	chdir(t, root)
	path := filepath.Join(root, "fieldnotes.sqlite")
	parent := legacyStore(t, path)
	_, err := parent.PutFact(context.Background(), factInput("private", "create"))
	requireNoError(t, err)
	child := filepath.Join(root, "nested")
	requireNoError(t, os.Mkdir(child, 0700))
	chdir(t, child)
	s, err := Open(path)
	requireNoError(t, err)
	defer s.Close()
	canonical, err := filepath.EvalSymlinks(child)
	requireNoError(t, err)
	requireEqual(t, s.project, canonical)
	read, err := s.Read(context.Background(), ReadInput{References: []Reference{{"fact", "private"}}})
	requireNoError(t, err)
	requireEqual(t, read.Items[0].Status, "missing")
}
