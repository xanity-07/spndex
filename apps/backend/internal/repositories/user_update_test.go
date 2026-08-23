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

func TestUserRepository_UpdateUser(t *testing.T) {
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

	email := "updated@example.com"
	password := "new-password-hash"
	firstName := "Jane"
	lastName := "Smith"

	result, err := repo.UpdateUser(ctx, userID, &user.UpdateUserPayload{
		Email:     &email,
		Password:  &password,
		FirstName: &firstName,
		LastName:  &lastName,
	})
	if err != nil {
		t.Fatalf("UpdateUser returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user, got nil")
	}

	if result.ID != userID {
		t.Errorf("expected user ID %q, got %q", userID, result.ID)
	}

	if result.Email != email {
		t.Errorf("expected email %q, got %q", email, result.Email)
	}

	if result.FirstName != firstName {
		t.Errorf("expected first name %q, got %q", firstName, result.FirstName)
	}

	if result.LastName != lastName {
		t.Errorf("expected last name %q, got %q", lastName, result.LastName)
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

func TestUserRepository_UpdateUser_PartialUpdate(t *testing.T) {
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

	firstName := "Jane"

	result, err := repo.UpdateUser(ctx, userID, &user.UpdateUserPayload{
		FirstName: &firstName,
	})
	if err != nil {
		t.Fatalf("UpdateUser returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user, got nil")
	}

	if result.FirstName != "Jane" {
		t.Errorf("expected first name %q, got %q", "Jane", result.FirstName)
	}

	if result.LastName != "Doe" {
		t.Errorf("expected last name %q, got %q", "Doe", result.LastName)
	}

	if result.Email != "john-"+userID.String()+"@example.com" {
		t.Errorf(
			"expected original email %q, got %q",
			"john-"+userID.String()+"@example.com",
			result.Email,
		)
	}
}

func TestUserRepository_UpdateUser_NoFields(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewUserRepository(srv)
	ctx := context.Background()

	result, err := repo.UpdateUser(
		ctx,
		uuid.New(),
		&user.UpdateUserPayload{},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil user, got %+v", result)
	}
}

func TestUserRepository_UpdateUser_NotFound(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewUserRepository(srv)
	ctx := context.Background()

	firstName := "Jane"

	result, err := repo.UpdateUser(
		ctx,
		uuid.New(),
		&user.UpdateUserPayload{
			FirstName: &firstName,
		},
	)

	if err == nil {
		t.Fatal("expected USER_NOT_FOUND error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil user, got %+v", result)
	}
}

func TestUserRepository_UpdateUser_DeletedUser(t *testing.T) {
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

	firstName := "Jane"

	result, err := repo.UpdateUser(
		ctx,
		userID,
		&user.UpdateUserPayload{
			FirstName: &firstName,
		},
	)

	if err == nil {
		t.Fatal("expected USER_NOT_FOUND error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil user, got %+v", result)
	}
}
