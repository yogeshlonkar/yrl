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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/yogeshlonkar/yrl/pkg/database"
	"github.com/yogeshlonkar/yrl/pkg/git"
	"github.com/yogeshlonkar/yrl/pkg/model"
)

const (
	branchMaxLen         = 50
	cacheDuration        = 5 * time.Second
	remoteUpdateInterval = 15 * time.Minute
	bgErr                = "160"
	fgErr                = "254"
	tmuxFlag             = "no-tmux"
	forceRemote          = "remote-update"
)

var (
	loading = fmt.Sprintf("#[fg=color033,bg=color%[1]s]\uE0B0#[fg=color%[2]s,bg=color%[1]s] \uF46A #[fg=color%[1]s,bg=colour235]\uE0B0\n", "056", "254")
	bugfix  = regexp.MustCompile("(bug)?fix(es)?/")
	feat    = regexp.MustCompile("feat(ures?)?/")
	hotfix  = regexp.MustCompile("hotfix?/")
	release = regexp.MustCompile("releases?/")
	start   = time.Now()
)

func gitStatus() *cli.Command {
	return &cli.Command{
		Name:    "git-status",
		Aliases: []string{"gst"},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    tmuxFlag,
				Aliases: []string{"t"},
				Usage:   "will print status with tmux powerline style fg/bg colours",
				Value:   os.Getenv("TMUX") == "",
			},
			&cli.BoolFlag{
				Name:    forceRemote,
				Aliases: []string{"r"},
				Usage:   "will for update remote",
			},
		},
		Action: gitStatusAction,
	}
}

func gitStatusAction(ctx *cli.Context) error {
	t, err := validate(ctx)
	if err != nil {
		return err
	}
	db := database.New()
	defer db.Close()
	id, c, updateTime, err := db.Git().GitStatus()
	if err == nil {
		isBefore := time.Now().Add(-cacheDuration).Before(updateTime)
		t.Cached = isBefore && id == t.ID
		if t.Cached {
			t.Status = c
		}
	} else if err != sql.ErrNoRows {
		log.Warn().Str("AbsPath", t.AbsPath).Err(err).Msg("Could not fetch cached git status")
	}
	noTmux := ctx.Bool(tmuxFlag)
	forceRemote := ctx.Bool(forceRemote)
	if !t.Cached || forceRemote {
		if !noTmux {
			fmt.Printf(loading)
		}
		remoteUpdate(t, forceRemote, db)
		getStatus(t)
		if err := db.Git().SaveGitStatus(t.ID, t.Status); err != nil {
			log.Fatal().Str("AbsPath", t.AbsPath).Err(err).Msg("Could insert/replace remote status")
		}
	}
	t.StatusString = getPowerlineStatus(t, noTmux)
	fmt.Println(t.StatusString)
	log.Info().Bool("cached", t.Cached).Str("took", time.Now().Sub(start).String()).Msg("Completed")
	return nil
}

func validate(ctx *cli.Context) (t *model.GitRepo, err error) {
	if ctx.Args().Len() != 1 {
		return nil, errors.New("expected single argument")
	}
	t = new(model.GitRepo)
	t.Status = new(git.Status)
	t.AbsPath = ctx.Args().First()
	fatal := log.Fatal().Str("AbsPath", t.AbsPath)
	if t.AbsPath, err = filepath.Abs(t.AbsPath); err != nil {
		fatal.Err(err).Msg("Could not get absolute AbsPath")
	}
	if isGit, _ := git.IsInsideWorkTree(t.AbsPath); !isGit {
		log.Debug().Msg("Not a git repo")
		os.Exit(0)
	}
	t.FileInfo, err = os.Stat(t.AbsPath)
	if err != nil || !t.IsDir() {
		fatal.Err(err).Msg("Could not stat")
	}
	t.ID = base64.StdEncoding.EncodeToString([]byte(t.AbsPath))
	return
}

func remoteUpdate(t *model.GitRepo, forceUpdate bool, db database.Database) {
	var err error
	log.Debug().Bool("RemoteSuccess", t.RemoteSuccess).Bool("forceUpdate", forceUpdate).Send()
	t.RemoteSuccess, err = db.Git().GitRemoteStatus(t.ID, time.Now().Add(-remoteUpdateInterval))
	noRows := err != nil && err == sql.ErrNoRows
	if forceUpdate || noRows {
		log.Debug().Msg("Updating remote")
		cmd := exec.Command("git", "remote", "update", "--prune")
		cmd.Dir = t.AbsPath
		_, err = cmd.Output()
		if t.RemoteSuccess = err == nil; err != nil {
			log.Warn().Str("AbsPath", t.AbsPath).Err(err).Msg("Could update remote")
		}
		if err = db.Git().SaveGitRemoteStatus(t.ID, t.RemoteSuccess); err != nil {
			log.Fatal().Str("AbsPath", t.AbsPath).Err(err).Msg("Could insert/replace remote status")
		}
		return
	} else if err != nil {
		log.Fatal().Str("AbsPath", t.AbsPath).Err(err).Msg("Could not fetch update status")
	}
	return
}

func getStatus(t *model.GitRepo) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		var buf = new(bytes.Buffer)
		cmd := exec.Command("git", "status", "--untracked-files=all", "--branch", "--porcelain=v2")
		cmd.Stdout = buf
		cmd.Dir = t.AbsPath
		if err := cmd.Run(); err != nil {
			log.Fatal().Err(err).Msg("Could not get git status")
		}
		t.Status.ParsePorcelainV2(buf)
		wg.Done()
	}()
	go func() {
		var buf = new(bytes.Buffer)
		cmd := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-dir")
		cmd.Stdout = buf
		cmd.Dir = t.AbsPath
		if err := cmd.Run(); err != nil {
			log.Fatal().Err(err).Msg("Could not get git Root")
		}
		wg.Done()
		t.Root = strings.Trim(buf.String(), "\n")
	}()
	wg.Wait()
	wg.Add(2)
	go func() {
		if t.Status.IsNew = t.Status.Upstream == ""; !t.Status.IsNew {
			remoteFile := t.Root + "/refs/remotes/" + t.Status.Upstream
			log.Debug().Str("remoteFile", remoteFile).Msg("Checking")
			if _, err := os.Stat(remoteFile); err != nil {
				if t.Status.IsGone = os.IsNotExist(err); !t.Status.IsGone {
					log.Fatal().Err(err).Msg("Could check remote branch")
				}
			}
		}
		wg.Done()
	}()
	stashFile := t.Root + "/logs/refs/stash"
	go func() {
		file, err := os.Open(stashFile)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Warn().Err(err).Msg("Could not open stash")
			}
		} else {
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Warn().Err(err).Msg("Error while closing stash file")
				}
			}(file)
			if _, err := file.Stat(); err == nil {
				reader := bufio.NewReader(file)
				t.Status.Stashed, err = lineCounter(reader)
				if err != nil {
					log.Warn().Err(err).Msg("Error while reading stash")
				}
			} else if !os.IsNotExist(err) {
				log.Warn().Err(err).Msg("Could not access stash")
			}
		}
		wg.Done()
	}()
	wg.Wait()
	return
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

type statusLine struct {
	bytes.Buffer
}

func (sl *statusLine) add(i interface{}, s ...string) {
	var err error
	switch v := i.(type) {
	case int:
		if v > 0 {
			_, err = sl.WriteString(strings.Join(append([]string{strconv.Itoa(v)}, s...), ""))
		}
	case bool:
		if v {
			_, err = sl.WriteString(strings.Join(s, ""))
		}
	}
	if err != nil {
		log.Warn().Err(err).Msgf("Error adding %s to status with %q", s, i)
	}
}

func shortBranch(branch string) string {
	to := branch
	if feat.Match([]byte(branch)) {
		to = feat.ReplaceAllString(branch, "F/")
	} else if hotfix.Match([]byte(branch)) {
		to = hotfix.ReplaceAllString(branch, "H/")
	} else if release.Match([]byte(branch)) {
		to = release.ReplaceAllString(branch, "R/")
	} else if bugfix.Match([]byte(branch)) {
		to = bugfix.ReplaceAllString(branch, "B/")
	}
	lastSpaceIx := -1
	strLen := 0
	for i, r := range branch {
		if unicode.IsSpace(r) {
			lastSpaceIx = i
		}
		strLen++
		if strLen >= branchMaxLen {
			if lastSpaceIx != -1 {
				to = branch[:lastSpaceIx] + "..."
			}
		}
	}
	return to
}

func getPowerlineStatus(t *model.GitRepo, noTmux bool) string {
	sl := new(statusLine)
	bg0 := t.Bg()
	errStr := ""
	if !t.RemoteSuccess {
		bg0 = bgErr
		color1 := fmt.Sprintf("#[fg=color%[2]s,bg=color%[1]s] ", bgErr, fgErr)
		color2 := fmt.Sprintf("#[fg=color%[1]s,bg=colour%[2]s]", bgErr, t.Bg())
		if noTmux {
			color1 = fmt.Sprintf("\033[38;5;%[2]sm\033[48;5;%[1]sm ", bgErr, fgErr)
			color2 = fmt.Sprintf("\033[38;5;%[1]sm\033[48;5;%[2]sm", bgErr, t.Bg())
		}
		errStr = fmt.Sprintf("%s\uF41C %s\uE0B0", color1, color2)
	}
	color1 := fmt.Sprintf("#[fg=color033,bg=color%[1]s]\uE0B0", bg0)
	color2 := fmt.Sprintf("#[fg=color%[2]s,bg=color%[1]s]", t.Bg(), t.Fg())
	if noTmux {
		color1 = fmt.Sprintf("\033[38;5;033m\033[48;5;%[1]sm", bg0)
		color2 = fmt.Sprintf("\033[38;5;%[2]sm\033[48;5;%[1]sm", t.Bg(), t.Fg())
	}
	prefix := fmt.Sprintf("%[1]s%[2]s", color1, errStr)
	prefix = fmt.Sprintf("%[1]s%[2]s \uE725 ", prefix, color2)
	sl.add(true, prefix+shortBranch(t.Branch)+" ")
	sl.add(t.IsNew, "", " ")
	sl.add(t.IsGone, "", " ")
	sl.add(t.Clean(), "\uF62B ")

	sl.add(t.Ahead, "ﯴ")
	sl.add(t.Behind, "ﯲ")
	sl.add(t.Ahead+t.Behind > 0, "|")
	sl.add(t.Unmerged, "", "|")
	sl.add(t.Untracked+t.UnStaged.Added, "")
	sl.add(t.UnStaged.Deleted, "-")
	sl.add(t.UnStaged.Modified, "*")
	sl.add(t.UnStaged.Renamed, "易")
	sl.add(t.UnStaged.Copied, "")
	// separator
	staged := t.Staged.Added + t.Staged.Deleted + t.Staged.Modified + t.Staged.Copied + t.Staged.Renamed
	sl.add(staged > 0, "|")
	sl.add(staged, "\uE257")
	sl.add(t.Stashed > 0, "|")
	sl.add(t.Stashed, "")
	if !noTmux {
		suffix := fmt.Sprintf(" #[fg=color%[1]s,bg=colour235]\uE0B0", t.Bg())
		sl.add(true, suffix)
	} else {
		sl.add(true, fmt.Sprintf(" \u001B[0m\033[38;5;%[1]sm\uE0B0\033[0m", t.Bg()))
	}
	return sl.String()
}
