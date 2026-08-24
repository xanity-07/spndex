package service

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/model"
	"github.com/xanity-07/spndex/internal/model/expense"
	"github.com/xanity-07/spndex/internal/model/user"
)

type fakeExpenseRepository struct {
	createExpenseFn  func(context.Context, uuid.UUID, *expense.CreateExpensePayload) (*expense.Expense, error)
	deleteExpenseFn  func(context.Context, uuid.UUID, *expense.DeleteExpense) error
	getExpenseByIDFn func(context.Context, uuid.UUID, *expense.GetExpenseByID) (*expense.Expense, error)
	getExpensesFn    func(context.Context, uuid.UUID, *expense.GetExpensesQuery) (*model.PaginatedResponse[expense.Expense], error)
	updateExpenseFn  func(context.Context, uuid.UUID, uuid.UUID, *expense.UpdateExpense) (*expense.Expense, error)
}

func (f *fakeExpenseRepository) CreateExpense(ctx context.Context, userID uuid.UUID, payload *expense.CreateExpensePayload) (*expense.Expense, error) {
	return f.createExpenseFn(ctx, userID, payload)
}

func (f *fakeExpenseRepository) DeleteExpense(ctx context.Context, userID uuid.UUID, payload *expense.DeleteExpense) error {
	return f.deleteExpenseFn(ctx, userID, payload)
}

func (f *fakeExpenseRepository) GetExpenseByID(ctx context.Context, userID uuid.UUID, payload *expense.GetExpenseByID) (*expense.Expense, error) {
	return f.getExpenseByIDFn(ctx, userID, payload)
}

func (f *fakeExpenseRepository) GetExpenses(ctx context.Context, userID uuid.UUID, query *expense.GetExpensesQuery) (*model.PaginatedResponse[expense.Expense], error) {
	return f.getExpensesFn(ctx, userID, query)
}

func (f *fakeExpenseRepository) UpdateExpense(ctx context.Context, userID uuid.UUID, expenseID uuid.UUID, payload *expense.UpdateExpense) (*expense.Expense, error) {
	return f.updateExpenseFn(ctx, userID, expenseID, payload)
}

type fakeUserRepository struct {
	checkUserExistsFn func(ctx context.Context, email string) (bool, error)
	createUserFn      func(ctx context.Context, payload *user.CreateUserPayload) (*user.User, error)
	deleteUserFn      func(ctx context.Context, payload *user.DeleteUserPayload) error
	getUserByEmailFn  func(ctx context.Context, email string) (*user.User, error)
	getUserByIDFn     func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error)
	getUsersFn        func(ctx context.Context, query *user.GetUsersQuery) (*model.PaginatedResponse[user.User], error)
	updateUserFn      func(ctx context.Context, userID uuid.UUID, payload *user.UpdateUserPayload) (*user.User, error)
}

func (f *fakeUserRepository) CheckUserExists(ctx context.Context, email string) (bool, error) {
	if f.checkUserExistsFn != nil {
		return f.checkUserExistsFn(ctx, email)
	}

	panic("unexpected call to CheckUserExists")
}

func (f *fakeUserRepository) CreateUser(ctx context.Context, payload *user.CreateUserPayload) (*user.User, error) {
	if f.createUserFn != nil {
		return f.createUserFn(ctx, payload)
	}

	panic("unexpected call to CreateUser")
}

func (f *fakeUserRepository) DeleteUser(ctx context.Context, payload *user.DeleteUserPayload) error {
	if f.deleteUserFn != nil {
		return f.deleteUserFn(ctx, payload)
	}

	panic("unexpected call to DeleteUser")
}

func (f *fakeUserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	if f.getUserByEmailFn != nil {
		return f.getUserByEmailFn(ctx, email)
	}

	panic("unexpected call to GetUserByEmail")
}

func (f *fakeUserRepository) GetUserByID(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
	if f.getUserByIDFn != nil {
		return f.getUserByIDFn(ctx, payload)
	}

	panic("unexpected call to GetUserByID")
}

func (f *fakeUserRepository) GetUsers(ctx context.Context, query *user.GetUsersQuery) (*model.PaginatedResponse[user.User], error) {
	if f.getUsersFn != nil {
		return f.getUsersFn(ctx, query)
	}

	panic("unexpected call to GetUsers")
}

func (f *fakeUserRepository) UpdateUser(ctx context.Context, userID uuid.UUID, payload *user.UpdateUserPayload) (*user.User, error) {
	if f.updateUserFn != nil {
		return f.updateUserFn(ctx, userID, payload)
	}

	panic("unexpected call to UpdateUser")
}

func TestExpenseService_CreateExpense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	payload := &expense.CreateExpensePayload{
		AmountCents: 2500,
	}

	expectedExpense := &expense.Expense{
		UserID:      userID.String(),
		AmountCents: 2500,
		Base: model.Base{
			BaseWithID: model.BaseWithID{
				ID: uuid.New(),
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		repo := &fakeExpenseRepository{
			createExpenseFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *expense.CreateExpensePayload) (*expense.Expense, error) {
				require.Equal(t, userID, gotUserID)
				require.Equal(t, payload, gotPayload)

				return expectedExpense, nil
			},
		}

		service := NewExpenseService(nil, repo, nil)

		result, err := service.CreateExpense(ctx, userID, payload)

		require.NoError(t, err)
		require.Equal(t, expectedExpense, result)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		repo := &fakeExpenseRepository{
			createExpenseFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *expense.CreateExpensePayload) (*expense.Expense, error) {
				return nil, expectedErr
			},
		}

		service := NewExpenseService(
			nil,
			repo,
			nil,
		)

		result, err := service.CreateExpense(ctx, userID, payload)

		require.Nil(t, result)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestExpenseService_GetExpenses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	query := &expense.GetExpensesQuery{}

	expectedResponse := &model.PaginatedResponse[expense.Expense]{
		Data: []expense.Expense{
			{
				UserID:      userID.String(),
				AmountCents: 2500,
				Base: model.Base{
					BaseWithID: model.BaseWithID{
						ID: uuid.New(),
					},
				},
			},
		},
		Page:       1,
		Limit:      10,
		Total:      1,
		TotalPages: 1,
	}

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		repo := &fakeExpenseRepository{
			getExpensesFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *expense.GetExpensesQuery) (*model.PaginatedResponse[expense.Expense], error) {
				require.Equal(t, userID, gotUserID)
				require.Equal(t, query, gotQuery)

				return expectedResponse, nil
			},
		}

		service := NewExpenseService(nil, repo, nil)

		result, err := service.GetExpenses(ctx, userID, query)

		require.NoError(t, err)
		require.Equal(t, expectedResponse, result)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		repo := &fakeExpenseRepository{
			getExpensesFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *expense.GetExpensesQuery) (*model.PaginatedResponse[expense.Expense], error) {
				return nil, expectedErr
			},
		}

		service := NewExpenseService(nil, repo, nil)

		result, err := service.GetExpenses(ctx, userID, query)

		require.Nil(t, result)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestExpenseService_GetExpenseByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	expenseID := uuid.New().String()

	payload := &expense.GetExpenseByID{
		ID: expenseID,
	}

	expectedExpense := &expense.Expense{
		UserID:      userID.String(),
		AmountCents: 2500,
		Base: model.Base{
			BaseWithID: model.BaseWithID{
				ID: uuid.MustParse(expenseID),
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		repo := &fakeExpenseRepository{
			getExpenseByIDFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *expense.GetExpenseByID) (*expense.Expense, error) {
				require.Equal(t, userID, gotUserID)
				require.Equal(t, payload, gotPayload)

				return expectedExpense, nil
			},
		}

		service := NewExpenseService(nil, repo, nil)

		result, err := service.GetExpenseByID(ctx, userID, payload)

		require.NoError(t, err)
		require.Equal(t, expectedExpense, result)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		repo := &fakeExpenseRepository{
			getExpenseByIDFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *expense.GetExpenseByID) (*expense.Expense, error) {
				return nil, expectedErr
			},
		}

		service := NewExpenseService(nil, repo, nil)

		result, err := service.GetExpenseByID(ctx, userID, payload)

		require.Nil(t, result)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestExpenseService_UpdateExpense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	expenseID := uuid.New()

	t.Run("expense not found", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := pgx.ErrNoRows

		repo := &fakeExpenseRepository{
			getExpenseByIDFn: func(ctx context.Context, gotUserID uuid.UUID, payload *expense.GetExpenseByID) (*expense.Expense, error) {
				require.Equal(t, userID, gotUserID)
				require.Equal(t, expenseID.String(), payload.ID)

				return nil, expectedErr
			},
		}

		service := NewExpenseService(nil, repo, nil)

		result, err := service.UpdateExpense(ctx, userID, expenseID.String(), &expense.UpdateExpense{})

		require.Nil(t, result)
		require.Error(t, err)

		var appErr *errs.AppError
		require.ErrorAs(t, err, &appErr)
		require.Equal(t, "EXPENSE_NOT_FOUND", appErr.Code)
	})

	t.Run("invalid category", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		invalidCategory := enums.ExpenseCategory("invalid")

		repo := &fakeExpenseRepository{
			getExpenseByIDFn: func(ctx context.Context, gotUserID uuid.UUID, payload *expense.GetExpenseByID) (*expense.Expense, error) {
				return &expense.Expense{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: expenseID,
						},
					},
					UserID: userID.String(),
				}, nil
			},
		}

		service := NewExpenseService(nil, repo, nil)

		result, err := service.UpdateExpense(ctx, userID, expenseID.String(), &expense.UpdateExpense{
			Category: &invalidCategory,
		})

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("negative amount", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		var negativeAmount int64 = -100

		repo := &fakeExpenseRepository{
			getExpenseByIDFn: func(ctx context.Context, gotUserID uuid.UUID, payload *expense.GetExpenseByID) (*expense.Expense, error) {
				return &expense.Expense{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: expenseID,
						},
					},
					UserID: userID.String(),
				}, nil
			},
		}

		service := NewExpenseService(nil, repo, nil)

		result, err := service.UpdateExpense(ctx, userID, expenseID.String(), &expense.UpdateExpense{
			AmountCents: &negativeAmount,
		})

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		var amount int64 = 5000
		payload := &expense.UpdateExpense{
			AmountCents: &amount,
		}

		expectedExpense := &expense.Expense{
			UserID:      userID.String(),
			AmountCents: int64(amount),
			Base: model.Base{
				BaseWithID: model.BaseWithID{
					ID: expenseID,
				},
			},
		}

		repo := &fakeExpenseRepository{
			getExpenseByIDFn: func(ctx context.Context, gotUserID uuid.UUID, payload *expense.GetExpenseByID) (*expense.Expense, error) {
				return &expense.Expense{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: expenseID,
						},
					},
					UserID: userID.String(),
				}, nil
			},
			updateExpenseFn: func(ctx context.Context, gotUserID uuid.UUID, gotExpenseID uuid.UUID, gotPayload *expense.UpdateExpense) (*expense.Expense, error) {
				require.Equal(t, userID, gotUserID)
				require.Equal(t, expenseID, gotExpenseID)
				require.Equal(t, payload, gotPayload)

				return expectedExpense, nil
			},
		}

		service := NewExpenseService(nil, repo, nil)

		result, err := service.UpdateExpense(ctx, userID, expenseID.String(), payload)

		require.NoError(t, err)
		require.Equal(t, expectedExpense, result)
	})
}

func TestExpenseService_DeleteExpense(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	payload := &expense.DeleteExpense{
		ID: uuid.New().String(),
	}

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				require.Equal(t, userID.String(), payload.ID)

				return &user.User{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: userID,
						},
					},
				}, nil
			},
		}

		deleteCalled := false

		expenseRepo := &fakeExpenseRepository{
			deleteExpenseFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *expense.DeleteExpense) error {
				deleteCalled = true

				require.Equal(t, userID, gotUserID)
				require.Equal(t, payload, gotPayload)

				return nil
			},
		}

		service := NewExpenseService(nil, expenseRepo, userRepo)

		err := service.DeleteExpense(ctx, userID, payload)

		require.NoError(t, err)
		require.True(t, deleteCalled)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return nil, errors.New("user not found")
			},
		}

		deleteCalled := false

		expenseRepo := &fakeExpenseRepository{
			deleteExpenseFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *expense.DeleteExpense) error {
				deleteCalled = true
				return nil
			},
		}

		service := NewExpenseService(nil, expenseRepo, userRepo)

		err := service.DeleteExpense(ctx, userID, payload)

		require.Error(t, err)
		require.False(t, deleteCalled)

		var appErr *errs.AppError
		require.ErrorAs(t, err, &appErr)
		require.Equal(t, "USER_NOT_FOUND", appErr.Code)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return &user.User{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: userID,
						},
					},
				}, nil
			},
		}

		expenseRepo := &fakeExpenseRepository{
			deleteExpenseFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *expense.DeleteExpense) error {
				return expectedErr
			},
		}

		service := NewExpenseService(nil, expenseRepo, userRepo)

		err := service.DeleteExpense(ctx, userID, payload)

		require.ErrorIs(t, err, expectedErr)
	})
}
