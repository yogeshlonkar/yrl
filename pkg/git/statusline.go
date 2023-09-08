package git

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
)

const (
	backgroundClean    = "120"
	backgroundDefault  = "209"
	backgroundError    = "160"
	backgroundGone     = "088"
	backgroundLoading  = "056"
	backgroundNew      = "251"
	backgroundTerminal = "235"
	branchMaxLen       = 50
	foregroundGreen    = "22"
	foregroundClean    = "000"
	foregroundDefault  = "235"
	foregroundError    = "254"
	foregroundGone     = "255"
	foregroundPrevious = "033"
	iconAdded          = "\U000F034C "
	iconAhead          = "\uF403"
	iconArrowRight     = "\uE0B0"
	iconBehind         = "\uF404 "
	iconChanged        = "\U000F1787 "
	iconClean          = "\uEBB1"
	iconCopied         = "\U000F0191 "
	iconDeleted        = "\U000F0AD3 "
	iconDirty          = "\uF256"
	iconGit            = "\uF418"
	iconGone           = " "
	iconLoading        = "\uF46A"
	iconNew            = " "
	iconRenamed        = "\U000F04E1 "
	iconSeparator      = "\uE621"
	iconStaged         = "\uF01C "
	iconStashed        = " "
	iconSyncFailed     = "\uF41C"
	iconUnmerged       = "\U000F1A98 "
	whiteSpace         = " "
)

var (
	bugfix         = regexp.MustCompile("(bug)?fix(es)?/")
	feat           = regexp.MustCompile("feat(ures?)?/")
	hotfix         = regexp.MustCompile("hotfix/")
	release        = regexp.MustCompile("releases?/")
	LoadingSegment = powerlineSegment(foregroundPrevious, "056", iconArrowRight) +
		powerlineSegment(foregroundError, backgroundLoading, " "+iconLoading+" ") +
		powerlineSegment(backgroundLoading, backgroundTerminal, iconArrowRight+"\n")
)

type statusLine struct {
	bytes.Buffer
}

func (sl *statusLine) add(s ...string) {
	sl.addIf(true, s...)
}

func (sl *statusLine) addIf(i interface{}, s ...string) {
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
		log.Warn().Err(err).Msgf("error adding %s to status with %q", s, i)
	}
}

func shortBranch(branch string) string {
	to := branch
	switch {
	case feat.Match([]byte(branch)):
		to = feat.ReplaceAllString(branch, "\uF0EB ")
	case hotfix.Match([]byte(branch)):
		to = hotfix.ReplaceAllString(branch, "\uF490 ")
	case release.Match([]byte(branch)):
		to = release.ReplaceAllString(branch, "\uF461 ")
	case bugfix.Match([]byte(branch)):
		to = bugfix.ReplaceAllString(branch, "\uF188 ")
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

func powerlineSegment(fg, bg, str string) string {
	return fmt.Sprint(fmt.Sprintf("#[fg=color%s,bg=color%s]", fg, bg), str)
}

func terminalSegment(fg, bg, str string) string {
	return fmt.Sprint(fmt.Sprintf("\033[38;5;%sm\033[48;5;%sm", fg, bg), str)
}

func segment(noTmux bool, fg, bg, str string) string {
	if noTmux {
		if str == iconArrowRight {
			return terminalSegment(fg, bg, "")
		}
		return terminalSegment(fg, bg, str)
	}
	return powerlineSegment(fg, bg, str)
}

type Segment []string

func (s *Segment) Counter(count int, segs ...string) {
	if count > 0 {
		*s = append(*s, fmt.Sprintf("%d%s", count, strings.Join(segs, "")))
	}
}

func (s *Segment) String() string {
	return strings.Join(*s, iconSeparator)
}

func (ss *Status) StatusLine(noTmux bool) string {
	sl := new(statusLine)
	if ss != nil && !ss.RemoteSuccess {
		startArrow := segment(noTmux, foregroundPrevious, backgroundError, iconArrowRight)
		startColor := segment(noTmux, foregroundError, backgroundError, whiteSpace)
		endColor := segment(noTmux, backgroundError, ss.Bg(), "")
		sl.add(startArrow)
		sl.add(startColor + iconSyncFailed + whiteSpace + endColor + iconArrowRight)
	} else {
		startArrow := segment(noTmux, foregroundPrevious, ss.Bg(), iconArrowRight)
		sl.add(startArrow)
	}
	if ss.Branch != "" {
		sl.add(segment(noTmux, ss.Fg(), ss.Bg(), "") + whiteSpace + iconGit + whiteSpace)
		sl.add(shortBranch(ss.Branch))
		sl.add(whiteSpace)
		switch {
		case ss.IsNew:
			sl.add(iconNew)
		case ss.IsGone:
			sl.add(iconGone)
		case ss.Clean():
			startColor := segment(noTmux, foregroundGreen, ss.Bg(), iconClean)
			endColor := segment(noTmux, backgroundTerminal, ss.Bg(), " ")
			sl.add(startColor + endColor)
			//sl.add(iconClean)
		case ss.Dirty():
			startColor := segment(noTmux, backgroundGone, ss.Bg(), iconDirty)
			endColor := segment(noTmux, backgroundTerminal, ss.Bg(), " ")
			sl.add(startColor + endColor)
			//sl.add(iconDirty)
		}
		// separator
		segments := &Segment{}

		segments.Counter(ss.Ahead, iconAhead)
		segments.Counter(ss.Behind, iconBehind)
		segments.Counter(ss.Unmerged, iconUnmerged)
		// separator

		segments.Counter(ss.Untracked+ss.UnStaged.Added, iconAdded)
		segments.Counter(ss.UnStaged.Deleted, iconDeleted)
		segments.Counter(ss.UnStaged.Modified, iconChanged)
		segments.Counter(ss.UnStaged.Renamed, iconRenamed)
		segments.Counter(ss.UnStaged.Copied, iconCopied)
		if ss.Staged.Renamed == 0 {
			segments.Counter(ss.Staged.Count(), iconStaged)
		} else {
			if ss.Staged.Count()-ss.Staged.Renamed > 0 {
				segments.Counter(ss.Staged.Renamed, iconRenamed)
				segments.Counter(ss.Staged.Count()-ss.Staged.Renamed, iconStaged)
			} else {
				segments.Counter(ss.Staged.Renamed, iconRenamed, iconStaged)
			}
		}
		segments.Counter(ss.Stashed, iconStashed)
		sl.addIf(len(*segments) > 0, segments.String())
		//sl.add(" ")
	}
	if noTmux {
		sl.add(fmt.Sprintf("\033[0m\033[38;5;%[1]sm%[2]s\033[0m", ss.Bg(), iconArrowRight))
	} else {
		sl.add(powerlineSegment(ss.Bg(), backgroundTerminal, iconArrowRight))
	}
	return sl.String()
}
