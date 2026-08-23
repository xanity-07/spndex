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

func TestExpenseRepository_DeleteExpense(t *testing.T) {
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

	tests.InsertTestUserEmail(
		t,
		ctx,
		testDB.Pool,
		userID,
		"john@example.com",
	)

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

	args := pgx.NamedArgs{
		"expense_id":  expenseID,
		"user_id":     userID,
		"amount":      int64(2500),
		"description": "Lunch",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
	}

	_, err := testDB.Pool.Exec(ctx, stmt, args)
	if err != nil {
		t.Fatalf("failed to insert test expense: %v", err)
	}

	payload := &expense.DeleteExpense{
		ID: expenseID.String(),
	}

	err = repo.DeleteExpense(ctx, userID, payload)
	if err != nil {
		t.Fatalf("DeleteExpense returned an error: %v", err)
	}

	var deletedAt *time.Time

	verifyStmt := `
		SELECT deleted_at
		FROM expenses
		WHERE id = @id
	`

	err = testDB.Pool.QueryRow(
		ctx,
		verifyStmt,
		pgx.NamedArgs{
			"id": expenseID,
		},
	).Scan(&deletedAt)

	if err != nil {
		t.Fatalf("failed to verify deleted expense: %v", err)
	}

	if deletedAt == nil {
		t.Fatal("expected deleted_at to be populated")
	}
}

func TestExpenseRepository_DeleteExpense_NotFound(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

	tests.InsertTestUserEmail(
		t,
		ctx,
		testDB.Pool,
		userID,
		"john@example.com",
	)

	payload := &expense.DeleteExpense{
		ID: uuid.NewString(),
	}

	err := repo.DeleteExpense(ctx, userID, payload)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExpenseRepository_DeleteExpense_WrongUser(t *testing.T) {
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

	tests.InsertTestUserEmail(
		t,
		ctx,
		testDB.Pool,
		userID,
		"john@example.com",
	)

	tests.InsertTestUserEmail(
		t,
		ctx,
		testDB.Pool,
		otherUserID,
		"jane@example.com",
	)

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

	args := pgx.NamedArgs{
		"expense_id":  expenseID,
		"user_id":     userID,
		"amount":      int64(2500),
		"description": "Lunch",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
	}

	_, err := testDB.Pool.Exec(ctx, stmt, args)
	if err != nil {
		t.Fatalf("failed to insert test expense: %v", err)
	}

	err = repo.DeleteExpense(
		ctx,
		otherUserID,
		&expense.DeleteExpense{
			ID: expenseID.String(),
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var deletedAt *time.Time

	verifyStmt := `
		SELECT deleted_at
		FROM expenses
		WHERE id = @id
	`

	err = testDB.Pool.QueryRow(
		ctx,
		verifyStmt,
		pgx.NamedArgs{
			"id": expenseID,
		},
	).Scan(&deletedAt)

	if err != nil {
		t.Fatalf("failed to verify expense: %v", err)
	}

	if deletedAt != nil {
		t.Fatal("expected expense to remain undeleted")
	}
}

func TestExpenseRepository_DeleteExpense_AlreadyDeleted(t *testing.T) {
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

	tests.InsertTestUserEmail(
		t,
		ctx,
		testDB.Pool,
		userID,
		"john@example.com",
	)

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

	args := pgx.NamedArgs{
		"expense_id":  expenseID,
		"user_id":     userID,
		"amount":      int64(2500),
		"description": "Lunch",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
	}

	_, err := testDB.Pool.Exec(ctx, stmt, args)
	if err != nil {
		t.Fatalf("failed to insert deleted expense: %v", err)
	}

	err = repo.DeleteExpense(
		ctx,
		userID,
		&expense.DeleteExpense{
			ID: expenseID.String(),
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExpenseRepository_DeleteExpense_DeletedUser(t *testing.T) {
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

	tests.InsertTestUserEmail(
		t,
		ctx,
		testDB.Pool,
		userID,
		"john@example.com",
	)

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

	expenseArgs := pgx.NamedArgs{
		"expense_id":  expenseID,
		"user_id":     userID,
		"amount":      int64(2500),
		"description": "Lunch",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
	}

	_, err := testDB.Pool.Exec(ctx, expenseStmt, expenseArgs)
	if err != nil {
		t.Fatalf("failed to insert test expense: %v", err)
	}

	deleteUserStmt := `
		UPDATE users
		SET deleted_at = NOW()
		WHERE id = @user_id
	`

	_, err = testDB.Pool.Exec(
		ctx,
		deleteUserStmt,
		pgx.NamedArgs{
			"user_id": userID,
		},
	)

	if err != nil {
		t.Fatalf("failed to soft delete test user: %v", err)
	}

	err = repo.DeleteExpense(
		ctx,
		userID,
		&expense.DeleteExpense{
			ID: expenseID.String(),
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var deletedAt *time.Time

	verifyStmt := `
		SELECT deleted_at
		FROM expenses
		WHERE id = @id
	`

	err = testDB.Pool.QueryRow(
		ctx,
		verifyStmt,
		pgx.NamedArgs{
			"id": expenseID,
		},
	).Scan(&deletedAt)

	if err != nil {
		t.Fatalf("failed to verify expense: %v", err)
	}

	if deletedAt != nil {
		t.Fatal("expected expense to remain undeleted")
	}
}
