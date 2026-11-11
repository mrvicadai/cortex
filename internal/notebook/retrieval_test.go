package notebook

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mrvicadai/cortex/internal/wire"
)

func TestWholeRecordReadPagination(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	body := strings.Repeat("\n\"🔐", 680)
	refs := []Reference{}
	for i := 0; i < 8; i++ {
		id := fmt.Sprintf("fact-%d", i)
		in := factInput(id, id)
		in.Fact.Body = body
		if _, err := s.PutFact(ctx, in); err != nil {
			t.Fatal(err)
		}
		refs = append(refs, Reference{"fact", id})
	}
	refs = append(refs, refs[0], Reference{"fact", "missing"})
	page, err := s.Read(ctx, ReadInput{References: refs})
	requireNoError(t, err)
	if !page.HasMore {
		t.Fatal("fixture did not force pagination")
	}
	items := []ReadItem{}
	for {
		if !wire.Fits(page) {
			t.Fatal("oversized page")
		}
		items = append(items, page.Items...)
		if !page.HasMore {
			break
		}
		page, err = s.Read(ctx, ReadInput{Cursor: page.NextCursor})
		requireNoError(t, err)
	}
	assertWholeItems(t, items, refs, body)
}

func assertWholeItems(t *testing.T, items []ReadItem, refs []Reference, body string) {
	t.Helper()
	if len(items) != len(refs) {
		t.Fatalf("items: %d", len(items))
	}
	for i, item := range items {
		requireEqual(t, item.Reference, refs[i])
		if i == len(refs)-1 {
			if item.Status != "missing" {
				t.Fatalf("missing: %+v", item)
			}
			continue
		}
		if item.Status != "complete" || item.Fact.Body != body {
			t.Fatalf("partial or altered body at %d", i)
		}
	}
}

func TestFindLiteralQueriesAndTopics(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	in := factInput("unicode", "unicode")
	in.Fact.Body = "ÄPFEL sqlite locking literal*"
	if _, err := s.PutFact(ctx, in); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AppendEvent(ctx, AppendInput{OperationID: "event", Body: "SQLite outcome", Evidence: []Evidence{}}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		query, topic string
		want         int
	}{
		{"äpfel LOCK", "", 1}, {"sqlite", "", 2}, {"sqlite", "storage", 1},
		{"sqlite missing", "", 0}, {"literal*", "", 1}, {"sqlite OR nonexistent", "", 0},
	} {
		page, err := s.Find(ctx, FindInput{Query: tc.query, Topic: tc.topic})
		if err != nil || len(page.Items) != tc.want {
			t.Fatalf("%q: %+v %v", tc.query, page, err)
		}
	}
}
