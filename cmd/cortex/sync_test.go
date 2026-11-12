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

	// Verify stale content replaced with fresh mandatory protocol block
	if strings.Contains(updated, "OLD STALE CONTENT") {
		t.Error("old stale content was not replaced")
	}
	if !strings.Contains(updated, "MANDATORY PROTOCOL") || !strings.Contains(updated, "cortex_bootstrap") {
		t.Error("updated file does not contain mandatory bootstrap protocol directive")
	}
	if !strings.Contains(updated, "RECORD PROGRESS MILESTONES") {
		t.Error("updated file does not contain RECORD PROGRESS MILESTONES directive")
	}
	if strings.Contains(updated, "beliefs_bootstrap") {
		t.Error("updated file unexpectedly contains legacy beliefs_bootstrap")
	}
	if strings.Contains(updated, "| Topic ID | Topic Name | When to Call |") {
		t.Error("updated file should not contain static topic table")
	}
}

func TestSyncFileWithoutMarkersInsertsBlock(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "AGENTS.md")

	content := "# Just a normal file without markers\n\nSome guidelines here."
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	bStore, _ := beliefs.NewStore("")
	err := syncFile(targetFile, bStore)
	if err != nil {
		t.Fatalf("expected sync to succeed and insert block, got error: %v", err)
	}

	updatedBytes, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("reading updated file: %v", err)
	}
	updated := string(updatedBytes)
	if !strings.Contains(updated, "<!-- CORTEX_BELIEFS_START -->") {
		t.Error("expected inserted block to contain start marker")
	}
	if !strings.Contains(updated, "Some guidelines here.") {
		t.Error("expected original file content to be preserved")
	}
}

func TestSyncClaudeMD(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "CLAUDE.md")

	content := "# Project Rules\n\nExisting rule."
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	bStore, _ := beliefs.NewStore("")
	err := syncFile(targetFile, bStore)
	if err != nil {
		t.Fatalf("expected sync on CLAUDE.md to succeed, got: %v", err)
	}

	updatedBytes, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("reading updated file: %v", err)
	}
	updated := string(updatedBytes)
	if !strings.Contains(updated, "<!-- CORTEX_BELIEFS_START -->") || !strings.Contains(updated, "cortex_bootstrap({})") {
		t.Errorf("expected CLAUDE.md to have synced instruction block, got: %s", updated)
	}
	if !strings.Contains(updated, "Existing rule.") {
		t.Error("expected original content of CLAUDE.md to be preserved")
	}
}

func TestSyncNonExistentFileCreatesFile(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "AGENTS.md")

	bStore, _ := beliefs.NewStore("")
	err := syncFile(targetFile, bStore)
	if err != nil {
		t.Fatalf("expected sync on non-existent file to succeed by creating it, got: %v", err)
	}

	updatedBytes, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("reading updated file: %v", err)
	}
	updated := string(updatedBytes)
	if !strings.Contains(updated, "<!-- CORTEX_BELIEFS_START -->") {
		t.Error("expected created file to contain cortex instruction block")
	}
}

func TestSyncMigratesLegacyPointer(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "AGENTS.md")

	content := `<!-- CORTEX_START -->
Legacy pointer block
<!-- CORTEX_END -->`
	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	bStore, _ := beliefs.NewStore("")
	if err := syncFile(targetFile, bStore); err != nil {
		t.Fatalf("syncFile failed on legacy pointer: %v", err)
	}

	updatedBytes, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("reading updated file: %v", err)
	}
	updated := string(updatedBytes)

	if !strings.Contains(updated, "MANDATORY PROTOCOL") || !strings.Contains(updated, "<!-- CORTEX_BELIEFS_START -->") {
		t.Errorf("expected legacy pointer to be migrated to mandatory protocol block, got: %s", updated)
	}
}
