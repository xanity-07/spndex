package repositories

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/model/analytic.go"
	"github.com/xanity-07/spndex/internal/model/expense"
	"github.com/xanity-07/spndex/internal/server"
)

type ExpenseAnalyticsRepository struct {
	server *server.Server
}

func NewAnalyticsRepository(s *server.Server) *ExpenseAnalyticsRepository {
	return &ExpenseAnalyticsRepository{
		server: s,
	}
}

func (r *ExpenseAnalyticsRepository) GetExpensesByCategory(ctx context.Context, userID uuid.UUID) ([]analytic.CategoryTotals, error) {
	stats := make(map[enums.ExpenseCategory]analytic.CategoryStats, len(enums.AllCategories))

	for _, c := range enums.AllCategories {
		stats[c] = analytic.CategoryStats{}
	}

	stmt := `
		SELECT
			e.category,
			SUM(e.amount) AS total,
			COUNT(*) AS count
		FROM
			expenses e
		INNER JOIN users u ON e.user_id = u.id
		WHERE
			e.user_id = @user_id
			AND u.deleted_at IS NULL
			AND e.deleted_at IS NULL
		GROUP BY
			e.category
	`
	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	rowResults, err := pgx.CollectRows(rows, pgx.RowToStructByName[analytic.CategoryTotals])
	if err != nil {
		return nil, fmt.Errorf("failed to collect rows from table expenses: %w", err)
	}

	for _, r := range rowResults {
		stats[r.Category] = analytic.CategoryStats{
			Count:      r.Count,
			TotalCents: r.TotalCents,
		}
	}

	var grandTotal int64
	for _, s := range stats {
		grandTotal += s.TotalCents
	}

	result := make([]analytic.CategoryTotals, 0, len(enums.AllCategories))
	for _, c := range enums.AllCategories {
		s := stats[c]
		var pct float64
		if grandTotal > 0 {
			pct = math.Round(float64(s.TotalCents)/float64(grandTotal)*100*100) / 100
		}

		result = append(result, analytic.CategoryTotals{
			Category:      c,
			CategoryStats: s,
			Percentage:    pct,
		})

	}
	slices.SortFunc(result, func(a, b analytic.CategoryTotals) int {
		return cmp.Compare(b.TotalCents, a.TotalCents)
	})
	return result, nil
}

func (r *ExpenseAnalyticsRepository) GetMonthlyExpenses(ctx context.Context, userID uuid.UUID, payload *analytic.GetMonthlyExpensesPayload) ([]analytic.MonthlyTotals, error) {
	stmt := `
		SELECT
			TO_CHAR(DATE_TRUNC('month', TO_DATE(e.date, 'YYYY-MM-DD')),	'MM-YYYY') AS month,
			SUM(e.amount) AS total,
			COUNT(*) AS count
		FROM
			expenses e
		INNER JOIN users u ON u.id = e.user_id
		WHERE
			e.user_id = @user_id
			AND TO_DATE(e.date, 'YYYY-MM-DD') >= MAKE_DATE(@year, 1, 1)
			AND TO_DATE(e.date, 'YYYY-MM-DD') < MAKE_DATE(@year + 1, 1, 1)
			AND u.deleted_at IS NULL
			AND e.deleted_at IS NULL
		GROUP BY
			DATE_TRUNC('month', TO_DATE(e.date, 'YYYY-MM-DD'))
		ORDER BY
			DATE_TRUNC('month', TO_DATE(e.date, 'YYYY-MM-DD'))
	`
	args := pgx.NamedArgs{
		"user_id": userID,
		"year":    payload.Year,
	}

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[analytic.MonthlyTotals])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table expenses: %w", err)
	}

	return results, nil
}

func (r *ExpenseAnalyticsRepository) GetHighestExpense(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (*expense.Expense, error) {
	stmt := `
		SELECT
			e.id,
			e.user_id,
    		e.description,
    		e.category,
    		e.date,
    		e.amount,
    		e.currency,
			e.created_at,
			e.updated_at,
			e.deleted_at
		FROM
			expenses e
		INNER JOIN users u ON u.id = e.user_id
		WHERE
			e.user_id = @user_id
			AND e.date >= @start_date
			AND e.date < @end_date
			AND e.deleted_at IS NULL
			AND u.deleted_at IS NULL
		ORDER BY
			e.amount DESC
		LIMIT 1
	`
	startDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, -(*query.Range - 1), 0).Format("2006-01-02")
	endDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, 1, 0).Format("2006-01-02")

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id":    userID,
		"start_date": startDate,
		"end_date":   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	exp, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[expense.Expense])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &expense.Expense{}, nil
		}
		return nil, fmt.Errorf("failed to collect row from table expenses: %w", err)
	}

	return &exp, nil
}

func (r *ExpenseAnalyticsRepository) GetLowestExpense(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (*expense.Expense, error) {
	stmt := `
		SELECT
			e.id,
			e.user_id,
    		e.description,
    		e.category,
    		e.date,
    		e.amount,
    		e.currency,
			e.created_at,
			e.updated_at,
			e.deleted_at
		FROM
			expenses e
		INNER JOIN users u ON u.id = e.user_id
		WHERE
			e.user_id = @user_id
			AND e.date >= @start_date
			AND e.date < @end_date
			AND e.deleted_at IS NULL
			AND u.deleted_at IS NULL
		ORDER BY
			e.amount ASC
		LIMIT 1
	`
	startDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, -(*query.Range - 1), 0).Format("2006-01-02")
	endDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, 1, 0).Format("2006-01-02")

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id":    userID,
		"start_date": startDate,
		"end_date":   endDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	exp, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[expense.Expense])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &expense.Expense{}, nil
		}
		return nil, fmt.Errorf("failed to collect row from table expenses: %w", err)
	}

	return &exp, nil
}

func (r *ExpenseAnalyticsRepository) GetExpenseCount(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int, error) {
	stmt := `
		SELECT
			COUNT(*)
		FROM
			expenses e
		INNER JOIN users u ON u.id = e.user_id
		WHERE
			e.user_id = @user_id
			AND e.date >= @start_date
			AND e.date < @end_date
			AND e.deleted_at IS NULL
			AND u.deleted_at IS NULL
	`
	startDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, -(*query.Range - 1), 0).Format("2006-01-02")
	endDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, 1, 0).Format("2006-01-02")

	var count int
	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"user_id":    userID,
		"start_date": startDate,
		"end_date":   endDate,
	}).Scan(&count)
	if err != nil {
		return -1, fmt.Errorf("failed to execute query: %w", err)
	}

	return count, nil
}

func (r *ExpenseAnalyticsRepository) GetTotalExpenses(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int64, error) {
	stmt := `
		SELECT
			COALESCE(SUM(amount), 0) as total
		FROM
			expenses e
		INNER JOIN users u ON u.id = e.user_id
		WHERE
			e.user_id = @user_id
			AND e.date >= @start_date
			AND e.date < @end_date
			AND e.deleted_at IS NULL
			AND u.deleted_at IS NULL
	`
	startDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, -(*query.Range - 1), 0).Format("2006-01-02")
	endDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, 1, 0).Format("2006-01-02")

	var total int
	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"user_id":    userID,
		"start_date": startDate,
		"end_date":   endDate,
	}).Scan(&total)

	if err != nil {
		return -1, fmt.Errorf("failed to execute query: %w", err)
	}

	return int64(total), nil
}

func (r *ExpenseAnalyticsRepository) GetAverageExpenseAmount(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int64, error) {
	stmt := `
		SELECT
			COALESCE(AVG(e.amount), 0)::bigint as average
		FROM
			expenses e
		INNER JOIN users u ON u.id = e.user_id
		WHERE
			e.user_id = @user_id
			AND e.date >= @start_date
			AND e.date < @end_date
			AND e.deleted_at IS NULL
			AND u.deleted_at IS NULL
	`
	startDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, -(*query.Range - 1), 0).Format("2006-01-02")
	endDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, 1, 0).Format("2006-01-02")

	var average int64
	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"user_id":    userID,
		"start_date": startDate,
		"end_date":   endDate,
	}).Scan(&average)

	if err != nil {
		return -1, fmt.Errorf("failed to execute query: %w", err)
	}

	return average, nil
}

func (r *ExpenseAnalyticsRepository) MonthlyTotals(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int64, error) {
	stmt := `
		SELECT
			COALESCE(SUM(e.amount), 0) as monthly_total
		FROM
			expenses e
		JOIN users u ON u.id = e.user_id
		WHERE
			e.user_id = @user_id
			AND e.date >= @start_date
			AND e.date < @end_date
			AND e.deleted_at IS NULL
			AND u.deleted_at IS NULL
	`
	startDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, -(*query.Range - 1), 0).Format("2006-01-02")
	endDate := time.Date(*query.Year, time.Month(*query.Month), 1, 0, 0, 0, 0, time.Now().Location()).AddDate(0, 1, 0).Format("2006-01-02")

	var monthlyTotal int64
	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{
		"user_id":    userID,
		"start_date": startDate,
		"end_date":   endDate,
	}).Scan((&monthlyTotal))

	if err != nil {
		return -1, fmt.Errorf("failed to execute query: %w", err)
	}

	return monthlyTotal, nil
}

func (r *ExpenseAnalyticsRepository) GetSpendingTrends(ctx context.Context, userID uuid.UUID) ([]analytic.MonthlyTotals, error) {
	stmt := `
		SELECT
			TO_CHAR(DATE_TRUNC('month', TO_DATE(e.date, 'YYYY-MM-DD')),'MM-YYYY') AS month,
			SUM(e.amount) AS total,
			COUNT(*) AS count
		FROM
			expenses e
		INNER JOIN users u ON u.id = e.user_id
		WHERE
			e.user_id = @user_id
			AND e.date >= @start_date
			AND e.date < @end_date
			AND e.deleted_at IS NULL
			AND u.deleted_at IS NULL
		GROUP BY
			DATE_TRUNC('month', TO_DATE(e.date, 'YYYY-MM-DD'))
		ORDER BY
			DATE_TRUNC('month', TO_DATE(e.date, 'YYYY-MM-DD'))
	`

	now := time.Now()

	startDate := time.Date(now.Year(), now.Month()-5, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	endDate := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")

	args := pgx.NamedArgs{
		"user_id":    userID,
		"start_date": startDate,
		"end_date":   endDate,
	}

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[analytic.MonthlyTotals])
	if err != nil {
		return nil, fmt.Errorf("failed to collect rows from table expenses: %w", err)
	}

	return results, nil
}
