package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelp_IncludesJmvnCommands(t *testing.T) {
	cmd := NewRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{":help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, ":init") {
		t.Fatalf("expected :init in help, got %q", output)
	}
	if !strings.Contains(output, ":info") {
		t.Fatalf("expected :info in help, got %q", output)
	}
}
