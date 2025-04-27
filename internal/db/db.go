package db

import (
	"auth-strategies/internal/config"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cfg *config.DbConfig) (*pgxpool.Pool, error) {
	connStr := buildConnStr(cfg)
	pgxConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	db, err := pgxpool.New(ctx, pgxConfig.ConnString())
	if err != nil {
		return nil, err
	}
	return db, nil
}

func buildConnStr(cfg *config.DbConfig) string {
	template := "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Europe/Vienna"
	return fmt.Sprintf(template, cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
}
