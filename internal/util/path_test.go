package util

import (
	"path/filepath"
	"runtime"
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

func TestResolveJavaBinary_UsesPlatformExecutableName(t *testing.T) {
	got := ResolveJavaBinary(`D:/jdks/jdk-17`)
	want := filepath.Clean(`D:/jdks/jdk-17/bin/java`)
	if runtime.GOOS == "windows" {
		want = filepath.Clean(`D:/jdks/jdk-17/bin/java.exe`)
	}
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
