package database

import (
	"time"
)

func (db *gitDatabase) GitRemoteStatus(ID string, updateTime time.Time) (remoteSuccess bool, err error) {
	selectFrom := "SELECT rs, ut FROM git_remote_update WHERE id=$1 AND ut>=$2"
	row := db.QueryRow(selectFrom, ID, updateTime)
	err = row.Scan(&remoteSuccess, &updateTime)
	return remoteSuccess, err
}

func (db *gitDatabase) SaveGitRemoteStatus(ID string, remoteSuccess bool) error {
	inertOrReplace := "INSERT OR REPLACE INTO git_remote_update (id, rs, ut) VALUES (?, ?, ?)"
	_, err := db.Exec(inertOrReplace, ID, remoteSuccess, time.Now())
	return err
}
