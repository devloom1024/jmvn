package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered JDK and Maven toolchains",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalCfg, err := deps.loadGlobal(filepath.Join(deps.userHomeDir(), ".jmvn", "config.toml"))
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "已注册的 JDK:"); err != nil {
				return err
			}
			for _, version := range sortedKeys(globalCfg.JDKs) {
				line := fmt.Sprintf("  %s  %s  %s", version, globalCfg.JDKs[version], existsMarker(globalCfg.JDKs[version]))
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
					return err
				}
			}

			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "\n已注册的 Maven:"); err != nil {
				return err
			}
			for _, version := range sortedKeys(globalCfg.Mavens) {
				line := fmt.Sprintf("  %s  %s  %s", version, globalCfg.Mavens[version], existsMarker(globalCfg.Mavens[version]))
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func existsMarker(path string) string {
	if _, err := os.Stat(path); err == nil {
		return "✓"
	}
	return "✗"
}
