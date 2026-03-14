package runner

import (
	"path/filepath"
	"strings"
	"testing"

	"jmvn/internal/config"
)

func TestBuildCommand_IncludesMavenLauncherAndOverrides(t *testing.T) {
	cfg := config.ResolvedConfig{
		JavaCmd:    filepath.Clean(`/jdks/jdk-17/bin/java`),
		MavenHome:  filepath.Clean(`/maven/apache-maven-3.9.6`),
		Settings:   filepath.Clean(`/work/demo/settings.xml`),
		LocalRepo:  filepath.Clean(`/work/demo/.m2/repository`),
		ProjectDir: filepath.Clean(`/work/demo`),
	}

	cmd, err := BuildCommand(cfg, []string{"clean", "install"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cmd.Path != cfg.JavaCmd {
		t.Fatalf("expected path %q, got %q", cfg.JavaCmd, cmd.Path)
	}
	joined := strings.Join(cmd.Args, " ")
	if !strings.Contains(joined, "org.codehaus.plexus.classworlds.launcher.Launcher") {
		t.Fatalf("expected launcher in args, got %q", joined)
	}
	if !strings.Contains(joined, "--settings") {
		t.Fatalf("expected settings flag in args, got %q", joined)
	}
	if !strings.Contains(joined, "-Dmaven.repo.local=") {
		t.Fatalf("expected local repo override in args, got %q", joined)
	}
	if cmd.Dir != cfg.ProjectDir {
		t.Fatalf("expected dir %q, got %q", cfg.ProjectDir, cmd.Dir)
	}
}

func TestBuildCommand_AddsNativeAccessForMaven4(t *testing.T) {
	cfg := config.ResolvedConfig{
		JavaCmd:    filepath.Clean(`/jdks/jdk-21/bin/java`),
		MavenHome:  filepath.Clean(`/maven/apache-maven-4.0.0`),
		ProjectDir: filepath.Clean(`/work/demo`),
	}

	cmd, err := BuildCommand(cfg, []string{"validate"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	joined := strings.Join(cmd.Args, " ")
	if !strings.Contains(joined, "--enable-native-access=ALL-UNNAMED") {
		t.Fatalf("expected native access flag for Maven 4, got %q", joined)
	}
}
