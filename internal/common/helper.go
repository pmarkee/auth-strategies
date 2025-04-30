package common

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net/http"
)

func GetUserIdFromContext(w http.ResponseWriter, r *http.Request) *uuid.UUID {
	id, ok := r.Context().Value("id").(*uuid.UUID)
	if !ok {
		// Should not be reached
		log.Error().Msg("failed to read user id from context")
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	return id
}
