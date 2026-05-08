package validate

import (
	"fmt"
	"os"
	"path/filepath"

	"jmvn/internal/config"
)

func ResolvedConfig(cfg config.ResolvedConfig) error {
	if cfg.JavaCmd == "" {
		return fmt.Errorf("java command is empty")
	}
	if _, err := os.Stat(cfg.JavaCmd); err != nil {
		return fmt.Errorf("java executable not found: %s", cfg.JavaCmd)
	}
	if cfg.MavenHome == "" {
		return fmt.Errorf("maven home is empty")
	}
	if info, err := os.Stat(cfg.MavenHome); err != nil || !info.IsDir() {
		return fmt.Errorf("maven home not found: %s", cfg.MavenHome)
	}
	if _, err := os.Stat(filepath.Join(cfg.MavenHome, "bin", "m2.conf")); err != nil {
		return fmt.Errorf("m2.conf not found under maven home: %s", cfg.MavenHome)
	}
	matches, err := filepath.Glob(filepath.Join(cfg.MavenHome, "boot", "plexus-classworlds-*.jar"))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("plexus-classworlds jar not found under maven home: %s", cfg.MavenHome)
	}
	if cfg.Settings != "" {
		if _, err := os.Stat(cfg.Settings); err != nil {
			return fmt.Errorf("settings.xml not found: %s", cfg.Settings)
		}
	}
	return nil
}
