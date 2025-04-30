package auth

import (
	"auth-strategies/internal/common"
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

func (api *Api) TokenAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			common.WriteJSON(w, http.StatusUnauthorized, common.ErrorResponse{Error: "Missing Authorization header"})
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		id, err := validateToken(api.hmacSecret, tokenString)
		if errors.Is(err, errInvalidToken) {
			common.WriteJSON(w, http.StatusUnauthorized, common.ErrorResponse{Error: "Invalid token"})
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error().Err(err).Msg("valid JWT but parsing claims failed")
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "id", id)))
	})
}

var (
	errInvalidToken     = errors.New("invalid token")
	errClaimsCastFailed = errors.New("failed to cast jwt claims to MapClaims")
	errInvalidClaims    = errors.New("invalid claims")
)

func validateToken(hmacSecret []byte, tokenString string) (*uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("%w: %w", errInvalidToken, err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); !ok {
		return nil, errClaimsCastFailed
	} else {
		idStr, err := claims.GetSubject()
		if err != nil {
			return nil, fmt.Errorf("%w: missing subject in JWT: %w", errInvalidClaims, err)
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("%w: subject is not a valid UUID: %w", errInvalidClaims, err)
		}
		return &id, nil
	}
}
