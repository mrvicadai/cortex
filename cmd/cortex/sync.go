package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrvicadai/cortex/internal/beliefs"
)

var (
	startMarkers = []string{"<!-- CORTEX_BELIEFS_START -->", "<!-- AGENT_BELIEFS_START -->", "<!-- CORTEX_START -->"}
	endMarkers   = []string{"<!-- CORTEX_BELIEFS_END -->", "<!-- AGENT_BELIEFS_END -->", "<!-- CORTEX_END -->"}
)

func syncFile(targetPath string, bStore *beliefs.Store) error {
	if targetPath == "" {
		targetPath = "AGENTS.md"
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", targetPath, err)
	}

	content := string(contentBytes)

	// Locate start and end markers
	startIdx := -1
	for _, m := range startMarkers {
		idx := strings.Index(content, m)
		if idx != -1 {
			startIdx = idx
			break
		}
	}

	endIdx := -1
	selectedEndMarker := ""
	for _, m := range endMarkers {
		idx := strings.Index(content, m)
		if idx != -1 {
			endIdx = idx
			selectedEndMarker = m
			break
		}
	}

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return fmt.Errorf("markers not found in %s. To enable in-place table sync, ensure both %s and %s exist in the file. Or use 'cortex snippet' for the zero-drift dynamic pointer", targetPath, startMarkers[0], endMarkers[0])
	}

	// Compute full replacement
	endOfBlock := endIdx + len(selectedEndMarker)
	newBlock := strings.TrimSpace(bStore.GetRenderedTableBlock())

	var updated strings.Builder
	updated.WriteString(content[:startIdx])
	updated.WriteString(newBlock)
	updated.WriteString(content[endOfBlock:])

	if err := os.WriteFile(absPath, []byte(updated.String()), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", targetPath, err)
	}

	return nil
}
