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
	errDb                 = errors.New("database error")
	errInvalidCredentials = errors.New("invalid credentials")
)

func (s *Service) checkPassword(ctx context.Context, email, password string) (*uuid.UUID, error) {
	repo := repository.New(s.pool)
	authInfo, err := repo.GetPasswordAuth(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		// Differentiate unknown email address from db error
		return nil, errInvalidCredentials
	} else if err != nil {
		return nil, fmt.Errorf("%w: %w", errDb, err)
	}

	inputPWHash := computeHash(password, authInfo.PwSalt)
	if !bytes.Equal(inputPWHash, authInfo.PwHash) {
		return nil, errInvalidCredentials
	}

	return &authInfo.ID, nil
}

type registerRq struct {
	email     string
	password  string
	firstName string
	lastName  string
}

var (
	errEmailTaken = errors.New("email already taken")
)

// register validate the email provided in rq is not taken, and create a new user
// with password-based authentication. Email verification is out-of-scope.
func (s *Service) register(ctx context.Context, rq *registerRq) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback(ctx)

	repo := repository.New(tx)

	emailTaken, err := repo.EmailTaken(ctx, rq.email)
	if err != nil {
		return fmt.Errorf("failed querying user by email: %w", err)
	}
	if emailTaken {
		return errEmailTaken
	}

	createUserParams := repository.CreateUserParams{
		Email:     rq.email,
		FirstName: rq.firstName,
		LastName:  rq.lastName,
	}
	userId, err := repo.CreateUser(ctx, createUserParams)
	if err != nil {
		return fmt.Errorf("failed creating user: %w", err)
	}

	pwSalt, err := generateSalt()
	if err != nil {
		return fmt.Errorf("failed generating salt: %w", err)
	}
	pwHash := computeHash(rq.password, pwSalt)

	createPasswordAuthParams := repository.CreatePasswordAuthParams{
		UserID: userId,
		PwHash: pwHash,
		PwSalt: pwSalt,
	}
	err = repo.CreatePasswordAuth(ctx, createPasswordAuthParams)
	if err != nil {
		return fmt.Errorf("failed to create password auth: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}
