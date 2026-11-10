package notebook

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestRejectUnrelatedDatabaseWithoutChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unrelated.sqlite")
	db, err := sql.Open("sqlite", path)
	requireNoError(t, err)
	if _, err := db.Exec("CREATE TABLE unrelated (value TEXT); INSERT INTO unrelated VALUES ('keep');"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err == nil {
		s.Close()
		t.Fatal("accepted an unrelated database")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("rejected database was changed")
	}
}

func TestDatabasePermissionsAndSpecialPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes?#%.sqlite")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("permissions: %v", info.Mode())
	}
}

func TestConcurrentStartup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new.sqlite")
	var wg sync.WaitGroup
	errors := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := Open(path)
			if err == nil {
				err = s.Close()
			}
			errors <- err
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
}
