package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"jmvn/internal/config"

	"github.com/spf13/cobra"
)

func handleInit(cmd *cobra.Command, args []string) error {
	global := false
	for _, arg := range args {
		if arg == "--global" {
			global = true
			break
		}
	}

	answers, err := deps.promptInit(global)
	if err != nil {
		return err
	}
	if global {
		return writeGlobalConfig(answers)
	}
	return writeProjectConfig(answers)
}

func writeProjectConfig(answers promptAnswers) error {
	cwd, err := deps.getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(cwd, ".jmvn.toml")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("configuration already exists: %s", path)
	}
	content := config.RenderProjectConfig(answers.JDK, answers.Maven, answers.Settings, answers.LocalRepo)
	return os.WriteFile(path, []byte(content), 0o644)
}

func writeGlobalConfig(answers promptAnswers) error {
	homeDir := deps.userHomeDir()
	configDir := filepath.Join(homeDir, ".jmvn")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(configDir, "config.toml")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("configuration already exists: %s", path)
	}
	content := config.RenderGlobalConfig(answers.JDK, answers.JDKHome, answers.Maven, answers.MavenHome, answers.Settings, answers.LocalRepo)
	return os.WriteFile(path, []byte(content), 0o644)
}

func defaultPromptInit(global bool) (promptAnswers, error) {
	reader := bufio.NewReader(os.Stdin)
	ask := func(label string) (string, error) {
		if _, err := fmt.Fprint(os.Stdout, label); err != nil {
			return "", err
		}
		value, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(value), nil
	}

	answers := promptAnswers{}
	var err error
	if answers.JDK, err = ask("JDK version: "); err != nil {
		return promptAnswers{}, err
	}
	if global {
		if answers.JDKHome, err = ask("JDK home for that version: "); err != nil {
			return promptAnswers{}, err
		}
		if answers.Maven, err = ask("Default Maven version: "); err != nil {
			return promptAnswers{}, err
		}
		if answers.MavenHome, err = ask("Default Maven home: "); err != nil {
			return promptAnswers{}, err
		}
	} else {
		if answers.Maven, err = ask("Maven version: "); err != nil {
			return promptAnswers{}, err
		}
	}
	if answers.Settings, err = ask("Settings path: "); err != nil {
		return promptAnswers{}, err
	}
	if answers.LocalRepo, err = ask("Local repo path: "); err != nil {
		return promptAnswers{}, err
	}
	return answers, nil
}
