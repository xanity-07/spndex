package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xanity-07/spndex/internal/database"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/tests"
)

func TestUserRepository_GetUserByID(t *testing.T) {
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

	result, err := repo.GetUserByID(ctx, &user.GetUserByIDPayload{
		ID: userID.String(),
	})
	if err != nil {
		t.Fatalf("GetUserByID returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user, got nil")
	}

	if result.ID != userID {
		t.Errorf("expected user ID %q, got %q", userID, result.ID)
	}

	if result.FirstName != "John" {
		t.Errorf("expected first name %q, got %q", "John", result.FirstName)
	}

	if result.LastName != "Doe" {
		t.Errorf("expected last name %q, got %q", "Doe", result.LastName)
	}

	if result.Email != "john-"+userID.String()+"@example.com" {
		t.Errorf(
			"expected email %q, got %q",
			"john-"+userID.String()+"@example.com",
			result.Email,
		)
	}

	if result.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be populated")
	}

	if result.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be populated")
	}

	if result.DeletedAt != nil {
		t.Error("expected DeletedAt to be nil")
	}
}

func TestUserRepository_GetUserByID_NotFound(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewUserRepository(srv)
	ctx := context.Background()

	result, err := repo.GetUserByID(ctx, &user.GetUserByIDPayload{
		ID: uuid.NewString(),
	})

	if err == nil {
		t.Fatal("expected USER_NOT_FOUND error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil user, got %+v", result)
	}
}

func TestUserRepository_GetUserByID_DeletedUser(t *testing.T) {
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

	result, err := repo.GetUserByID(ctx, &user.GetUserByIDPayload{
		ID: userID.String(),
	})

	if err == nil {
		t.Fatal("expected USER_NOT_FOUND error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil user, got %+v", result)
	}
}

func TestUserRepository_GetUserByID_MultipleUsers(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewUserRepository(srv)
	ctx := context.Background()

	firstUserID := uuid.New()
	secondUserID := uuid.New()

	tests.InsertTestUser(t, ctx, testDB.Pool, firstUserID)
	tests.InsertTestUser(t, ctx, testDB.Pool, secondUserID)

	result, err := repo.GetUserByID(ctx, &user.GetUserByIDPayload{
		ID: secondUserID.String(),
	})
	if err != nil {
		t.Fatalf("GetUserByID returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user, got nil")
	}

	if result.ID != secondUserID {
		t.Errorf(
			"expected user ID %q, got %q",
			secondUserID,
			result.ID,
		)
	}

	if result.ID == firstUserID {
		t.Error("expected second user, got first user")
	}
}
