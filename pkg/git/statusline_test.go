package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusLine(t *testing.T) {
	tests := []struct {
		name     string
		status   *Status
		noTmux   bool
		expected string
	}{
		{
			"Remote-Fail-Clean",
			&Status{Branch: "x"},
			false,
			"#[fg=color033,bg=color160]\uE0B0#[fg=color254,bg=color160] \uF41C #[fg=color160,bg=color120]\uE0B0#[fg=color000,bg=color120] \uE725 x \uF62B #[fg=color120,bg=color235]\uE0B0",
		},
		{
			"Clean",
			&Status{Branch: "x", RemoteSuccess: true},
			false,
			"#[fg=color033,bg=color120]\uE0B0#[fg=color000,bg=color120] \uE725 x \uF62B #[fg=color120,bg=color235]\uE0B0",
		},
		{
			"Clean-New",
			&Status{Branch: "x", IsNew: true, RemoteSuccess: true},
			false,
			"#[fg=color033,bg=color251]\uE0B0#[fg=color235,bg=color251] \uE725 x \uF403 #[fg=color251,bg=color235]\uE0B0",
		},
		{
			"Clean-Gone",
			&Status{Branch: "x", IsGone: true, RemoteSuccess: true},
			false,
			"#[fg=color033,bg=color088]\uE0B0#[fg=color255,bg=color088] \uE725 x \uF48E #[fg=color088,bg=color235]\uE0B0",
		},
		{
			"Ahead",
			&Status{Branch: "x", RemoteSuccess: true, IsNew: true, Ahead: 1},
			false,
			"#[fg=color033,bg=color251]\uE0B0#[fg=color235,bg=color251] \uE725 x \uF403 \uE6211ﯴ #[fg=color251,bg=color235]\uE0B0",
		},
		{
			"Behind",
			&Status{Branch: "x", RemoteSuccess: true, IsNew: true, Behind: 1},
			false,
			"#[fg=color033,bg=color251]\uE0B0#[fg=color235,bg=color251] \uE725 x \uF403 \uE6211ﯲ #[fg=color251,bg=color235]\uE0B0",
		},
		{
			"Unmerged",
			&Status{Branch: "x", RemoteSuccess: true, IsNew: true, Unmerged: 1},
			false,
			"#[fg=color033,bg=color251]\uE0B0#[fg=color235,bg=color251] \uE725 x \uF403 \uE6211\uF12A #[fg=color251,bg=color235]\uE0B0",
		},
		{
			"New-UnStaged",
			&Status{Branch: "x", RemoteSuccess: true, IsNew: true,
				UnStaged: Area{Modified: 1, Added: 1, Deleted: 1}},
			false,
			"#[fg=color033,bg=color251]\uE0B0#[fg=color235,bg=color251] \uE725 x \uF403 \uE6211\uF4A71-1* #[fg=color251,bg=color235]\uE0B0",
		},
		{
			"New-Stashed",
			&Status{Branch: "x", RemoteSuccess: true, IsNew: true, Stashed: 1},
			false,
			"#[fg=color033,bg=color251]\uE0B0#[fg=color235,bg=color251] \uE725 x \uF403 \uE6211\uE257 #[fg=color251,bg=color235]\uE0B0",
		},
		{
			"Dirty-All",
			&Status{Ahead: 1, Behind: 1, Branch: "x", Unmerged: 1, Untracked: 1, RemoteSuccess: true, Stashed: 1,
				Staged:   Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1},
				UnStaged: Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1}},
			false,
			"#[fg=color033,bg=color209]\uE0B0#[fg=color235,bg=color209] \uE725 x \uF256 \uE6211ﯴ1ﯲ1\uF12A2\uF4A71-1*1易1\uF6905\uF01C1\uE257 #[fg=color209,bg=color235]\uE0B0",
		},
		{
			"No-Tmux-Remote-Fail-Clean",
			&Status{Branch: "x"},
			true,
			"\033[38;5;033m\033[48;5;160m\033[38;5;254m\033[48;5;160m \uF41C \033[38;5;160m\033[48;5;120m\uE0B0\033[38;5;000m\033[48;5;120m \uE725 x \uF62B \033[0m\033[38;5;120m\uE0B0\033[0m",
		},
		{
			"No-Tmux-Clean",
			&Status{Branch: "x", RemoteSuccess: true},
			true,
			"\033[38;5;033m\033[48;5;120m\033[38;5;000m\033[48;5;120m \uE725 x \uF62B \033[0m\033[38;5;120m\uE0B0\033[0m",
		},
		{
			"No-Tmux-Clean-New",
			&Status{Branch: "x", IsNew: true, RemoteSuccess: true},
			true,
			"\033[38;5;033m\033[48;5;251m\033[38;5;235m\033[48;5;251m \uE725 x \uF403 \033[0m\033[38;5;251m\uE0B0\033[0m",
		},
		{
			"No-Tmux-Clean-Gone",
			&Status{Branch: "x", IsGone: true, RemoteSuccess: true},
			true,
			"\033[38;5;033m\033[48;5;088m\033[38;5;255m\033[48;5;088m \uE725 x \uF48E \033[0m\033[38;5;088m\uE0B0\033[0m",
		},
		{
			"No-Tmux-Ahead",
			&Status{Branch: "x", RemoteSuccess: true, IsNew: true, Ahead: 1},
			true,
			"\033[38;5;033m\033[48;5;251m\033[38;5;235m\033[48;5;251m \uE725 x \uF403 \uE6211ﯴ \033[0m\033[38;5;251m\uE0B0\033[0m",
		},
		{
			"No-Tmux-Behind",
			&Status{Branch: "x", RemoteSuccess: true, IsNew: true, Behind: 1},
			true,
			"\033[38;5;033m\033[48;5;251m\033[38;5;235m\033[48;5;251m \uE725 x \uF403 \uE6211ﯲ \033[0m\033[38;5;251m\uE0B0\033[0m",
		},
		{
			"No-Tmux-Unmerged",
			&Status{Branch: "x", RemoteSuccess: true, IsNew: true, Unmerged: 1},
			true,
			"\033[38;5;033m\033[48;5;251m\033[38;5;235m\033[48;5;251m \uE725 x \uF403 \uE6211\uF12A \033[0m\033[38;5;251m\uE0B0\033[0m",
		},
		{
			"No-Tmux-New-UnStaged",
			&Status{Branch: "x", RemoteSuccess: true, IsNew: true,
				UnStaged: Area{Modified: 1, Added: 1, Deleted: 1}},
			true,
			"\033[38;5;033m\033[48;5;251m\033[38;5;235m\033[48;5;251m \uE725 x \uF403 \uE6211\uF4A71-1* \033[0m\033[38;5;251m\uE0B0\033[0m",
		},
		{
			"No-Tmux-New-Stashed",
			&Status{Branch: "x", RemoteSuccess: true, IsNew: true, Stashed: 1},
			true,
			"\033[38;5;033m\033[48;5;251m\033[38;5;235m\033[48;5;251m \uE725 x \uF403 \uE6211\uE257 \033[0m\033[38;5;251m\uE0B0\033[0m",
		},
		{
			"No-Tmux-Dirty-All",
			&Status{Ahead: 1, Behind: 1, Branch: "x", Unmerged: 1, Untracked: 1, RemoteSuccess: true, Stashed: 1,
				Staged:   Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1},
				UnStaged: Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1}},
			true,
			"\033[38;5;033m\033[48;5;209m\033[38;5;235m\033[48;5;209m \uE725 x \uF256 \uE6211ﯴ1ﯲ1\uF12A2\uF4A71-1*1易1\uF6905\uF01C1\uE257 \033[0m\033[38;5;209m\uE0B0\033[0m",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.status.StatusLine(test.noTmux)
			assert.Equal(t, test.expected, actual)
		})
	}
}
