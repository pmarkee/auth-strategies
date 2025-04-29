package auth

import (
	"auth-strategies/internal/db/repository"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool}
}

var (
	DbError            = errors.New("database error")
	InvalidCredentials = errors.New("invalid credentials")
)

func (s *Service) CheckPassword(ctx context.Context, email, password string) (*uuid.UUID, error) {
	repo := repository.New(s.pool)
	authInfo, err := repo.GetPasswordAuth(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		// Differentiate unknown email address from db error
		return nil, InvalidCredentials
	} else if err != nil {
		return nil, fmt.Errorf("%w: %w", DbError, err)
	}

	inputPWHash := ComputeHash(password, authInfo.PwSalt)
	if !bytes.Equal(inputPWHash, authInfo.PwHash) {
		return nil, InvalidCredentials
	}

	return &authInfo.ID, nil
}
