package notebook

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/mrvicadai/cortex/internal/wire"
)

type FindInput struct {
	Kind   string `json:"kind,omitempty" jsonschema:"fact, event, or all (default)."`
	Query  string `json:"query,omitempty" jsonschema:"Case-insensitive literal substrings, whitespace-separated and ANDed; no ranking, operators, stemming, or semantic search."`
	Topic  string `json:"topic,omitempty" jsonschema:"Optional exact case-sensitive fact topic; excludes events."`
	Cursor string `json:"cursor,omitempty" jsonschema:"Resume with cursor only."`
}

type findState struct {
	Kind   string `json:"kind"`
	Query  string `json:"query"`
	Topic  string `json:"topic"`
	Before int64  `json:"before"`
}

type DiscoveryItem struct {
	Reference
	Status     string `json:"status"`
	Revision   string `json:"revision,omitempty"`
	Title      string `json:"title,omitempty"`
	RecordedAt string `json:"recorded_at,omitempty"`
	Excerpt    string `json:"excerpt"`
}

type FindResult struct {
	Items      []DiscoveryItem `json:"items"`
	HasMore    bool            `json:"has_more"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

type discovered struct {
	seq  int64
	item DiscoveryItem
}

func (s *Store) Find(ctx context.Context, in FindInput) (FindResult, error) {
	state, err := s.findState(in)
	if err != nil {
		return FindResult{}, err
	}
	items, err := s.discover(ctx, s.project, state)
	if err != nil {
		return FindResult{}, err
	}
	return s.findPage(s.project, state, items)
}

func (s *Store) findPage(project string, state findState, items []discovered) (FindResult, error) {
	result := FindResult{Items: []DiscoveryItem{}}
	for i, found := range items {
		if i == 10 {
			break
		}
		candidate := FindResult{Items: append(result.Items[:len(result.Items):len(result.Items)], found.item), HasMore: i+1 < len(items)}
		if candidate.HasMore {
			state.Before = found.seq
			var err error
			candidate.NextCursor, err = s.encodeCursor(project, "find", state)
			if err != nil {
				return FindResult{}, err
			}
		}
		if !wire.Fits(candidate) {
			break
		}
		result = candidate
	}
	return result, nil
}

func prefix(body string) string {
	if len(body) <= 192 {
		return body
	}
	end := 192
	for !utf8.RuneStart(body[end]) {
		end--
	}
	return body[:end]
}

func queryTerms(query string) []string { return strings.Fields(strings.ToLower(query)) }
