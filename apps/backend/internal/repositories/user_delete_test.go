package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xanity-07/spndex/internal/database"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/tests"
)

func TestUserRepository_DeleteUser(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewUserRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

	tests.InsertTestUser(t, ctx, testDB.Pool, userID)

	err := repo.DeleteUser(ctx, &user.DeleteUserPayload{
		ID: userID.String(),
	})
	if err != nil {
		t.Fatalf("DeleteUser returned an error: %v", err)
	}

	var deletedAt *time.Time

	err = testDB.Pool.QueryRow(
		ctx,
		`
			SELECT deleted_at
			FROM users
			WHERE id = @user_id
		`,
		pgx.NamedArgs{
			"user_id": userID,
		},
	).Scan(&deletedAt)

	if err != nil {
		t.Fatalf("failed to query deleted user: %v", err)
	}

	if deletedAt == nil {
		t.Fatal("expected deleted_at to be populated")
	}
}

func TestUserRepository_DeleteUser_NotFound(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewUserRepository(srv)
	ctx := context.Background()

	err := repo.DeleteUser(ctx, &user.DeleteUserPayload{
		ID: uuid.NewString(),
	})

	if err == nil {
		t.Fatal("expected USER_NOT_FOUND error, got nil")
	}
}

func TestUserRepository_DeleteUser_AlreadyDeleted(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewUserRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

	tests.InsertTestUser(t, ctx, testDB.Pool, userID)

	_, err := testDB.Pool.Exec(
		ctx,
		`
			UPDATE users
			SET deleted_at = NOW()
			WHERE id = @user_id
		`,
		pgx.NamedArgs{
			"user_id": userID,
		},
	)
	if err != nil {
		t.Fatalf("failed to soft delete test user: %v", err)
	}

	err = repo.DeleteUser(ctx, &user.DeleteUserPayload{
		ID: userID.String(),
	})

	if err == nil {
		t.Fatal("expected USER_NOT_FOUND error, got nil")
	}
}
