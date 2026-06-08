package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func handleInfo(cmd *cobra.Command) error {
	ctx, resolved, err := resolveCommandConfig()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n%s %s [%s]\n%s %s [%s]\n%s %s [%s]\n%s %s [%s]\n%s %s\n%s\n  %s %s  %s\n  %s %s  %s\n",
		styledHeader("jmvn 配置解析"),
		styledLabel("JDK       "), resolved.JavaCmd, resolved.JavaCmdSource,
		styledLabel("Maven     "), resolved.MavenHome, resolved.MavenHomeSource,
		styledLabel("Settings  "), resolved.Settings, resolved.SettingsSource,
		styledLabel("Local Repo"), resolved.LocalRepo, resolved.LocalRepoSource,
		styledLabel("Project Dir"), resolved.ProjectDir,
		styledHeader("Config Files:"),
		styledLabel("Global: "), ctx.globalPath, styledStatus(fileExists(ctx.globalPath)),
		styledLabel("Project:"), ctx.projectPath, styledStatus(fileExists(ctx.projectPath)),
	)
	return err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
