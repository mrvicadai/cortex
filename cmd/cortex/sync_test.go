package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrvicadai/cortex/internal/beliefs"
)

func TestSyncUpdatesTableInPlace(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "AGENTS.md")

	initialContent := `# Project Documentation

User defined pre-amble rule that must not be deleted.

<!-- CORTEX_BELIEFS_START -->
OLD STALE CONTENT
<!-- CORTEX_BELIEFS_END -->

User defined post-amble rule that must be preserved.
`
	if err := os.WriteFile(targetFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	bStore, err := beliefs.NewStore("")
	if err != nil {
		t.Fatalf("creating beliefs store: %v", err)
	}

	if err := syncFile(targetFile, bStore); err != nil {
		t.Fatalf("syncFile failed: %v", err)
	}

	updatedBytes, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("reading updated file: %v", err)
	}
	updated := string(updatedBytes)

	// Verify preamble and postamble preserved
	if !strings.Contains(updated, "User defined pre-amble rule that must not be deleted.") {
		t.Error("pre-amble was unexpectedly modified or removed")
	}
	if !strings.Contains(updated, "User defined post-amble rule that must be preserved.") {
		t.Error("post-amble was unexpectedly modified or removed")
	}

	// Verify stale content replaced with fresh topics table
	if strings.Contains(updated, "OLD STALE CONTENT") {
		t.Error("old stale content was not replaced")
	}
	if !strings.Contains(updated, "coding-standards") || !strings.Contains(updated, "mr-rules") {
		t.Error("updated file does not contain fresh topic table")
	}
}

func TestSyncFileWithoutMarkersFails(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "AGENTS.md")

	content := "# Just a normal file without markers"
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	bStore, _ := beliefs.NewStore("")
	err := syncFile(targetFile, bStore)
	if err == nil {
		t.Error("expected error when markers are missing, got nil")
	}
}

func TestSyncFileWithZeroDriftPointerInformsUser(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "AGENTS.md")

	content := `<!-- CORTEX_START -->
Zero-drift pointer active
<!-- CORTEX_END -->`
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	bStore, _ := beliefs.NewStore("")
	err := syncFile(targetFile, bStore)
	if err == nil || !strings.Contains(err.Error(), "zero-drift dynamic pointer") {
		t.Errorf("expected zero-drift pointer notification error, got: %v", err)
	}
}
