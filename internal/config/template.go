package config

import (
	"fmt"
	"strings"
)

func RenderProjectConfig(jdk string, maven string, settings string, localRepo string) string {
	lines := make([]string, 0, 4)
	if jdk != "" {
		lines = append(lines, fmt.Sprintf("jdk = %q", jdk))
	}
	if maven != "" {
		lines = append(lines, fmt.Sprintf("maven = %q", maven))
	}
	if settings != "" {
		lines = append(lines, fmt.Sprintf("settings = %q", settings))
	}
	if localRepo != "" {
		lines = append(lines, fmt.Sprintf("local_repo = %q", localRepo))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func RenderGlobalConfig(jdk string, mavenHome string, settings string, localRepo string) string {
	lines := []string{"[defaults]"}
	if jdk != "" {
		lines = append(lines, fmt.Sprintf("jdk = %q", jdk))
	}
	if mavenHome != "" {
		lines = append(lines, fmt.Sprintf("maven_home = %q", mavenHome))
	}
	if settings != "" {
		lines = append(lines, fmt.Sprintf("settings = %q", settings))
	}
	if localRepo != "" {
		lines = append(lines, fmt.Sprintf("local_repo = %q", localRepo))
	}
	lines = append(lines, "", "[jdks]", "", "[mavens]")
	return strings.Join(lines, "\n") + "\n"
}
