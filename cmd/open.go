package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pkg/browser"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/yogeshlonkar/yrl/pkg/database"
	"github.com/yogeshlonkar/yrl/pkg/git"
)

func openCommand() *cli.Command {
	return &cli.Command{
		Name:      "open",
		Usage:     "open repo",
		UsageText: fmt.Sprintf("open github location of %s", workingDirectory),
		ArgsUsage: "[path]",
		Action: func(ctx *cli.Context) error {
			log.Debug().Msg("Started open")
			absPath := workingDirectory
			if ctx.Args().Len() == 1 {
				absPath = ctx.Args().First()
			} else if ctx.Args().Len() > 1 {
				return errors.New("expected single argument")
			}
			db := database.New()
			defer db.Close()
			url, err := db.Git().GetRemoteURL(absPath)
			var toOpen string
			if err != nil {
				if url == "" {
					log.Debug().Msg("no url in database")
					if is, err := git.IsInsideWorkTree(absPath); err == nil && is {
						remotes := git.GetRemotes(absPath)
						log.Debug().Str("remotes", strings.Join(remotes, ",")).Msg("got remotes")
						if len(remotes) > 0 {
							toOpen = strings.Replace(remotes[0], "git@github.com:", "https://github.com/", 1)
						}
					}
				} else {
					toOpen = strings.Replace(url, "git@github.com:", "https://github.com/", 1)
				}
			} else {
				log.Warn().Err(err).Msg("could not get remote URL from database")
			}
			log.Info().Str("toOpen", toOpen).Send()
			if toOpen != "" {
				browser.OpenURL(toOpen)
			} else {
				log.Info().Msg("Nothing to open")
			}
			return nil
		},
	}
}
