package notebook

import (
	"encoding/json"
	"strings"

	"github.com/mrvicadai/cortex/internal/wire"
)

const contextLineTarget = 10
const contextRecentByteLimit = 8192

type contextState struct {
	Before int64 `json:"before"`
}

func (s *Store) contextState(in ContextInput) (contextState, error) {
	state := contextState{}
	if in.Cursor != "" {
		if err := s.decodeCursor(in.Cursor, s.project, "context", &state); err != nil {
			return state, err
		}
		if state.Before <= 0 {
			return state, invalid("context cursor position")
		}
	}
	return state, nil
}

func (s *Store) contextPage(result ContextResult, events []recentEvent) (ContextResult, error) {
	if len(events) == 0 {
		return result, nil
	}
	result, err := s.contextContinuation(result, events[0].seq+1, true)
	if err != nil {
		return ContextResult{}, err
	}
	lines := 0
	for i, event := range events {
		candidate := result
		candidate.Recent = append(result.Recent[:len(result.Recent):len(result.Recent)], event.item)
		candidate, err = s.contextContinuation(candidate, event.seq, i+1 < len(events))
		if err != nil {
			return ContextResult{}, err
		}
		nextLines := lines + strings.Count(event.item.Event.Body, "\n") + 1
		if !contextFits(candidate, nextLines) {
			break
		}
		result, lines = candidate, nextLines
	}
	return result, nil
}

func (s *Store) contextContinuation(result ContextResult, before int64, more bool) (ContextResult, error) {
	result.HasMore, result.NextCursor = more, ""
	var err error
	if more {
		result.NextCursor, err = s.encodeCursor(s.project, "context", contextState{Before: before})
	}
	return result, err
}

func contextFits(result ContextResult, lines int) bool {
	if len(result.Recent) > 1 && lines > contextLineTarget {
		return false
	}
	encoded, err := json.Marshal(result.Recent)
	return err == nil && len(encoded) <= contextRecentByteLimit && wire.Fits(result)
}
