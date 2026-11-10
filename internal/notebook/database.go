package notebook

import (
	"database/sql"
	"fmt"
)

const applicationID = 0x464e4f54

func identifyDatabase(tx *sql.Tx) error {
	var application, version int
	if err := tx.QueryRow("PRAGMA application_id").Scan(&application); err != nil {
		return err
	}
	if err := tx.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if application == applicationID {
		if version != 1 {
			return fmt.Errorf("unsupported Fieldnotes schema version")
		}
		return nil
	}
	if application != 0 || version != 0 {
		return fmt.Errorf("not a Fieldnotes database; choose a new separate path")
	}
	return initializeEmptyDatabase(tx)
}

func initializeEmptyDatabase(tx *sql.Tx) error {
	var tables int
	if err := tx.QueryRow("SELECT count(*) FROM sqlite_schema WHERE name NOT LIKE 'sqlite_%'").Scan(&tables); err != nil {
		return err
	}
	if tables != 0 {
		return fmt.Errorf("refusing a nonempty unrelated database; choose a new separate path")
	}
	_, err := tx.Exec("PRAGMA application_id = 0x464e4f54")
	return err
}
