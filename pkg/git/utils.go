package git

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

var ErrNotAGitRepo = errors.New("not a git repo")

const notRepoStatus string = "exit status 128"

func IsInsideWorkTree(cwd string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 128 {
					return false, ErrNotAGitRepo
				}
			}
		}
		if cmd.ProcessState.String() == notRepoStatus {
			return false, ErrNotAGitRepo
		}
		return false, err
	}
	return strconv.ParseBool(strings.TrimSpace(string(out)))
}
