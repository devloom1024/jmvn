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
