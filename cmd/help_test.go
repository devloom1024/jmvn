package cmd

import "testing"

func TestRootHelp_IncludesInitGlobalExample(t *testing.T) {
	// placeholder - will be filled in Task 4
}

func TestInfoCommand_AcceptsPersistentRootFlags(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"info", "--jdk", "8"})

	_, _, err := executeForTest(cmd)
	if err != nil {
		t.Fatalf("expected inherited root flag support, got %v", err)
	}
}
