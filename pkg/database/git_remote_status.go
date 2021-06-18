package database

import (
	"time"
)

func (db *gitDatabase) GitRemoteStatus(id string, updateTime time.Time) (remoteSuccess bool, err error) {
	selectFrom := "SELECT rs, ut FROM git_remote_update WHERE id=$1 AND ut>=$2"
	row := db.QueryRow(selectFrom, id, updateTime)
	err = row.Scan(&remoteSuccess, &updateTime)
	return remoteSuccess, err
}

func (db *gitDatabase) SaveGitRemoteStatus(id string, remoteSuccess bool) error {
	inertOrReplace := "INSERT OR REPLACE INTO git_remote_update (id, rs, ut) VALUES (?, ?, ?)"
	_, err := db.Exec(inertOrReplace, id, remoteSuccess, time.Now())
	return err
}
