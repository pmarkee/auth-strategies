package user

import (
	"auth-strategies/internal/db/repository"
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool}
}

type userDataRs struct {
	FirstName string
	LastName  string
}

func (s *Service) getUserData(ctx context.Context, id uuid.UUID) (*userDataRs, error) {
	repo := repository.New(s.pool)
	info, err := repo.GetUserInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	return &userDataRs{
		FirstName: info.FirstName,
		LastName:  info.LastName,
	}, nil
}
