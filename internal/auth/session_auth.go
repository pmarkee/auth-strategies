package auth

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (api *Api) SessionAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIdStr := api.sessionStore.GetString(r.Context(), "user_id")
		if userIdStr == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userId, err := uuid.Parse(userIdStr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error().Err(err).Msg("user id stored in session is not a valid UUID")
			if err := api.sessionStore.Destroy(r.Context()); err != nil {
				log.Error().Err(err).Msg("failed to destroy faulty session (invalid UUID)")
			}
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "id", &userId)))
	})
}
