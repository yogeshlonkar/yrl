package ansi

import (
	"fmt"
	"strings"
)

type Segment []string

func (s *Segment) Counter(count int, segs ...string) {
	if count > 0 {
		s.Add(fmt.Sprintf("%d%s", count, strings.Join(segs, "")))
	}
}

func (s *Segment) When(condition bool, segs ...string) {
	if condition {
		s.Add(strings.Join(segs, ""))
	}
}

func (s *Segment) Add(segs ...string) {
	*s = append(*s, strings.Join(segs, ""))
}

func (s *Segment) Append(segs ...string) {
	if len(*s) > 0 {
		(*s)[len(*s)-1] += strings.Join(segs, "")
	} else {
		s.Add(segs...)
	}
}

func (s *Segment) AppendOnly(segs ...string) {
	if len(*s) > 0 {
		s.Append(segs...)
	}
}

func (s *Segment) AppendWhen(condition bool, segs ...string) {
	if condition {
		s.Append(segs...)
	}
}

func (s *Segment) Prepend(segs ...string) {
	if len(*s) > 0 {
		(*s)[0] = strings.Join(segs, "") + (*s)[0]
	} else {
		s.Add(segs...)
	}
}

func (s *Segment) PrependOnly(segs ...string) {
	if len(*s) > 0 {
		s.Prepend(segs...)
	}
}

func (s *Segment) PrependWhen(condition bool, segs ...string) {
	if condition {
		s.Prepend(segs...)
	}
}

func (s *Segment) String() string {
	return s.Join("")
}

func (s *Segment) Join(separator string) string {
	return strings.Join(*s, separator)
}
