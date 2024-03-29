package cmd

import (
	"bufio"
	"bytes"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	syslog "log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/yogeshlonkar/yrl/pkg/database"
	"github.com/yogeshlonkar/yrl/pkg/git"
	"github.com/yogeshlonkar/yrl/pkg/model"
)

const (
	gitStatusLogFile     = "/tmp/.yrl_git_status.log"
	cacheDuration        = 2 * time.Second
	remoteUpdateInterval = 15 * time.Minute
	tmuxFlag             = "no-tmux"
	forceRemote          = "remote-update"
)

var (
	start               = time.Now()
	workingDirectory, _ = os.Getwd()
)

func gitStatus() *cli.Command {
	return &cli.Command{
		Name:      "git-status",
		Aliases:   []string{"gst"},
		Usage:     "git status for repository",
		UsageText: fmt.Sprintf("yrl git-status [command options] [path]\n\n   path\t\tgit directory (default: %s)", workingDirectory),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  tmuxFlag,
				Usage: "will print status with tmux powerline style fg/bg colours",
				Value: os.Getenv("TMUX") == "",
			},
			&cli.BoolFlag{
				Name:    forceRemote,
				Aliases: []string{"r"},
				Usage:   "will for update remote",
			},
		},
		ArgsUsage: "[path]",
		Action:    gitStatusAction,
		Before: func(c *cli.Context) error {
			if !c.Bool("trace") && !c.Bool("debug") {
				f, err := os.OpenFile(gitStatusLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
				if err != nil {
					syslog.Printf("Could not create log file %q", err)
				} else if s, _ := f.Stat(); s != nil && s.Size() > maxLogSizeInKb {
					err = f.Truncate(0)
					if err != nil {
						syslog.Printf("Could not truncate log file %q", err)
					}
					_, err = f.Seek(0, 0)
					if err != nil {
						syslog.Printf("Could not seek file to start %q", err)
					}
				}
				if time.Now().Unix()%logSamplePerMinute == 0 {
					log.Logger = log.Output(zerolog.ConsoleWriter{Out: f}).Level(zerolog.DebugLevel)
				}
			}
			return nil
		},
	}
}

func gitStatusAction(ctx *cli.Context) error {
	noTmux := ctx.Bool(tmuxFlag)
	g, err := validate(ctx)
	if err != nil {
		if errors.Is(err, git.ErrNotInAGitRepo) {
			return nil
		}
		return err
	}
	db := database.New()
	defer db.Close()
	id, c, updateTime, err := db.Git().GitStatus()
	if err == nil {
		isBefore := time.Now().Add(-cacheDuration).Before(updateTime)
		g.Cached = isBefore && id == g.ID
		if g.Cached {
			g.Status = c
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		log.Warn().Str("AbsPath", g.AbsPath).Err(err).Msg("could not fetch cached git status")
	}
	forceRemote := ctx.Bool(forceRemote)
	if !g.Cached || forceRemote {
		g.Loading = true
		getStatus(g)
		if err = remoteUpdate(g, forceRemote, noTmux, db); err != nil {
			g.StatusString = g.StatusLine(noTmux)
			fmt.Println(g.StatusString)
			return nil
		}
		g.Loading = false
		getStatus(g)
		// log.Trace().Err(err).Msg("Continuing without remote update")
		if err = db.Git().SaveGitStatus(g.ID, g.Status); err != nil {
			g.StatusString = g.StatusLine(noTmux)
			fmt.Println(g.StatusString)
			log.Panic().Str("AbsPath", g.AbsPath).Err(err).Msg("could insert/replace remote status")
		}
	} else {
		g.RemoteSuccess = true
	}
	g.StatusString = g.StatusLine(noTmux)
	fmt.Println(g.StatusString)
	log.Info().Bool("cached", g.Cached).Str("took", time.Since(start).String()).Msg("completed")
	return nil
}

func validate(ctx *cli.Context) (g *model.GitRepo, err error) {
	g = new(model.GitRepo)
	g.AbsPath = workingDirectory
	if ctx.Args().Len() == 1 {
		g.AbsPath = ctx.Args().First()
	} else if ctx.Args().Len() > 1 {
		return nil, errors.New("expected single argument")
	}
	g.Status = new(git.Status)
	fatal := log.Fatal().Str("AbsPath", g.AbsPath)
	if g.AbsPath, err = filepath.Abs(g.AbsPath); err != nil {
		fatal.Err(err).Msg("could not get absolute AbsPath")
	}
	if isGit, err := git.IsInsideWorkTree(g.AbsPath); err != nil {
		return nil, err
	} else if !isGit {
		log.Debug().Msg("not a git repo")
		os.Exit(0)
	}
	g.FileInfo, err = os.Stat(g.AbsPath)
	if err != nil || !g.IsDir() {
		fatal.Err(err).Msg("could not stat")
	}
	g.ID = base64.StdEncoding.EncodeToString([]byte(g.AbsPath))
	return
}

func remoteUpdate(g *model.GitRepo, forceUpdate, noTmux bool, db database.Database) error {
	log.Debug().Bool("RemoteSuccess", g.RemoteSuccess).Bool("forceUpdate", forceUpdate).Send()
	savedRemote, err := db.Git().GetRemoteURL(g.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Trace().Str("AbsPath", g.AbsPath).Msg("no git status saved for repository")
			forceUpdate = true
		} else {
			return err
		}
	}
	if !forceUpdate {
		g.Remote = []string{savedRemote}
		g.RemoteSuccess, err = db.Git().GitRemoteStatus(g.ID, time.Now().Add(-remoteUpdateInterval))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Trace().Str("AbsPath", g.AbsPath).Err(err).Msg("no update required at this time")
				g.RemoteSuccess = true
				return nil
			}
			return err
		}
	}
	if len(g.Remote) == 0 {
		g.Remote = git.GetRemotes(g.AbsPath)
	}
	getStatus(g)
	g.Loading = true
	g.StatusString = g.StatusLine(noTmux)
	fmt.Println(g.StatusString)
	log.Debug().Msg("updating remote")
	cmd := exec.Command("git", "remote", "update", "--prune")
	cmd.Dir = g.AbsPath
	_, err = cmd.Output()
	if g.RemoteSuccess = err == nil; err != nil {
		log.Warn().Str("AbsPath", g.AbsPath).Err(err).Msg("could update remote")
		return err
	}
	if err = db.Git().SaveGitRemoteStatus(g.ID, g.GetFirstRemote(), g.RemoteSuccess, time.Now().Add(remoteUpdateInterval)); err != nil {
		// log.Fatal().Str("AbsPath", g.AbsPath).Err(err).Msg("could insert/replace remote status")
		return err
	}
	return nil
}

func getStatus(g *model.GitRepo) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		var buf = new(bytes.Buffer)
		cmd := exec.Command("git", "status", "--untracked-files=all", "--branch", "--porcelain=v2")
		cmd.Stdout = buf
		cmd.Dir = g.AbsPath
		if err := cmd.Run(); err != nil {
			log.Fatal().Err(err).Msg("could not get git status")
		}
		s := new(git.Status)
		s.ParsePorcelainV2(buf)
		s.RemoteSuccess = g.RemoteSuccess
		g.Status = s
		wg.Done()
	}()
	go func() {
		var buf = new(bytes.Buffer)
		cmd := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-dir")
		cmd.Stdout = buf
		cmd.Dir = g.AbsPath
		if err := cmd.Run(); err != nil {
			log.Fatal().Err(err).Msg("could not get git Root")
		}
		wg.Done()
		g.Root = strings.Trim(buf.String(), "\n")
	}()
	wg.Wait()
	wg.Add(2)
	go func() {
		if g.Status.IsNew = g.Status.Upstream == ""; !g.Status.IsNew {
			var buf = new(bytes.Buffer)
			cmd := exec.Command("git", "branch", "-r")
			cmd.Stdout = buf
			cmd.Dir = g.AbsPath
			if err := cmd.Run(); err != nil {
				log.Warn().Err(err).Msg("could check remote branch")
			} else {
				g.Status.IsGone = !strings.Contains(buf.String(), g.Status.Upstream)
			}
		}
		wg.Done()
	}()
	stashFile := g.Root + "/logs/refs/stash"
	go func() {
		file, err := os.Open(stashFile)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Warn().Err(err).Msg("could not open stash")
			}
		} else {
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Warn().Err(err).Msg("error while closing stash file")
				}
			}(file)
			if _, err := file.Stat(); err == nil {
				reader := bufio.NewReader(file)
				g.Status.Stashed, err = lineCounter(reader)
				if err != nil {
					log.Warn().Err(err).Msg("error while reading stash")
				}
			} else if !os.IsNotExist(err) {
				log.Warn().Err(err).Msg("could not access stash")
			}
		}
		wg.Done()
	}()
	wg.Wait()
}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}
	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)
		switch {
		case err == io.EOF:
			return count, nil
		case err != nil:
			return count, err
		}
	}
}
