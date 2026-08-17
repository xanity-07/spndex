package service

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/model"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/repositories"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/validation"
)

type UserService struct {
	server   *server.Server
	userRepo *repositories.UserRepository
}

func NewUserService(s *server.Server, userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		server:   s,
		userRepo: userRepo,
	}
}

func (s *UserService) GetUsers(ctx *gin.Context, query *user.GetUsersQuery) (*model.PaginatedResponse[user.User], error) {
	logger := middleware.GetLogger(ctx)
	userList, err := s.userRepo.GetUsers(ctx, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch users")
		return nil, err
	}
	return userList, nil
}

func (s *UserService) GetUserByID(ctx *gin.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
	logger := middleware.GetLogger(ctx)

	user, err := s.userRepo.GetUserByID(ctx, payload)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			code := "USER_NOT_FOUND"
			return nil, errs.NewBadRequestError("user not found", false, &code, nil, nil)
		}
		logger.Error().Err(err).Msg("failed to fetch user by ID")
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx *gin.Context, userID uuid.UUID, payload *user.UpdateUserPayload) (*user.User, error) {
	logger := middleware.GetLogger(ctx)

	changed := false

	// Check if user exists
	user, err := s.userRepo.GetUserByID(ctx, &user.GetUserByIDPayload{ID: userID.String()})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			code := "USER_NOT_FOUND"
			return nil, errs.NewBadRequestError("user not found", false, &code, nil, nil)
		}
		logger.Error().Err(err).Msg("failed to fetch user")
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if payload.FirstName != nil {
		if err = validation.ValidateName(*payload.FirstName); err != nil {
			logger.Error().Err(err).Msg("invalid first name")
			return nil, errs.NewBadRequestError(err.Error(), false, nil, nil, nil)
		}

		if user.FirstName != *payload.FirstName {
			changed = true
		}
	}

	if payload.LastName != nil {
		if err = validation.ValidateName(*payload.LastName); err != nil {
			logger.Error().Err(err).Msg("invalid last name")
			return nil, errs.NewBadRequestError(err.Error(), false, nil, nil, nil)
		}

		if user.LastName != *payload.LastName {
			changed = true
		}
	}

	// Check if email is already in use
	var exists bool
	if payload.Email != nil && user.Email != *payload.Email {
		changed = true

		exists, err = s.userRepo.CheckUserExists(ctx, *payload.Email)
		if err != nil {
			logger.Error().Err(err).Msg("failed to check if user exists")
			return nil, err
		}

		if exists {
			return nil, errs.NewBadRequestError("user with this email already exists", false, nil, nil, nil)
		}
	}

	// Check password strength
	var matches bool
	if payload.Password != nil {
		if err = validation.ValidatePasswordStrength(*payload.Password); err != nil {
			logger.Error().Err(err).Msg("invalid password")
			return nil, errs.NewBadRequestError(err.Error(), false, nil, nil, nil)
		}

		matches, err = validation.ComparePassword(user.Password, *payload.Password)
		if err != nil {
			logger.Error().Err(err).Msg("failed to compare passwords")
			return nil, fmt.Errorf("failed to compare password: %w", err)
		}

		if !matches {
			passwordHash, err := validation.HashPassword(*payload.Password)
			if err != nil {
				logger.Error().Err(err).Msg("failed to hash password")
				return nil, fmt.Errorf("failed to hash password")
			}
			payload.Password = &passwordHash
			changed = true
		}

	}

	if !changed {
		code := "NO_FIELDS_UPDATED"
		return nil, errs.NewBadRequestError("no fields updated", false, &code, nil, nil)
	}

	updatedUser, err := s.userRepo.UpdateUser(ctx, userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create user")
		return nil, err
	}

	// Business event logs
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "user_updated").
		Str("user_id", updatedUser.ID.String()).
		Str("first_name", updatedUser.LastName).
		Str("last_name", updatedUser.LastName).
		Str("email", updatedUser.Email)

	return updatedUser, nil
}

func (s *UserService) DeleteUser(ctx *gin.Context, payload *user.DeleteUserPayload) error {
	err := s.userRepo.DeleteUser(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Business event logs
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "user_deleted").
		Str("user_id", payload.ID).
		Msg("User deleted successfully")

	return nil
}
