package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xanity-07/spndex/internal/database"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/model/expense"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/tests"
)

func TestExpenseRepository_CreateExpense(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)
	// tests.SetupSpndexSchema(t, testDB.Pool)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)

	ctx := context.Background()

	userID := uuid.New()

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
		"email":         "john@example.com",
		"password_hash": "hashed-password",
	}

	_, err := testDB.Pool.Exec(ctx, stmt, args)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	description := "Lunch"

	payload := &expense.CreateExpensePayload{
		Date:         "2026-08-22",
		Description:  &description,
		Category:     enums.ExpenseCategory("food"),
		CurrencyCode: enums.CurrencyCode("USD"),
		AmountCents:  2500,
	}

	result, err := repo.CreateExpense(ctx, userID, payload)
	if err != nil {
		t.Fatalf("CreateExpense returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected expense, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected expense ID to be populated")
	}

	if result.UserID != userID.String() {
		t.Errorf("expected user ID %q, got %q", userID, result.UserID)
	}

	if result.AmountCents != 2500 {
		t.Errorf("expected amount 2500, got %v", result.AmountCents)
	}

	if result.Date != "2026-08-22" {
		t.Errorf("expected date 2026-08-22, got %q", result.Date)
	}

	if result.Description == nil {
		t.Fatal("expected description, got nil")
	}

	if *result.Description != "Lunch" {
		t.Errorf(
			"expected description %q, got %q",
			"Lunch",
			*result.Description,
		)
	}

	if result.Category != enums.ExpenseCategory("food") {
		t.Errorf("expected category food, got %q", result.Category)
	}

	if result.CurrencyCode != enums.CurrencyCode("USD") {
		t.Errorf("expected currency USD, got %q", result.CurrencyCode)
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

func TestExpenseRepository_CreateExpense_SoftDeletedUser(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)
	// tests.SetupSpndexSchema(t, testDB.Pool)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

	stmt := `
		INSERT INTO users (
			id,
			first_name,
			last_name,
			email,
			password_hash,
			deleted_at
		)
		VALUES (
			@user_id,
			@first_name,
			@last_name,
			@email,
			@password_hash,
			@deleted_at
		)
	`

	args := pgx.NamedArgs{
		"user_id":       userID,
		"first_name":    "John",
		"last_name":     "Doe",
		"email":         "deleted@example.com",
		"password_hash": "hashed-password",
		"deleted_at":    time.Now(),
	}

	_, err := testDB.Pool.Exec(ctx, stmt, args)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	description := "Lunch"

	payload := &expense.CreateExpensePayload{
		Date:         "2026-08-22",
		Description:  &description,
		Category:     enums.ExpenseCategory("food"),
		CurrencyCode: enums.CurrencyCode("USD"),
		AmountCents:  2500,
	}

	result, err := repo.CreateExpense(ctx, userID, payload)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil expense, got %+v", result)
	}

	if err.Error() != "user is deleted or does not exist" {
		t.Errorf("expected user deleted error, got %q", err.Error())
	}
}

func TestExpenseRepository_CreateExpense_NilDescription(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)
	// tests.SetupSpndexSchema(t, testDB.Pool)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

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
		"email":         "nodescription@example.com",
		"password_hash": "hashed-password",
	}

	_, err := testDB.Pool.Exec(ctx, stmt, args)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	payload := &expense.CreateExpensePayload{
		Date:         "2026-08-22",
		Description:  nil,
		Category:     enums.ExpenseCategory("food"),
		CurrencyCode: enums.CurrencyCode("USD"),
		AmountCents:  2500,
	}

	result, err := repo.CreateExpense(ctx, userID, payload)

	if err != nil {
		t.Fatalf("CreateExpense returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected expense, got nil")
	}

	if result.Description != nil {
		t.Errorf(
			"expected nil description, got %q",
			*result.Description,
		)
	}
}
