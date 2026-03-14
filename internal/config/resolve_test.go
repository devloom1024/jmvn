package config

import (
	"path/filepath"
	"testing"

	"jmvn/internal/cli"
)

func TestResolve_PrefersCLIThenProjectThenGlobal(t *testing.T) {
	globalCfg := GlobalConfig{
		Defaults: DefaultsConfig{
			JDK:       "17",
			MavenHome: `/global/maven-default`,
			Settings:  `~/global-settings.xml`,
			LocalRepo: `~/global-repo`,
		},
		JDKs: map[string]string{
			"11": `/jdks/jdk-11`,
			"17": `/jdks/jdk-17`,
		},
		Mavens: map[string]string{
			"3.6": `/mavens/apache-maven-3.6`,
			"3.9": `/mavens/apache-maven-3.9`,
		},
	}
	projectCfg := ProjectConfig{
		JDK:       "11",
		Maven:     "3.6",
		Settings:  `./project-settings.xml`,
		LocalRepo: `./project-repo`,
	}
	cliOpts := cli.Options{
		JDK: "17",
	}

	resolved, err := Resolve(cliOpts, projectCfg, globalCfg, map[string]string{}, `D:/work/demo`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resolved.JavaCmd != filepath.Clean(`/jdks/jdk-17/bin/java`) {
		t.Fatalf("unexpected java command: %q", resolved.JavaCmd)
	}
	if resolved.JavaCmdSource != "cli" {
		t.Fatalf("expected java source cli, got %q", resolved.JavaCmdSource)
	}
	if resolved.MavenHome != filepath.Clean(`/mavens/apache-maven-3.6`) {
		t.Fatalf("unexpected maven home: %q", resolved.MavenHome)
	}
	if resolved.MavenHomeSource != "project" {
		t.Fatalf("expected maven source project, got %q", resolved.MavenHomeSource)
	}
	if resolved.Settings != filepath.Clean(`D:/work/demo/project-settings.xml`) {
		t.Fatalf("unexpected settings path: %q", resolved.Settings)
	}
	if resolved.SettingsSource != "project" {
		t.Fatalf("expected settings source project, got %q", resolved.SettingsSource)
	}
	if resolved.LocalRepo != filepath.Clean(`D:/work/demo/project-repo`) {
		t.Fatalf("unexpected local repo path: %q", resolved.LocalRepo)
	}
	if resolved.LocalRepoSource != "project" {
		t.Fatalf("expected local repo source project, got %q", resolved.LocalRepoSource)
	}
}

func TestResolve_FallsBackToEnvWhenConfigMissing(t *testing.T) {
	resolved, err := Resolve(cli.Options{}, ProjectConfig{}, GlobalConfig{}, map[string]string{
		"JAVA_HOME":  `/env/jdk-21`,
		"MAVEN_HOME": `/env/maven-3.9`,
	}, `D:/work/demo`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resolved.JavaCmd != filepath.Clean(`/env/jdk-21/bin/java`) {
		t.Fatalf("unexpected java command: %q", resolved.JavaCmd)
	}
	if resolved.JavaCmdSource != "env" {
		t.Fatalf("expected java source env, got %q", resolved.JavaCmdSource)
	}
	if resolved.MavenHome != filepath.Clean(`/env/maven-3.9`) {
		t.Fatalf("unexpected maven home: %q", resolved.MavenHome)
	}
	if resolved.MavenHomeSource != "env" {
		t.Fatalf("expected maven source env, got %q", resolved.MavenHomeSource)
	}
}
