package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xanity-07/spndex/internal/database"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/model/analytic.go"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/tests"
)

func TestExpenseAnalyticsRepository_GetAverageExpenseAmount(t *testing.T) {
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

	amounts := []int64{
		1000,
		2000,
		3000,
	}

	for _, amount := range amounts {
		_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
			"id":          uuid.New(),
			"user_id":     userID,
			"amount":      amount,
			"description": "Test expense",
			"category":    enums.ExpenseCategory("food"),
			"currency":    enums.CurrencyCode("USD"),
			"date":        "2026-07-15",
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

	result, err := repo.GetAverageExpenseAmount(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetAverageExpenseAmount returned an error: %v", err)
	}

	if result != 2000 {
		t.Errorf("expected average expense amount 2000, got %d", result)
	}
}

func TestExpenseAnalyticsRepository_GetAverageExpenseAmount_Rounding(t *testing.T) {
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

	amounts := []int64{
		1000,
		2000,
		2500,
	}

	for _, amount := range amounts {
		_, err := testDB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
			"id":          uuid.New(),
			"user_id":     userID,
			"amount":      amount,
			"description": "Test expense",
			"category":    enums.ExpenseCategory("food"),
			"currency":    enums.CurrencyCode("USD"),
			"date":        "2026-07-15",
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

	result, err := repo.GetAverageExpenseAmount(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetAverageExpenseAmount returned an error: %v", err)
	}

	if result != 1833 {
		t.Errorf("expected average expense amount 1833, got %d", result)
	}
}
