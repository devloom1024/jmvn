package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [maven-args...]",
		Short: "Run Maven with the resolved JDK",
		Long: `Run Maven with the selected JDK, settings.xml and local repository.

This is the explicit form of the root command. Both 'jmvn clean install'
and 'jmvn run clean install' are equivalent.`,
		Example: strings.Join([]string{
			"  jmvn run clean install",
			"  jmvn run --jdk 11 test",
			"  jmvn run --dry-run package",
		}, "\n"),
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			state := executionStateFromContext(cmd.Root().Context())
			state.mavenArgs = append([]string(nil), args...)
			return runRootCommand(cmd, state)
		},
	}
	cmd.Flags().SetInterspersed(false)
	return cmd
}
