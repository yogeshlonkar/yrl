package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/yogeshlonkar/yrl/cmd"
)

func main() {
	app := cmd.NewApp()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}
