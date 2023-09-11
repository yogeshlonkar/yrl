package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusLine_tmux(t *testing.T) {
	tests := []struct {
		name     string
		status   *Status
		expected string
	}{
		{
			"Loading",
			&Status{Branch: "branch1", Loading: true},
			"#[fg=color254,bg=color056] \U000f1378 #[fg=color056,bg=color120]\ue0b0#[fg=color000,bg=color120]\uf418 branch1\ue621#[fg=color22,bg=color120]\uebb1#[fg=color235,bg=color120] #[fg=color120,bg=color235]\ue0b0",
		},
		{
			"Remote-Fail",
			&Status{Branch: "branch1"},
			"#[fg=color025,bg=color160]\ue0b0#[fg=color254,bg=color160] \U000f04e7 #[fg=color160,bg=color120]\ue0b0#[fg=color000,bg=color120]\uf418 branch1\ue621#[fg=color22,bg=color120]\uebb1#[fg=color235,bg=color120] #[fg=color120,bg=color235]\ue0b0",
		},
		{
			"Clean",
			&Status{Branch: "branch1", RemoteSuccess: true},
			"#[fg=color025,bg=color120] #[fg=color000,bg=color120] branch1#[fg=color22,bg=color120]#[fg=color235,bg=color120] #[fg=color120,bg=color235]",
		},
		{
			"New",
			&Status{Branch: "branch1", IsNew: true, RemoteSuccess: true},
			"#[fg=color025,bg=color251] #[fg=color235,bg=color251] branch1#[fg=color33,bg=color251] #[fg=color235,bg=color251]#[fg=color251,bg=color235]",
		},
		{
			"Gone",
			&Status{Branch: "branch1", IsGone: true, RemoteSuccess: true},
			"#[fg=color025,bg=color088] #[fg=color255,bg=color088] branch1 #[fg=color088,bg=color235]",
		},
		{
			"Feature",
			&Status{Branch: "feat/branch1", RemoteSuccess: true},
			"#[fg=color025,bg=color120] #[fg=color000,bg=color120]  branch1#[fg=color22,bg=color120]#[fg=color235,bg=color120] #[fg=color120,bg=color235]",
		},
		{
			"Bugfix",
			&Status{Branch: "bugfix/branch1", RemoteSuccess: true},
			"#[fg=color025,bg=color120] #[fg=color000,bg=color120]  branch1#[fg=color22,bg=color120]#[fg=color235,bg=color120] #[fg=color120,bg=color235]",
		},
		{
			"Hotfix",
			&Status{Branch: "hotfix/branch1", RemoteSuccess: true},
			"#[fg=color025,bg=color120] #[fg=color000,bg=color120]  branch1#[fg=color22,bg=color120]#[fg=color235,bg=color120] #[fg=color120,bg=color235]",
		},
		{
			"Chore",
			&Status{Branch: "chore/branch1", RemoteSuccess: true},
			"#[fg=color025,bg=color120] #[fg=color000,bg=color120]\uf418 \U000f19a1 branch1\ue621#[fg=color22,bg=color120]\uebb1#[fg=color235,bg=color120] #[fg=color120,bg=color235]\ue0b0",
		},
		{
			"Branch-Too-Long",
			&Status{Branch: "this-is-a-very-long-branch-name", RemoteSuccess: true},
			"#[fg=color025,bg=color120] #[fg=color000,bg=color120] this-is-...ranch-name#[fg=color22,bg=color120]#[fg=color235,bg=color120] #[fg=color120,bg=color235]",
		},
		{
			"Ahead",
			&Status{Branch: "branch1", RemoteSuccess: true, Ahead: 1},
			"#[fg=color025,bg=color209] #[fg=color235,bg=color209] branch1#[fg=color088,bg=color209]#[fg=color235,bg=color209]1#[fg=color24,bg=color209] #[fg=color235,bg=color209]#[fg=color235,bg=color209]#[fg=color209,bg=color235]",
		},
		{
			"Behind",
			&Status{Branch: "branch1", RemoteSuccess: true, Behind: 1},
			"#[fg=color025,bg=color209] #[fg=color235,bg=color209] branch1#[fg=color088,bg=color209]#[fg=color235,bg=color209]1#[fg=color24,bg=color209] #[fg=color235,bg=color209]#[fg=color235,bg=color209]#[fg=color209,bg=color235]",
		},
		{
			"Unmerged",
			&Status{Branch: "branch1", RemoteSuccess: true, Unmerged: 1},
			"#[fg=color025,bg=color209] #[fg=color235,bg=color209]\uf418 branch1\ue621#[fg=color088,bg=color209]#[fg=color235,bg=color209]1#[fg=color160,bg=color209]\U000f1a98 #[fg=color235,bg=color209]#[fg=color209,bg=color235]\ue0b0",
		},
		{
			"UnStaged",
			&Status{Branch: "branch1", RemoteSuccess: true, UnStaged: Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1}},
			"#[fg=color025,bg=color209] #[fg=color235,bg=color209]\uf418 branch1\ue621#[fg=color088,bg=color209]#[fg=color235,bg=color209]1\U000f034c 1\U000f0ad3 1\uebcb 1\U000f0191 1\U000f1787 #[fg=color209,bg=color235]\ue0b0",
		},
		{
			"Staged",
			&Status{Branch: "branch1", RemoteSuccess: true, Staged: Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1}},
			"#[fg=color025,bg=color209] #[fg=color235,bg=color209]\uf418 branch1\ue621#[fg=color088,bg=color209]#[fg=color235,bg=color209]#[fg=color22,bg=color209]\uf01c #[fg=color235,bg=color209]1\U000f034c 1\U000f0ad3 1\uebcb 1\U000f0191 1\U000f1787 #[fg=color209,bg=color235]\ue0b0",
		},
		{
			"Stashed",
			&Status{Branch: "branch1", RemoteSuccess: true, Stashed: 1},
			"#[fg=color025,bg=color120] #[fg=color000,bg=color120] branch1#[fg=color22,bg=color120]#[fg=color235,bg=color120] 1#[fg=color53,bg=color120] #[fg=color235,bg=color120]#[fg=color120,bg=color235]",
		},
		{
			"Dirty-All",
			&Status{Ahead: 1, Behind: 1, Branch: "branch1", Unmerged: 1, Untracked: 1, RemoteSuccess: true, Stashed: 1, Staged: Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1}, UnStaged: Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1}},
			"#[fg=color025,bg=color209] #[fg=color235,bg=color209]\uf418 branch1\ue621#[fg=color088,bg=color209]#[fg=color235,bg=color209]1#[fg=color24,bg=color209]\uf40a #[fg=color235,bg=color209]1#[fg=color24,bg=color209]\uf409 #[fg=color235,bg=color209]1#[fg=color160,bg=color209]\U000f1a98 #[fg=color235,bg=color209]|2\U000f034c 1\U000f0ad3 1\uebcb 1\U000f0191 1\U000f1787 |#[fg=color22,bg=color209]\uf01c #[fg=color235,bg=color209]1\U000f034c 1\U000f0ad3 1\uebcb 1\U000f0191 1\U000f1787 |1#[fg=color53,bg=color209]\ue257 #[fg=color235,bg=color209]#[fg=color209,bg=color235]\ue0b0",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.status.StatusLine(false)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStatusLine_no_tmux(t *testing.T) {
	tests := []struct {
		name     string
		status   *Status
		expected string
	}{
		{
			"Loading",
			&Status{Branch: "branch1", Loading: true},
			"\x1b[38;5;254m\x1b[48;5;056m \U000f1378 \x1b[38;5;056m\x1b[48;5;120m\ue0b0\x1b[38;5;000m\x1b[48;5;120m\uf418 branch1\ue621\x1b[38;5;22m\x1b[48;5;120m\uebb1\x1b[38;5;235m\x1b[48;5;120m \x1b[0m\x1b[38;5;120m\ue0b0\x1b[0m",
		},
		{
			"Remote-Fail",
			&Status{Branch: "branch1"},
			"\x1b[38;5;025m\x1b[48;5;160m\x1b[38;5;254m\x1b[48;5;160m \U000f04e7 \x1b[38;5;160m\x1b[48;5;120m\ue0b0\x1b[38;5;000m\x1b[48;5;120m\uf418 branch1\ue621\x1b[38;5;22m\x1b[48;5;120m\uebb1\x1b[38;5;235m\x1b[48;5;120m \x1b[0m\x1b[38;5;120m\ue0b0\x1b[0m",
		},
		{
			"Clean",
			&Status{Branch: "branch1", RemoteSuccess: true},
			"[38;5;025m[48;5;120m [38;5;000m[48;5;120m branch1[38;5;22m[48;5;120m[38;5;235m[48;5;120m [0m[38;5;120m[0m",
		},
		{
			"New",
			&Status{Branch: "branch1", IsNew: true, RemoteSuccess: true},
			"[38;5;025m[48;5;251m [38;5;235m[48;5;251m branch1[38;5;33m[48;5;251m [38;5;235m[48;5;251m[0m[38;5;251m[0m",
		},
		{
			"Gone",
			&Status{Branch: "branch1", IsGone: true, RemoteSuccess: true},
			"[38;5;025m[48;5;088m [38;5;255m[48;5;088m branch1 [0m[38;5;088m[0m",
		},
		{
			"Feature",
			&Status{Branch: "feat/branch1", RemoteSuccess: true},
			"[38;5;025m[48;5;120m [38;5;000m[48;5;120m  branch1[38;5;22m[48;5;120m[38;5;235m[48;5;120m [0m[38;5;120m[0m",
		},
		{
			"Bugfix",
			&Status{Branch: "bugfix/branch1", RemoteSuccess: true},
			"[38;5;025m[48;5;120m [38;5;000m[48;5;120m  branch1[38;5;22m[48;5;120m[38;5;235m[48;5;120m [0m[38;5;120m[0m",
		},
		{
			"Hotfix",
			&Status{Branch: "hotfix/branch1", RemoteSuccess: true},
			"[38;5;025m[48;5;120m [38;5;000m[48;5;120m  branch1[38;5;22m[48;5;120m[38;5;235m[48;5;120m [0m[38;5;120m[0m",
		},
		{
			"Chore",
			&Status{Branch: "chore/branch1", RemoteSuccess: true},
			"\x1b[38;5;025m\x1b[48;5;120m \x1b[38;5;000m\x1b[48;5;120m\uf418 \U000f19a1 branch1\ue621\x1b[38;5;22m\x1b[48;5;120m\uebb1\x1b[38;5;235m\x1b[48;5;120m \x1b[0m\x1b[38;5;120m\ue0b0\x1b[0m",
		},
		{
			"Branch-Too-Long",
			&Status{Branch: "this-is-a-very-long-branch-name", RemoteSuccess: true},
			"[38;5;025m[48;5;120m [38;5;000m[48;5;120m this-is-...ranch-name[38;5;22m[48;5;120m[38;5;235m[48;5;120m [0m[38;5;120m[0m",
		},
		{
			"Ahead",
			&Status{Branch: "branch1", RemoteSuccess: true, Ahead: 1},
			"[38;5;025m[48;5;209m [38;5;235m[48;5;209m branch1[38;5;088m[48;5;209m[38;5;235m[48;5;209m1[38;5;24m[48;5;209m [38;5;235m[48;5;209m[38;5;235m[48;5;209m[0m[38;5;209m[0m",
		},
		{
			"Behind",
			&Status{Branch: "branch1", RemoteSuccess: true, Behind: 1},
			"[38;5;025m[48;5;209m [38;5;235m[48;5;209m branch1[38;5;088m[48;5;209m[38;5;235m[48;5;209m1[38;5;24m[48;5;209m [38;5;235m[48;5;209m[38;5;235m[48;5;209m[0m[38;5;209m[0m",
		},
		{
			"Unmerged",
			&Status{Branch: "branch1", RemoteSuccess: true, Unmerged: 1},
			"\x1b[38;5;025m\x1b[48;5;209m \x1b[38;5;235m\x1b[48;5;209m\uf418 branch1\ue621\x1b[38;5;088m\x1b[48;5;209m\x1b[38;5;235m\x1b[48;5;209m1\x1b[38;5;160m\x1b[48;5;209m\U000f1a98 \x1b[38;5;235m\x1b[48;5;209m\x1b[0m\x1b[38;5;209m\ue0b0\x1b[0m",
		},
		{
			"UnStaged",
			&Status{Branch: "branch1", RemoteSuccess: true, UnStaged: Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1}},
			"\x1b[38;5;025m\x1b[48;5;209m \x1b[38;5;235m\x1b[48;5;209m\uf418 branch1\ue621\x1b[38;5;088m\x1b[48;5;209m\x1b[38;5;235m\x1b[48;5;209m1\U000f034c 1\U000f0ad3 1\uebcb 1\U000f0191 1\U000f1787 \x1b[0m\x1b[38;5;209m\ue0b0\x1b[0m",
		},
		{
			"Staged",
			&Status{Branch: "branch1", RemoteSuccess: true, Staged: Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1}},
			"\x1b[38;5;025m\x1b[48;5;209m \x1b[38;5;235m\x1b[48;5;209m\uf418 branch1\ue621\x1b[38;5;088m\x1b[48;5;209m\x1b[38;5;235m\x1b[48;5;209m\x1b[38;5;22m\x1b[48;5;209m\uf01c \x1b[38;5;235m\x1b[48;5;209m1\U000f034c 1\U000f0ad3 1\uebcb 1\U000f0191 1\U000f1787 \x1b[0m\x1b[38;5;209m\ue0b0\x1b[0m",
		},
		{
			"Stashed",
			&Status{Branch: "branch1", RemoteSuccess: true, Stashed: 1},
			"[38;5;025m[48;5;120m [38;5;000m[48;5;120m branch1[38;5;22m[48;5;120m[38;5;235m[48;5;120m 1[38;5;53m[48;5;120m [38;5;235m[48;5;120m[0m[38;5;120m[0m",
		},
		{
			"Dirty-All",
			&Status{Ahead: 1, Behind: 1, Branch: "branch1", Unmerged: 1, Untracked: 1, RemoteSuccess: true, Stashed: 1, Staged: Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1}, UnStaged: Area{Modified: 1, Added: 1, Deleted: 1, Renamed: 1, Copied: 1}},
			"\x1b[38;5;025m\x1b[48;5;209m \x1b[38;5;235m\x1b[48;5;209m\uf418 branch1\ue621\x1b[38;5;088m\x1b[48;5;209m\x1b[38;5;235m\x1b[48;5;209m1\x1b[38;5;24m\x1b[48;5;209m\uf40a \x1b[38;5;235m\x1b[48;5;209m1\x1b[38;5;24m\x1b[48;5;209m\uf409 \x1b[38;5;235m\x1b[48;5;209m1\x1b[38;5;160m\x1b[48;5;209m\U000f1a98 \x1b[38;5;235m\x1b[48;5;209m|2\U000f034c 1\U000f0ad3 1\uebcb 1\U000f0191 1\U000f1787 |\x1b[38;5;22m\x1b[48;5;209m\uf01c \x1b[38;5;235m\x1b[48;5;209m1\U000f034c 1\U000f0ad3 1\uebcb 1\U000f0191 1\U000f1787 |1\x1b[38;5;53m\x1b[48;5;209m\ue257 \x1b[38;5;235m\x1b[48;5;209m\x1b[0m\x1b[38;5;209m\ue0b0\x1b[0m",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.status.StatusLine(true)
			assert.Equal(t, test.expected, actual)
		})
	}
}
