package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

func (api *Api) ApiKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawKey := r.Header.Get("X-API-Key")
		if rawKey == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		key, err := parseApiKey(rawKey)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userId, err := api.s.validateApiKey(r.Context(), key)
		if errors.Is(err, errApiKeyInvalid) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err != nil {
			log.Error().Err(err).Msg("admin API key validation failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "id", userId)))
	})
}

func parseApiKey(rawKey string) (*apiKey, error) {
	parts := strings.Split(rawKey, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid api key format")
	}
	return &apiKey{publicId: parts[0], secret: parts[1]}, nil
}
