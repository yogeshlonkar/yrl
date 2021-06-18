package statusline

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"

	"github.com/yogeshlonkar/yrl/pkg/model"
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
		log.Warn().Err(err).Msgf("Error adding %s to status with %q", s, i)
	}
}

func shortBranch(branch string) string {
	to := branch
	if feat.Match([]byte(branch)) {
		to = feat.ReplaceAllString(branch, "\uF893 ")
	} else if hotfix.Match([]byte(branch)) {
		to = hotfix.ReplaceAllString(branch, "\uF490 ")
	} else if release.Match([]byte(branch)) {
		to = release.ReplaceAllString(branch, "\uF461 ")
	} else if bugfix.Match([]byte(branch)) {
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

func GenerateStatus(g *model.GitRepo, noTmux bool) string {
	sl := new(statusLine)
	if !g.RemoteSuccess {
		startArrow := powerlineSegment(foregroundPrevious, backgroundError, iconArrowRight)
		if noTmux {
			startArrow = terminalSegment(foregroundPrevious, backgroundError)
		}
		sl.add(true, startArrow)
		errorSegmentStartColor := powerlineSegment(foregroundError, backgroundError, whiteSpace)
		errorSegmentEndColor := powerlineSegment(backgroundError, g.Bg())
		if noTmux {
			errorSegmentStartColor = terminalSegment(foregroundError, backgroundError, whiteSpace)
			errorSegmentEndColor = terminalSegment(backgroundError, g.Bg())
		}
		sl.add(true, errorSegmentStartColor+iconSyncFailed+whiteSpace+errorSegmentEndColor+iconArrowRight)
	} else {
		startArrow := powerlineSegment(foregroundPrevious, g.Bg(), iconArrowRight)
		if noTmux {
			startArrow = terminalSegment(foregroundPrevious, g.Bg())
		}
		sl.add(true, startArrow)
	}
	segmentColor := powerlineSegment(g.Fg(), g.Bg())
	if noTmux {
		segmentColor = terminalSegment(g.Fg(), g.Bg())
	}
	sl.add(true, segmentColor+whiteSpace+iconGit+whiteSpace)
	sl.add(true, shortBranch(g.Branch))
	sl.add(true, whiteSpace)
	if g.IsNew {
		sl.add(true, iconNew)
	} else if g.IsGone {
		sl.add(true, iconGone)
	} else if g.Clean() {
		sl.add(true, iconClean)
	} else if g.Dirty() {
		sl.add(true, iconDirty)
	}
	sl.add(g.Count() > 0, whiteSpace, iconSeparator)
	// separator

	sl.add(g.Ahead, iconAhead)
	sl.add(g.Ahead > 0 && g.Behind > 0, iconSeparator)
	sl.add(g.Behind, iconBehind)
	sl.add((g.Ahead > 0 || g.Behind > 0) && g.Unmerged > 0, iconSeparator)
	sl.add(g.Unmerged, iconUnmerged)
	sl.add(g.Ahead+g.Behind+g.Unmerged > 0 && g.Untracked+g.UnStaged.Count() > 0, iconSeparator)
	// separator

	sl.add(g.Untracked+g.UnStaged.Added, iconAdded)
	sl.add(g.UnStaged.Deleted, iconDeleted)
	sl.add(g.UnStaged.Modified, iconChanged)
	sl.add(g.UnStaged.Renamed, iconRenamed)
	sl.add(g.UnStaged.Copied, iconCopied)
	sl.add(g.Untracked+g.UnStaged.Count() > 0 && g.Staged.Count()+g.Stashed > 0, iconSeparator)
	// separator

	sl.add(g.Staged.Count(), iconStaged)
	sl.add(g.Staged.Count() > 0 && g.Stashed > 0, iconSeparator)
	sl.add(g.Stashed, "")
	sl.add(true, " ")
	if !noTmux {
		suffix := powerlineSegment(g.Bg(), backgroundTerminal, iconArrowRight)
		sl.add(true, suffix)
	} else {
		sl.add(true, fmt.Sprintf("\033[0m\033[38;5;%[1]sm%[2]s\033[0m", g.Bg(), iconArrowRight))
	}
	return sl.String()
}
