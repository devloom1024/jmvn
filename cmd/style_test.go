package cmd

import (
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestStyledHeader_UsesColorEscape(t *testing.T) {
	original := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = original }()

	got := styledHeader("jmvn")
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("expected ANSI color escape, got %q", got)
	}
}
