package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xanity-07/spndex/internal/database"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/model/analytic.go"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/tests"
)

func TestExpenseAnalyticsRepository_GetLowestExpense_NoResults(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewAnalyticsRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

	tests.InsertTestUserEmail(
		t,
		ctx,
		testDB.Pool,
		userID,
		"john@example.com",
	)

	year := 2026
	month := 8
	rng := 3

	query := &analytic.GetDashboardQuery{
		Year:  &year,
		Month: &month,
		Range: &rng,
	}

	result, err := repo.GetLowestExpense(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetLowestExpense returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected empty expense, got nil")
	}

	if result.ID != uuid.Nil {
		t.Errorf("expected empty expense ID, got %q", result.ID)
	}

	if result.AmountCents != 0 {
		t.Errorf("expected amount 0, got %d", result.AmountCents)
	}
}

func TestExpenseAnalyticsRepository_GetLowestExpense(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewAnalyticsRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

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
			@id,
			@user_id,
			@amount,
			@description,
			@category,
			@currency,
			@date
		)
	`

	expenses := []struct {
		description string
		date        string
		amount      int64
	}{
		{
			amount:      7500,
			description: "Large expense",
			date:        "2026-06-05",
		},
		{
			amount:      2500,
			description: "Lowest expense",
			date:        "2026-07-15",
		},
		{
			amount:      5000,
			description: "Medium expense",
			date:        "2026-08-10",
		},
	}

	for _, e := range expenses {
		_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
			"id":          uuid.New(),
			"user_id":     userID,
			"amount":      e.amount,
			"description": e.description,
			"category":    enums.ExpenseCategory("food"),
			"currency":    enums.CurrencyCode("USD"),
			"date":        e.date,
		})
		if err != nil {
			t.Fatalf("failed to insert test expense: %v", err)
		}
	}

	year := 2026
	month := 8
	rng := 3

	query := &analytic.GetDashboardQuery{
		Year:  &year,
		Month: &month,
		Range: &rng,
	}

	result, err := repo.GetLowestExpense(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetLowestExpense returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected expense, got nil")
	}

	if result.AmountCents != 2500 {
		t.Errorf("expected amount 2500, got %d", result.AmountCents)
	}

	if result.Description == nil {
		t.Fatal("expected description, got nil")
	}

	if *result.Description != "Lowest expense" {
		t.Errorf(
			"expected description %q, got %q",
			"Lowest expense",
			*result.Description,
		)
	}

	if result.Date != "2026-07-15" {
		t.Errorf(
			"expected date 2026-07-15, got %q",
			result.Date,
		)
	}
}

func TestExpenseAnalyticsRepository_GetLowestExpense_RangeBoundary(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewAnalyticsRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

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
			@id,
			@user_id,
			@amount,
			@description,
			@category,
			@currency,
			@date
		)
	`

	expenses := []struct {
		amount      int64
		description string
		date        string
	}{
		{
			amount:      100,
			description: "Before range",
			date:        "2026-05-31",
		},
		{
			amount:      2500,
			description: "Start boundary",
			date:        "2026-06-01",
		},
		{
			amount:      7500,
			description: "Inside range",
			date:        "2026-07-15",
		},
		{
			amount:      50,
			description: "End boundary",
			date:        "2026-09-01",
		},
	}

	for _, e := range expenses {
		_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
			"id":          uuid.New(),
			"user_id":     userID,
			"amount":      e.amount,
			"description": e.description,
			"category":    enums.ExpenseCategory("food"),
			"currency":    enums.CurrencyCode("USD"),
			"date":        e.date,
		})
		if err != nil {
			t.Fatalf("failed to insert test expense: %v", err)
		}
	}

	year := 2026
	month := 8
	rng := 3

	query := &analytic.GetDashboardQuery{
		Year:  &year,
		Month: &month,
		Range: &rng,
	}

	result, err := repo.GetLowestExpense(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetLowestExpense returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected expense, got nil")
	}

	if result.AmountCents != 2500 {
		t.Errorf("expected amount 2500, got %d", result.AmountCents)
	}

	if result.Description == nil {
		t.Fatal("expected description, got nil")
	}

	if *result.Description != "Start boundary" {
		t.Errorf(
			"expected description %q, got %q",
			"Start boundary",
			*result.Description,
		)
	}

	if result.Date != "2026-06-01" {
		t.Errorf(
			"expected date 2026-06-01, got %q",
			result.Date,
		)
	}
}

func TestExpenseAnalyticsRepository_GetLowestExpense_FiltersDeletedAndOtherUsers(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewAnalyticsRepository(srv)
	ctx := context.Background()

	userID := uuid.New()
	otherUserID := uuid.New()
	deletedUserID := uuid.New()

	tests.InsertTestUserEmail(t, ctx, testDB.Pool, userID, "john@example.com")

	tests.InsertTestUserEmail(t, ctx, testDB.Pool, otherUserID, "jane@example.com")

	tests.InsertTestUserEmail(t, ctx, testDB.Pool, deletedUserID, "deleted@example.com")

	_, err := testDB.Pool.Exec(ctx, `UPDATE users SET deleted_at = NOW() WHERE id = @user_id`,
		pgx.NamedArgs{
			"user_id": deletedUserID,
		})
	if err != nil {
		t.Fatalf("failed to soft delete test user: %v", err)
	}

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
			@id,
			@user_id,
			@amount,
			@description,
			@category,
			@currency,
			@date,
			@deleted_at
		)
	`

	// Valid expense.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     userID,
		"amount":      int64(5000),
		"description": "Valid expense",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-07-15",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert valid expense: %v", err)
	}

	// Smaller expense belonging to another user.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     otherUserID,
		"amount":      int64(100),
		"description": "Other user expense",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-07-15",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert other user expense: %v", err)
	}

	// Smaller soft-deleted expense.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     userID,
		"amount":      int64(200),
		"description": "Deleted expense",
		"category":    enums.ExpenseCategory("shopping"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-07-15",
		"deleted_at":  time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to insert deleted expense: %v", err)
	}

	// Smaller expense belonging to a soft-deleted user.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     deletedUserID,
		"amount":      int64(300),
		"description": "Deleted user expense",
		"category":    enums.ExpenseCategory("shopping"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-07-15",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert deleted user expense: %v", err)
	}

	year := 2026
	month := 8
	rng := 3

	query := &analytic.GetDashboardQuery{
		Year:  &year,
		Month: &month,
		Range: &rng,
	}

	result, err := repo.GetLowestExpense(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetLowestExpense returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected expense, got nil")
	}

	if result.AmountCents != 5000 {
		t.Errorf("expected amount 5000, got %d", result.AmountCents)
	}

	if result.Description == nil {
		t.Fatal("expected description, got nil")
	}

	if *result.Description != "Valid expense" {
		t.Errorf(
			"expected description %q, got %q",
			"Valid expense",
			*result.Description,
		)
	}
}
