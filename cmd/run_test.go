package cmd

import (
	"strings"
	"testing"

	"jmvn/internal/cli"

	"github.com/spf13/cobra"
)

func TestRunCommand_ExecutesMavenWithArgs(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"run", "--dry-run", "clean", "install"})

	opts, mvnArgs, err := executeForTest(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !opts.DryRun {
		t.Fatalf("expected dry-run to be true")
	}
	if len(mvnArgs) != 2 || mvnArgs[0] != "clean" || mvnArgs[1] != "install" {
		t.Fatalf("expected Maven args [clean install], got %#v", mvnArgs)
	}
}

func TestRunCommand_ParsesOwnFlagsAndLeavesMavenArgs(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"run", "--jdk", "17", "--dry-run", "clean", "install"})

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

func TestRunCommand_PassesMavenDashDFlagAsArg(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"run", "--jdk", "17", "clean", "install", "-DskipTests"})

	opts, mvnArgs, err := executeForTest(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if opts.JDK != "17" {
		t.Fatalf("expected JDK 17, got %q", opts.JDK)
	}
	if len(mvnArgs) != 3 || mvnArgs[2] != "-DskipTests" {
		t.Fatalf("expected Maven args [clean install -DskipTests], got %#v", mvnArgs)
	}
}

func TestRunCommand_ShowsInHelp(t *testing.T) {
	cmd := NewRootCmd()
	output := captureHelpOutput(cmd)

	if !strings.Contains(output, "run") {
		t.Fatalf("expected help output to mention 'run' subcommand, got %q", output)
	}
	if !strings.Contains(output, "jmvn run clean install") {
		t.Fatalf("expected help to show 'jmvn run clean install' example, got %q", output)
	}
}

func TestRunCommand_SameBehaviorAsRoot(t *testing.T) {
	runTest := func(args []string) (cli.Options, []string) {
		cmd := NewRootCmd()
		cmd.SetArgs(args)
		opts, mvnArgs, err := executeForTest(cmd)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		return opts, mvnArgs
	}

	rootOpts, rootArgs := runTest([]string{"--jdk", "11", "--dry-run", "compile"})
	runOpts, runArgs := runTest([]string{"run", "--jdk", "11", "--dry-run", "compile"})

	if rootOpts.JDK != runOpts.JDK || rootOpts.DryRun != runOpts.DryRun {
		t.Fatalf("expected same option parsing: root=%+v run=%+v", rootOpts, runOpts)
	}
	if len(rootArgs) != len(runArgs) || rootArgs[0] != runArgs[0] {
		t.Fatalf("expected same maven args: root=%v run=%v", rootArgs, runArgs)
	}
}

func TestRunCommand_HelpShowsPersistentFlags(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"run", "--help"})

	var stdout strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "--jdk") {
		t.Fatalf("expected run help to show --jdk flag, got %q", output)
	}
	if !strings.Contains(output, "--dry-run") {
		t.Fatalf("expected run help to show --dry-run flag, got %q", output)
	}
}

func captureHelpOutput(_ *cobra.Command) string {
	helpCmd := NewRootCmd()
	helpCmd.SetArgs([]string{"--help"})
	var stdout strings.Builder
	helpCmd.SetOut(&stdout)
	helpCmd.SetErr(&stdout)
	_ = helpCmd.Execute()
	return stdout.String()
}
