package notebook

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func factInput(id, operation string) PutInput {
	return PutInput{OperationID: operation, Fact: Fact{ID: id, Title: "Storage", Body: "Use reliable local locking.\n", Topics: []string{"storage"}, Evidence: []Evidence{{Source: "README.md", VerifiedAt: "2026-11-06"}}}}
}

func TestFactCASAndRetry(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	in := factInput("locking", "create")
	first, err := s.PutFact(ctx, in)
	requireNoError(t, err)
	retry, err := s.PutFact(ctx, in)
	requireNoError(t, err)
	requireEqual(t, retry, first)
	in.Fact.Body = "Changed"
	_, err = s.PutFact(ctx, in)
	faultCode(t, err, "operation_conflict")
	in.OperationID = "stale"
	_, err = s.PutFact(ctx, in)
	faultCode(t, err, "revision_conflict")
	in.ExpectedRevision = first.Revision
	updated, err := s.PutFact(ctx, in)
	if err != nil || updated.Revision == first.Revision {
		t.Fatalf("update: %+v %v", updated, err)
	}
	retired, err := s.RetireFact(ctx, RetireInput{"retire", "locking", updated.Revision})
	if err != nil || !retired.Retired {
		t.Fatalf("retire: %+v %v", retired, err)
	}
	in.OperationID = "resurrect"
	in.ExpectedRevision = ""
	_, err = s.PutFact(ctx, in)
	faultCode(t, err, "revision_conflict")
	in.ExpectedRevision = retired.Revision
	if _, err := s.PutFact(ctx, in); err != nil {
		t.Fatal(err)
	}
}

func TestFactValidation(t *testing.T) {
	s := openTest(t)
	for name, change := range map[string]func(*PutInput){
		"body":     func(in *PutInput) { in.Fact.Body = strings.Repeat("x", 4097) },
		"escaping": func(in *PutInput) { in.Fact.Body = strings.Repeat("\x00", 4000) },
		"utf8":     func(in *PutInput) { in.Fact.Body = string([]byte{0xff}) },
		"date":     func(in *PutInput) { in.Fact.Evidence[0].VerifiedAt = "2026-02-30" },
		"topics":   func(in *PutInput) { in.Fact.Topics = []string{"same", "same"} },
	} {
		t.Run(name, func(t *testing.T) {
			in := factInput(name, name)
			change(&in)
			_, err := s.PutFact(context.Background(), in)
			faultCode(t, err, "invalid_input")
		})
	}
}

func TestConcurrentFactCreation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shared.sqlite")
	a, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	var wg sync.WaitGroup
	errors := make(chan error, 2)
	for i, store := range []*Store{a, b} {
		wg.Add(1)
		go func(s *Store, operation string) {
			defer wg.Done()
			_, err := s.PutFact(context.Background(), factInput("same", operation))
			errors <- err
		}(store, string(rune('a'+i)))
	}
	wg.Wait()
	close(errors)
	successes := 0
	for err := range errors {
		if err == nil {
			successes++
		} else {
			faultCode(t, err, "revision_conflict")
		}
	}
	if successes != 1 {
		t.Fatalf("successful creates: %d", successes)
	}
}
