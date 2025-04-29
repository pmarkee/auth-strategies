package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

func (api *Api) BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="user"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		payload, err := parseBasicAuth(auth)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="user"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		id, err := api.s.checkPassword(r.Context(), payload.Email, payload.Password)
		if errors.Is(err, errInvalidCredentials) {
			w.Header().Set("WWW-Authenticate", `Basic realm="user"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error().Err(err).Msgf("basic auth failed: %s", err)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "id", id)))
	})
}

type basicAuthPayload struct {
	Email    string
	Password string
}

func parseBasicAuth(auth string) (basicAuthPayload, error) {
	if !strings.HasPrefix(auth, "Basic ") {
		return basicAuthPayload{}, fmt.Errorf("invalid authentication method")
	}

	enc := strings.TrimPrefix(auth, "Basic ")
	dec, err := base64Decode(enc)
	if err != nil {
		return basicAuthPayload{}, err
	}

	s := strings.SplitN(dec, ":", 2)
	if len(s) != 2 {
		return basicAuthPayload{}, fmt.Errorf("invalid basic auth format")
	}

	return basicAuthPayload{
		Email:    s[0],
		Password: s[1],
	}, nil
}

func base64Decode(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
