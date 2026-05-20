package main

import "testing"

type fakeExitCoder struct {
	code int
}

func (f fakeExitCoder) Error() string {
	return "boom"
}

func (f fakeExitCoder) ExitCode() int {
	return f.code
}

func TestExitCodeForError_DefaultsToOne(t *testing.T) {
	if got := exitCodeForError(fakePlainError("boom")); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestExitCodeForError_UsesExitCoder(t *testing.T) {
	if got := exitCodeForError(fakeExitCoder{code: 7}); got != 7 {
		t.Fatalf("expected 7, got %d", got)
	}
}

type fakePlainError string

func (f fakePlainError) Error() string {
	return string(f)
}

func TestPreprocessArgs_PassesThroughCleanTest(t *testing.T) {
	result := preprocessArgs([]string{"clean", "test"})
	assertArgs(t, result, "clean", "test")
}

func TestPreprocessArgs_InsertsDashDashBeforeMavenFlags(t *testing.T) {
	result := preprocessArgs([]string{"-pl", "module", "test"})
	assertArgs(t, result, "--", "-pl", "module", "test")
}

func TestPreprocessArgs_InsertsDashDashAfterJmvnFlags(t *testing.T) {
	result := preprocessArgs([]string{"--jdk", "17", "-pl", "module", "test"})
	assertArgs(t, result, "--jdk", "17", "--", "-pl", "module", "test")
}

func TestPreprocessArgs_NoDashDashWhenNonFlagSeenFirst(t *testing.T) {
	result := preprocessArgs([]string{"test", "-pl", "module"})
	assertArgs(t, result, "test", "-pl", "module")
}

func TestPreprocessArgs_DryRunBeforeMavenFlags(t *testing.T) {
	result := preprocessArgs([]string{"--dry-run", "-pl", "module", "test"})
	assertArgs(t, result, "--dry-run", "--", "-pl", "module", "test")
}

func TestPreprocessArgs_ResetsAfterSubcommand(t *testing.T) {
	result := preprocessArgs([]string{"run", "-pl", "module", "test"})
	assertArgs(t, result, "run", "--", "-pl", "module", "test")
}

func TestPreprocessArgs_ResetsAfterSubcommandWithJmvnFlags(t *testing.T) {
	result := preprocessArgs([]string{"run", "--jdk", "17", "-pl", "module", "test"})
	assertArgs(t, result, "run", "--jdk", "17", "--", "-pl", "module", "test")
}

func TestPreprocessArgs_PreservesExistingDashDash(t *testing.T) {
	result := preprocessArgs([]string{"--", "-pl", "module", "test"})
	assertArgs(t, result, "--", "-pl", "module", "test")
}

func TestPreprocessArgs_PreservesMvnPrefix(t *testing.T) {
	result := preprocessArgs([]string{"mvn", "-pl", "module", "test"})
	assertArgs(t, result, "mvn", "-pl", "module", "test")
}

func TestPreprocessArgs_JmvnFlagsWithEqualsSign(t *testing.T) {
	result := preprocessArgs([]string{"--jdk=17", "-pl", "module", "test"})
	assertArgs(t, result, "--jdk=17", "--", "-pl", "module", "test")
}

func TestPreprocessArgs_EmptyArgs(t *testing.T) {
	result := preprocessArgs([]string{})
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %#v", result)
	}
}

func assertArgs(t *testing.T, actual []string, expected ...string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("expected %#v, got %#v", expected, actual)
	}
	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("expected %#v at index %d, got %#v", expected[i], i, actual[i])
		}
	}
}
