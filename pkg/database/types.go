package database

import (
	"database/sql"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/yogeshlonkar/yrl/pkg/git"
)

type Database interface {
	Git() Git
	Close()
}

type database struct {
	*gitDatabase
}

func (d database) Git() Git {
	return d.gitDatabase
}

func (d database) Close() {
	err := d.gitDatabase.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("error while closing gitDatabase")
	}
}

type gitDatabase struct {
	*sql.DB
}

type Git interface {
	GitStatus
	GitRemote
}

type GitStatus interface {
	GitStatus() (ID string, status *git.Status, updateTime time.Time, err error)
	SaveGitStatus(ID string, status *git.Status) error
}

type GitRemote interface {
	GitRemoteStatus(ID string, updateTime time.Time) (remoteSuccess bool, err error)
	SaveGitRemoteStatus(ID string, remoteSuccess bool) error
}
