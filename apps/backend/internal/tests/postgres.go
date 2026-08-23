package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/xanity-07/spndex/internal/config"
	"github.com/xanity-07/spndex/internal/database"
)

type PostgresTestDatabase struct {
	Database  *database.Database
	Pool      *pgxpool.Pool
	Container *postgres.PostgresContainer
}

func NewPostgresTestDatabase(t *testing.T) *PostgresTestDatabase {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:17",
		postgres.WithDatabase("spndex_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate postgres container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get postgres host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get postgres port: %v", err)
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     host,
			Port:     port.Port(),
			User:     "postgres",
			Password: "postgres",
			Name:     "spndex_test",
			SSLMode:  "disable",
		},
	}

	logger := zerolog.Nop()

	if err := database.Migrate(ctx, cfg, &logger); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create postgres pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to ping postgres: %v", err)
	}

	return &PostgresTestDatabase{
		Database:  &database.Database{Pool: pool},
		Pool:      pool,
		Container: container,
	}
}
