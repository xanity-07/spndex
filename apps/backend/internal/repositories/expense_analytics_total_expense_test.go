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

func TestExpenseAnalyticsRepository_GetTotalExpenses(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewAnalyticsRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

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
			date:   "2026-06-05",
		},
		{
			amount: 5000,
			date:   "2026-07-15",
		},
		{
			amount: 2500,
			date:   "2026-08-10",
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
	month := 8
	rng := 3

	query := &analytic.GetDashboardQuery{
		Year:  &year,
		Month: &month,
		Range: &rng,
	}

	result, err := repo.GetTotalExpenses(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetTotalExpenses returned an error: %v", err)
	}

	if result != 10000 {
		t.Errorf("expected total expenses 10000, got %d", result)
	}
}

func TestExpenseAnalyticsRepository_GetTotalExpenses_FiltersResults(t *testing.T) {
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

	insertExpense := func(userID uuid.UUID, amount int64, date string, deletedAt *time.Time) {
		t.Helper()

		_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
			"id":          uuid.New(),
			"user_id":     userID,
			"amount":      amount,
			"description": "Test expense",
			"category":    enums.ExpenseCategory("food"),
			"currency":    enums.CurrencyCode("USD"),
			"date":        date,
			"deleted_at":  deletedAt,
		})
		if err != nil {
			t.Fatalf("failed to insert test expense: %v", err)
		}
	}

	// Valid expenses — should contribute to the total.
	insertExpense(userID, 1000, "2026-06-10", nil)
	insertExpense(userID, 2000, "2026-07-15", nil)
	insertExpense(userID, 3000, "2026-08-10", nil)

	// Outside the range — should NOT contribute.
	insertExpense(userID, 4000, "2026-05-31", nil)

	// Another user's expense — should NOT contribute.
	insertExpense(otherUserID, 5000, "2026-07-10", nil)

	// Soft-deleted expense — should NOT contribute.
	deletedAt := time.Now()
	insertExpense(userID, 6000, "2026-07-20", &deletedAt)

	// Expense belonging to a soft-deleted user — should NOT contribute.
	insertExpense(deletedUserID, 7000, "2026-07-20", nil)

	year := 2026
	month := 8
	rng := 3

	query := &analytic.GetDashboardQuery{
		Year:  &year,
		Month: &month,
		Range: &rng,
	}

	result, err := repo.GetTotalExpenses(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetTotalExpenses returned an error: %v", err)
	}

	if result != 6000 {
		t.Errorf("expected total expenses 6000, got %d", result)
	}
}

func TestExpenseAnalyticsRepository_GetTotalExpenses_NoResults(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewAnalyticsRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

	tests.InsertTestUserEmail(t, ctx, testDB.Pool, userID, "john@example.com")

	year := 2026
	month := 8
	rng := 3

	query := &analytic.GetDashboardQuery{
		Year:  &year,
		Month: &month,
		Range: &rng,
	}

	result, err := repo.GetTotalExpenses(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetTotalExpenses returned an error: %v", err)
	}

	if result != 0 {
		t.Errorf("expected total expenses 0, got %d", result)
	}
}
