package cmd

import "testing"

func TestRootCommand_PassesMavenArgs(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"clean", "install"})

	_, mvnArgs, err := executeForTest(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mvnArgs) != 2 || mvnArgs[0] != "clean" || mvnArgs[1] != "install" {
		t.Fatalf("expected Maven args [clean install], got %#v", mvnArgs)
	}
}

func TestRootCommand_PassesMavenArgsWithFlags(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"-pl", "module", "-DskipTests"})

	_, mvnArgs, err := executeForTest(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mvnArgs) != 3 || mvnArgs[0] != "-pl" || mvnArgs[1] != "module" || mvnArgs[2] != "-DskipTests" {
		t.Fatalf("expected Maven args [-pl module -DskipTests], got %#v", mvnArgs)
	}
}

func TestRootCommand_DryRunPassesMavenArgs(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{":dry-run", "clean", "install"})

	_, mvnArgs, err := executeForTest(cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mvnArgs) != 2 || mvnArgs[0] != "clean" || mvnArgs[1] != "install" {
		t.Fatalf("expected Maven args [clean install], got %#v", mvnArgs)
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
