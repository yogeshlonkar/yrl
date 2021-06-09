package model

import (
	"os"

	"github.com/yogeshlonkar/yrl/pkg/git"
)

type GitRepo struct {
	*git.Status
	os.FileInfo
	ID           string
	Root         string
	AbsPath      string
	Cached       bool
	StatusString string
}
