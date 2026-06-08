package main

import (
	"errors"
	"fmt"
	"os"

	"jmvn/cmd"
)

var version = "dev"

func main() {
	cmd.SetBuildVersion(version)
	rootCmd := cmd.NewRootCmd()
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
