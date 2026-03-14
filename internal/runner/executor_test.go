package runner

import (
	"errors"
	"os/exec"
	"testing"
)

func TestExec_ReturnsExitCodeErrorForFailedCommand(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "exit", "7")
	err := Exec(cmd)
	if err == nil {
		t.Fatal("expected error")
	}
	var coded interface{ ExitCode() int }
	if !errors.As(err, &coded) {
		t.Fatalf("expected exit code error, got %T", err)
	}
	if coded.ExitCode() != 7 {
		t.Fatalf("expected exit code 7, got %d", coded.ExitCode())
	}
}
