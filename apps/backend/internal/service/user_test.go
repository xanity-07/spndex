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

	"github.com/xanity-07/spndex/internal/model"
	"github.com/xanity-07/spndex/internal/model/user"
)

func TestUserService_GetUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedUsers := &model.PaginatedResponse[user.User]{
			Data: []user.User{
				{
					Email: "test@example.com",
				},
			},
			Page:       1,
			Limit:      10,
			Total:      1,
			TotalPages: 1,
		}

		query := &user.GetUsersQuery{}

		userRepo := &fakeUserRepository{
			getUsersFn: func(ctx context.Context, gotQuery *user.GetUsersQuery) (*model.PaginatedResponse[user.User], error) {
				require.Equal(t, query, gotQuery)

				return expectedUsers, nil
			},
		}

		service := NewUserService(nil, userRepo)

		result, err := service.GetUsers(ctx, query)

		require.NoError(t, err)
		require.Equal(t, expectedUsers, result)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		userRepo := &fakeUserRepository{
			getUsersFn: func(ctx context.Context, gotQuery *user.GetUsersQuery) (*model.PaginatedResponse[user.User], error) {
				return nil, expectedErr
			},
		}

		service := NewUserService(nil, userRepo)

		result, err := service.GetUsers(ctx, &user.GetUsersQuery{})

		require.Nil(t, result)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	payload := &user.GetUserByIDPayload{
		ID: userID.String(),
	}

	expectedUser := &user.User{
		Email: "test@example.com",
	}

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, gotPayload *user.GetUserByIDPayload) (*user.User, error) {
				require.Equal(t, payload, gotPayload)

				return expectedUser, nil
			},
		}

		service := NewUserService(nil, userRepo)

		result, err := service.GetUserByID(ctx, payload)

		require.NoError(t, err)
		require.Equal(t, expectedUser, result)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, gotPayload *user.GetUserByIDPayload) (*user.User, error) {
				return nil, pgx.ErrNoRows
			},
		}

		service := NewUserService(nil, userRepo)

		result, err := service.GetUserByID(ctx, payload)

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, gotPayload *user.GetUserByIDPayload) (*user.User, error) {
				return nil, expectedErr
			},
		}

		service := NewUserService(nil, userRepo)

		result, err := service.GetUserByID(ctx, payload)

		require.Nil(t, result)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	existingUser := &user.User{
		Base: model.Base{
			BaseWithID: model.BaseWithID{
				ID: userID,
			},
		},
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		firstName := "Jane"

		payload := &user.UpdateUserPayload{
			FirstName: &firstName,
		}

		updatedUser := &user.User{
			Base: model.Base{
				BaseWithID: model.BaseWithID{
					ID: userID,
				},
			},
			FirstName: "Jane",
			LastName:  "Doe",
			Email:     "john@example.com",
		}

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return existingUser, nil
			},

			updateUserFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *user.UpdateUserPayload) (*user.User, error) {
				require.Equal(t, userID, gotUserID)
				require.Equal(t, "Jane", *gotPayload.FirstName)

				return updatedUser, nil
			},
		}

		service := NewUserService(nil, userRepo)

		result, err := service.UpdateUser(ctx, userID, payload)

		require.NoError(t, err)
		require.Equal(t, updatedUser, result)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		firstName := "Jane"

		payload := &user.UpdateUserPayload{
			FirstName: &firstName,
		}

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return nil, pgx.ErrNoRows
			},
		}

		service := NewUserService(nil, userRepo)

		result, err := service.UpdateUser(ctx, userID, payload)

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("no fields updated", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		payload := &user.UpdateUserPayload{}

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return existingUser, nil
			},
		}

		service := NewUserService(nil, userRepo)

		result, err := service.UpdateUser(ctx, userID, payload)

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("email already exists", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		email := "existing@example.com"

		payload := &user.UpdateUserPayload{
			Email: &email,
		}

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return existingUser, nil
			},

			checkUserExistsFn: func(ctx context.Context, email string) (bool, error) {
				require.Equal(t, "existing@example.com", email)
				return true, nil
			},
		}

		service := NewUserService(nil, userRepo)

		result, err := service.UpdateUser(ctx, userID, payload)

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		firstName := "Jane"

		payload := &user.UpdateUserPayload{
			FirstName: &firstName,
		}

		expectedErr := errors.New("database error")

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return existingUser, nil
			},

			updateUserFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *user.UpdateUserPayload) (*user.User, error) {
				return nil, expectedErr
			},
		}

		service := NewUserService(nil, userRepo)

		result, err := service.UpdateUser(ctx, userID, payload)

		require.Nil(t, result)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New().String()

	payload := &user.DeleteUserPayload{
		ID: userID,
	}

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			deleteUserFn: func(ctx context.Context, gotPayload *user.DeleteUserPayload) error {
				require.Equal(t, payload, gotPayload)

				return nil
			},
		}

		service := NewUserService(nil, userRepo)

		err := service.DeleteUser(ctx, payload)

		require.NoError(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		userRepo := &fakeUserRepository{
			deleteUserFn: func(ctx context.Context, gotPayload *user.DeleteUserPayload) error {
				return expectedErr
			},
		}

		service := NewUserService(nil, userRepo)

		err := service.DeleteUser(ctx, payload)

		require.Error(t, err)
		require.ErrorIs(t, err, expectedErr)
		require.ErrorContains(t, err, "failed to delete user")
	})
}
