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

func TestExpenseRepository_GetExpenseByID(t *testing.T) {
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

	// Create user.
	userStmt := `
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

	userArgs := pgx.NamedArgs{
		"user_id":       userID,
		"first_name":    "John",
		"last_name":     "Doe",
		"email":         "get-expense@example.com",
		"password_hash": "hashed-password",
	}

	_, err := testDB.Pool.Exec(ctx, userStmt, userArgs)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// Create expense.
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

	description := "Lunch"

	expenseArgs := pgx.NamedArgs{
		"expense_id":  expenseID,
		"user_id":     userID,
		"amount":      2500,
		"description": description,
		"category":    enums.ExpenseCategory("food"),
		"currency":    enums.CurrencyCode("USD"),
		"date":        "2026-08-22",
	}

	_, err = testDB.Pool.Exec(ctx, expenseStmt, expenseArgs)
	if err != nil {
		t.Fatalf("failed to insert test expense: %v", err)
	}

	payload := &expense.GetExpenseByID{
		ID: expenseID.String(),
	}

	result, err := repo.GetExpenseByID(ctx, userID, payload)
	if err != nil {
		t.Fatalf("GetExpenseByID returned an error: %v", err)
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

	if result.AmountCents != 2500 {
		t.Errorf("expected amount 2500, got %v", result.AmountCents)
	}

	if result.Description == nil {
		t.Fatal("expected description, got nil")
	}

	if *result.Description != description {
		t.Errorf("expected description %q, got %q", description, *result.Description)
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

func TestExpenseRepository_GetExpenseByID_NotFound(t *testing.T) {
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

	payload := &expense.GetExpenseByID{
		ID: expenseID.String(),
	}

	result, err := repo.GetExpenseByID(ctx, userID, payload)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil expense, got %+v", result)
	}

	if err.Error() != "expense not found" {
		t.Errorf("expected expense not found error, got %q", err.Error())
	}
}

func TestExpenseRepository_GetExpenseByID_WrongUser(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	ownerID := uuid.New()
	otherUserID := uuid.New()
	expenseID := uuid.New()

	userStmt := `
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

	for _, user := range []struct {
		id    uuid.UUID
		email string
	}{
		{ownerID, "owner@example.com"},
		{otherUserID, "other@example.com"},
	} {
		_, err := testDB.Pool.Exec(ctx, userStmt, pgx.NamedArgs{
			"user_id":       user.id,
			"first_name":    "John",
			"last_name":     "Doe",
			"email":         user.email,
			"password_hash": "hashed-password",
		})
		if err != nil {
			t.Fatalf("failed to insert test user: %v", err)
		}
	}

	expenseStmt := `
		INSERT INTO expenses (
			id,
			user_id,
			amount,
			category,
			currency,
			date
		)
		VALUES (
			@expense_id,
			@user_id,
			@amount,
			@category,
			@currency,
			@date
		)
	`

	_, err := testDB.Pool.Exec(ctx, expenseStmt, pgx.NamedArgs{
		"expense_id": expenseID,
		"user_id":    ownerID,
		"amount":     2500,
		"category":   enums.ExpenseCategory("food"),
		"currency":   enums.CurrencyCode("USD"),
		"date":       "2026-08-22",
	})
	if err != nil {
		t.Fatalf("failed to insert test expense: %v", err)
	}

	payload := &expense.GetExpenseByID{
		ID: expenseID.String(),
	}

	result, err := repo.GetExpenseByID(ctx, otherUserID, payload)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil expense, got %+v", result)
	}

	if err.Error() != "expense not found" {
		t.Errorf("expected expense not found error, got %q", err.Error())
	}
}

func TestExpenseRepository_GetExpenseByID_DeletedExpense(t *testing.T) {
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

	userStmt := `
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

	_, err := testDB.Pool.Exec(ctx, userStmt, pgx.NamedArgs{
		"user_id":       userID,
		"first_name":    "John",
		"last_name":     "Doe",
		"email":         "deleted-expense@example.com",
		"password_hash": "hashed-password",
	})
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	expenseStmt := `
		INSERT INTO expenses (
			id,
			user_id,
			amount,
			category,
			currency,
			date,
			deleted_at
		)
		VALUES (
			@expense_id,
			@user_id,
			@amount,
			@category,
			@currency,
			@date,
			@deleted_at
		)
	`

	_, err = testDB.Pool.Exec(ctx, expenseStmt, pgx.NamedArgs{
		"expense_id": expenseID,
		"user_id":    userID,
		"amount":     2500,
		"category":   enums.ExpenseCategory("food"),
		"currency":   enums.CurrencyCode("USD"),
		"date":       "2026-08-22",
		"deleted_at": time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to insert deleted expense: %v", err)
	}

	payload := &expense.GetExpenseByID{
		ID: expenseID.String(),
	}

	result, err := repo.GetExpenseByID(ctx, userID, payload)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Errorf("expected nil expense, got %+v", result)
	}

	if err.Error() != "expense not found" {
		t.Errorf("expected expense not found error, got %q", err.Error())
	}
}

func TestExpenseRepository_GetExpenses(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)

	ctx := context.Background()

	userID := uuid.New()

	userStmt := `
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

	userArgs := pgx.NamedArgs{
		"user_id":       userID,
		"first_name":    "John",
		"last_name":     "Doe",
		"email":         "john@example.com",
		"password_hash": "hashed-password",
	}

	_, err := testDB.Pool.Exec(ctx, userStmt, userArgs)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
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
		Description string
		Category    enums.ExpenseCategory
		Currency    enums.CurrencyCode
		Date        string
		Amount      int64
		ID          uuid.UUID
	}{
		{
			ID:          uuid.New(),
			Amount:      2500,
			Description: "Lunch",
			Category:    enums.ExpenseCategory("food"),
			Currency:    enums.CurrencyCode("USD"),
			Date:        "2026-08-20",
		},
		{
			ID:          uuid.New(),
			Amount:      5000,
			Description: "Gas",
			Category:    enums.ExpenseCategory("transport"),
			Currency:    enums.CurrencyCode("USD"),
			Date:        "2026-08-21",
		},
	}

	for _, exp := range expenses {
		args := pgx.NamedArgs{
			"id":          exp.ID,
			"user_id":     userID,
			"amount":      exp.Amount,
			"description": exp.Description,
			"category":    exp.Category,
			"currency":    exp.Currency,
			"date":        exp.Date,
		}

		_, err = testDB.Pool.Exec(ctx, expenseStmt, args)
		if err != nil {
			t.Fatalf("failed to insert test expense: %v", err)
		}
	}

	currency := enums.CurrencyCode("USD")

	page := 1
	limit := 5

	query := &expense.GetExpensesQuery{
		Page:         &page,
		Limit:        &limit,
		CurrencyCode: &currency,
	}

	result, err := repo.GetExpenses(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetExpenses returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(result.Data))
	}

	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}

	if result.Limit != 5 {
		t.Errorf("expected limit 5, got %d", result.Limit)
	}

	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}

	if result.TotalPages != 1 {
		t.Errorf("expected total pages 1, got %d", result.TotalPages)
	}

	if result.Data[0].CurrencyCode != enums.CurrencyCode("USD") {
		t.Errorf(
			"expected currency USD, got %q",
			result.Data[0].CurrencyCode,
		)
	}
}

func TestExpenseRepository_GetExpenses_SearchDescription(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

	tests.InsertTestUser(t, ctx, testDB.Pool, userID)

	expense1ID := uuid.New()
	expense2ID := uuid.New()

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
		VALUES
			(
				@id1,
				@user_id,
				1000,
				@description1,
				'food',
				'USD',
				'2026-08-20'
			),
			(
				@id2,
				@user_id,
				2000,
				@description2,
				'transport',
				'USD',
				'2026-08-21'
			)
	`

	args := pgx.NamedArgs{
		"id1":          expense1ID,
		"id2":          expense2ID,
		"user_id":      userID,
		"description1": "Lunch at restaurant",
		"description2": "Gas station",
	}

	_, err := testDB.Pool.Exec(ctx, stmt, args)
	if err != nil {
		t.Fatalf("failed to insert test expenses: %v", err)
	}

	page := 1
	limit := 5
	search := "Lunch"

	query := &expense.GetExpensesQuery{
		Page:   &page,
		Limit:  &limit,
		Search: &search,
	}

	result, err := repo.GetExpenses(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetExpenses returned an error: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 expense, got %d", len(result.Data))
	}

	if result.Data[0].ID != expense1ID {
		t.Errorf(
			"expected expense %s, got %s",
			expense1ID,
			result.Data[0].ID,
		)
	}
}

func TestExpenseRepository_GetExpenses_SearchCategory(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()

	tests.InsertTestUser(t, ctx, testDB.Pool, userID)

	expense1ID := uuid.New()
	expense2ID := uuid.New()

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
		VALUES
			(
				@id1,
				@user_id,
				1000,
				'Lunch',
				'food',
				'USD',
				'2026-08-20'
			),
			(
				@id2,
				@user_id,
				2000,
				'Gas',
				'transport',
				'USD',
				'2026-08-21'
			)
	`

	args := pgx.NamedArgs{
		"id1":     expense1ID,
		"id2":     expense2ID,
		"user_id": userID,
	}

	_, err := testDB.Pool.Exec(ctx, stmt, args)
	if err != nil {
		t.Fatalf("failed to insert test expenses: %v", err)
	}

	page := 1
	limit := 5
	search := "food"

	query := &expense.GetExpensesQuery{
		Page:   &page,
		Limit:  &limit,
		Search: &search,
	}

	result, err := repo.GetExpenses(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetExpenses returned an error: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 expense, got %d", len(result.Data))
	}

	if result.Data[0].ID != expense1ID {
		t.Errorf(
			"expected expense %s, got %s",
			expense1ID,
			result.Data[0].ID,
		)
	}
}

func TestExpenseRepository_GetExpenses_CategoryFilter(t *testing.T) {
	testDB := tests.NewPostgresTestDatabase(t)

	srv := &server.Server{
		DB: &database.Database{
			Pool: testDB.Pool,
		},
	}

	repo := NewExpenseRepository(srv)
	ctx := context.Background()

	userID := uuid.New()
	tests.InsertTestUser(t, ctx, testDB.Pool, userID)

	expense1ID := uuid.New()
	expense2ID := uuid.New()
	expense3ID := uuid.New()

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
		VALUES
			(
				@expense1_id,
				@user_id,
				1000,
				'Lunch',
				'food',
				'USD',
				'2026-08-20'
			),
			(
				@expense2_id,
				@user_id,
				5000,
				'Gas',
				'transport',
				'USD',
				'2026-08-21'
			),
			(
				@expense3_id,
				@user_id,
				2000,
				'Dinner',
				'food',
				'USD',
				'2026-08-22'
			)
	`

	args := pgx.NamedArgs{
		"expense1_id": expense1ID,
		"expense2_id": expense2ID,
		"expense3_id": expense3ID,
		"user_id":     userID,
	}

	_, err := testDB.Pool.Exec(ctx, stmt, args)
	if err != nil {
		t.Fatalf("failed to insert test expenses: %v", err)
	}

	page := 1
	limit := 5
	category := enums.ExpenseCategory("food")

	query := &expense.GetExpensesQuery{
		Page:     &page,
		Limit:    &limit,
		Category: &category,
	}

	result, err := repo.GetExpenses(ctx, userID, query)
	if err != nil {
		t.Fatalf("GetExpenses returned an error: %v", err)
	}

	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(result.Data))
	}

	for _, exp := range result.Data {
		if exp.Category != enums.ExpenseCategory("food") {
			t.Errorf(
				"expected category food, got %q",
				exp.Category,
			)
		}
	}

	if result.TotalPages != 1 {
		t.Errorf(
			"expected total pages 1, got %d",
			result.TotalPages,
		)
	}

	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}

	if result.Limit != 5 {
		t.Errorf("expected limit 5, got %d", result.Limit)
	}
}
