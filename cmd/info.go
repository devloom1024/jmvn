package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"jmvn/internal/cli"

	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show resolved jmvn configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := deps.getwd()
			if err != nil {
				return err
			}
			globalCfg, err := deps.loadGlobal(filepath.Join(deps.userHomeDir(), ".jmvn", "config.toml"))
			if err != nil {
				return err
			}
			projectCfg, err := deps.loadProject(filepath.Join(cwd, ".jmvn.toml"))
			if err != nil {
				return err
			}
			resolved, err := deps.resolve(cli.Options{}, projectCfg, globalCfg, deps.lookupEnv(), cwd)
			if err != nil {
				return err
			}
			resolved.ProjectDir = cwd
			globalPath := filepath.Join(deps.userHomeDir(), ".jmvn", "config.toml")
			projectPath := filepath.Join(cwd, ".jmvn.toml")
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n%s %s [%s]\n%s %s [%s]\n%s %s [%s]\n%s %s [%s]\n%s %s\n%s\n  %s %s  %s\n  %s %s  %s\n",
				styledHeader("jmvn 配置解析"),
				styledLabel("JDK       "), resolved.JavaCmd, resolved.JavaCmdSource,
				styledLabel("Maven     "), resolved.MavenHome, resolved.MavenHomeSource,
				styledLabel("Settings  "), resolved.Settings, resolved.SettingsSource,
				styledLabel("Local Repo"), resolved.LocalRepo, resolved.LocalRepoSource,
				styledLabel("Project Dir"), resolved.ProjectDir,
				styledHeader("Config Files:"),
				styledLabel("Global: "), globalPath, styledStatus(fileExists(globalPath)),
				styledLabel("Project:"), projectPath, styledStatus(fileExists(projectPath)),
			)
			return err
		},
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
