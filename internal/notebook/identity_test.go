package notebook

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "fieldnotes.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func faultCode(t *testing.T, err error, want string) {
	t.Helper()
	var fault *Fault
	if !errors.As(err, &fault) || fault.Code != want {
		t.Fatalf("want %s, got %v", want, err)
	}
}

func chdir(t *testing.T, directory string) {
	t.Helper()
	previous, err := os.Getwd()
	requireNoError(t, err)
	requireNoError(t, os.Chdir(directory))
	t.Cleanup(func() { requireNoError(t, os.Chdir(previous)) })
}

func TestIdentityIsLaunchDirectory(t *testing.T) {
	directory := t.TempDir()
	chdir(t, directory)
	s := openTest(t)
	canonical, err := filepath.EvalSymlinks(directory)
	requireNoError(t, err)
	requireEqual(t, s.project, canonical)
	chdir(t, t.TempDir())
	got, err := s.Context(context.Background(), ContextInput{})
	requireNoError(t, err)
	requireEqual(t, got.Directory, canonical)
	requireEqual(t, got.Context, nil)
	_, err = s.PutFact(context.Background(), factInput("fixed", "create"))
	requireNoError(t, err)
	item, err := readOne(context.Background(), s.db, canonical, Reference{"fact", "fixed"})
	requireNoError(t, err)
	requireEqual(t, item.Status, "complete")
}

func TestDirectoryAliases(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "project")
	requireNoError(t, os.Mkdir(directory, 0700))
	alias := filepath.Join(root, "alias")
	requireNoError(t, os.Symlink(directory, alias))
	chdir(t, alias)
	s := openTest(t)
	canonical, err := filepath.EvalSymlinks(directory)
	requireNoError(t, err)
	requireEqual(t, s.project, canonical)
}
