package notebook

import (
	"context"
	"database/sql"
)

func (s *Store) findState(in FindInput) (findState, error) {
	state := findState{Kind: in.Kind, Query: in.Query, Topic: in.Topic}
	if in.Cursor != "" {
		if in.Kind != "" || in.Query != "" || in.Topic != "" {
			return state, invalid("cursor is exclusive with filters")
		}
		if err := s.decodeCursor(in.Cursor, s.project, "find", &state); err != nil {
			return state, err
		}
	}
	if state.Kind == "" {
		state.Kind = "all"
	}
	return state, validateFind(state)
}

func validateFind(state findState) error {
	if state.Kind != "all" && state.Kind != "fact" && state.Kind != "event" {
		return invalid("kind")
	}
	if state.Topic != "" {
		if !validName(state.Topic, 64) || state.Kind == "event" {
			return invalid("topic or kind")
		}
	}
	return validateQuery(state.Query)
}

func validateQuery(query string) error {
	if query == "" {
		return nil
	}
	if !validName(query, 256) {
		return invalid("query: at most 256 UTF-8 bytes without controls or surrounding whitespace")
	}
	if len(queryTerms(query)) > 16 {
		return invalid("query: at most sixteen terms")
	}
	return nil
}

func (s *Store) discover(ctx context.Context, project string, state findState) ([]discovered, error) {
	statement, args := findSQL(project, state)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []discovered{}
	for rows.Next() {
		item, err := scanDiscovery(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func findSQL(project string, state findState) (string, []any) {
	statement := "SELECT seq,kind,id,revision,document FROM records WHERE project_id=? AND retired=0 AND (?=0 OR seq<?)"
	args := []any{project, state.Before, state.Before}
	if state.Kind != "all" {
		statement += " AND kind=?"
		args = append(args, state.Kind)
	}
	if state.Topic != "" {
		statement += " AND kind='fact' AND EXISTS (SELECT 1 FROM json_each(records.document,'$.topics') WHERE value=?)"
		args = append(args, state.Topic)
	}
	for _, term := range queryTerms(state.Query) {
		statement += " AND instr(search_text,?)>0"
		args = append(args, term)
	}
	return statement + " ORDER BY seq DESC LIMIT 11", args
}

func scanDiscovery(rows *sql.Rows) (discovered, error) {
	var seq int64
	var ref Reference
	var revision, document string
	if err := rows.Scan(&seq, &ref.Kind, &ref.ID, &revision, &document); err != nil {
		return discovered{}, err
	}
	item, err := decodeRecord(ReadItem{Reference: ref, Revision: revision, Status: "complete"}, document)
	if err != nil {
		return discovered{}, err
	}
	result := DiscoveryItem{Reference: ref, Status: "reference", Revision: revision}
	if item.Fact != nil {
		result.Title = item.Fact.Title
		result.Excerpt = prefix(item.Fact.Body)
	}
	if item.Event != nil {
		result.RecordedAt = item.Event.RecordedAt
		result.Excerpt = prefix(item.Event.Body)
	}
	return discovered{seq: seq, item: result}, nil
}
