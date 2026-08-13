package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/errs"
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

func (s *UserService) CreateUser(ctx *gin.Context, payload *user.CreateUserPayload) (*user.User, error) {
	_, err := s.userRepo.CheckUserExists(ctx, payload.Email)
	if err == nil {
		code := "EMAIL_ALREADY_EXISTS"
		return nil, errs.NewBadRequestError("email already exists ", false, &code, nil, nil)
	}

	if err = validation.ValidateName(payload.FirstName); err != nil {
		code := "INVALID_FIRST_NAME"
		fieldError := []errs.FieldError{
			{
				Field: "first name",
				Error: err.Error(),
			},
		}
		return nil, errs.NewBadRequestError("invalid first name", false, &code, fieldError, nil)
	}

	if err = validation.ValidateName(payload.LastName); err != nil {
		code := "INVALID_LAST_NAME"
		fieldError := []errs.FieldError{
			{
				Field: "last name",
				Error: err.Error(),
			},
		}
		return nil, errs.NewBadRequestError("invalid last name", false, &code, fieldError, nil)
	}

	if err = validation.ValidatePasswordStrength(payload.Password); err != nil {
		code := "INVALID_PASSWORD"
		fieldError := []errs.FieldError{
			{
				Field: "password",
				Error: err.Error(),
			},
		}
		return nil, errs.NewBadRequestError("invalid password", false, &code, fieldError, nil)
	}

	passwordHash, err := validation.HashPassword(payload.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password")
	}

	payload.Password = passwordHash

	user, err := s.userRepo.CreateUser(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUsers(ctx *gin.Context, query *user.GetUsersQuery) (*model.PaginatedResponse[user.User], error) {
	userList, err := s.userRepo.GetUsers(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	return userList, nil
}
