package notebook

import (
	"context"
	"os"
	"path/filepath"
)

func canonicalDirectory(directory string) (string, error) {
	canonical, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(canonical)
	if err != nil || !info.IsDir() {
		return "", invalid("existing directory")
	}
	if !validName(canonical, 1024) {
		return "", invalid("directory length or characters")
	}
	return filepath.Clean(canonical), nil
}

func (s *Store) Context(ctx context.Context, in ContextInput) (ContextResult, error) {
	state, err := s.contextState(in)
	if err != nil {
		return ContextResult{}, err
	}
	result := ContextResult{Directory: s.directory, Recent: []ReadItem{}}
	if in.Cursor == "" {
		item, err := readOne(ctx, s.db, s.project, Reference{Kind: "fact", ID: ContextID})
		if err != nil {
			return ContextResult{}, err
		}
		if item.Status != "missing" {
			result.Context = &item
		}
	}
	events, err := s.recentEvents(ctx, state.Before)
	if err != nil {
		return ContextResult{}, err
	}
	return s.contextPage(result, events)
}
