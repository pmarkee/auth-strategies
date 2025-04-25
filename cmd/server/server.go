package main

import (
	"auth-strategies/internal/config"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	r := SetupRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r)
}
