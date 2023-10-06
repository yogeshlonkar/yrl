package git

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

var ErrNotInAGitRepo = errors.New("not in a git repo")

const notRepoStatus string = "exit status 128"

func IsInsideWorkTree(cwd string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 128 {
					return false, ErrNotInAGitRepo
				}
			}
		}
		if cmd.ProcessState.String() == notRepoStatus {
			return false, ErrNotInAGitRepo
		}
		return false, err
	}
	return strconv.ParseBool(strings.TrimSpace(string(out)))
}
