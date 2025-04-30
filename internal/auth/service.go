package auth

import (
	"auth-strategies/internal/db/repository"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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

func (s *Service) generateApiKey(ctx context.Context, userId *uuid.UUID) (string, error) {
	repo := repository.New(s.pool)

	key, err := _generateApiKey(ctx, repo.ApiKeyPublicIdTaken)
	if err != nil {
		return "", err
	}

	secretSalt, err := generateSalt()
	if err != nil {
		return "", err
	}

	secretHash := computeHash(key.secret, secretSalt)
	params := repository.CreateApiKeyParams{
		UserID:     *userId,
		PublicID:   key.publicId,
		SecretHash: secretHash,
		SecretSalt: secretSalt,
	}
	if err := repo.CreateApiKey(ctx, params); err != nil {
		return "", fmt.Errorf("failed to create api key: %w", err)
	}

	return fmt.Sprintf("%s.%s", key.publicId, key.secret), nil
}

type publicIdTakenFunc func(context.Context, string) (bool, error)

type apiKey struct {
	publicId string
	secret   string
}

func _generateApiKey(ctx context.Context, checkTaken publicIdTakenFunc) (*apiKey, error) {
	publicId, err := generateApiKeyPublicId(ctx, checkTaken, 16, 10)
	if err != nil {
		return nil, err
	}

	secret, err := generateRandomHex(32)
	if err != nil {
		return nil, err
	}

	return &apiKey{publicId, secret}, nil
}

func generateApiKeyPublicId(ctx context.Context, checkTaken publicIdTakenFunc, length int, retries int) (string, error) {
	for range retries {
		id, err := generateRandomHex(length)
		if err != nil {
			return "", err
		}

		taken, err := checkTaken(ctx, id)
		if err != nil {
			return "", err
		}
		if !taken {
			return id, nil
		}
	}

	return "", fmt.Errorf("no unused api key public id of length %d found in %d retries", length, retries)
}

func generateRandomHex(length int) (string, error) {
	var b = make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	s := hex.EncodeToString(b)
	return s, nil
}

var (
	errApiKeyInvalid = errors.New("invalid api key")
)

func (s *Service) validateApiKey(ctx context.Context, inputApiKey *apiKey) (*uuid.UUID, error) {
	repo := repository.New(s.pool)
	dbApiKey, err := repo.FindApiKey(ctx, inputApiKey.publicId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: admin api key not found", errApiKeyInvalid)
	} else if err != nil {
		return nil, fmt.Errorf("error fetching admin api key: %w", err)
	}

	inputHash := computeHash(inputApiKey.secret, dbApiKey.SecretSalt)
	if !bytes.Equal(dbApiKey.SecretHash, inputHash) {
		return nil, fmt.Errorf("%w: admin api key secret invalid", errApiKeyInvalid)
	}

	return &dbApiKey.UserID, nil
}
