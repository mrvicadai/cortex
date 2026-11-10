package notebook

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
)

func validateRevision(revision string, allowEmpty bool) error {
	if revision == "" && allowEmpty {
		return nil
	}
	decoded, err := hex.DecodeString(revision)
	if err != nil || len(decoded) != 16 {
		return invalid("expected_revision: use the opaque revision from a complete read")
	}
	return nil
}

func checkRevision(ctx context.Context, tx *sql.Tx, project, id, expected string) error {
	var current string
	err := tx.QueryRowContext(ctx, "SELECT revision FROM records WHERE project_id=? AND kind='fact' AND id=?", project, id).Scan(&current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if current != expected {
		return &Fault{Code: "revision_conflict", Message: "read the complete current record and reconcile before retrying", CurrentRevision: current}
	}
	return nil
}
