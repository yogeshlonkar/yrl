package git

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type Status struct {
	Ahead         int
	Behind        int
	Branch        string
	Commit        string
	IsGone        bool
	IsNew         bool
	Staged        Area
	Stashed       int
	UnStaged      Area
	Unmerged      int
	Untracked     int
	Upstream      string
	RemoteSuccess bool
}

type Area struct {
	Modified int
	Added    int
	Deleted  int
	Renamed  int
	Copied   int
}

func (a *Area) HasChanged() bool {
	return a.Count() != 0
}

func (a *Area) parseSymbol(s string) {
	switch s {
	case "M":
		a.Modified++
	case "A":
		a.Added++
	case "D":
		a.Deleted++
	case "R":
		a.Renamed++
	case "C":
		a.Copied++
	}
}

func (a *Area) Count() int {
	return a.Added + a.Deleted + a.Modified + a.Copied + a.Renamed
}

func consumeNext(s *bufio.Scanner) string {
	if s.Scan() {
		return s.Text()
	}
	return ""
}

func (ss *Status) ParsePorcelainV2(r io.Reader) {
	for s := bufio.NewScanner(r); s.Scan(); {
		if len(s.Text()) < 1 {
			continue
		}
		ss.ParseLine(s.Text())
	}
}

func (ss *Status) ParseLine(line string) {
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	for s.Scan() {
		switch s.Text() {
		case "#":
			if err := ss.parseBranchInfo(s); err != nil {
				log.Warn().Err(err).Msg("error parsing Branch")
			}
		case "1", "2":
			ss.parseTrackedFile(s)
		case "u":
			ss.Unmerged++
		case "?":
			ss.Untracked++
		}
	}
}

func (ss *Status) parseBranchInfo(s *bufio.Scanner) (err error) {
	for s.Scan() {
		switch s.Text() {
		case "branch.oid":
			ss.Commit = consumeNext(s)
		case "branch.head":
			ss.Branch = consumeNext(s)
		case "branch.upstream":
			ss.Upstream = consumeNext(s)
		case "branch.ab":
			err = ss.parseAheadBehind(s)
			if err != nil {
				log.Warn().Err(err).Msg("error parsing Branch.ab")
			}
		}
	}
	return err
}

func (ss *Status) parseAheadBehind(s *bufio.Scanner) error {
	for s.Scan() {
		i, err := strconv.Atoi(s.Text()[1:])
		if err != nil {
			return err
		}
		switch s.Text()[:1] {
		case "+":
			ss.Ahead = i
		case "-":
			ss.Behind = i
		}
	}
	return nil
}

func (ss *Status) parseTrackedFile(s *bufio.Scanner) {
	for index := 0; s.Scan(); index++ {
		switch index {
		case 0:
			ss.parseXY(s.Text())
		default:
			break
		}
	}
}

func (ss *Status) parseXY(xy string) {
	ss.Staged.parseSymbol(xy[:1])
	ss.UnStaged.parseSymbol(xy[1:])
}

func (ss *Status) Clean() bool {
	return !ss.IsNew && !ss.IsGone && ss.Count() == 0
}

func (ss *Status) Dirty() bool {
	return !ss.IsNew && !ss.IsGone && ss.Count() > 0
}

func (ss *Status) Count() int {
	return ss.Staged.Count() + ss.UnStaged.Count() +
		ss.Stashed + ss.Behind + ss.Ahead + ss.Unmerged + ss.Untracked
}

func (ss *Status) Bg() string {
	switch {
	case ss.Clean():
		return backgroundClean
	case ss.IsNew:
		return backgroundNew
	case ss.IsGone:
		return backgroundGone
	default:
		return backgroundDefault
	}
}

func (ss *Status) Fg() string {
	switch {
	case ss.Clean():
		return foregroundClean
	case ss.IsGone:
		return foregroundGone
	default:
		return foregroundDefault
	}
}
