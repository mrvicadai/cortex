package notebook

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type Receipt struct {
	Reference
	Revision   string `json:"revision,omitempty"`
	Retired    bool   `json:"retired"`
	RecordedAt string `json:"recorded_at,omitempty"`
}

func (s *Store) mutate(ctx context.Context, project, operation, action string, request any, apply func(*sql.Tx) (Receipt, error)) (Receipt, error) {
	if !validName(operation, 128) {
		return Receipt{}, invalid("operation_id")
	}
	encoded, err := json.Marshal(struct {
		Action  string
		Request any
	}{action, request})
	if err != nil {
		return Receipt{}, err
	}
	var receipt Receipt
	err = s.transaction(ctx, func(tx *sql.Tx) error {
		if err := ensureProjectRow(ctx, tx, project); err != nil {
			return err
		}
		cached, err := replay(ctx, tx, project, operation, string(encoded))
		if err != nil {
			return err
		}
		if cached != nil {
			receipt = *cached
			return nil
		}
		receipt, err = apply(tx)
		if err != nil {
			return err
		}
		result, err := json.Marshal(receipt)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO receipts(project_id,operation_id,request,result) VALUES (?,?,?,?)", project, operation, string(encoded), string(result))
		return err
	})
	return receipt, err
}

func ensureProjectRow(ctx context.Context, tx *sql.Tx, project string) error {
	_, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO projects(id) VALUES (?)", project)
	return err
}

func replay(ctx context.Context, tx *sql.Tx, project, operation, request string) (*Receipt, error) {
	var previous, result string
	err := tx.QueryRowContext(ctx, "SELECT request,result FROM receipts WHERE project_id=? AND operation_id=?", project, operation).Scan(&previous, &result)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if previous != request && withoutLegacyProject(previous) != withoutLegacyProject(request) {
		return nil, &Fault{Code: "operation_conflict", Message: "operation_id was already used with a different request"}
	}
	var receipt Receipt
	if err := json.Unmarshal([]byte(result), &receipt); err != nil {
		return nil, err
	}
	return &receipt, nil
}

func withoutLegacyProject(request string) string {
	var envelope struct {
		Action  string
		Request map[string]json.RawMessage
	}
	if err := json.Unmarshal([]byte(request), &envelope); err != nil {
		return ""
	}
	delete(envelope.Request, "project_id")
	data, err := json.Marshal(envelope)
	if err != nil {
		return ""
	}
	return string(data)
}
