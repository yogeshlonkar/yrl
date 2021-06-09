package cmd

import (
	"errors"
	"io/ioutil"
	syslog "log"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var (
	version = "0.0.0" // populated at compile time
	commit  = "HEAD"  // populated at compile time
)

const (
	logFile            = "/tmp/.git_status_bar.log"
	logSamplePerMinute = 29 * 60 * 23
	maxLogSizeInKb     = 1024 * 20
)

func NewApp() *cli.App {
	return &cli.App{
		Name:    "yrl",
		Usage:   "personal utility for yogesh",
		Version: version,
		Action:  mainAction,
		Authors: []*cli.Author{
			{
				Name:  "Yogesh Lonkar",
				Email: "yogesh@lonkar.org",
			},
		},
		Before: func(c *cli.Context) error {
			var err error
			syslog.SetOutput(os.Stderr)
			lvl := zerolog.WarnLevel
			if c.Bool("debug") {
				lvl = zerolog.DebugLevel
			} else if c.Bool("trace") {
				lvl = zerolog.TraceLevel
			} else {
				if time.Now().Unix()%logSamplePerMinute == 0 {
					lvl = zerolog.DebugLevel
				}
			}
			f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
			if err != nil {
				syslog.Printf("Could not create log file %q", err)
			} else {
				if s, _ := f.Stat(); s != nil && s.Size() > maxLogSizeInKb {
					err = f.Truncate(0)
					if err != nil {
						syslog.Printf("Could not truncate log file %q", err)
					}
					_, err = f.Seek(0, 0)
					if err != nil {
						syslog.Printf("Could not seek file to strat %q", err)
					}
				}
				log.Logger = log.Output(zerolog.ConsoleWriter{Out: f}).Level(lvl)
			}
			syslog.SetFlags(0)
			syslog.SetOutput(ioutil.Discard)
			log.Debug().Str("Commit", commit).Str("Version", version).Send()
			return nil
		},
		Commands: []*cli.Command{
			gitStatus(),
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

func mainAction(c *cli.Context) error {
	return errors.New("not implemented yet")
}
