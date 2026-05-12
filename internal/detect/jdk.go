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
	Entries map[string]string
}

func (p *pomProperties) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p.Entries = make(map[string]string)
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			var val string
			if err := d.DecodeElement(&val, &t); err != nil {
				return err
			}
			p.Entries[t.Name.Local] = val
		case xml.EndElement:
			if t == start.End() {
				return nil
			}
		}
	}
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
	props := project.Properties.Entries

	if v := resolveProperty(props["maven.compiler.release"], props); v != "" {
		return normalizeVersion(v)
	}
	if v := resolveProperty(props["maven.compiler.source"], props); v != "" {
		return normalizeVersion(v)
	}
	if v := resolveProperty(props["java.version"], props); v != "" {
		return normalizeVersion(v)
	}
	for _, plugin := range project.Build.Plugins {
		if plugin.ArtifactID != "maven-compiler-plugin" {
			continue
		}
		if v := resolveProperty(plugin.Configuration.Release, props); v != "" {
			return normalizeVersion(v)
		}
		if v := resolveProperty(plugin.Configuration.Source, props); v != "" {
			return normalizeVersion(v)
		}
	}
	return ""
}

var placeholderRe = regexp.MustCompile(`^\$\{([^}]+)\}$`)

func resolveProperty(value string, props map[string]string) string {
	if value == "" {
		return ""
	}
	matches := placeholderRe.FindStringSubmatch(value)
	if matches == nil {
		return value
	}
	key := matches[1]
	resolved, ok := props[key]
	if !ok {
		return ""
	}
	// 简单循环引用保护：解析后的值如果仍是占位符，放弃解析
	if placeholderRe.MatchString(resolved) {
		return ""
	}
	return resolved
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
	// 安全网：防御性过滤未解析的占位符（正常流程已在 resolveProperty 中处理）
	if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
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
