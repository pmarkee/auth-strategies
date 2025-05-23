package common

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"net/http"
)

// WriteJSON writes any struct into an http response as json
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("failed to json-stringify response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// SuccessResponse generic HTTP response returned in a happy-case, contains a status message
type SuccessResponse struct {
	Status string `json:"status" example:"success"`
}

// ErrorResponse generic HTTP response returned in an error-case, contains an error message
type ErrorResponse struct {
	Error string `json:"error" example:"error""`
}
