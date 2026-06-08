package cmd

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"jmvn/internal/config"
)

func TestRootCommand_DryRunPrintsResolvedJavaCommand(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	deps = baseTestDeps()
	deps.getwd = func() (string, error) { return `D:/work/demo`, nil }
	deps.resolve = func(projectCfg config.ProjectConfig, globalCfg config.GlobalConfig, env map[string]string, projectDir string) (config.ResolvedConfig, error) {
		return config.ResolvedConfig{
			JavaCmd:    filepath.Clean(`/jdks/jdk-17/bin/java`),
			MavenHome:  filepath.Clean(`/maven/apache-maven-3.9.6`),
			Settings:   filepath.Clean(`/work/demo/settings.xml`),
			LocalRepo:  filepath.Clean(`/work/demo/.m2/repository`),
			ProjectDir: filepath.Clean(`/work/demo`),
		}, nil
	}
	deps.buildCommand = func(cfg config.ResolvedConfig, mavenArgs []string) (*exec.Cmd, error) {
		cmd := exec.Command(cfg.JavaCmd, append([]string{"org.codehaus.plexus.classworlds.launcher.Launcher"}, mavenArgs...)...)
		cmd.Dir = cfg.ProjectDir
		return cmd, nil
	}

	cmd := NewRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{":dry-run", "clean", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "org.codehaus.plexus.classworlds.launcher.Launcher") {
		t.Fatalf("expected launcher in output, got %q", output)
	}
	if !strings.Contains(output, "clean") || !strings.Contains(output, "test") {
		t.Fatalf("expected goals in output, got %q", output)
	}
}
