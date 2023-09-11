package git

import (
	"fmt"
	"regexp"

	"github.com/yogeshlonkar/yrl/pkg/ansi"
	"github.com/yogeshlonkar/yrl/pkg/git/icons"
)

const (
	branchMaxLen = 20
	tailLen      = 9
	headLen      = 8
)

var (
	bugfix  = regexp.MustCompile("(bug)?fix(es)?/")
	feat    = regexp.MustCompile("feat(ures?)?/")
	hotfix  = regexp.MustCompile("hotfix/")
	chore   = regexp.MustCompile("chores?/")
	release = regexp.MustCompile("releases?/")
)

func shortBranch(branch string) string {
	var icon string
	switch {
	case feat.MatchString(branch):
		icon = icons.Feature
		branch = feat.ReplaceAllString(branch, "")
	case hotfix.MatchString(branch):
		icon = icons.Hotfix
		branch = hotfix.ReplaceAllString(branch, "")
	case release.MatchString(branch):
		icon = icons.Release
		branch = release.ReplaceAllString(branch, "")
	case chore.MatchString(branch):
		icon = icons.Chore
		branch = chore.ReplaceAllString(branch, "")
	case bugfix.MatchString(branch):
		icon = icons.Bugfix
		branch = bugfix.ReplaceAllString(branch, "")
	}
	for strLen := range branch {
		if strLen >= branchMaxLen {
			lastIndex := len(branch) - 1
			branch = branch[:headLen] + "..." + branch[lastIndex-tailLen:]
			break
		}
	}
	return icon + branch
}

func (ss *Status) StatusLine(noTmux bool) string {
	statusLine := &ansi.Segment{}
	remote := &ansi.Segment{}
	if ss.Loading {
		remote.Add(ansi.ColoredSegment(noTmux, ansi.ForegroundGrey89, ansi.BackgroundLoading, ansi.WhiteSpace, icons.Sync, ansi.WhiteSpace))
		remote.Add(ansi.ColoredSegment(noTmux, ansi.BackgroundLoading, ss.Bg(), icons.ArrowRight, ""))
	} else if ss.RemoteSuccess {
		remote.Add(ansi.ColoredSegment(noTmux, ansi.ForegroundPrevious, ss.Bg(), ansi.WhiteSpace))
	} else {
		remote.Add(ansi.ColoredSegment(noTmux, ansi.ForegroundPrevious, ansi.BackgroundError, icons.ArrowRight))
		remote.Add(ansi.ColoredSegment(noTmux, ansi.ForegroundGrey89, ansi.BackgroundError, ansi.WhiteSpace))
		remote.Add(icons.Failed, ansi.WhiteSpace)
		remote.Add(ansi.ColoredSegment(noTmux, ansi.BackgroundError, ss.Bg()), icons.ArrowRight)
	}
	remote.Add(ansi.ColoredSegment(noTmux, ss.Fg(), ss.Bg()), icons.Git, shortBranch(ss.Branch), icons.Separator)
	reset := ansi.ColoredSegment(noTmux, ansi.BackgroundTerminal, ss.Bg())
	resetWhitespace := ansi.ColoredSegment(noTmux, ansi.BackgroundTerminal, ss.Bg(), ansi.WhiteSpace)
	switch {
	case ss.IsNew:
		remote.Add(ansi.ColoredSegment(noTmux, ansi.ForegroundBlue, ss.Bg(), icons.New), reset)
	case ss.IsGone:
		remote.Add(icons.Gone)
	case ss.Clean():
		remote.Add(ansi.ColoredSegment(noTmux, ansi.ForegroundGreen, ss.Bg(), icons.Clean), resetWhitespace)
	case ss.Dirty():
		remote.Add(ansi.ColoredSegment(noTmux, ansi.BackgroundGone, ss.Bg()), reset)
	}
	statusLine.Add(remote.String())
	// sub info
	sub := &ansi.Segment{}
	// branch info
	branch := &ansi.Segment{}
	branch.Counter(ss.Ahead, ansi.ColoredSegment(noTmux, ansi.ForegroundDarkBlue, ss.Bg(), icons.Ahead), reset)
	branch.Counter(ss.Behind, ansi.ColoredSegment(noTmux, ansi.ForegroundDarkBlue, ss.Bg(), icons.Behind), reset)
	branch.Counter(ss.Unmerged, ansi.ColoredSegment(noTmux, ansi.BackgroundError, ss.Bg(), icons.Unmerged))
	branch.AppendOnly(reset)
	sub.When(len(*branch) > 0, branch.String())
	// unStaged info
	unStaged := &ansi.Segment{}
	unStaged.Counter(ss.Untracked+ss.UnStaged.Added, icons.Added)
	unStaged.Counter(ss.UnStaged.Deleted, icons.Deleted)
	unStaged.Counter(ss.UnStaged.Renamed, icons.Renamed)
	unStaged.Counter(ss.UnStaged.Copied, icons.Copied)
	unStaged.Counter(ss.UnStaged.Modified, icons.Modified)
	sub.When(len(*unStaged) > 0, unStaged.String())
	// staged info
	staged := &ansi.Segment{}
	staged.Counter(ss.Staged.Added, icons.Added)
	staged.Counter(ss.Staged.Deleted, icons.Deleted)
	staged.Counter(ss.Staged.Renamed, icons.Renamed)
	staged.Counter(ss.Staged.Copied, icons.Copied)
	staged.Counter(ss.Staged.Modified, icons.Modified)
	staged.PrependOnly(ansi.ColoredSegment(noTmux, ansi.ForegroundGreen, ss.Bg(), icons.Staged), reset)
	sub.When(len(*staged) > 0, staged.String())
	// stashed info
	sub.Counter(ss.Stashed, ansi.ColoredSegment(noTmux, ansi.ForegroundPurple, ss.Bg(), icons.Stashed), reset)
	statusLine.Add(sub.Join(icons.Divider))

	if noTmux {
		statusLine.Append(fmt.Sprintf("\033[0m\033[38;5;%[1]sm%[2]s\033[0m", ss.Bg(), icons.ArrowRight))
		return statusLine.String()
	}
	statusLine.Append(ansi.PowerlineSegment(ss.Bg(), ansi.BackgroundTerminal, icons.ArrowRight))
	return statusLine.String()
}
