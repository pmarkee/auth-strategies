//go:generate sqlc generate
package db

import (
	"auth-strategies/internal/config"
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect establish connection to the database with the provided config
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

// ApplyMigrations apply migrations from the "migrations" embed FS to the provided db
// Return a boolean signaling whether migrations were applied and an error.
func ApplyMigrations(cfg *config.DbConfig) (bool, error) {
	migDir, err := iofs.New(Migrations, "migrations")
	if err != nil {
		return false, fmt.Errorf("reading migrations directory from embed failed: %w", err)
	}

	dbUrl := buildDbUrl(cfg)
	m, err := migrate.NewWithSourceInstance("iofs", migDir, dbUrl)
	if err != nil {
		return false, fmt.Errorf("creating migrate instance failed: %w", err)
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return true, nil
}

func buildDbUrl(cfg *config.DbConfig) string {
	urlTemplate := "pgx://%s:%s@%s:%d/%s?sslmode=disable"
	return fmt.Sprintf(urlTemplate, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}
