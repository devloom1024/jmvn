package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jmvn/internal/cli"
	"jmvn/internal/config"
)

func TestInfoCommand_UsesDetectedJDKWhenProjectConfigMissing(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	projectDir := t.TempDir()
	capturedProjectJDK := ""
	deps = commandDeps{
		getwd:            func() (string, error) { return projectDir, nil },
		userHomeDir:      func() string { return `D:/home` },
		loadGlobal:       func(string) (config.GlobalConfig, error) { return config.GlobalConfig{}, nil },
		loadProject:      func(string) (config.ProjectConfig, error) { return config.ProjectConfig{}, nil },
		detectJDKVersion: func(string) string { return "8" },
		resolve: func(cliOpts cli.Options, projectCfg config.ProjectConfig, globalCfg config.GlobalConfig, env map[string]string, projectDir string) (config.ResolvedConfig, error) {
			capturedProjectJDK = projectCfg.JDK
			return config.ResolvedConfig{JavaCmd: `java`, MavenHome: `maven`, ProjectDir: projectDir}, nil
		},
		lookupEnv:  func() map[string]string { return map[string]string{} },
		promptInit: func(bool) (promptAnswers, error) { return promptAnswers{}, nil },
	}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"info"})
	_ = cmd.Execute()

	if capturedProjectJDK != "8" {
		t.Fatalf("expected info to reuse detected JDK, got %q", capturedProjectJDK)
	}
}

func TestInfoCommand_PrintsResolvedConfigSources(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	homeDir := t.TempDir()
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(homeDir, ".jmvn"), 0o755); err != nil {
		t.Fatalf("mkdir global config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(homeDir, ".jmvn", "config.toml"), []byte("[defaults]\n"), 0o644); err != nil {
		t.Fatalf("write global config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".jmvn.toml"), []byte("jdk = \"17\"\n"), 0o644); err != nil {
		t.Fatalf("write project config: %v", err)
	}

	deps = commandDeps{
		getwd:            func() (string, error) { return projectDir, nil },
		userHomeDir:      func() string { return homeDir },
		loadGlobal:       func(string) (config.GlobalConfig, error) { return config.GlobalConfig{}, nil },
		loadProject:      func(string) (config.ProjectConfig, error) { return config.ProjectConfig{}, nil },
		detectJDKVersion: func(string) string { return "" },
		resolve: func(cliOpts cli.Options, projectCfg config.ProjectConfig, globalCfg config.GlobalConfig, env map[string]string, projectDir string) (config.ResolvedConfig, error) {
			return config.ResolvedConfig{
				JavaCmd:         filepath.Clean(`D:/jdks/jdk-17/bin/java`),
				MavenHome:       filepath.Clean(`D:/mavens/apache-maven-3.9.6`),
				Settings:        filepath.Join(projectDir, "settings.xml"),
				LocalRepo:       filepath.Join(projectDir, ".m2", "repository"),
				ProjectDir:      projectDir,
				JavaCmdSource:   "project",
				MavenHomeSource: "global",
				SettingsSource:  "project",
				LocalRepoSource: "global",
			}, nil
		},
		lookupEnv:  func() map[string]string { return map[string]string{} },
		promptInit: func(bool) (promptAnswers, error) { return promptAnswers{}, nil },
	}

	cmd := NewRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"info"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "jmvn 配置解析") {
		t.Fatalf("expected header, got %q", output)
	}
	if !strings.Contains(output, "JDK") || !strings.Contains(output, "[project]") {
		t.Fatalf("expected JDK source output, got %q", output)
	}
	if !strings.Contains(output, "Maven") || !strings.Contains(output, "[global]") {
		t.Fatalf("expected Maven source output, got %q", output)
	}
	if !strings.Contains(output, projectDir) {
		t.Fatalf("expected project dir in output, got %q", output)
	}
	if !strings.Contains(output, "Config Files:") || !strings.Contains(output, "Global:") || !strings.Contains(output, "Project:") {
		t.Fatalf("expected config files section, got %q", output)
	}
	if !strings.Contains(output, "found") {
		t.Fatalf("expected found markers in config files section, got %q", output)
	}
}
