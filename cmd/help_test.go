package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelp_IncludesInitGlobalExample(t *testing.T) {
	cmd := NewRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(stdout.String(), "jmvn init --global") {
		t.Fatalf("expected init --global example in root help, got %q", stdout.String())
	}
}

func TestInfoCommand_AcceptsPersistentRootFlags(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"info", "--jdk", "8"})

	_, _, err := executeForTest(cmd)
	if err != nil {
		t.Fatalf("expected inherited root flag support, got %v", err)
	}
}
