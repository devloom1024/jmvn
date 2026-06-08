package cmd

import (
	"bytes"
	"strings"
	"testing"

	"jmvn/internal/config"
)

func TestListCommand_PrintsRegisteredToolchains(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	deps = baseTestDeps()
	deps.userHomeDir = func() string { return `D:/home` }
	deps.loadGlobal = func(string) (config.GlobalConfig, error) {
		return config.GlobalConfig{
			JDKs: map[string]string{
				"17": `D:/jdks/jdk-17`,
			},
			Mavens: map[string]string{
				"3.9": `D:/mavens/apache-maven-3.9.6`,
			},
		}, nil
	}

	cmd := NewRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{":list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := stdout.String()
	if !strings.Contains(output, "已注册的 JDK") {
		t.Fatalf("expected JDK header, got %q", output)
	}
	if !strings.Contains(output, "17") || !strings.Contains(output, `D:/jdks/jdk-17`) {
		t.Fatalf("expected JDK entry, got %q", output)
	}
	if !strings.Contains(output, "已注册的 Maven") {
		t.Fatalf("expected Maven header, got %q", output)
	}
	if !strings.Contains(output, "3.9") || !strings.Contains(output, `D:/mavens/apache-maven-3.9.6`) {
		t.Fatalf("expected Maven entry, got %q", output)
	}
}
