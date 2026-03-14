package cmd

import (
	"os/exec"
	"testing"

	"jmvn/internal/cli"
	"jmvn/internal/config"
)

func TestRootCommand_UsesDetectedJDKWhenProjectConfigMissing(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	capturedProjectJDK := ""
	deps = commandDeps{
		getwd:            func() (string, error) { return `D:/work/demo`, nil },
		userHomeDir:      func() string { return `D:/home` },
		loadGlobal:       func(string) (config.GlobalConfig, error) { return config.GlobalConfig{}, nil },
		loadProject:      func(string) (config.ProjectConfig, error) { return config.ProjectConfig{}, nil },
		detectJDKVersion: func(string) string { return "21" },
		resolve: func(cliOpts cli.Options, projectCfg config.ProjectConfig, globalCfg config.GlobalConfig, env map[string]string, projectDir string) (config.ResolvedConfig, error) {
			capturedProjectJDK = projectCfg.JDK
			return config.ResolvedConfig{JavaCmd: `java`, MavenHome: `maven`, ProjectDir: `D:/work/demo`}, nil
		},
		validateResolved: func(config.ResolvedConfig) error { return nil },
		buildCommand: func(cfg config.ResolvedConfig, mavenArgs []string) (*exec.Cmd, error) {
			cmd := exec.Command(cfg.JavaCmd, mavenArgs...)
			cmd.Dir = cfg.ProjectDir
			return cmd, nil
		},
		lookupEnv:      func() map[string]string { return map[string]string{} },
		promptInit:     func(bool) (promptAnswers, error) { return promptAnswers{}, nil },
		executeCommand: func(*exec.Cmd) error { return nil },
	}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"clean"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedProjectJDK != "21" {
		t.Fatalf("expected detected JDK to be forwarded, got %q", capturedProjectJDK)
	}
}
