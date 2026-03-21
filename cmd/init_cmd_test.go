package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommand_GlobalWritesRegisteredDefaultJDK(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	homeDir := t.TempDir()
	deps = commandDeps{
		userHomeDir: func() string { return homeDir },
		promptInit: func(global bool) (promptAnswers, error) {
			return promptAnswers{
				JDK:       "17",
				JDKHome:   `D:/jdks/jdk-17`,
				Maven:     "3.9",
				MavenHome: `D:/mavens/apache-maven-3.9.6`,
				Settings:  `D:/users/demo/.m2/settings.xml`,
				LocalRepo: `D:/users/demo/.m2/repository`,
			}, nil
		},
	}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init", "--global"})
	_ = cmd.Execute()

	content, _ := os.ReadFile(filepath.Join(homeDir, ".jmvn", "config.toml"))
	text := string(content)
	if !strings.Contains(text, `"17" = "D:/jdks/jdk-17"`) {
		t.Fatalf("expected registered default JDK mapping, got %q", text)
	}
	if !strings.Contains(text, `"3.9" = "D:/mavens/apache-maven-3.9.6"`) {
		t.Fatalf("expected registered default Maven mapping, got %q", text)
	}
}

func TestInitCommand_WritesProjectConfigTemplate(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	projectDir := t.TempDir()
	deps = commandDeps{
		getwd: func() (string, error) { return projectDir, nil },
		promptInit: func(global bool) (promptAnswers, error) {
			return promptAnswers{
				JDK:      "17",
				Maven:    "3.9",
				Settings: "./maven/settings.xml",
			}, nil
		},
	}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content, err := os.ReadFile(filepath.Join(projectDir, ".jmvn.toml"))
	if err != nil {
		t.Fatalf("read project config: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, `jdk = "17"`) {
		t.Fatalf("expected jdk field, got %q", text)
	}
	if !strings.Contains(text, `maven = "3.9"`) {
		t.Fatalf("expected maven field, got %q", text)
	}
	if !strings.Contains(text, `settings = "./maven/settings.xml"`) {
		t.Fatalf("expected settings field, got %q", text)
	}
}

func TestInitCommand_GlobalWritesConfigToml(t *testing.T) {
	original := deps
	defer func() { deps = original }()

	homeDir := t.TempDir()
	deps = commandDeps{
		userHomeDir: func() string { return homeDir },
		promptInit: func(global bool) (promptAnswers, error) {
			if !global {
				t.Fatal("expected global init")
			}
			return promptAnswers{
				JDK:       "17",
				MavenHome: `D:/mavens/apache-maven-3.9.6`,
				Settings:  `D:/users/demo/.m2/settings.xml`,
				LocalRepo: `D:/users/demo/.m2/repository`,
			}, nil
		},
	}

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"init", "--global"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content, err := os.ReadFile(filepath.Join(homeDir, ".jmvn", "config.toml"))
	if err != nil {
		t.Fatalf("read global config: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, `[defaults]`) {
		t.Fatalf("expected defaults section, got %q", text)
	}
	if !strings.Contains(text, `jdk = "17"`) {
		t.Fatalf("expected jdk field, got %q", text)
	}
	if !strings.Contains(text, `maven_home = "D:/mavens/apache-maven-3.9.6"`) {
		t.Fatalf("expected maven_home field, got %q", text)
	}
}
