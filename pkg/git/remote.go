package git

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

var (
	gitRemoteRegex = regexp.MustCompile(`\w+\s+([^\/]+\/[^\.]+)\.git\s.*`)
)

func GetRemotes(cwd string) (remotes []string) {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = cwd
	o, err := cmd.Output()
	output := string(o)
	log.Debug().Str("output", output)
	if err == nil {
		for _, line := range strings.Split(output, "\n") {
			if gitRemoteRegex.Match([]byte(line)) {
				groups := gitRemoteRegex.FindAllSubmatch([]byte(line), -1)
				remotes = append(remotes, string(groups[0][1]))
			}
		}
	} else {
		log.Warn().Err(err).Msg("could not find remotes for repo")
	}
	return
}
