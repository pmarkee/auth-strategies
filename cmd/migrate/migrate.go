package main

import (
	"auth-strategies/internal/config"
	"auth-strategies/internal/db"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.ParseConfig()

	applied, err := db.ApplyMigrations(&cfg.Db)
	if err != nil {
		log.Error().Err(err).Msg("applying database migrations failed")
		return
	} else if !applied {
		log.Info().Msg("database already up-to-date, no migrations applied")
	} else {
		log.Info().Msg("migrations applied")
	}
}
