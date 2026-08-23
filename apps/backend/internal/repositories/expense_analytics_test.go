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

func TestExpenseAnalyticsRepository_GetSpendingTrends(t *testing.T) {
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
		id          uuid.UUID
		amount      int64
		description string
		date        string
	}{
		{
			id:          uuid.New(),
			amount:      2500,
			description: "March expense",
			date:        "2026-03-15",
		},
		{
			id:          uuid.New(),
			amount:      5000,
			description: "March expense 2",
			date:        "2026-03-20",
		},
		{
			id:          uuid.New(),
			amount:      3000,
			description: "April expense",
			date:        "2026-04-10",
		},
		{
			id:          uuid.New(),
			amount:      7500,
			description: "May expense",
			date:        "2026-05-05",
		},
	}

	for _, e := range expenses {
		_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
			"id":          e.id,
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

	result, err := repo.GetSpendingTrends(ctx, userID)
	if err != nil {
		t.Fatalf("GetSpendingTrends returned an error: %v", err)
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

	if result[1].Month != "04-2026" {
		t.Errorf("expected second month 04-2026, got %q", result[1].Month)
	}

	if result[1].TotalCents != 3000 {
		t.Errorf("expected April total 3000, got %d", result[1].TotalCents)
	}

	if result[1].Count != 1 {
		t.Errorf("expected April count 1, got %d", result[1].Count)
	}

	if result[2].Month != "05-2026" {
		t.Errorf("expected third month 05-2026, got %q", result[2].Month)
	}

	if result[2].TotalCents != 7500 {
		t.Errorf("expected May total 7500, got %d", result[2].TotalCents)
	}

	if result[2].Count != 1 {
		t.Errorf("expected May count 1, got %d", result[2].Count)
	}
}

func TestExpenseAnalyticsRepository_GetSpendingTrends_DateRange(t *testing.T) {
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
		date   string
		amount int64
	}{
		{
			amount: 1000,
			date:   "2026-02-28",
		},
		{
			amount: 2000,
			date:   "2026-03-01",
		},
		{
			amount: 3000,
			date:   "2026-08-31",
		},
		{
			amount: 4000,
			date:   "2026-09-01",
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

	result, err := repo.GetSpendingTrends(ctx, userID)
	if err != nil {
		t.Fatalf("GetSpendingTrends returned an error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 monthly results, got %d", len(result))
	}

	if result[0].Month != "03-2026" {
		t.Errorf("expected first month 03-2026, got %q", result[0].Month)
	}

	if result[0].TotalCents != 2000 {
		t.Errorf("expected March total 2000, got %d", result[0].TotalCents)
	}

	if result[1].Month != "08-2026" {
		t.Errorf("expected second month 08-2026, got %q", result[1].Month)
	}

	if result[1].TotalCents != 3000 {
		t.Errorf("expected August total 3000, got %d", result[1].TotalCents)
	}
}

func TestExpenseAnalyticsRepository_GetSpendingTrends_FiltersDeletedAndOtherUsers(t *testing.T) {
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

	// Soft-delete one of the users.
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

	// This expense should be included.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     userID,
		"amount":      int64(2500),
		"description": "Valid expense",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-05-15",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert valid expense: %v", err)
	}

	// This expense belongs to another user and should be excluded.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     otherUserID,
		"amount":      int64(5000),
		"description": "Other user expense",
		"category":    enums.ExpenseCategory("shopping"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-05-15",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert other user expense: %v", err)
	}

	// This expense belongs to the requested user but is soft-deleted.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     userID,
		"amount":      int64(10000),
		"description": "Deleted expense",
		"category":    enums.ExpenseCategory("shopping"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-05-15",
		"deleted_at":  time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to insert deleted expense: %v", err)
	}

	// This expense belongs to a soft-deleted user and should be excluded.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     deletedUserID,
		"amount":      int64(20000),
		"description": "Deleted user expense",
		"category":    enums.ExpenseCategory("shopping"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-05-15",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert deleted user expense: %v", err)
	}

	result, err := repo.GetSpendingTrends(ctx, userID)
	if err != nil {
		t.Fatalf("GetSpendingTrends returned an error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 monthly result, got %d", len(result))
	}

	if result[0].Month != "05-2026" {
		t.Errorf("expected month 05-2026, got %q", result[0].Month)
	}

	if result[0].TotalCents != 2500 {
		t.Errorf("expected total 2500, got %d", result[0].TotalCents)
	}

	if result[0].Count != 1 {
		t.Errorf("expected count 1, got %d", result[0].Count)
	}
}

func TestExpenseAnalyticsRepository_GetSpendingTrends_NoExpenses(t *testing.T) {
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

	result, err := repo.GetSpendingTrends(ctx, userID)
	if err != nil {
		t.Fatalf("GetSpendingTrends returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected empty slice, got nil")
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result))
	}
}

func TestExpenseAnalyticsRepository_GetHighestExpense_RangeBoundary(t *testing.T) {
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
			amount:      50000,
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
			amount:      10000,
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

	if *result.Description != "Inside range" {
		t.Errorf(
			"expected description %q, got %q",
			"Inside range",
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

func TestExpenseAnalyticsRepository_GetHighestExpense_NoResults(t *testing.T) {
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

	result, err := repo.GetHighestExpense(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetHighestExpense returned an error: %v", err)
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

func TestExpenseAnalyticsRepository_GetHighestExpense_FiltersDeletedAndOtherUsers(t *testing.T) {
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

	// Larger expense belonging to another user.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     otherUserID,
		"amount":      int64(50000),
		"description": "Other user expense",
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-07-15",
		"deleted_at":  nil,
	})
	if err != nil {
		t.Fatalf("failed to insert other user expense: %v", err)
	}

	// Larger expense belonging to the requested user, but deleted.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     userID,
		"amount":      int64(100000),
		"description": "Deleted expense",
		"category":    enums.ExpenseCategory("shopping"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-07-15",
		"deleted_at":  time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to insert deleted expense: %v", err)
	}

	// Larger expense belonging to a deleted user.
	_, err = testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     deletedUserID,
		"amount":      int64(200000),
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

	result, err := repo.GetHighestExpense(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetHighestExpense returned an error: %v", err)
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
