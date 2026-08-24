package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertTestUser(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
) {
	t.Helper()

	stmt := `
		INSERT INTO users (
			id,
			first_name,
			last_name,
			email,
			password_hash
		)
		VALUES (
			@user_id,
			@first_name,
			@last_name,
			@email,
			@password_hash
		)
	`

	args := pgx.NamedArgs{
		"user_id":       userID,
		"first_name":    "John",
		"last_name":     "Doe",
		"email":         "john-" + userID.String() + "@example.com",
		"password_hash": "hashed-password",
	}

	_, err := pool.Exec(ctx, stmt, args)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
}

func InsertTestUserEmail(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
	email string,
) {
	t.Helper()

	stmt := `
		INSERT INTO users (
			id,
			first_name,
			last_name,
			email,
			password_hash
		)
		VALUES (
			@user_id,
			@first_name,
			@last_name,
			@email,
			@password_hash
		)
	`

	args := pgx.NamedArgs{
		"user_id":       userID,
		"first_name":    "John",
		"last_name":     "Doe",
		"email":         email,
		"password_hash": "hashed-password",
	}

	_, err := pool.Exec(ctx, stmt, args)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
}
