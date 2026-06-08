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

func handleVersion(cmd *cobra.Command) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "jmvn v%s (%s, %s/%s)\n", buildVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return err
}
