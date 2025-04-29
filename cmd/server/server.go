package main

import (
	"auth-strategies/internal/config"
	"auth-strategies/internal/db"
	"auth-strategies/internal/user"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	slogchi "github.com/samber/slog-chi"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"

	_ "auth-strategies/docs"
)

func SetupRouter(pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	sc := slogchi.Config{
		WithRequestID: true,
	}
	r.Use(slogchi.NewWithConfig(slog.Default(), sc))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))

	userRouter := chi.NewRouter()
	userApi := user.NewApi(user.NewService(pool))
	userRouter.Get("/", userApi.GetUserInfo)
	r.Mount("/user", userRouter)

	return r
}

// @title			Auth Strategies Showcase
// @version		1
// @description	These are the API docs for my showcase of auth strategies in Go.
// @contact.name	Peter Marki
// @contact.url	https://github.com/pmarkee
// @host			localhost:8080
// @BasePath		/
// @accept			json
// @produce		json
func main() {
	config.SetupLogger("debug")

	cfg := config.ParseConfig()

	pool, err := db.Connect(context.Background(), &cfg.Db)
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

	r := SetupRouter(pool)
	r.Get("/docs/*", httpSwagger.Handler())

	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), r)
}
