package cmd

import "testing"

func TestRootCommand_ParsesOwnFlagsAndLeavesMavenArgs(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--jdk", "17", "--dry-run", "clean", "install"})

	opts, mvnArgs, err := executeForTest(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if opts.JDK != "17" {
		t.Fatalf("expected JDK 17, got %q", opts.JDK)
	}
	if !opts.DryRun {
		t.Fatalf("expected dry-run to be true")
	}
	if len(mvnArgs) != 2 || mvnArgs[0] != "clean" || mvnArgs[1] != "install" {
		t.Fatalf("expected Maven args [clean install], got %#v", mvnArgs)
	}
}
