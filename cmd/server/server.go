package main

import (
	"auth-strategies/internal/config"
	"auth-strategies/internal/db"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	slogchi "github.com/samber/slog-chi"
	"log/slog"
	"net/http"
)

func SetupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	sc := slogchi.Config{
		WithRequestID: true,
	}
	r.Use(slogchi.NewWithConfig(slog.Default(), sc))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))

	return r
}

func main() {
	config.SetupLogger("debug")

	cfg := config.ParseConfig()

	_, err := db.Connect(context.Background(), &cfg.Db)
	if err != nil {
		log.Error().Err(err).Msg("database connection failed")
		return
	}
	log.Info().Msg("database connection established")

	applied, err := db.ApplyMigrations(&cfg.Db)
	if err != nil {
		log.Error().Err(err).Msg("applying database migrations failed")
		return
	} else if !applied {
		log.Info().Msg("database already up-to-date, no migrations applied")
	} else {
		log.Info().Msg("migrations applied")
	}

	r := SetupRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), r)
}
