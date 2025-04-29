package config

import (
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"time"
)

func InitSessionStore(pool *pgxpool.Pool) *scs.SessionManager {
	sessionStore := scs.New()
	sessionStore.Store = pgxstore.New(pool)
	sessionStore.IdleTimeout = 7 * 24 * time.Hour
	sessionStore.Lifetime = 24 * time.Hour

	log.Info().Msg("session storage initialized")
	return sessionStore
}
