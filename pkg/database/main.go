package database

import (
	"database/sql"
	"embed"
	"io/fs"
	"os"

	"github.com/rs/zerolog/log"
)

//go:embed assets/*
var assets embed.FS

const (
	dbFile = "/tmp/.git_status_bar.db"
	dsn    = "file:" + dbFile + "?cache=shared"
)

func New() Database {
	return &database{
		&gitDatabase{newDB()},
	}
}

func newDB() *sql.DB {
	setup := false
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		setup = true
	}
	fatal := log.Fatal().Str("dsn", dsn)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		fatal.Err(err).Msg("Could not open assets")
	}

	//setup = true // Force update
	if setup {
		log.Debug().Msg("Setting up assets")
		files, _ := fs.Glob(assets, "assets/schemas/*")
		for _, file := range files {
			query, err := assets.ReadFile(file)
			if err != nil {
				log.Fatal().Err(err).Msg("Could not read schema files")
			}
			_, err = db.Exec(string(query))
			if err != nil {
				fatal.Err(err).Str("schema", file).Msg("Could created schema")
			}
		}
	}
	return db
}