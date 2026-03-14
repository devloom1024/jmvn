package util

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePath_ExpandsHomeAndProjectRelative(t *testing.T) {
	homeResult := ResolvePath("~/settings.xml", `D:/work/demo`)
	if !strings.HasSuffix(homeResult, filepath.Clean("settings.xml")) {
		t.Fatalf("expected expanded home path to end with settings.xml, got %q", homeResult)
	}

	relativeResult := ResolvePath("./maven/settings.xml", `D:/work/demo`)
	expected := filepath.Clean(`D:/work/demo/maven/settings.xml`)
	if relativeResult != expected {
		t.Fatalf("expected %q, got %q", expected, relativeResult)
	}
}
