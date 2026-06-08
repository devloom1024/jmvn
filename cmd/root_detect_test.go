package cmd

import (
	"testing"

	"jmvn/internal/config"
)

func TestRootCommand_UsesDetectedJDKWhenProjectConfigMissing(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	capturedProjectJDK := ""
	deps = baseTestDeps()
	deps.detectJDKVersion = func(string) string { return "21" }
	deps.resolve = func(projectCfg config.ProjectConfig, globalCfg config.GlobalConfig, env map[string]string, projectDir string) (config.ResolvedConfig, error) {
		capturedProjectJDK = projectCfg.JDK
		return config.ResolvedConfig{JavaCmd: `java`, MavenHome: `maven`, ProjectDir: `D:/work/demo`}, nil
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
