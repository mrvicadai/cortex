package notebook

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/mrvicadai/cortex/internal/wire"
)

const contextTodoLimit = 20

type Todo struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Title     string `json:"title"`
	Notes     string `json:"notes,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type AddTodoInput struct {
	OperationID string `json:"operation_id" jsonschema:"Stable retry key; reuse with the identical request after response loss."`
	Title       string `json:"title" jsonschema:"Short, actionable summary of the deferred task (1–256 UTF-8 bytes)."`
	Notes       string `json:"notes,omitempty" jsonschema:"Optional context, file paths, or details (up to 2048 UTF-8 bytes)."`
}

type DoneTodoInput struct {
	OperationID string     `json:"operation_id" jsonschema:"Stable retry key; reuse with the identical request after response loss."`
	ID          string     `json:"id" jsonschema:"ID of the todo item to mark done."`
	Outcome     string     `json:"outcome,omitempty" jsonschema:"Optional completion summary for the auto-appended milestone event (defaults to 'Completed todo: <title>')."`
	Evidence    []Evidence `json:"evidence,omitempty" jsonschema:"Optional structured evidence pointers (e.g. git commit hash, test logs, PR link) attached to the auto-appended event."`
	SkipEvent   bool       `json:"skip_event,omitempty" jsonschema:"Set true ONLY when dropping an obsolete, duplicate, or cancelled todo without doing real work. Leave false (default) when completing a task so the milestone event is recorded automatically."`
}

type ListTodosInput struct {
	Status string `json:"status,omitempty" jsonschema:"Filter by status: 'open' (default), 'done', or 'all'."`
}

type ListTodosResult struct {
	Items     []Todo `json:"items"`
	OpenCount int    `json:"open_count"`
	Truncated bool   `json:"truncated,omitempty"`
}

// AddTodo appends a new open todo item for the active project.
func (s *Store) AddTodo(ctx context.Context, in AddTodoInput) (Receipt, error) {
	if !validName(in.Title, 256) {
		return Receipt{}, invalid("title: require 1–256 UTF-8 bytes without controls or surrounding whitespace")
	}
	if len(in.Notes) > 2048 || !utf8.ValidString(in.Notes) {
		return Receipt{}, invalid("notes: at most 2048 UTF-8 bytes")
	}
	return s.mutate(ctx, s.project, in.OperationID, "todo_add", in, func(tx *sql.Tx) (Receipt, error) {
		id := token()[:12]
		now := time.Now().UTC().Format(time.RFC3339Nano)
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO todos(project_id,id,status,title,notes,created_at,updated_at) VALUES (?,?,'open',?,?,?,?)`,
			s.project, id, in.Title, in.Notes, now, now,
		)
		return Receipt{Reference: Reference{Kind: "todo", ID: id}, RecordedAt: now}, err
	})
}

// DoneTodo marks an existing todo item as done and appends a milestone event unless skip_event is true.
func (s *Store) DoneTodo(ctx context.Context, in DoneTodoInput) (Receipt, error) {
	if err := validateDoneTodoInput(in); err != nil {
		return Receipt{}, err
	}
	return s.mutate(ctx, s.project, in.OperationID, "todo_done", in, func(tx *sql.Tx) (Receipt, error) {
		var prevStatus, title string
		err := tx.QueryRowContext(ctx, "SELECT status, title FROM todos WHERE project_id=? AND id=?", s.project, in.ID).Scan(&prevStatus, &title)
		if errors.Is(err, sql.ErrNoRows) {
			return Receipt{}, &Fault{Code: "not_found", Message: "todo not found"}
		}
		if err != nil {
			return Receipt{}, err
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		if _, err := tx.ExecContext(ctx, "UPDATE todos SET status='done', updated_at=? WHERE project_id=? AND id=?", now, s.project, in.ID); err != nil {
			return Receipt{}, err
		}
		if !in.SkipEvent && prevStatus != "done" {
			if err := appendTodoEvent(ctx, tx, s.project, now, title, in); err != nil {
				return Receipt{}, err
			}
		}
		return Receipt{Reference: Reference{Kind: "todo", ID: in.ID}, Retired: true, RecordedAt: now}, nil
	})
}

func validateDoneTodoInput(in DoneTodoInput) error {
	if !validName(in.ID, 128) {
		return invalid("id")
	}
	if in.Outcome != "" {
		if err := validateBody(in.Outcome); err != nil {
			return err
		}
	}
	if in.Evidence != nil {
		return validateEvidence(in.Evidence)
	}
	return nil
}

func appendTodoEvent(ctx context.Context, tx *sql.Tx, project, now, title string, in DoneTodoInput) error {
	body := in.Outcome
	if body == "" {
		body = "Completed todo: " + title
	}
	return appendMilestoneEvent(ctx, tx, project, now, body, in.Evidence)
}

func appendMilestoneEvent(ctx context.Context, tx *sql.Tx, project, now, body string, evidence []Evidence) error {
	if evidence == nil {
		evidence = []Evidence{}
	}
	event := Event{ID: token(), RecordedAt: now, Body: body, Evidence: evidence}
	document, err := encodeDocument(event)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO records(project_id,kind,id,revision,document,search_text) VALUES (?,'event',?,'',?,?)",
		project, event.ID, document, eventSearch(event),
	)
	return err
}

// ListTodos returns todos for the active project filtered by status ('open', 'done', or 'all').
func (s *Store) ListTodos(ctx context.Context, in ListTodosInput) (ListTodosResult, error) {
	status := in.Status
	if status == "" {
		status = "open"
	}
	if status != "open" && status != "done" && status != "all" {
		return ListTodosResult{}, invalid("status: require 'open', 'done', or 'all'")
	}
	openCount, err := s.countOpenTodos(ctx)
	if err != nil {
		return ListTodosResult{}, err
	}
	todos, err := s.queryTodos(ctx, status, 0)
	if err != nil {
		return ListTodosResult{}, err
	}
	result := ListTodosResult{Items: []Todo{}, OpenCount: openCount}
	for _, todo := range todos {
		candidate := result
		candidate.Items = append(result.Items[:len(result.Items):len(result.Items)], todo)
		if !wire.Fits(candidate) {
			result.Truncated = true
			break
		}
		result = candidate
	}
	return result, nil
}

func (s *Store) openTodosForContext(ctx context.Context) ([]Todo, int, error) {
	openCount, err := s.countOpenTodos(ctx)
	if err != nil || openCount == 0 {
		return nil, 0, err
	}
	todos, err := s.queryTodos(ctx, "open", contextTodoLimit)
	return todos, openCount, err
}

func (s *Store) countOpenTodos(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM todos WHERE project_id=? AND status='open'", s.project).Scan(&count)
	return count, err
}

func (s *Store) queryTodos(ctx context.Context, status string, limit int) ([]Todo, error) {
	query := "SELECT id,status,title,notes,created_at,updated_at FROM todos WHERE project_id=?"
	args := []any{s.project}
	if status != "all" {
		query += " AND status=?"
		args = append(args, status)
	}
	query += " ORDER BY seq ASC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	todos := []Todo{}
	for rows.Next() {
		var item Todo
		if err := rows.Scan(&item.ID, &item.Status, &item.Title, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, item)
	}
	return todos, rows.Err()
}
