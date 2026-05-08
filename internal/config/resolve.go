package config

import (
	"fmt"
	"path/filepath"

	"jmvn/internal/cli"
	"jmvn/internal/util"
)

func Resolve(cliOpts cli.Options, projectCfg ProjectConfig, globalCfg GlobalConfig, env map[string]string, projectDir string) (ResolvedConfig, error) {
	resolved := ResolvedConfig{}

	jdkVersion, jdkSource := firstNonEmpty(
		cliOpts.JDK, "cli",
		projectCfg.JDK, "project",
		globalCfg.Defaults.JDK, "global",
	)
	if jdkVersion != "" {
		jdkHome, ok := globalCfg.JDKs[jdkVersion]
		if !ok {
			return ResolvedConfig{}, fmt.Errorf("jdk %s is not registered", jdkVersion)
		}
		resolved.JavaCmd = util.ResolveJavaBinary(jdkHome)
		resolved.JavaCmdSource = jdkSource
	} else if javaHome := firstEnv(env, "JAVA_HOME"); javaHome != "" {
		resolved.JavaCmd = util.ResolveJavaBinary(javaHome)
		resolved.JavaCmdSource = "env"
	}

	mavenVersion, mavenVersionSource := firstNonEmpty(
		cliOpts.Maven, "cli",
		projectCfg.Maven, "project",
	)
	if mavenVersion != "" {
		mavenHome, ok := globalCfg.Mavens[mavenVersion]
		if !ok {
			return ResolvedConfig{}, fmt.Errorf("maven %s is not registered", mavenVersion)
		}
		resolved.MavenHome = filepath.Clean(mavenHome)
		resolved.MavenHomeSource = mavenVersionSource
	} else if globalCfg.Defaults.MavenHome != "" {
		resolved.MavenHome = filepath.Clean(globalCfg.Defaults.MavenHome)
		resolved.MavenHomeSource = "global"
	} else if mavenHome := firstEnv(env, "MAVEN_HOME", "M2_HOME"); mavenHome != "" {
		resolved.MavenHome = filepath.Clean(mavenHome)
		resolved.MavenHomeSource = "env"
	}

	if settings, source := firstNonEmpty(cliOpts.Settings, "cli", projectCfg.Settings, "project", globalCfg.Defaults.Settings, "global"); settings != "" {
		resolved.Settings = util.ResolvePath(settings, projectDir)
		resolved.SettingsSource = source
	}
	if repo, source := firstNonEmpty(cliOpts.LocalRepo, "cli", projectCfg.LocalRepo, "project", globalCfg.Defaults.LocalRepo, "global"); repo != "" {
		resolved.LocalRepo = util.ResolvePath(repo, projectDir)
		resolved.LocalRepoSource = source
	}

	return resolved, nil
}

func firstNonEmpty(values ...string) (string, string) {
	for i := 0; i+1 < len(values); i += 2 {
		if values[i] != "" {
			return values[i], values[i+1]
		}
	}
	return "", ""
}

func firstEnv(env map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := env[key]; value != "" {
			return value
		}
	}
	return ""
}
