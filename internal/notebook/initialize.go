package notebook

import (
	"context"
	"database/sql"
	"errors"
)

func (s *Store) initializeProject() error {
	return s.transaction(context.Background(), func(tx *sql.Tx) error {
		var legacy string
		err := tx.QueryRow("SELECT project_id FROM bindings WHERE directory=?", s.directory).Scan(&legacy)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if legacy != "" {
			s.project = legacy
		}
		return ensureProjectRow(context.Background(), tx, s.project)
	})
}
