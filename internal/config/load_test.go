package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGlobal_ReadsDefaultsAndMaps(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "config.toml")
	content := `[defaults]
jdk = "17"
maven_home = "/opt/maven"
[jdks]
"17" = "/opt/jdk17"
[mavens]
"3.9" = "/opt/maven-3.9"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadGlobal(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Defaults.JDK != "17" {
		t.Fatalf("expected default JDK 17, got %q", cfg.Defaults.JDK)
	}
	if cfg.JDKs["17"] != "/opt/jdk17" {
		t.Fatalf("expected JDK map entry, got %#v", cfg.JDKs)
	}
	if cfg.Mavens["3.9"] != "/opt/maven-3.9" {
		t.Fatalf("expected Maven map entry, got %#v", cfg.Mavens)
	}
}

func TestLoadProject_ReadsOverrides(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, ".jmvn.toml")
	content := `jdk = "11"
maven = "3.6"
settings = "./settings.xml"
local_repo = "./repo"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write project config: %v", err)
	}

	cfg, err := LoadProject(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.JDK != "11" || cfg.Maven != "3.6" {
		t.Fatalf("unexpected project config: %#v", cfg)
	}
	if cfg.Settings != "./settings.xml" || cfg.LocalRepo != "./repo" {
		t.Fatalf("unexpected path fields: %#v", cfg)
	}
}
