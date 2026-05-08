package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jmvn/internal/config"
)

func TestBuildCommand_IncludesJvmConfigArgs(t *testing.T) {
	tempDir := t.TempDir()
	mavenHome := filepath.Join(tempDir, "maven")
	projectDir := filepath.Join(tempDir, "project")
	if err := os.MkdirAll(filepath.Join(mavenHome, "boot"), 0o755); err != nil {
		t.Fatalf("mkdir boot: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(mavenHome, "bin"), 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".mvn"), 0o755); err != nil {
		t.Fatalf("mkdir .mvn: %v", err)
	}
	if err := os.WriteFile(filepath.Join(mavenHome, "boot", "plexus-classworlds-2.0.jar"), []byte(""), 0o644); err != nil {
		t.Fatalf("write jar: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".mvn", "jvm.config"), []byte("-Xmx512m\n-Ddemo=true\n"), 0o644); err != nil {
		t.Fatalf("write jvm.config: %v", err)
	}

	cmd, err := BuildCommand(config.ResolvedConfig{
		JavaCmd:    filepath.Clean(`/jdks/jdk-17/bin/java`),
		MavenHome:  mavenHome,
		ProjectDir: projectDir,
	}, []string{"clean"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	joined := strings.Join(cmd.Args, " ")
	if !strings.Contains(joined, "-Xmx512m") || !strings.Contains(joined, "-Ddemo=true") {
		t.Fatalf("expected jvm.config args in command, got %q", joined)
	}
}
