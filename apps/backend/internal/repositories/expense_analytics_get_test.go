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

func TestExpenseAnalyticsRepository_GetExpensesByCategory(t *testing.T) {
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
		category enums.ExpenseCategory
		amount   int64
	}{
		{
			amount:   2500,
			category: enums.ExpenseCategory("food"),
		},
		{
			amount:   5000,
			category: enums.ExpenseCategory("food"),
		},
		{
			amount:   7500,
			category: enums.ExpenseCategory("shopping"),
		},
	}

	for _, e := range expenses {
		_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
			"id":          uuid.New(),
			"user_id":     userID,
			"amount":      e.amount,
			"description": "Test expense",
			"category":    e.category,
			"currency":    enums.CurrencyCode("USD"),
			"date":        "2026-08-22",
		})
		if err != nil {
			t.Fatalf("failed to insert test expense: %v", err)
		}
	}

	result, err := repo.GetExpensesByCategory(ctx, userID)
	if err != nil {
		t.Fatalf("GetExpensesByCategory returned an error: %v", err)
	}

	if len(result) != len(enums.AllCategories) {
		t.Fatalf(
			"expected %d categories, got %d",
			len(enums.AllCategories),
			len(result),
		)
	}

	var food *analytic.CategoryTotals
	var shopping *analytic.CategoryTotals

	for i := range result {
		switch result[i].Category {
		case enums.ExpenseCategory("food"):
			food = &result[i]
		case enums.ExpenseCategory("shopping"):
			shopping = &result[i]
		}
	}

	if food == nil {
		t.Fatal("expected food category")
	}

	if food.TotalCents != 7500 {
		t.Errorf("expected food total 7500, got %d", food.TotalCents)
	}

	if food.Count != 2 {
		t.Errorf("expected food count 2, got %d", food.Count)
	}

	if food.Percentage != 50 {
		t.Errorf("expected food percentage 50, got %v", food.Percentage)
	}

	if shopping == nil {
		t.Fatal("expected shopping category")
	}

	if shopping.TotalCents != 7500 {
		t.Errorf("expected shopping total 7500, got %d", shopping.TotalCents)
	}

	if shopping.Count != 1 {
		t.Errorf("expected shopping count 1, got %d", shopping.Count)
	}

	if shopping.Percentage != 50 {
		t.Errorf("expected shopping percentage 50, got %v", shopping.Percentage)
	}

	if result[0].TotalCents < result[len(result)-1].TotalCents {
		t.Error("expected categories to be sorted by total descending")
	}
}

func TestExpenseAnalyticsRepository_GetExpensesByCategory_EmptyCategories(t *testing.T) {
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

	_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     userID,
		"amount":      int64(5000),
		"description": "Food expense",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
	})
	if err != nil {
		t.Fatalf("failed to insert test expense: %v", err)
	}

	result, err := repo.GetExpensesByCategory(ctx, userID)
	if err != nil {
		t.Fatalf("GetExpensesByCategory returned an error: %v", err)
	}

	if len(result) != len(enums.AllCategories) {
		t.Fatalf(
			"expected %d categories, got %d",
			len(enums.AllCategories),
			len(result),
		)
	}

	for _, category := range result {
		if category.Category == enums.ExpenseCategory("food") {
			if category.TotalCents != 5000 {
				t.Errorf(
					"expected food total 5000, got %d",
					category.TotalCents,
				)
			}

			if category.Count != 1 {
				t.Errorf(
					"expected food count 1, got %d",
					category.Count,
				)
			}

			if category.Percentage != 100 {
				t.Errorf(
					"expected food percentage 100, got %v",
					category.Percentage,
				)
			}

			continue
		}

		if category.TotalCents != 0 {
			t.Errorf(
				"expected %q total 0, got %d",
				category.Category,
				category.TotalCents,
			)
		}

		if category.Count != 0 {
			t.Errorf(
				"expected %q count 0, got %d",
				category.Category,
				category.Count,
			)
		}

		if category.Percentage != 0 {
			t.Errorf(
				"expected %q percentage 0, got %v",
				category.Category,
				category.Percentage,
			)
		}
	}
}

func TestExpenseAnalyticsRepository_GetExpensesByCategory_FiltersDeletedAndOtherUsers(t *testing.T) {
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

	tests.InsertTestUserEmail(
		t,
		ctx,
		testDB.Pool,
		deletedUserID,
		"deleted@example.com",
	)

	_, err := testDB.Pool.Exec(ctx, `
		UPDATE users
		SET deleted_at = NOW()
		WHERE id = @user_id
	`, pgx.NamedArgs{
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

	// Valid expense for the requested user.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     userID,
		"amount":      int64(2500),
		"description": "Valid expense",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert valid expense: %v", err)
	}

	// Expense belonging to another user.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     otherUserID,
		"amount":      int64(5000),
		"description": "Other user expense",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert other user expense: %v", err)
	}

	// Soft-deleted expense belonging to the requested user.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     userID,
		"amount":      int64(10000),
		"description": "Deleted expense",
		"category":    enums.ExpenseCategory("shopping"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
		"deleted_at":  time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to insert deleted expense: %v", err)
	}

	// Expense belonging to a soft-deleted user.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     deletedUserID,
		"amount":      int64(20000),
		"description": "Deleted user expense",
		"category":    enums.ExpenseCategory("shopping"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert deleted user expense: %v", err)
	}

	result, err := repo.GetExpensesByCategory(ctx, userID)
	if err != nil {
		t.Fatalf("GetExpensesByCategory returned an error: %v", err)
	}

	if len(result) != len(enums.AllCategories) {
		t.Fatalf(
			"expected %d categories, got %d",
			len(enums.AllCategories),
			len(result),
		)
	}

	var food *analytic.CategoryTotals
	var shopping *analytic.CategoryTotals

	for i := range result {
		switch result[i].Category {
		case enums.ExpenseCategory("food"):
			food = &result[i]
		case enums.ExpenseCategory("shopping"):
			shopping = &result[i]
		}
	}

	if food == nil {
		t.Fatal("expected food category")
	}

	if food.TotalCents != 2500 {
		t.Errorf("expected food total 2500, got %d", food.TotalCents)
	}

	if food.Count != 1 {
		t.Errorf("expected food count 1, got %d", food.Count)
	}

	if food.Percentage != 100 {
		t.Errorf("expected food percentage 100, got %v", food.Percentage)
	}

	if shopping == nil {
		t.Fatal("expected shopping category")
	}

	if shopping.TotalCents != 0 {
		t.Errorf("expected shopping total 0, got %d", shopping.TotalCents)
	}

	if shopping.Count != 0 {
		t.Errorf("expected shopping count 0, got %d", shopping.Count)
	}

	if shopping.Percentage != 0 {
		t.Errorf("expected shopping percentage 0, got %v", shopping.Percentage)
	}
}

func TestExpenseAnalyticsRepository_GetMonthlyExpenses(t *testing.T) {
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
		amount int64
		date   string
	}{
		{
			amount: 2500,
			date:   "2026-03-10",
		},
		{
			amount: 5000,
			date:   "2026-03-20",
		},
		{
			amount: 3000,
			date:   "2026-05-05",
		},
		{
			amount: 7500,
			date:   "2026-08-22",
		},
	}

	for _, e := range expenses {
		_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
			"id":          uuid.New(),
			"user_id":     userID,
			"amount":      e.amount,
			"description": "Test expense",
			"category":    enums.ExpenseCategory("food"),
			"currency":    enums.CurrencyCode("USD"),
			"date":        e.date,
		})
		if err != nil {
			t.Fatalf("failed to insert test expense: %v", err)
		}
	}

	year := 2026

	payload := &analytic.GetMonthlyExpensesPayload{
		Year: &year,
	}

	result, err := repo.GetMonthlyExpenses(ctx, userID, payload)
	if err != nil {
		t.Fatalf("GetMonthlyExpenses returned an error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 monthly results, got %d", len(result))
	}

	if result[0].Month != "03-2026" {
		t.Errorf("expected first month 03-2026, got %q", result[0].Month)
	}

	if result[0].TotalCents != 7500 {
		t.Errorf("expected March total 7500, got %d", result[0].TotalCents)
	}

	if result[0].Count != 2 {
		t.Errorf("expected March count 2, got %d", result[0].Count)
	}

	if result[1].Month != "05-2026" {
		t.Errorf("expected second month 05-2026, got %q", result[1].Month)
	}

	if result[1].TotalCents != 3000 {
		t.Errorf("expected May total 3000, got %d", result[1].TotalCents)
	}

	if result[1].Count != 1 {
		t.Errorf("expected May count 1, got %d", result[1].Count)
	}

	if result[2].Month != "08-2026" {
		t.Errorf("expected third month 08-2026, got %q", result[2].Month)
	}

	if result[2].TotalCents != 7500 {
		t.Errorf("expected August total 7500, got %d", result[2].TotalCents)
	}

	if result[2].Count != 1 {
		t.Errorf("expected August count 1, got %d", result[2].Count)
	}
}

func TestExpenseAnalyticsRepository_GetMonthlyExpenses_YearBoundary(t *testing.T) {
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
		amount int64
		date   string
	}{
		{
			amount: 1000,
			date:   "2025-12-31",
		},
		{
			amount: 2500,
			date:   "2026-01-01",
		},
		{
			amount: 5000,
			date:   "2026-12-31",
		},
		{
			amount: 10000,
			date:   "2027-01-01",
		},
	}

	for _, e := range expenses {
		_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
			"id":          uuid.New(),
			"user_id":     userID,
			"amount":      e.amount,
			"description": "Test expense",
			"category":    enums.ExpenseCategory("food"),
			"currency":    enums.CurrencyCode("USD"),
			"date":        e.date,
		})
		if err != nil {
			t.Fatalf("failed to insert test expense: %v", err)
		}
	}

	year := 2026
	payload := &analytic.GetMonthlyExpensesPayload{
		Year: &year,
	}

	result, err := repo.GetMonthlyExpenses(ctx, userID, payload)
	if err != nil {
		t.Fatalf("GetMonthlyExpenses returned an error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 monthly results, got %d", len(result))
	}

	if result[0].Month != "01-2026" {
		t.Errorf("expected first month 01-2026, got %q", result[0].Month)
	}

	if result[0].TotalCents != 2500 {
		t.Errorf("expected January total 2500, got %d", result[0].TotalCents)
	}

	if result[0].Count != 1 {
		t.Errorf("expected January count 1, got %d", result[0].Count)
	}

	if result[1].Month != "12-2026" {
		t.Errorf("expected second month 12-2026, got %q", result[1].Month)
	}

	if result[1].TotalCents != 5000 {
		t.Errorf("expected December total 5000, got %d", result[1].TotalCents)
	}

	if result[1].Count != 1 {
		t.Errorf("expected December count 1, got %d", result[1].Count)
	}
}

func TestExpenseAnalyticsRepository_GetMonthlyExpenses_FiltersDeletedAndOtherUsers(t *testing.T) {
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

	tests.InsertTestUserEmail(
		t,
		ctx,
		testDB.Pool,
		deletedUserID,
		"deleted@example.com",
	)

	_, err := testDB.Pool.Exec(ctx, `
		UPDATE users
		SET deleted_at = NOW()
		WHERE id = @user_id
	`, pgx.NamedArgs{
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
		"amount":      int64(2500),
		"description": "Valid expense",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-06-15",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert valid expense: %v", err)
	}

	// Other user's expense.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     otherUserID,
		"amount":      int64(5000),
		"description": "Other user expense",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-06-15",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert other user expense: %v", err)
	}

	// Deleted expense belonging to the requested user.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     userID,
		"amount":      int64(10000),
		"description": "Deleted expense",
		"category":    enums.ExpenseCategory("shopping"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-06-15",
		"deleted_at":  time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to insert deleted expense: %v", err)
	}

	// Expense belonging to a deleted user.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     deletedUserID,
		"amount":      int64(20000),
		"description": "Deleted user expense",
		"category":    enums.ExpenseCategory("shopping"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-06-15",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert deleted user expense: %v", err)
	}

	year := 2026
	payload := &analytic.GetMonthlyExpensesPayload{
		Year: &year,
	}

	result, err := repo.GetMonthlyExpenses(ctx, userID, payload)
	if err != nil {
		t.Fatalf("GetMonthlyExpenses returned an error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 monthly result, got %d", len(result))
	}

	if result[0].Month != "06-2026" {
		t.Errorf("expected month 06-2026, got %q", result[0].Month)
	}

	if result[0].TotalCents != 2500 {
		t.Errorf("expected total 2500, got %d", result[0].TotalCents)
	}

	if result[0].Count != 1 {
		t.Errorf("expected count 1, got %d", result[0].Count)
	}
}

func TestExpenseAnalyticsRepository_GetHighestExpense(t *testing.T) {
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
			amount:      2500,
			description: "Small expense",
			date:        "2026-06-05",
		},
		{
			amount:      7500,
			description: "Highest expense",
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

	result, err := repo.GetHighestExpense(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetHighestExpense returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected expense, got nil")
	}

	if result.AmountCents != 7500 {
		t.Errorf("expected amount 7500, got %d", result.AmountCents)
	}

	if result.Description == nil {
		t.Fatal("expected description, got nil")
	}

	if *result.Description != "Highest expense" {
		t.Errorf(
			"expected description %q, got %q",
			"Highest expense",
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
