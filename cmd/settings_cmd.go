package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/yogeshlonkar/yrl/pkg/gmail"
)

func gmailSettings() *cli.Command {
	return &cli.Command{
		Name:  "gmail-settings",
		Usage: "update gmail filters, labels",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:     "credentials",
				Usage:    "google application credentials (json) file path",
				EnvVars:  []string{"GOOGLE_APPLICATION_CREDENTIALS"},
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"d"},
				Usage:   "just print probable changes to settings",
			},
			&cli.BoolFlag{
				Name:  "push",
				Usage: "push local settings if diverged in gmail",
				Value: false,
			},
			&cli.PathFlag{
				Name:  "settings",
				Usage: "yaml file with account settings",
				Value: "gmail-settings.yml",
			},
			&cli.StringFlag{
				Name:     "user",
				Usage:    "google user email address",
				EnvVars:  []string{"GOOGLE_USER"},
				Required: true,
			},
		},
		Action: settingsAction,
	}
}

func settingsAction(c *cli.Context) error {
	user := c.String("user")
	credentials := c.Path("credentials")
	dryRun := c.Bool("dry-run")
	push := c.Bool("push")
	settingsFile := c.Path("settings")
	log.Info().Str("credentials", credentials).Str("user", user).Msg("Started sync gmail-settings")
	gmailSvc := gmail.NewService(c.Context, credentials, user)
	syncSvc, closeSvc := gmail.NewSyncService(settingsFile, gmailSvc, dryRun, push)
	defer closeSvc()
	if err := syncSvc.SyncLabels(); err != nil {
		return err
	}
	if err := syncSvc.SyncFilters(); err != nil {
		return err
	}
	return nil
}
