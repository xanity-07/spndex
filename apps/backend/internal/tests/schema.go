package tests

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupSpndexSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()

	statements := []string{
		`
		CREATE TYPE expense_category AS ENUM (
			'food',
			'transport',
			'education',
			'utilities',
			'entertainment',
			'healthcare',
			'shopping',
			'other'
		)
		`,
		`
		CREATE TYPE user_roles AS ENUM (
			'user',
			'admin'
		)
		`,
		`
		CREATE TYPE currency_code AS ENUM (
			'USD',
			'EUR',
			'GBP',
			'CAD',
			'AUD'
		)
		`,
		`
		CREATE TABLE users (
			id UUID PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			user_role user_roles NOT NULL DEFAULT 'user',
			deleted_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
		`,
		`
		CREATE TABLE expenses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id),
			amount BIGINT NOT NULL,
			description TEXT,
			currency currency_code NOT NULL,
			category expense_category NOT NULL,
			date TEXT NOT NULL,
			deleted_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
		`,
	}

	for _, stmt := range statements {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("failed to execute schema statement: %v", err)
		}
	}
}
