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
	foregroundClean    = "000"
	foregroundDefault  = "235"
	foregroundError    = "254"
	foregroundGone     = "255"
	foregroundPrevious = "033"
	iconAdded          = ""
	iconAhead          = "ﯴ"
	iconArrowRight     = "\uE0B0"
	iconBehind         = "ﯲ"
	iconChanged        = "*"
	iconClean          = "\uF62B"
	iconCopied         = ""
	iconDeleted        = "-"
	iconDirty          = "\uF256"
	iconGit            = "\uE725"
	iconGone           = ""
	iconLoading        = "\uF46A"
	iconNew            = ""
	iconRenamed        = "易"
	iconSeparator      = "\uE621"
	iconStaged         = "\uF01C"
	iconSyncFailed     = "\uF41C"
	iconUnmerged       = ""
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

func (ss *Status) StatusLine(noTmux bool) string {
	sl := new(statusLine)
	if !ss.RemoteSuccess {
		startArrow := segment(noTmux, foregroundPrevious, backgroundError, iconArrowRight)
		startColor := segment(noTmux, foregroundError, backgroundError, whiteSpace)
		endColor := segment(noTmux, backgroundError, ss.Bg(), "")
		sl.add(startArrow)
		sl.add(startColor + iconSyncFailed + whiteSpace + endColor + iconArrowRight)
	} else {
		startArrow := segment(noTmux, foregroundPrevious, ss.Bg(), iconArrowRight)
		sl.add(startArrow)
	}
	sl.add(segment(noTmux, ss.Fg(), ss.Bg(), "") + whiteSpace + iconGit + whiteSpace)
	sl.add(shortBranch(ss.Branch))
	sl.add(whiteSpace)
	switch {
	case ss.IsNew:
		sl.add(iconNew)
	case ss.IsGone:
		sl.add(iconGone)
	case ss.Clean():
		sl.add(iconClean)
	case ss.Dirty():
		sl.add(iconDirty)
	}
	sl.addIf(ss.Count() > 0, whiteSpace, iconSeparator)
	// separator

	sl.addIf(ss.Ahead, iconAhead)
	sl.addIf(ss.Ahead > 0 && ss.Behind > 0, iconSeparator)
	sl.addIf(ss.Behind, iconBehind)
	sl.addIf((ss.Ahead > 0 || ss.Behind > 0) && ss.Unmerged > 0, iconSeparator)
	sl.addIf(ss.Unmerged, iconUnmerged)
	sl.addIf(ss.Ahead+ss.Behind+ss.Unmerged > 0 && ss.Untracked+ss.UnStaged.Count() > 0, iconSeparator)
	// separator

	sl.addIf(ss.Untracked+ss.UnStaged.Added, iconAdded)
	sl.addIf(ss.UnStaged.Deleted, iconDeleted)
	sl.addIf(ss.UnStaged.Modified, iconChanged)
	sl.addIf(ss.UnStaged.Renamed, iconRenamed)
	sl.addIf(ss.UnStaged.Copied, iconCopied)
	sl.addIf(ss.Untracked+ss.UnStaged.Count() > 0 && ss.Staged.Count()+ss.Stashed > 0, iconSeparator)
	// separator

	sl.addIf(ss.Staged.Count(), iconStaged)
	sl.addIf(ss.Staged.Count() > 0 && ss.Stashed > 0, iconSeparator)
	sl.addIf(ss.Stashed, "")
	sl.add(" ")
	if noTmux {
		sl.add(fmt.Sprintf("\033[0m\033[38;5;%[1]sm%[2]s\033[0m", ss.Bg(), iconArrowRight))
	} else {
		sl.add(powerlineSegment(ss.Bg(), backgroundTerminal, iconArrowRight))
	}
	return sl.String()
}
