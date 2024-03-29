package database

import (
	"time"
)

func (db *gitDatabase) GitRemoteStatus(id string, updateTime time.Time) (remoteSuccess bool, err error) {
	selectFrom := "SELECT rs, ut FROM git_remote_update WHERE id=$1 AND ut<=$2"
	row := db.QueryRow(selectFrom, id, updateTime)
	err = row.Scan(&remoteSuccess, &updateTime)
	return remoteSuccess, err
}

func (db *gitDatabase) GetRemoteURL(id string) (remote string, err error) {
	selectFrom := "SELECT r FROM git_remote_update WHERE id=$1"
	row := db.QueryRow(selectFrom, id)
	err = row.Scan(&remote)
	return remote, err
}

func (db *gitDatabase) SaveGitRemoteStatus(id, remote string, remoteSuccess bool, now time.Time) error {
	inertOrReplace := "INSERT OR REPLACE INTO git_remote_update (id, r, rs, ut) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(inertOrReplace, id, remote, remoteSuccess, now)
	return err
}
