package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xanity-07/spndex/internal/database"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/model/expense"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/tests"
)

func TestExpenseRepository_UpdateExpense(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()
	expenseID := uuid.New()

	tests.InsertTestUserEmail(t, ctx, testDB.Pool, userID, "john@example.com")

	description := "Lunch"

	stmt := `
		INSERT INTO expenses (
			id,
			user_id,
			amount,
			description,
			category,
			currency,
			date
		)
		VALUES (
			@expense_id,
			@user_id,
			@amount,
			@description,
			@category,
			@currency,
			@date
		)
	`

	_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"expense_id":  expenseID,
		"user_id":     userID,
		"amount":      int64(2500),
		"description": description,
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
	})
	if err != nil {
		t.Fatalf("failed to insert test expense: %v", err)
	}

	newAmount := int64(5000)
	newCategory := enums.ExpenseCategory("shopping")
	newCurrency := enums.CurrencyCode("EUR")
	newDate := "2026-08-23"
	newDescription := "New expense"

	payload := &expense.UpdateExpense{
		AmountCents:  &newAmount,
		Category:     &newCategory,
		CurrencyCode: &newCurrency,
		Date:         &newDate,
		Description:  &newDescription,
	}

	result, err := repo.UpdateExpense(ctx, userID, expenseID, payload)
	if err != nil {
		t.Fatalf("UpdateExpense returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected expense, got nil")
	}

	if result.ID != expenseID {
		t.Errorf("expected expense ID %q, got %q", expenseID, result.ID)
	}

	if result.UserID != userID.String() {
		t.Errorf("expected user ID %q, got %q", userID, result.UserID)
	}

	if result.AmountCents != 5000 {
		t.Errorf("expected amount 5000, got %v", result.AmountCents)
	}

	if result.Category != enums.ExpenseCategory("shopping") {
		t.Errorf("expected category shopping, got %q", result.Category)
	}

	if result.CurrencyCode != enums.CurrencyCode("EUR") {
		t.Errorf("expected currency EUR, got %q", result.CurrencyCode)
	}

	if result.Date != "2026-08-23" {
		t.Errorf("expected date 2026-08-23, got %q", result.Date)
	}

	if result.Description == nil {
		t.Fatal("expected description, got nil")
	}

	if *result.Description != "New expense" {
		t.Errorf("expected description %q, got %q", "New expense", *result.Description)
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

func TestExpenseRepository_UpdateExpense_PartialUpdate(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()
	expenseID := uuid.New()

	tests.InsertTestUserEmail(t, ctx, testDB.Pool, userID, "john@example.com")

	stmt := `
		INSERT INTO expenses (
			id,
			user_id,
			amount,
			description,
			category,
			currency,
			date
		)
		VALUES (
			@expense_id,
			@user_id,
			@amount,
			@description,
			@category,
			@currency,
			@date
		)
	`

	_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"expense_id":  expenseID,
		"user_id":     userID,
		"amount":      int64(2500),
		"description": "Lunch",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
	})
	if err != nil {
		t.Fatalf("failed to insert test expense: %v", err)
	}

	newAmount := int64(5000)

	payload := &expense.UpdateExpense{
		AmountCents: &newAmount,
	}

	result, err := repo.UpdateExpense(ctx, userID, expenseID, payload)
	if err != nil {
		t.Fatalf("UpdateExpense returned an error: %v", err)
	}

	if result.AmountCents != 5000 {
		t.Errorf("expected amount 5000, got %v", result.AmountCents)
	}

	if result.Category != enums.ExpenseCategory("food") {
		t.Errorf("expected category food, got %q", result.Category)
	}

	if result.CurrencyCode != enums.CurrencyCode("USD") {
		t.Errorf("expected currency USD, got %q", result.CurrencyCode)
	}

	if result.Date != "2026-08-22" {
		t.Errorf("expected date 2026-08-22, got %q", result.Date)
	}

	if result.Description == nil {
		t.Fatal("expected description, got nil")
	}

	if *result.Description != "Lunch" {
		t.Errorf("expected description %q, got %q", "Lunch", *result.Description)
	}
}

func TestExpenseRepository_UpdateExpense_NoFields(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	result, err := repo.UpdateExpense(
		ctx,
		uuid.New(),
		uuid.New(),
		&expense.UpdateExpense{},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil expense, got %+v", result)
	}
}

func TestExpenseRepository_UpdateExpense_NotFound(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()
	tests.InsertTestUserEmail(t, ctx, testDB.Pool, userID, "john@example.com")

	amount := int64(5000)

	result, err := repo.UpdateExpense(
		ctx,
		userID,
		uuid.New(),
		&expense.UpdateExpense{
			AmountCents: &amount,
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil expense, got %+v", result)
	}
}

func TestExpenseRepository_UpdateExpense_WrongUser(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()
	otherUserID := uuid.New()
	expenseID := uuid.New()

	tests.InsertTestUserEmail(t, ctx, testDB.Pool, userID, "john@example.com")
	tests.InsertTestUserEmail(t, ctx, testDB.Pool, otherUserID, "jane@example.com")

	stmt := `
		INSERT INTO expenses (
			id,
			user_id,
			amount,
			description,
			category,
			currency,
			date
		)
		VALUES (
			@expense_id,
			@user_id,
			@amount,
			@description,
			@category,
			@currency,
			@date
		)
	`

	_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"expense_id":  expenseID,
		"user_id":     userID,
		"amount":      int64(2500),
		"description": "Lunch",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
	})
	if err != nil {
		t.Fatalf("failed to insert test expense: %v", err)
	}

	amount := int64(5000)

	result, err := repo.UpdateExpense(
		ctx,
		otherUserID,
		expenseID,
		&expense.UpdateExpense{
			AmountCents: &amount,
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil expense, got %+v", result)
	}
}

func TestExpenseRepository_UpdateExpense_DeletedExpense(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()
	expenseID := uuid.New()

	tests.InsertTestUserEmail(t, ctx, testDB.Pool, userID, "john@example.com")

	stmt := `
		INSERT INTO expenses (
			id,
			user_id,
			amount,
			description,
			category,
			currency,
			date,
			deleted_at
		)
		VALUES (
			@expense_id,
			@user_id,
			@amount,
			@description,
			@category,
			@currency,
			@date,
			NOW()
		)
	`

	_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"expense_id":  expenseID,
		"user_id":     userID,
		"amount":      int64(2500),
		"description": "Lunch",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
	})
	if err != nil {
		t.Fatalf("failed to insert deleted expense: %v", err)
	}

	amount := int64(5000)

	result, err := repo.UpdateExpense(
		ctx,
		userID,
		expenseID,
		&expense.UpdateExpense{
			AmountCents: &amount,
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil expense, got %+v", result)
	}
}

func TestExpenseRepository_UpdateExpense_DeletedUser(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()
	expenseID := uuid.New()

	tests.InsertTestUserEmail(t, ctx, testDB.Pool, userID, "john@example.com")

	stmt := `
		UPDATE users
		SET deleted_at = NOW()
		WHERE id = @user_id
	`

	_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
	})
	if err != nil {
		t.Fatalf("failed to soft delete test user: %v", err)
	}

	expenseStmt := `
		INSERT INTO expenses (
			id,
			user_id,
			amount,
			description,
			category,
			currency,
			date
		)
		VALUES (
			@expense_id,
			@user_id,
			@amount,
			@description,
			@category,
			@currency,
			@date
		)
	`

	_, err = testDB.Pool.Exec(ctx, expenseStmt, pgx.NamedArgs{
		"expense_id":  expenseID,
		"user_id":     userID,
		"amount":      int64(2500),
		"description": "Lunch",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
	})
	if err != nil {
		t.Fatalf("failed to insert test expense: %v", err)
	}

	amount := int64(5000)

	result, err := repo.UpdateExpense(
		ctx,
		userID,
		expenseID,
		&expense.UpdateExpense{
			AmountCents: &amount,
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil expense, got %+v", result)
	}
}
