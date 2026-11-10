package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLIHelpAndVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// --version
	err := run([]string{"--version"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error on --version: %v", err)
	}
	if strings.TrimSpace(stdout.String()) != "0.1.0" {
		t.Errorf("expected 0.1.0, got %s", stdout.String())
	}

	// -v
	stdout.Reset()
	err = run([]string{"-v"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error on -v: %v", err)
	}
	if strings.TrimSpace(stdout.String()) != "0.1.0" {
		t.Errorf("expected 0.1.0, got %s", stdout.String())
	}

	// --help
	stdout.Reset()
	err = run([]string{"--help"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error on --help: %v", err)
	}
	if !strings.Contains(stdout.String(), "@mrvicadai/cortex") {
		t.Errorf("expected help text, got %s", stdout.String())
	}
}

func TestCLISnippetAndInstructions(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// snippet
	err := run([]string{"snippet"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error on snippet: %v", err)
	}
	if !strings.Contains(stdout.String(), "CORTEX_START") {
		t.Errorf("expected snippet block, got %s", stdout.String())
	}

	// instructions
	stdout.Reset()
	err = run([]string{"instructions"}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error on instructions: %v", err)
	}
	if !strings.Contains(stdout.String(), "Cortex Agent Context: Beliefs Bootstrap") {
		t.Errorf("expected bootstrap instructions, got %s", stdout.String())
	}
}

func TestCLIRejectsUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"unknown-subcommand"}, &stdout, &stderr)
	if err == nil {
		t.Error("expected error for unknown subcommand, got nil")
	}
}
