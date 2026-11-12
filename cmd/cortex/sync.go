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
	var content string
	fileExists := true
	if err != nil {
		if os.IsNotExist(err) {
			fileExists = false
		} else {
			return fmt.Errorf("reading %s: %w", targetPath, err)
		}
	} else {
		content = string(contentBytes)
	}

	newBlock := strings.TrimSpace(bStore.GetRenderedTableBlock())

	if !fileExists {
		if err := os.WriteFile(absPath, []byte(newBlock+"\n"), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", targetPath, err)
		}
		return nil
	}

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

	var updated strings.Builder
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		// In-place replacement of existing block
		endOfBlock := endIdx + len(selectedEndMarker)
		updated.WriteString(content[:startIdx])
		updated.WriteString(newBlock)
		updated.WriteString(content[endOfBlock:])
	} else {
		// Markers missing: prepend the instruction block to the top of the file
		updated.WriteString(newBlock)
		if strings.TrimSpace(content) != "" {
			updated.WriteString("\n\n")
			updated.WriteString(strings.TrimSpace(content))
		}
		updated.WriteString("\n")
	}

	if err := os.WriteFile(absPath, []byte(updated.String()), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", targetPath, err)
	}

	return nil
}
