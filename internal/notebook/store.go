package notebook

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	_ "modernc.org/sqlite"
)

type Store struct {
	db        *sql.DB
	secret    []byte
	directory string
	project   string
}

const schema = `
CREATE TABLE IF NOT EXISTS projects (id TEXT PRIMARY KEY);
CREATE TABLE IF NOT EXISTS bindings (
 directory TEXT PRIMARY KEY, project_id TEXT NOT NULL REFERENCES projects(id)
);
CREATE TABLE IF NOT EXISTS records (
 seq INTEGER PRIMARY KEY AUTOINCREMENT,
 project_id TEXT NOT NULL REFERENCES projects(id),
 kind TEXT NOT NULL CHECK(kind IN ('fact','event')),
 id TEXT NOT NULL, revision TEXT NOT NULL, retired INTEGER NOT NULL DEFAULT 0,
 document TEXT NOT NULL, search_text TEXT NOT NULL,
 UNIQUE(project_id, kind, id)
);
CREATE TABLE IF NOT EXISTS receipts (
 project_id TEXT NOT NULL REFERENCES projects(id), operation_id TEXT NOT NULL,
 request TEXT NOT NULL, result TEXT NOT NULL, PRIMARY KEY(project_id, operation_id)
);
CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT NOT NULL);
PRAGMA user_version = 1;
`

func Open(path string) (*Store, error) {
	if !filepath.IsAbs(path) {
		return nil, invalid("absolute database path")
	}
	directory, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	directory, err = canonicalDirectory(directory)
	if err != nil {
		return nil, err
	}
	return openStore(path, directory)
}

func openStore(path, directory string) (*Store, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	dsn := url.URL{Scheme: "file", Path: path}
	query := url.Values{"_pragma": {"busy_timeout(10000)", "foreign_keys(1)"}, "_txlock": {"immediate"}}
	dsn.RawQuery = query.Encode()
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, directory: directory, project: directory}
	if err := s.initialize(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.initializeProject(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) initialize() error {
	return s.transaction(context.Background(), func(tx *sql.Tx) error {
		if err := identifyDatabase(tx); err != nil {
			return err
		}
		if _, err := tx.Exec(schema); err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT OR IGNORE INTO settings(key,value) VALUES ('cursor_secret',?)", token()+token()); err != nil {
			return err
		}
		var secret string
		if err := tx.QueryRow("SELECT value FROM settings WHERE key='cursor_secret'").Scan(&secret); err != nil {
			return err
		}
		s.secret = []byte(secret)
		return nil
	})
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) transaction(ctx context.Context, run func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := run(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func token() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("cryptographic random source unavailable")
	}
	return hex.EncodeToString(b[:])
}

func validName(value string, max int) bool {
	return value != "" && len(value) <= max && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value && !strings.ContainsFunc(value, unicode.IsControl)
}
