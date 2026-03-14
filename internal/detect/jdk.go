package detect

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type pomProject struct {
	Properties pomProperties `xml:"properties"`
	Build      pomBuild      `xml:"build"`
}

type pomProperties struct {
	CompilerRelease string `xml:"maven.compiler.release"`
	CompilerSource  string `xml:"maven.compiler.source"`
	JavaVersion     string `xml:"java.version"`
}

type pomBuild struct {
	Plugins []pomPlugin `xml:"plugins>plugin"`
}

type pomPlugin struct {
	ArtifactID    string                 `xml:"artifactId"`
	Configuration pomPluginConfiguration `xml:"configuration"`
}

type pomPluginConfiguration struct {
	Release string `xml:"release"`
	Source  string `xml:"source"`
}

func DetectJDKVersion(projectDir string) string {
	if version := detectFromJavaVersion(projectDir); version != "" {
		return version
	}
	if version := detectFromPom(projectDir); version != "" {
		return version
	}
	if version := detectFromMvnJdkConfig(projectDir); version != "" {
		return version
	}
	return ""
}

func detectFromJavaVersion(projectDir string) string {
	path := filepath.Join(projectDir, ".java-version")
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return normalizeVersion(strings.TrimSpace(string(content)))
}

func detectFromPom(projectDir string) string {
	path := filepath.Join(projectDir, "pom.xml")
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var project pomProject
	if err := xml.Unmarshal(content, &project); err != nil {
		return ""
	}
	if project.Properties.CompilerRelease != "" {
		return normalizeVersion(project.Properties.CompilerRelease)
	}
	if project.Properties.CompilerSource != "" {
		return normalizeVersion(project.Properties.CompilerSource)
	}
	if project.Properties.JavaVersion != "" {
		return normalizeVersion(project.Properties.JavaVersion)
	}
	for _, plugin := range project.Build.Plugins {
		if plugin.ArtifactID != "maven-compiler-plugin" {
			continue
		}
		if plugin.Configuration.Release != "" {
			return normalizeVersion(plugin.Configuration.Release)
		}
		if plugin.Configuration.Source != "" {
			return normalizeVersion(plugin.Configuration.Source)
		}
	}
	return ""
}

func detectFromMvnJdkConfig(projectDir string) string {
	path := filepath.Join(projectDir, ".mvn", "jdk.config")
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return ""
	}
	re := regexp.MustCompile(`(?i)(?:jdk|java)[-_]?([0-9]+(?:\.[0-9]+)*)`)
	if matches := re.FindStringSubmatch(trimmed); len(matches) >= 2 {
		return normalizeVersion(matches[1])
	}
	return normalizeVersion(filepath.Base(trimmed))
}

func normalizeVersion(v string) string {
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "1.") {
		parts := strings.Split(v, ".")
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	parts := strings.Split(v, ".")
	return parts[0]
}
