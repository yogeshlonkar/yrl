package cmd

import (
	"errors"
	"io/ioutil"
	syslog "log"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var (
	version = "0.0.0" // populated at compile time
	commit  = "HEAD"  // populated at compile time
)

const (
	logSamplePerMinute = 29 * 60 * 23
	maxLogSizeInKb     = 1024 * 20
)

func NewApp() *cli.App {
	return &cli.App{
		Name:                 "yrl",
		Usage:                "personal utility for yogesh",
		Version:              version,
		Action:               mainAction,
		EnableBashCompletion: true,
		Authors: []*cli.Author{
			{
				Name:  "Yogesh Lonkar",
				Email: "yogesh@lonkar.org",
			},
		},
		Before: func(c *cli.Context) error {
			syslog.SetOutput(os.Stderr)
			lvl := zerolog.WarnLevel
			f := os.Stderr
			if c.Bool("trace") {
				lvl = zerolog.TraceLevel
			} else if c.Bool("debug") {
				lvl = zerolog.DebugLevel
			}
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: f}).Level(lvl)
			syslog.SetFlags(0)
			syslog.SetOutput(ioutil.Discard)
			log.Debug().Str("Commit", commit).Str("Version", version).Send()
			return nil
		},
		Commands: []*cli.Command{
			gitStatus(),
			gmailSettings(),
			openCommand(),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:   "debug",
				Usage:  "enables debug level logs written to stderr",
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:   "trace",
				Usage:  "enables trace level logs written to stderr",
				Hidden: true,
			},
		},
	}
}

func mainAction(_ *cli.Context) error {
	return errors.New("not implemented yet")
}
