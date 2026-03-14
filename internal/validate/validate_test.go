package validate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jmvn/internal/config"
	"jmvn/internal/validate"
)

func TestResolvedConfig_ReturnsHelpfulErrorForMissingJava(t *testing.T) {
	err := validate.ResolvedConfig(config.ResolvedConfig{JavaCmd: filepath.Clean(`D:/missing/java.exe`)})
	if err == nil {
		t.Fatal("expected error for missing java")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "java") {
		t.Fatalf("expected helpful java error, got %q", got)
	}
}

func TestResolvedConfig_AcceptsMissingOptionalPaths(t *testing.T) {
	tempDir := t.TempDir()
	javaPath := filepath.Join(tempDir, "bin", "java")
	mavenHome := filepath.Join(tempDir, "maven")
	bootDir := filepath.Join(mavenHome, "boot")
	binDir := filepath.Join(mavenHome, "bin")
	mustMkdirAll(t, filepath.Dir(javaPath))
	mustMkdirAll(t, bootDir)
	mustMkdirAll(t, binDir)
	mustWriteFile(t, javaPath, "")
	mustWriteFile(t, filepath.Join(bootDir, "plexus-classworlds-2.0.jar"), "")
	mustWriteFile(t, filepath.Join(binDir, "m2.conf"), "")

	err := validate.ResolvedConfig(config.ResolvedConfig{
		JavaCmd:   javaPath,
		MavenHome: mavenHome,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
