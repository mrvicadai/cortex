package notebook

import (
	"context"
	"database/sql"
)

type recentEvent struct {
	seq  int64
	item ReadItem
}

func (s *Store) recentEvents(ctx context.Context, before int64) ([]recentEvent, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT seq,id,document FROM records WHERE project_id=? AND kind='event' AND (?=0 OR seq<?) ORDER BY seq DESC LIMIT 11", s.project, before, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []recentEvent{}
	for rows.Next() {
		event, err := scanRecentEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func scanRecentEvent(rows *sql.Rows) (recentEvent, error) {
	event := recentEvent{item: ReadItem{Reference: Reference{Kind: "event"}, Status: "complete"}}
	var document string
	if err := rows.Scan(&event.seq, &event.item.ID, &document); err != nil {
		return event, err
	}
	item, err := decodeEvent(event.item, document)
	event.item = item
	return event, err
}
