package notebook

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestTodoLifecycleAndContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cortex.sqlite")
	s, err := Open(path)
	requireNoError(t, err)
	t.Cleanup(func() { s.Close() })
	ctx := context.Background()

	first, err := s.AddTodo(ctx, AddTodoInput{
		OperationID: "todo-1",
		Title:       "Fix nil pointer in auth middleware",
		Notes:       "internal/auth/middleware.go",
	})
	requireNoError(t, err)
	requireEqual(t, first.Kind, "todo")
	if first.ID == "" {
		t.Fatalf("expected non-empty todo ID")
	}

	second, err := s.AddTodo(ctx, AddTodoInput{
		OperationID: "todo-2",
		Title:       "Add retry test for webhook handler",
	})
	requireNoError(t, err)

	// Open list preserves FIFO order
	openList, err := s.ListTodos(ctx, ListTodosInput{})
	requireNoError(t, err)
	requireEqual(t, openList.OpenCount, 2)
	requireEqual(t, len(openList.Items), 2)
	requireEqual(t, openList.Items[0].ID, first.ID)
	requireEqual(t, openList.Items[1].ID, second.ID)

	// Context surfaces open todos on initial page
	ctxRes, err := s.Context(ctx, ContextInput{})
	requireNoError(t, err)
	requireEqual(t, ctxRes.OpenTodoCount, 2)
	requireEqual(t, len(ctxRes.OpenTodos), 2)
	requireEqual(t, ctxRes.OpenTodos[0].Title, "Fix nil pointer in auth middleware")

	// Mark first done
	doneReceipt, err := s.DoneTodo(ctx, DoneTodoInput{OperationID: "done-1", ID: first.ID})
	requireNoError(t, err)
	requireEqual(t, doneReceipt.Kind, "todo")
	requireEqual(t, doneReceipt.Retired, true)

	// Open list now has only second item
	openAfter, err := s.ListTodos(ctx, ListTodosInput{Status: "open"})
	requireNoError(t, err)
	requireEqual(t, openAfter.OpenCount, 1)
	requireEqual(t, len(openAfter.Items), 1)
	requireEqual(t, openAfter.Items[0].ID, second.ID)

	// Done list has first item
	doneList, err := s.ListTodos(ctx, ListTodosInput{Status: "done"})
	requireNoError(t, err)
	requireEqual(t, doneList.OpenCount, 1)
	requireEqual(t, len(doneList.Items), 1)
	requireEqual(t, doneList.Items[0].ID, first.ID)

	// All list has both
	allList, err := s.ListTodos(ctx, ListTodosInput{Status: "all"})
	requireNoError(t, err)
	requireEqual(t, len(allList.Items), 2)

	// Context reflects remaining open item and auto-appended completion event
	ctxAfter, err := s.Context(ctx, ContextInput{})
	requireNoError(t, err)
	requireEqual(t, ctxAfter.OpenTodoCount, 1)
	requireEqual(t, len(ctxAfter.OpenTodos), 1)
	requireEqual(t, ctxAfter.OpenTodos[0].ID, second.ID)
	requireEqual(t, len(ctxAfter.Recent), 1)
	requireEqual(t, ctxAfter.Recent[0].Event.Body, "Completed todo: Fix nil pointer in auth middleware")

	// Completing an already-done todo with a new operation_id does not duplicate the event
	_, err = s.DoneTodo(ctx, DoneTodoInput{OperationID: "done-1-again", ID: first.ID})
	requireNoError(t, err)
	ctxNoDup, err := s.Context(ctx, ContextInput{})
	requireNoError(t, err)
	requireEqual(t, len(ctxNoDup.Recent), 1)

	// Completing second item with SkipEvent: true does not append an event
	_, err = s.DoneTodo(ctx, DoneTodoInput{OperationID: "done-2-skip", ID: second.ID, SkipEvent: true})
	requireNoError(t, err)
	ctxSkipped, err := s.Context(ctx, ContextInput{})
	requireNoError(t, err)
	requireEqual(t, ctxSkipped.OpenTodoCount, 0)
	requireEqual(t, len(ctxSkipped.Recent), 1)

	// Completing with custom outcome and evidence appends them to the event
	third, err := s.AddTodo(ctx, AddTodoInput{OperationID: "todo-3", Title: "Investigate timeout"})
	requireNoError(t, err)
	_, err = s.DoneTodo(ctx, DoneTodoInput{
		OperationID: "done-3",
		ID:          third.ID,
		Outcome:     "Fixed HTTP client timeout",
		Evidence:    []Evidence{{Source: "git:abc123", VerifiedAt: "2026-11-23"}},
	})
	requireNoError(t, err)
	ctxCustom, err := s.Context(ctx, ContextInput{})
	requireNoError(t, err)
	requireEqual(t, len(ctxCustom.Recent), 2)
	requireEqual(t, ctxCustom.Recent[0].Event.Body, "Fixed HTTP client timeout")
	requireEqual(t, len(ctxCustom.Recent[0].Event.Evidence), 1)
	requireEqual(t, ctxCustom.Recent[0].Event.Evidence[0].Source, "git:abc123")
}

func TestTodoIdempotencyAndValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cortex.sqlite")
	s, err := Open(path)
	requireNoError(t, err)
	t.Cleanup(func() { s.Close() })
	ctx := context.Background()

	// Replay with same operation_id returns identical receipt
	in := AddTodoInput{OperationID: "op-add", Title: "Check metrics"}
	r1, err := s.AddTodo(ctx, in)
	requireNoError(t, err)
	r2, err := s.AddTodo(ctx, in)
	requireNoError(t, err)
	requireEqual(t, r1.ID, r2.ID)

	// Reused operation_id with different request triggers operation_conflict
	_, err = s.AddTodo(ctx, AddTodoInput{OperationID: "op-add", Title: "Different title"})
	faultCode(t, err, "operation_conflict")

	// Invalid inputs
	_, err = s.AddTodo(ctx, AddTodoInput{OperationID: "bad-1", Title: ""})
	faultCode(t, err, "invalid_input")
	_, err = s.AddTodo(ctx, AddTodoInput{OperationID: "bad-2", Title: "  padded  "})
	faultCode(t, err, "invalid_input")
	_, err = s.AddTodo(ctx, AddTodoInput{OperationID: "bad-3", Title: "ok", Notes: strings.Repeat("a", 2049)})
	faultCode(t, err, "invalid_input")
	_, err = s.ListTodos(ctx, ListTodosInput{Status: "bogus"})
	faultCode(t, err, "invalid_input")
	_, err = s.DoneTodo(ctx, DoneTodoInput{OperationID: "bad-4", ID: ""})
	faultCode(t, err, "invalid_input")
	_, err = s.DoneTodo(ctx, DoneTodoInput{OperationID: "missing-1", ID: "nonexistent"})
	faultCode(t, err, "not_found")
	_, err = s.DoneTodo(ctx, DoneTodoInput{OperationID: "bad-5", ID: r1.ID, Outcome: strings.Repeat("a", 4097)})
	faultCode(t, err, "invalid_input")
	_, err = s.DoneTodo(ctx, DoneTodoInput{OperationID: "bad-6", ID: r1.ID, Evidence: []Evidence{{Source: ""}}})
	faultCode(t, err, "invalid_input")
}
