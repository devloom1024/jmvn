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
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "jmvn 配置解析\nJDK        %s [%s]\nMaven      %s [%s]\nSettings   %s [%s]\nLocal Repo %s [%s]\nProject Dir %s\nConfig Files:\n  Global:  %s  %s\n  Project: %s  %s\n",
				resolved.JavaCmd,
				resolved.JavaCmdSource,
				resolved.MavenHome,
				resolved.MavenHomeSource,
				resolved.Settings,
				resolved.SettingsSource,
				resolved.LocalRepo,
				resolved.LocalRepoSource,
				resolved.ProjectDir,
				globalPath,
				fileStatus(globalPath),
				projectPath,
				fileStatus(projectPath),
			)
			return err
		},
	}
}

func fileStatus(path string) string {
	if _, err := os.Stat(path); err == nil {
		return "found"
	}
	return "missing"
}
