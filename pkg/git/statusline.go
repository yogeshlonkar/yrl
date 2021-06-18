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
	backgroundError    = "160"
	backgroundTerminal = "235"
	branchMaxLen       = 50
	foregroundPrevious = "033"
	foregroundError    = "254"
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
	hotfix         = regexp.MustCompile("hotfix?/")
	release        = regexp.MustCompile("releases?/")
	LoadingSegment = powerlineSegment(foregroundPrevious, "056", iconArrowRight) +
		powerlineSegment("254", "056", " ", iconLoading, " ") +
		powerlineSegment("056", backgroundTerminal, iconArrowRight+"\n")
)

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
		log.Warn().Err(err).Msgf("error adding %s to status with %q", s, i)
	}
}

func shortBranch(branch string) string {
	to := branch
	switch {
	case feat.Match([]byte(branch)):
		to = feat.ReplaceAllString(branch, "\uF893 ")
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

func powerlineSegment(fg, bg string, str ...string) string {
	str1 := strings.Join(str, "")
	return fmt.Sprint(fmt.Sprintf("#[fg=color%s,bg=color%s]", fg, bg), str1)
}

func terminalSegment(fg, bg string, str ...string) string {
	str1 := strings.Join(str, "")
	return fmt.Sprint(fmt.Sprintf("\033[38;5;%sm\033[48;5;%sm", fg, bg), str1)
}

func (ss *Status) StatusLine(noTmux bool) string {
	sl := new(statusLine)
	if !ss.RemoteSuccess {
		startArrow := powerlineSegment(foregroundPrevious, backgroundError, iconArrowRight)
		if noTmux {
			startArrow = terminalSegment(foregroundPrevious, backgroundError)
		}
		sl.add(true, startArrow)
		errorSegmentStartColor := powerlineSegment(foregroundError, backgroundError, whiteSpace)
		errorSegmentEndColor := powerlineSegment(backgroundError, ss.Bg())
		if noTmux {
			errorSegmentStartColor = terminalSegment(foregroundError, backgroundError, whiteSpace)
			errorSegmentEndColor = terminalSegment(backgroundError, ss.Bg())
		}
		sl.add(true, errorSegmentStartColor+iconSyncFailed+whiteSpace+errorSegmentEndColor+iconArrowRight)
	} else {
		startArrow := powerlineSegment(foregroundPrevious, ss.Bg(), iconArrowRight)
		if noTmux {
			startArrow = terminalSegment(foregroundPrevious, ss.Bg())
		}
		sl.add(true, startArrow)
	}
	segmentColor := powerlineSegment(ss.Fg(), ss.Bg())
	if noTmux {
		segmentColor = terminalSegment(ss.Fg(), ss.Bg())
	}
	sl.add(true, segmentColor+whiteSpace+iconGit+whiteSpace)
	sl.add(true, shortBranch(ss.Branch))
	sl.add(true, whiteSpace)
	switch {
	case ss.IsNew:
		sl.add(true, iconNew)
	case ss.IsGone:
		sl.add(true, iconGone)
	case ss.Clean():
		sl.add(true, iconClean)
	case ss.Dirty():
		sl.add(true, iconDirty)
	}
	sl.add(ss.Count() > 0, whiteSpace, iconSeparator)
	// separator

	sl.add(ss.Ahead, iconAhead)
	sl.add(ss.Ahead > 0 && ss.Behind > 0, iconSeparator)
	sl.add(ss.Behind, iconBehind)
	sl.add((ss.Ahead > 0 || ss.Behind > 0) && ss.Unmerged > 0, iconSeparator)
	sl.add(ss.Unmerged, iconUnmerged)
	sl.add(ss.Ahead+ss.Behind+ss.Unmerged > 0 && ss.Untracked+ss.UnStaged.Count() > 0, iconSeparator)
	// separator

	sl.add(ss.Untracked+ss.UnStaged.Added, iconAdded)
	sl.add(ss.UnStaged.Deleted, iconDeleted)
	sl.add(ss.UnStaged.Modified, iconChanged)
	sl.add(ss.UnStaged.Renamed, iconRenamed)
	sl.add(ss.UnStaged.Copied, iconCopied)
	sl.add(ss.Untracked+ss.UnStaged.Count() > 0 && ss.Staged.Count()+ss.Stashed > 0, iconSeparator)
	// separator

	sl.add(ss.Staged.Count(), iconStaged)
	sl.add(ss.Staged.Count() > 0 && ss.Stashed > 0, iconSeparator)
	sl.add(ss.Stashed, "")
	sl.add(true, " ")
	if !noTmux {
		suffix := powerlineSegment(ss.Bg(), backgroundTerminal, iconArrowRight)
		sl.add(true, suffix)
	} else {
		sl.add(true, fmt.Sprintf("\033[0m\033[38;5;%[1]sm%[2]s\033[0m", ss.Bg(), iconArrowRight))
	}
	return sl.String()
}
