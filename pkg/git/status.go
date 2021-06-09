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
	return a.Added+a.Deleted+a.Modified+a.Copied+a.Renamed != 0
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

func consumeNext(s *bufio.Scanner) string {
	if s.Scan() {
		return s.Text()
	}
	return ""
}

func (pi *Status) ParsePorcelainV2(r io.Reader) {
	for s := bufio.NewScanner(r); s.Scan(); {
		if len(s.Text()) < 1 {
			continue
		}
		pi.ParseLine(s.Text())
	}
}

func (pi *Status) ParseLine(line string) {
	s := bufio.NewScanner(strings.NewReader(line))
	s.Split(bufio.ScanWords)
	for s.Scan() {
		switch s.Text() {
		case "#":
			if err := pi.parseBranchInfo(s); err != nil {
				log.Warn().Err(err).Msg("error parsing Branch")
			}
		case "1", "2":
			pi.parseTrackedFile(s)
		case "u":
			pi.Unmerged++
		case "?":
			pi.Untracked++
		}
	}
}

func (pi *Status) parseBranchInfo(s *bufio.Scanner) (err error) {
	for s.Scan() {
		switch s.Text() {
		case "branch.oid":
			pi.Commit = consumeNext(s)
		case "branch.head":
			pi.Branch = consumeNext(s)
		case "branch.upstream":
			pi.Upstream = consumeNext(s)
		case "branch.ab":
			err = pi.parseAheadBehind(s)
			if err != nil {
				log.Warn().Err(err).Msg("error parsing Branch.ab")
			}
		}
	}
	return err
}

func (pi *Status) parseAheadBehind(s *bufio.Scanner) error {
	for s.Scan() {
		i, err := strconv.Atoi(s.Text()[1:])
		if err != nil {
			return err
		}
		switch s.Text()[:1] {
		case "+":
			pi.Ahead = i
		case "-":
			pi.Behind = i
		}
	}
	return nil
}

func (pi *Status) parseTrackedFile(s *bufio.Scanner) {
	for index := 0; s.Scan(); index++ {
		switch index {
		case 0:
			pi.parseXY(s.Text())
		default:
			break
		}
	}
}

func (pi *Status) parseXY(xy string) {
	pi.Staged.parseSymbol(xy[:1])
	pi.UnStaged.parseSymbol(xy[1:])
}

func (pi *Status) Clean() bool {
	return !pi.IsGone && !pi.Staged.HasChanged() && !pi.UnStaged.HasChanged() &&
		pi.Stashed+pi.Behind+pi.Ahead+pi.Unmerged+pi.Untracked == 0
}

func (pi *Status) Bg() string {
	if pi.Clean() {
		return "120"
	}
	if pi.IsNew {
		return "251"
	}
	if pi.IsGone {
		return "088"
	}
	return "209"
}

func (pi *Status) Fg() string {
	if pi.Clean() {
		return "000"
	}
	if pi.IsGone {
		return "255"
	}
	return "235"
}
