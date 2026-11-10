package notebook

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type rowQuery interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func readOne(ctx context.Context, query rowQuery, project string, ref Reference) (ReadItem, error) {
	item := ReadItem{Reference: ref, Status: "missing"}
	var document string
	var retired bool
	err := query.QueryRowContext(ctx, "SELECT revision,retired,document FROM records WHERE project_id=? AND kind=? AND id=?", project, ref.Kind, ref.ID).Scan(&item.Revision, &retired, &document)
	if errors.Is(err, sql.ErrNoRows) {
		return item, nil
	}
	if err != nil {
		return item, err
	}
	if retired {
		item.Status = "retired"
		return item, nil
	}
	item.Status = "complete"
	return decodeRecord(item, document)
}

var recordDecoders = map[string]func(ReadItem, string) (ReadItem, error){
	"fact":  decodeFact,
	"event": decodeEvent,
}

func decodeRecord(item ReadItem, document string) (ReadItem, error) {
	decode := recordDecoders[item.Kind]
	if decode == nil {
		return item, invalid("record kind")
	}
	return decode(item, document)
}

func decodeFact(item ReadItem, document string) (ReadItem, error) {
	item.Fact = &Fact{}
	return item, json.Unmarshal([]byte(document), item.Fact)
}

func decodeEvent(item ReadItem, document string) (ReadItem, error) {
	item.Event = &Event{}
	return item, json.Unmarshal([]byte(document), item.Event)
}
