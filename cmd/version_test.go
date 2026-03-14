package cmd

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func TestVersionCommand_PrintsBuildVersion(t *testing.T) {
	originalVersion := buildVersion
	defer func() { buildVersion = originalVersion }()
	buildVersion = "1.0.0"

	cmd := NewRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "jmvn v1.0.0") {
		t.Fatalf("expected version string, got %q", output)
	}
	if !strings.Contains(output, runtime.Version()) {
		t.Fatalf("expected Go version in output, got %q", output)
	}
	if !strings.Contains(output, runtime.GOOS+"/"+runtime.GOARCH) {
		t.Fatalf("expected platform in output, got %q", output)
	}
}
