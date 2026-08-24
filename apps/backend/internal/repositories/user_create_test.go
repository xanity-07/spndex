package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/xanity-07/spndex/internal/database"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/tests"
)

func TestUserRepository_CreateUser(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewUserRepository(srv)
	ctx := context.Background()

	payload := &user.CreateUserPayload{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Password:  "hashed-password",
	}

	result, err := repo.CreateUser(ctx, payload)
	if err != nil {
		t.Fatalf("CreateUser returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected user ID to be populated")
	}

	if result.FirstName != "John" {
		t.Errorf(
			"expected first name %q, got %q",
			"John",
			result.FirstName,
		)
	}

	if result.LastName != "Doe" {
		t.Errorf(
			"expected last name %q, got %q",
			"Doe",
			result.LastName,
		)
	}

	if result.Email != "john@example.com" {
		t.Errorf(
			"expected email %q, got %q",
			"john@example.com",
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

func TestUserRepository_CreateUser_DuplicateEmail(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewUserRepository(srv)
	ctx := context.Background()

	firstPayload := &user.CreateUserPayload{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Password:  "hashed-password",
	}

	_, err := repo.CreateUser(ctx, firstPayload)
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	secondPayload := &user.CreateUserPayload{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "john@example.com",
		Password:  "another-password",
	}

	result, err := repo.CreateUser(ctx, secondPayload)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil user, got %+v", result)
	}
}
