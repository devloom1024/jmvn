package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"jmvn/cmd"
)

var version = "dev"

func main() {
	cmd.SetBuildVersion(version)
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs(preprocessArgs(os.Args[1:]))
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCodeForError(err))
	}
}

type exitCoder interface {
	error
	ExitCode() int
}

func exitCodeForError(err error) int {
	if err == nil {
		return 0
	}
	var coded exitCoder
	if errors.As(err, &coded) {
		return coded.ExitCode()
	}
	return 1
}

func preprocessArgs(args []string) []string {
	noValueFlags := map[string]bool{
		"--dry-run": true, "-n": true,
		"--verbose": true, "-v": true,
		"--help": true, "-h": true,
	}
	valueFlags := map[string]bool{
		"--jdk": true, "-j": true,
		"--maven": true, "-m": true,
		"--settings": true, "-s": true,
		"--local-repo": true, "-r": true,
	}
	subcommands := map[string]bool{
		"run": true, "version": true, "list": true, "info": true, "init": true,
		"help": true,
	}

	result := make([]string, 0, len(args))
	seenNonFlag := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			result = append(result, args[i:]...)
			break
		}

		if subcommands[arg] {
			result = append(result, arg)
			seenNonFlag = false
			continue
		}

		flagName := arg
		if idx := strings.Index(arg, "="); idx > 0 {
			flagName = arg[:idx]
		}

		if noValueFlags[flagName] {
			result = append(result, arg)
			continue
		}

		if valueFlags[flagName] {
			result = append(result, arg)
			if !strings.Contains(arg, "=") && i+1 < len(args) {
				i++
				result = append(result, args[i])
			}
			continue
		}

		if strings.HasPrefix(arg, "-") {
			if !seenNonFlag {
				result = append(result, "--")
			}
			result = append(result, args[i:]...)
			break
		}

		seenNonFlag = true
		result = append(result, arg)
	}

	return result
}
