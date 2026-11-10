package notebook

import (
	"context"
	"testing"
)

func TestRetireRetryAndContextProtection(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	created, err := s.PutFact(ctx, factInput("old", "create-old"))
	requireNoError(t, err)
	in := RetireInput{"retire-old", "old", created.Revision}
	retired, err := s.RetireFact(ctx, in)
	requireNoError(t, err)
	if retry, err := s.RetireFact(ctx, in); err != nil || retry != retired {
		t.Fatalf("retire retry: %+v %v", retry, err)
	}
	item, err := readOne(ctx, s.db, s.project, Reference{"fact", "old"})
	if err != nil || item.Status != "retired" || item.Fact != nil || item.Revision != retired.Revision {
		t.Fatalf("tombstone: %+v %v", item, err)
	}
	_, err = s.RetireFact(ctx, RetireInput{"bad", ContextID, retired.Revision})
	faultCode(t, err, "context_required")
	_, err = s.PutFact(ctx, factInput("another", "retire-old"))
	faultCode(t, err, "operation_conflict")
}

func TestMutationRollbackAndProjectIsolation(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	first, err := s.PutFact(ctx, factInput("same", "key"))
	requireNoError(t, err)
	in := factInput("same", "key")
	other := openTest(t)
	second, err := other.PutFact(ctx, in)
	requireNoError(t, err)
	if second.Revision == first.Revision {
		t.Fatal("projects shared a revision")
	}
	if _, err := s.db.Exec("CREATE TRIGGER fail_receipt BEFORE INSERT ON receipts BEGIN SELECT RAISE(ABORT,'test failure'); END"); err != nil {
		t.Fatal(err)
	}
	_, err = s.PutFact(ctx, factInput("rollback", "failed"))
	if err == nil {
		t.Fatal("injected receipt failure did not fail")
	}
	item, err := readOne(ctx, s.db, s.project, Reference{"fact", "rollback"})
	if err != nil || item.Status != "missing" {
		t.Fatalf("uncommitted write escaped: %+v %v", item, err)
	}
}
