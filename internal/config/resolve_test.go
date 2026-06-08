package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolve_PrefersEnvThenProjectThenGlobal(t *testing.T) {
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
	env := map[string]string{
		"JMVN_JDK": "17",
	}

	resolved, err := Resolve(projectCfg, globalCfg, env, `D:/work/demo`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resolved.JavaCmd != expectedJavaCmd(`/jdks/jdk-17`) {
		t.Fatalf("unexpected java command: %q", resolved.JavaCmd)
	}
	if resolved.JavaCmdSource != "env" {
		t.Fatalf("expected java source env, got %q", resolved.JavaCmdSource)
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
	resolved, err := Resolve(ProjectConfig{}, GlobalConfig{}, map[string]string{
		"JAVA_HOME":  `/env/jdk-21`,
		"MAVEN_HOME": `/env/maven-3.9`,
	}, `D:/work/demo`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resolved.JavaCmd != expectedJavaCmd(`/env/jdk-21`) {
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

func TestResolve_JmvnEnvOverridesProject(t *testing.T) {
	globalCfg := GlobalConfig{
		JDKs:   map[string]string{"21": `/jdks/jdk-21`},
		Mavens: map[string]string{"3.9": `/mavens/maven-3.9`},
	}
	projectCfg := ProjectConfig{
		JDK:       "17",
		Maven:     "3.6",
		Settings:  `./project-settings.xml`,
		LocalRepo: `./project-repo`,
	}
	env := map[string]string{
		"JMVN_JDK":        "21",
		"JMVN_MAVEN":      "3.9",
		"JMVN_SETTINGS":   `/env/settings.xml`,
		"JMVN_LOCAL_REPO": `/env/repo`,
	}

	resolved, err := Resolve(projectCfg, globalCfg, env, `D:/work/demo`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resolved.JavaCmdSource != "env" {
		t.Fatalf("expected java source env, got %q", resolved.JavaCmdSource)
	}
	if resolved.MavenHomeSource != "env" {
		t.Fatalf("expected maven source env, got %q", resolved.MavenHomeSource)
	}
	if resolved.SettingsSource != "env" {
		t.Fatalf("expected settings source env, got %q", resolved.SettingsSource)
	}
	if resolved.LocalRepoSource != "env" {
		t.Fatalf("expected local repo source env, got %q", resolved.LocalRepoSource)
	}
}

func TestResolve_JmvnMavenHomeEnv(t *testing.T) {
	resolved, err := Resolve(ProjectConfig{}, GlobalConfig{}, map[string]string{
		"JMVN_MAVEN_HOME": `/env/custom-maven`,
	}, `D:/work/demo`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resolved.MavenHome != filepath.Clean(`/env/custom-maven`) {
		t.Fatalf("unexpected maven home: %q", resolved.MavenHome)
	}
	if resolved.MavenHomeSource != "env" {
		t.Fatalf("expected maven source env, got %q", resolved.MavenHomeSource)
	}
}

func expectedJavaCmd(jdkHome string) string {
	javaName := "java"
	if runtime.GOOS == "windows" {
		javaName = "java.exe"
	}
	return filepath.Clean(filepath.Join(jdkHome, "bin", javaName))
}
