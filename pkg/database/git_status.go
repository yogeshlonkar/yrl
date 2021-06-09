package database

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/yogeshlonkar/yrl/pkg/git"
)

func (db *gitDatabase) GitStatus() (ID string, status *git.Status, updateTime time.Time, err error) {
	data := new([]byte)
	status = new(git.Status)
	row := db.QueryRow("SELECT id, ss, ut FROM git_status WHERE pk=0")
	err = row.Scan(&ID, data, &updateTime)
	if err != nil {
		return
	}
	buf := new(bytes.Buffer)
	buf.Write(*data)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(status)
	if err != nil {
		return
	}
	return ID, status, updateTime, err
}

func (db *gitDatabase) SaveGitStatus(ID string, status *git.Status) (err error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(status)
	if err != nil {
		return
	}
	_, err = db.Exec("INSERT OR REPLACE INTO git_status (pk, id, ss, ut) VALUES (0, ?, ?, ?)", ID, buf.Bytes(), time.Now())
	return
}
