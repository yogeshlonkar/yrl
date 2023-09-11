package ansi

import (
	"fmt"
	"strings"

	"github.com/yogeshlonkar/yrl/pkg/git/icons"
)

const (
	BackgroundClean    = "120"
	BackgroundDefault  = "209"
	BackgroundError    = "160"
	BackgroundGone     = "088"
	BackgroundLoading  = "056"
	BackgroundNew      = "251"
	BackgroundTerminal = "235"
	ForegroundBlue     = "33"
	ForegroundClean    = "000"
	ForegroundDarkBlue = "24"
	ForegroundDefault  = "235"
	ForegroundGone     = "255"
	ForegroundGreen    = "22"
	ForegroundGrey89   = "254"
	ForegroundPrevious = "025"
	ForegroundPurple   = "53"

	WhiteSpace = " "
)

func PowerlineSegment(fg, bg string, str ...string) string {
	return fmt.Sprint(fmt.Sprintf("#[fg=color%s,bg=color%s]", fg, bg), strings.Join(str, ""))
}

func TerminalSegment(fg, bg, str string) string {
	return fmt.Sprint(fmt.Sprintf("\033[38;5;%sm\033[48;5;%sm", fg, bg), str)
}

func ColoredSegment(noTmux bool, fg, bg string, strs ...string) string {
	if noTmux {
		if len(strs) == 1 && strs[0] == icons.ArrowRight {
			return TerminalSegment(fg, bg, "")
		}
		return TerminalSegment(fg, bg, strings.Join(strs, ""))
	}
	return PowerlineSegment(fg, bg, strings.Join(strs, ""))
}
