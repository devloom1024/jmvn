package cmd

import (
	"os/exec"

	"jmvn/internal/config"
)

func baseTestDeps() commandDeps {
	return commandDeps{
		getwd:            func() (string, error) { return `D:/test`, nil },
		userHomeDir:      func() string { return `D:/home` },
		loadGlobal:       func(string) (config.GlobalConfig, error) { return config.GlobalConfig{}, nil },
		loadProject:      func(string) (config.ProjectConfig, error) { return config.ProjectConfig{}, nil },
		detectJDKVersion: func(string) string { return "" },
		resolve: func(config.ProjectConfig, config.GlobalConfig, map[string]string, string) (config.ResolvedConfig, error) {
			return config.ResolvedConfig{JavaCmd: `java`, MavenHome: `maven`, ProjectDir: `D:/test`}, nil
		},
		validateResolved: func(config.ResolvedConfig) error { return nil },
		buildCommand: func(cfg config.ResolvedConfig, mavenArgs []string) (*exec.Cmd, error) {
			cmd := exec.Command(cfg.JavaCmd, mavenArgs...)
			cmd.Dir = cfg.ProjectDir
			return cmd, nil
		},
		lookupEnv:      func() map[string]string { return map[string]string{} },
		promptInit:     func(bool) (promptAnswers, error) { return promptAnswers{}, nil },
		executeCommand: func(*exec.Cmd) error { return nil },
	}
}
