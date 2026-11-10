package notebook

import (
	"context"

	"github.com/mrvicadai/cortex/internal/wire"
)

type ReadInput struct {
	References []Reference `json:"references,omitempty" jsonschema:"One to ten fact/event references, in desired order; exclusive with cursor."`
	Cursor     string      `json:"cursor,omitempty" jsonschema:"Opaque continuation; resume with cursor only."`
}

type ReadResult struct {
	Items      []ReadItem `json:"items"`
	HasMore    bool       `json:"has_more"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

func (s *Store) Read(ctx context.Context, in ReadInput) (ReadResult, error) {
	refs, err := s.readReferences(in)
	if err != nil {
		return ReadResult{}, err
	}
	result := ReadResult{Items: []ReadItem{}}
	for i, ref := range refs {
		item, err := readOne(ctx, s.db, s.project, ref)
		if err != nil {
			return ReadResult{}, err
		}
		candidate, err := s.readCandidate(s.project, result.Items, item, refs[i+1:])
		if err != nil {
			return ReadResult{}, err
		}
		if !wire.Fits(candidate) {
			break
		}
		result = candidate
	}
	if len(result.Items) == 0 {
		return ReadResult{}, invalid("record cannot fit a complete read")
	}
	return result, nil
}

func (s *Store) readCandidate(project string, previous []ReadItem, item ReadItem, pending []Reference) (ReadResult, error) {
	result := ReadResult{Items: append(previous[:len(previous):len(previous)], item), HasMore: len(pending) > 0}
	var err error
	if result.HasMore {
		result.NextCursor, err = s.encodeCursor(project, "read", pending)
	}
	return result, err
}

func (s *Store) readReferences(in ReadInput) ([]Reference, error) {
	refs := in.References
	if in.Cursor != "" {
		if len(refs) != 0 {
			return nil, invalid("references and cursor are exclusive")
		}
		if err := s.decodeCursor(in.Cursor, s.project, "read", &refs); err != nil {
			return nil, err
		}
	}
	return refs, validateReferences(refs)
}

func validateReferences(refs []Reference) error {
	if len(refs) < 1 || len(refs) > 10 {
		return invalid("references: require one to ten")
	}
	for _, ref := range refs {
		if err := validateReference(ref); err != nil {
			return err
		}
	}
	return nil
}

func validateReference(ref Reference) error {
	if ref.Kind != "fact" && ref.Kind != "event" {
		return invalid("reference kind")
	}
	if !validName(ref.ID, 128) {
		return invalid("reference id")
	}
	return nil
}
