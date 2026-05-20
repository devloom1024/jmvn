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

func TestRootCommand_PassesMavenDashDFlagAsArg(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--jdk", "17", "clean", "install", "-DskipTests"})

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

func TestStripLeadingMvnPrefix_StripsMvn(t *testing.T) {
	args := []string{"mvn", "-pl", "module"}
	stripLeadingMvnPrefix(&args)
	if len(args) != 2 || args[0] != "-pl" || args[1] != "module" {
		t.Fatalf("expected [-pl module], got %#v", args)
	}
}

func TestStripLeadingMvnPrefix_StripsMvnw(t *testing.T) {
	args := []string{"mvnw", "clean", "test"}
	stripLeadingMvnPrefix(&args)
	if len(args) != 2 || args[0] != "clean" || args[1] != "test" {
		t.Fatalf("expected [clean test], got %#v", args)
	}
}

func TestStripLeadingMvnPrefix_EmptyArgs(t *testing.T) {
	args := []string{}
	stripLeadingMvnPrefix(&args)
	if len(args) != 0 {
		t.Fatalf("expected empty args, got %#v", args)
	}
}

func TestStripLeadingMvnPrefix_NoMatch(t *testing.T) {
	args := []string{"clean", "test"}
	stripLeadingMvnPrefix(&args)
	if len(args) != 2 || args[0] != "clean" || args[1] != "test" {
		t.Fatalf("expected [clean test], got %#v", args)
	}
}

func TestRootCommand_StripsMvnPrefixWhenPresent(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"mvn", "clean", "install"})

	_, mvnArgs, err := executeForTest(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mvnArgs) != 2 || mvnArgs[0] != "clean" || mvnArgs[1] != "install" {
		t.Fatalf("expected Maven args [clean install] after stripping mvn, got %#v", mvnArgs)
	}
}
