package service

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/xanity-07/spndex/internal/config"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/model"
	"github.com/xanity-07/spndex/internal/model/authmodel"
	"github.com/xanity-07/spndex/internal/model/session"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/server"
)

type fakeSessionRepository struct {
	createFn func(ctx context.Context, session *session.Session, ttl time.Duration) error
	deleteFn func(ctx context.Context, sessionID string) error
	getFn    func(ctx context.Context, sessionID string) (*session.Session, error)
}

func (f *fakeSessionRepository) Create(ctx context.Context, session *session.Session, ttl time.Duration) error {
	if f.createFn != nil {
		return f.createFn(ctx, session, ttl)
	}

	return nil
}

func (f *fakeSessionRepository) Delete(ctx context.Context, sessionID string) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, sessionID)
	}

	return nil
}

func (f *fakeSessionRepository) Get(ctx context.Context, sessionID string) (*session.Session, error) {
	if f.getFn != nil {
		return f.getFn(ctx, sessionID)
	}

	return nil, nil
}

func newTestServer() *server.Server {
	return &server.Server{
		Config: &config.Config{
			Auth: config.AuthConfig{
				TTLHours:  24,
				SecretKey: "test-secret-key",
			},
		},
	}
}

func TestAuthService_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		payload := &user.CreateUserPayload{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Password:  "StrongPassword123!",
		}

		expectedUser := &user.User{
			Base: model.Base{
				BaseWithID: model.BaseWithID{
					ID: uuid.New(),
				},
			},
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		}

		userRepo := &fakeUserRepository{
			checkUserExistsFn: func(ctx context.Context, email string) (bool, error) {
				require.Equal(t, "john@example.com", email)

				return false, nil
			},

			createUserFn: func(ctx context.Context, gotPayload *user.CreateUserPayload) (*user.User, error) {
				require.Equal(t, "john@example.com", gotPayload.Email)
				require.Equal(t, "John", gotPayload.FirstName)
				require.Equal(t, "Doe", gotPayload.LastName)

				require.NotEqual(t, "StrongPassword123!", gotPayload.Password)
				require.NotEmpty(t, gotPayload.Password)

				return expectedUser, nil
			},
		}

		service := NewAuthService(nil, userRepo, nil)

		result, err := service.Register(ctx, payload)

		require.NoError(t, err)
		require.Equal(t, expectedUser, result)
	})

	t.Run("email already exists", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		payload := &user.CreateUserPayload{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Password:  "StrongPassword123!",
		}

		userRepo := &fakeUserRepository{
			checkUserExistsFn: func(ctx context.Context, email string) (bool, error) {
				return true, nil
			},
		}

		service := NewAuthService(nil, userRepo, nil)

		result, err := service.Register(ctx, payload)

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("check user exists error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		payload := &user.CreateUserPayload{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Password:  "StrongPassword123!",
		}

		userRepo := &fakeUserRepository{
			checkUserExistsFn: func(ctx context.Context, email string) (bool, error) {
				return false, expectedErr
			},
		}

		service := NewAuthService(nil, userRepo, nil)

		result, err := service.Register(ctx, payload)

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("invalid first name", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		payload := &user.CreateUserPayload{
			FirstName: "",
			LastName:  "Doe",
			Email:     "john@example.com",
			Password:  "StrongPassword123!",
		}

		userRepo := &fakeUserRepository{
			checkUserExistsFn: func(ctx context.Context, email string) (bool, error) {
				return false, nil
			},
		}

		service := NewAuthService(nil, userRepo, nil)

		result, err := service.Register(ctx, payload)

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("invalid last name", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		payload := &user.CreateUserPayload{
			FirstName: "John",
			LastName:  "",
			Email:     "john@example.com",
			Password:  "StrongPassword123!",
		}

		userRepo := &fakeUserRepository{
			checkUserExistsFn: func(ctx context.Context, email string) (bool, error) {
				return false, nil
			},
		}

		service := NewAuthService(nil, userRepo, nil)

		result, err := service.Register(ctx, payload)

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("invalid password", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		payload := &user.CreateUserPayload{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Password:  "weak",
		}

		userRepo := &fakeUserRepository{
			checkUserExistsFn: func(ctx context.Context, email string) (bool, error) {
				return false, nil
			},
		}

		service := NewAuthService(nil, userRepo, nil)

		result, err := service.Register(ctx, payload)

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("create user error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		payload := &user.CreateUserPayload{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Password:  "StrongPassword123!",
		}

		userRepo := &fakeUserRepository{
			checkUserExistsFn: func(ctx context.Context, email string) (bool, error) {
				return false, nil
			},

			createUserFn: func(ctx context.Context, payload *user.CreateUserPayload) (*user.User, error) {
				return nil, expectedErr
			},
		}

		service := NewAuthService(nil, userRepo, nil)

		result, err := service.Register(ctx, payload)

		require.Nil(t, result)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestAuthService_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	password := "StrongPassword123!"

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	require.NoError(t, err)

	foundUser := &user.User{
		Base: model.Base{
			BaseWithID: model.BaseWithID{
				ID: userID,
			},
		},
		Email:    "john@example.com",
		Password: string(passwordHash),
		Role:     "user",
	}

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		payload := &authmodel.LoginPayload{
			Email:    "john@example.com",
			Password: password,
		}

		userRepo := &fakeUserRepository{
			getUserByEmailFn: func(ctx context.Context, email string) (*user.User, error) {
				require.Equal(t, "john@example.com", email)

				return foundUser, nil
			},
		}

		sessionRepo := &fakeSessionRepository{
			createFn: func(ctx context.Context, gotSession *session.Session, ttl time.Duration) error {
				require.NotEmpty(t, gotSession.ID)
				require.Equal(t, userID.String(), gotSession.UserID)
				require.Greater(t, ttl, time.Duration(0))

				return nil
			},
		}

		service := NewAuthService(newTestServer(), userRepo, sessionRepo)

		result, err := service.Login(ctx, payload)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, foundUser, result.User)
		require.NotEmpty(t, result.Token)
	})

	t.Run("user lookup fails", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		payload := &authmodel.LoginPayload{
			Email:    "missing@example.com",
			Password: password,
		}

		userRepo := &fakeUserRepository{
			getUserByEmailFn: func(ctx context.Context, email string) (*user.User, error) {
				return nil, errors.New("user not found")
			},
		}

		service := NewAuthService(nil, userRepo, nil)

		result, err := service.Login(ctx, payload)

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("incorrect password", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		payload := &authmodel.LoginPayload{
			Email:    "john@example.com",
			Password: "WrongPassword123!",
		}

		userRepo := &fakeUserRepository{
			getUserByEmailFn: func(ctx context.Context, email string) (*user.User, error) {
				return foundUser, nil
			},
		}

		service := NewAuthService(nil, userRepo, nil)

		result, err := service.Login(ctx, payload)

		require.Nil(t, result)
		require.Error(t, err)
	})

	t.Run("session creation fails", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("session creation failed")

		payload := &authmodel.LoginPayload{
			Email:    "john@example.com",
			Password: password,
		}

		userRepo := &fakeUserRepository{
			getUserByEmailFn: func(ctx context.Context, email string) (*user.User, error) {
				return foundUser, nil
			},
		}

		sessionRepo := &fakeSessionRepository{
			createFn: func(ctx context.Context, gotSession *session.Session, ttl time.Duration) error {
				return expectedErr
			},
		}

		service := NewAuthService(newTestServer(), userRepo, sessionRepo)

		result, err := service.Login(ctx, payload)

		require.Nil(t, result)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestAuthService_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("session deletion fails", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		sessionID := uuid.NewString()

		ctx.Set(middleware.SessionIDKey, sessionID)

		expectedErr := errors.New("session deletion failed")

		sessionRepo := &fakeSessionRepository{
			deleteFn: func(ctx context.Context, gotSessionID string) error {
				require.Equal(t, sessionID, gotSessionID)

				return expectedErr
			},
		}

		service := NewAuthService(nil, nil, sessionRepo)

		err := service.Logout(ctx)

		require.ErrorIs(t, err, expectedErr)
		require.ErrorContains(t, err, "failed to delete session")
	})

	t.Run("missing session id", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		sessionRepo := &fakeSessionRepository{
			deleteFn: func(ctx context.Context, sessionID string) error {
				t.Fatal("Delete should not be called")

				return nil
			},
		}

		service := NewAuthService(nil, nil, sessionRepo)

		err := service.Logout(ctx)

		require.Error(t, err)
	})

	t.Run("session deletion fails", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		sessionID := uuid.NewString()

		ctx.Set(middleware.SessionIDKey, sessionID)

		expectedErr := errors.New("session deletion failed")

		sessionRepo := &fakeSessionRepository{
			deleteFn: func(ctx context.Context, gotSessionID string) error {
				require.Equal(t, sessionID, gotSessionID)

				return expectedErr
			},
		}

		service := NewAuthService(nil, nil, sessionRepo)

		err := service.Logout(ctx)

		require.ErrorIs(t, err, expectedErr)
		require.ErrorContains(t, err, "failed to delete session")
	})
}
