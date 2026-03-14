package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"jmvn/internal/config"
)

func BuildCommand(cfg config.ResolvedConfig, mavenArgs []string) (*exec.Cmd, error) {
	classworldsJar, err := findClassworldsJar(cfg.MavenHome)
	if err != nil {
		return nil, err
	}

	args := []string{}
	jvmConfigArgs, err := readJvmConfig(cfg.ProjectDir)
	if err != nil {
		return nil, err
	}
	args = append(args, jvmConfigArgs...)
	if isMaven4(cfg.MavenHome) {
		args = append(args, "--enable-native-access=ALL-UNNAMED")
	}

	args = append(args,
		"-classpath", classworldsJar,
		fmt.Sprintf("-Dclassworlds.conf=%s", filepath.Join(cfg.MavenHome, "bin", "m2.conf")),
		fmt.Sprintf("-Dmaven.home=%s", cfg.MavenHome),
		fmt.Sprintf("-Dmaven.multiModuleProjectDirectory=%s", cfg.ProjectDir),
		"org.codehaus.plexus.classworlds.launcher.Launcher",
	)
	if cfg.Settings != "" {
		args = append(args, "--settings", cfg.Settings)
	}
	if cfg.LocalRepo != "" {
		args = append(args, fmt.Sprintf("-Dmaven.repo.local=%s", cfg.LocalRepo))
	}
	args = append(args, mavenArgs...)

	cmd := exec.Command(cfg.JavaCmd, args...)
	cmd.Dir = cfg.ProjectDir
	return cmd, nil
}

func readJvmConfig(projectDir string) ([]string, error) {
	if projectDir == "" {
		return nil, nil
	}
	path := filepath.Join(projectDir, ".mvn", "jvm.config")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	args := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		args = append(args, trimmed)
	}
	return args, nil
}

func isMaven4(mavenHome string) bool {
	return strings.Contains(filepath.Base(mavenHome), "maven-4.")
}

func findClassworldsJar(mavenHome string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(mavenHome, "boot", "plexus-classworlds-*.jar"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return filepath.Join(mavenHome, "boot", "plexus-classworlds.jar"), nil
	}
	return matches[0], nil
}
