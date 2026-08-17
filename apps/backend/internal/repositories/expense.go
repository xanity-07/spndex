package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/model"
	"github.com/xanity-07/spndex/internal/model/expense"
	"github.com/xanity-07/spndex/internal/server"
)

type ExpenseRepository struct {
	server *server.Server
}

func NewExpenseRepository(s *server.Server) *ExpenseRepository {
	return &ExpenseRepository{
		server: s,
	}
}

func (r *ExpenseRepository) CreateExpense(
	ctx context.Context,
	userID uuid.UUID,
	payload *expense.CreateExpensePayload,
) (*expense.Expense, error) {
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
	SELECT
		@id,
		u.id,
		@amount,
		@description,
		@category,
		@currency,
		@date
	FROM
		users u
	WHERE
		u.id = @user_id
	    AND u.deleted_at IS NULL
	RETURNING
		id,
		user_id,
		amount,
		description,
		category,
		currency,
		date,
		created_at,
		updated_at,
		deleted_at
`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":          uuid.New(),
		"user_id":     userID,
		"description": payload.Description,
		"amount":      payload.Amount,
		"category":    payload.Category,
		"currency":    payload.CurrencyCode,
		"date":        payload.Date,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	exp, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[expense.Expense])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewUnauthorizedError("user is deleted or does not exist", false)
		}

		return nil, fmt.Errorf("failed to collect row from table expenses: %w", err)
	}

	return &exp, nil
}

func (r *ExpenseRepository) GetExpenses(
	ctx context.Context,
	userID uuid.UUID,
	query *expense.GetExpensesQuery,
) (*model.PaginatedResponse[expense.Expense], error) {
	stmt := `
		SELECT
			e.id,
			e.user_id,
			e.amount,
			e.description,
			e.category,
			e.currency,
			e.date,
			e.created_at,
			e.updated_at,
			e.deleted_at
		FROM
			expenses e
		INNER JOIN users u ON e.user_id = u.id
	`

	conditions := []string{"e.deleted_at IS NULL AND u.deleted_at IS NULL"}
	args := pgx.NamedArgs{
		"user_id": userID,
	}

	if query.Search != nil {
		conditions = append(conditions, "(e.description ILIKE @search OR e.category ILIKE @search)")
		args["search"] = "%" + *query.Search + "%"
	}

	if query.Category != nil {
		conditions = append(conditions, "e.category = @category")
		args["category"] = *query.Category
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause += " WHERE e.user_id = @user_id AND " + strings.Join(conditions, " AND ")
	}

	stmt += whereClause

	var total int
	countStmt := "SELECT COUNT(*) FROM expenses e INNER JOIN users u ON e.user_id = u.id " + whereClause

	err := r.server.DB.Pool.QueryRow(ctx, countStmt, args).Scan(&total)

	if err != nil {
		return nil, fmt.Errorf("failed to get expenses count: %w", err)
	}

	orderColumn := map[string]string{
		"amount":     "e.amount",
		"created_at": "e.created_at",
		"category":   "e.category",
		"date":       "e.date",
	}

	if query.Order != nil {
		stmt += " ORDER BY " + orderColumn[*query.Order]
		if query.Sort != nil && *query.Sort == "desc" {
			stmt += " DESC"
		} else {
			stmt += " ASC"
		}
	} else {
		stmt += " ORDER BY created_at DESC"
	}

	stmt += " OFFSET @offset LIMIT @limit"
	args["limit"] = query.Limit
	args["offset"] = (*query.Page - 1) * *query.Limit

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	expenseList, err := pgx.CollectRows(rows, pgx.RowToStructByName[expense.Expense])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PaginatedResponse[expense.Expense]{
				Data:       []expense.Expense{},
				Page:       1,
				Limit:      0,
				Total:      0,
				TotalPages: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to collect rows from table expenses: %w", err)
	}

	return &model.PaginatedResponse[expense.Expense]{
		Data:       expenseList,
		Page:       *query.Page,
		Limit:      *query.Limit,
		Total:      total,
		TotalPages: (total + *query.Limit - 1) / *query.Limit,
	}, nil
}

func (r *ExpenseRepository) GetExpenseByID(ctx context.Context, userID uuid.UUID, payload *expense.GetExpenseByID) (*expense.Expense, error) {
	stmt := `
		SELECT
			e.id,
			e.user_id,
			e.description,
			e.amount,
			e.category,
			e.currency,
			e.date,
			e.created_at,
			e.updated_at,
			e.deleted_at
		FROM
			expenses e
		INNER JOIN users u ON e.user_id = u.id
		WHERE
			e.id = @id
			AND e.user_id = @user_id
			AND e.deleted_at IS NULL
			AND u.deleted_at IS NULL
`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":      payload.ID,
		"user_id": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	exp, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[expense.Expense])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			code := "EXPENSE_NOT_FOUND"
			return nil, errs.NewNotFoundError("expense not found", false, &code)
		}
		return nil, fmt.Errorf("failed to collect row from table expenses: %w", err)
	}

	return &exp, nil
}

func (r *ExpenseRepository) UpdateExpense(ctx context.Context, userID uuid.UUID, expenseID uuid.UUID, payload *expense.UpdateExpense) (*expense.Expense, error) {
	stmt := "UPDATE expenses e SET "

	args := pgx.NamedArgs{
		"id":      expenseID,
		"user_id": userID,
	}

	setClauses := []string{}

	if payload.Amount != nil {
		setClauses = append(setClauses, "amount = @amount")
		args["amount"] = payload.Amount
	}

	if payload.Category != nil {
		setClauses = append(setClauses, "category = @category")
		args["category"] = payload.Category
	}

	if payload.CurrencyCode != nil {
		setClauses = append(setClauses, "currency = @currency")
		args["currency"] = payload.CurrencyCode
	}

	if payload.Date != nil {
		setClauses = append(setClauses, "date = @date")
		args["date"] = payload.Date
	}

	if payload.Description != nil {
		setClauses = append(setClauses, "description = @description")
		args["description"] = payload.Description
	}

	if len(setClauses) == 0 {
		return nil, errs.NewBadRequestError(
			"no fields to update", false, nil, nil, nil,
		)
	}

	setClauses = append(setClauses, "updated_at = @updated_at")
	args["updated_at"] = time.Now()
	stmt += strings.Join(setClauses, ", ")
	stmt += `
		FROM
			users u
		WHERE
			e.id = @id
			AND e.user_id = @user_id
			AND e.deleted_at IS NULL
			AND u.id = e.user_id
			AND u.deleted_at IS NULL
		RETURNING
			e.id,
			e.user_id,
			e.amount,
			e.description,
			e.category,
			e.currency,
			e.date,
			e.created_at,
			e.updated_at,
			e.deleted_at
	`
	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	updatedExpense, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[expense.Expense])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			code := "EXPENSE_NOT_FOUND"
			return nil, errs.NewNotFoundError("expense not found", false, &code)
		}
		return nil, fmt.Errorf("failed to collect row from table expenses: %w", err)
	}

	return &updatedExpense, nil
}

func (r *ExpenseRepository) DeleteExpense(ctx context.Context, userID uuid.UUID, payload *expense.DeleteExpense) error {
	stmt := `
		UPDATE expenses e SET
			deleted_at = NOW()
		FROM
			users u
		WHERE
			e.user_id = u.id
			AND e.id = @id
			AND e.user_id = @user_id
			AND e.deleted_at IS NULL
			AND u.deleted_at IS NULL
	`
	result, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"id":      payload.ID,
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}

	if result.RowsAffected() == 0 {
		code := "EXPENSE_NOT_FOUND"
		return errs.NewNotFoundError("expense not found", false, &code)
	}
	return nil
}
