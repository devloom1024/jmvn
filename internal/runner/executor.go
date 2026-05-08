package runner

import (
	"fmt"
	"os/exec"
)

type ExitCodeError struct {
	Code int
	Err  error
}

func (e ExitCodeError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("process exited with code %d", e.Code)
	}
	return e.Err.Error()
}

func (e ExitCodeError) Unwrap() error {
	return e.Err
}

func (e ExitCodeError) ExitCode() int {
	return e.Code
}

func Exec(cmd *exec.Cmd) error {
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if ok := AsExitError(err, &exitErr); ok {
			return ExitCodeError{Code: exitErr.ExitCode(), Err: err}
		}
		return err
	}
	return nil
}

func AsExitError(err error, target **exec.ExitError) bool {
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}
	*target = exitErr
	return true
}
