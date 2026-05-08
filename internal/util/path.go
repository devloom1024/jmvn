package util

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func ResolvePath(raw string, baseDir string) string {
	if raw == "" {
		return ""
	}

	resolved := raw
	if strings.HasPrefix(resolved, "~/") || resolved == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if resolved == "~" {
				resolved = home
			} else {
				resolved = filepath.Join(home, resolved[2:])
			}
		}
	}

	if filepath.IsAbs(resolved) {
		return filepath.Clean(resolved)
	}
	if baseDir == "" {
		return filepath.Clean(resolved)
	}
	return filepath.Clean(filepath.Join(baseDir, resolved))
}

func ResolveJavaBinary(jdkHome string) string {
	javaName := "java"
	if runtime.GOOS == "windows" {
		javaName = "java.exe"
	}
	return filepath.Clean(filepath.Join(jdkHome, "bin", javaName))
}
