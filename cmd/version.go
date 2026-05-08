package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var buildVersion = "dev"

func SetBuildVersion(version string) {
	if version != "" {
		buildVersion = version
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print jmvn version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "jmvn v%s (%s, %s/%s)\n", buildVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH)
			return err
		},
	}
}
