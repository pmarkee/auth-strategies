package main

import (
	"auth-strategies/internal/auth"
	"auth-strategies/internal/config"
	"auth-strategies/internal/db"
	"auth-strategies/internal/user"
	"context"
	"fmt"
	"github.com/alexedwards/scs/v2"
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

func SetupRouter(pool *pgxpool.Pool, sessionStore *scs.SessionManager, hmacSecret []byte) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	sc := slogchi.Config{
		WithRequestID: true,
	}
	r.Use(slogchi.NewWithConfig(slog.Default(), sc))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))

	authRouter := chi.NewRouter()
	authService := auth.NewService(pool)
	authApi := auth.NewApi(authService, sessionStore, hmacSecret)
	authRouter.Post("/register", authApi.Register)
	authRouter.Post("/login", authApi.Login)
	authRouter.Post("/token/login", authApi.LoginToken)
	authRouter.Post("/logout", authApi.Logout)
	authRouter.With(authApi.SessionAuth).Get("/api-key", authApi.GenerateApiKey)
	r.Mount("/auth", authRouter)

	userRouter := chi.NewRouter()
	userApi := user.NewApi(user.NewService(pool))
	userRouter.With(authApi.BasicAuth).Get("/basic", userApi.GetUserInfoBasic)
	userRouter.With(authApi.SessionAuth).Get("/session", userApi.GetUserInfoSession)
	userRouter.With(authApi.TokenAuth).Get("/token", userApi.GetUserInfoToken)
	userRouter.With(authApi.ApiKeyAuth).Get("/api-key", userApi.GetUserInfoApiKey)
	r.Mount("/user", userRouter)

	return r
}

// @title						Auth Strategies Showcase
// @version					1
// @description				These are the API docs for my showcase of auth strategies in Go.
// @contact.name				Peter Marki
// @contact.url				https://github.com/pmarkee
// @host						localhost:8080
// @BasePath					/
// @accept						json
// @produce					json
//
// @securityDefinitions.basic	BasicAuth
// @in							header
// @name						X-API-KEY
// @description				API key passed in header X-API-KEY
//
// @securityDefinitions.apiKey	session
// @in							cookie
// @name						session
// @description				session cookie
//
// @securityDefinitions.apiKey	Bearer
// @in							header
// @name						Authorization
// @description				Enter the token with the "Bearer " prefix
//
// @securityDefinitions.apiKey	ApiKey
// @in							header
// @name						X-API-Key
// @description				API key passed in the X-API-Key header
func main() {
	config.SetupLogger("debug")

	cfg := config.ParseConfig()

	pool, err := db.Connect(context.Background(), &cfg.Db)
	if err != nil {
		log.Fatal().Err(err).Msg("database connection failed")
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

	sessionStore := config.InitSessionStore(pool)

	r := SetupRouter(pool, sessionStore, []byte(cfg.Server.HmacSecret))
	r.Get("/*", httpSwagger.Handler())

	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), sessionStore.LoadAndSave(r))
}
