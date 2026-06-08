package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

func handleList(cmd *cobra.Command) error {
	globalCfg, err := deps.loadGlobal(filepath.Join(deps.userHomeDir(), ".jmvn", "config.toml"))
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), styledHeader("已注册的 JDK:")); err != nil {
		return err
	}
	for _, version := range sortedKeys(globalCfg.JDKs) {
		line := fmt.Sprintf("  %s  %s  %s", styledLabel(version), globalCfg.JDKs[version], styledMarker(pathExists(globalCfg.JDKs[version])))
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "\n"+styledHeader("已注册的 Maven:")); err != nil {
		return err
	}
	for _, version := range sortedKeys(globalCfg.Mavens) {
		line := fmt.Sprintf("  %s  %s  %s", styledLabel(version), globalCfg.Mavens[version], styledMarker(pathExists(globalCfg.Mavens[version])))
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
			return err
		}
	}
	return nil
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
