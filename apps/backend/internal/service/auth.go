package service

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xanity-07/spndex/internal/auth"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/model/authmodel"
	"github.com/xanity-07/spndex/internal/model/session"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/repositories"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/validation"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	server      *server.Server
	userRepo    *repositories.UserRepository
	sessionRepo *repositories.SessionRepository
}

func NewAuthService(s *server.Server, userRepo *repositories.UserRepository, sessionRepo *repositories.SessionRepository) *AuthService {
	return &AuthService{
		server:      s,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (auth *AuthService) Register(ctx *gin.Context, payload *user.CreateUserPayload) (*user.User, error) {
	logger := middleware.GetLogger(ctx)

	exists, err := auth.userRepo.CheckUserExists(ctx, payload.Email)
	if err != nil {
		logger.Error().Err(err).Msg("failed to check if user exists")
		return nil, errs.NewBadRequestError("failed to check if user exists", false, nil, nil, nil)
	}

	if exists {
		logger.Warn().Msg("user with this email already exists")
		code := "USER_WITH_EMAIL_EXISTS"
		return nil, errs.NewBadRequestError("user with this email already exists", false, &code, nil, nil)
	}

	if err = validation.ValidateName(payload.FirstName); err != nil {
		logger.Error().Err(err).Msg("first name validation failed")
		return nil, errs.NewBadRequestError(err.Error(), false, nil, nil, nil)
	}

	if err = validation.ValidateName(payload.LastName); err != nil {
		logger.Error().Err(err).Msg("last name validation failed")
		return nil, errs.NewBadRequestError(err.Error(), false, nil, nil, nil)
	}

	if err = validation.ValidatePasswordStrength(payload.Password); err != nil {
		logger.Error().Err(err).Msg("invalid password")
		return nil, errs.NewBadRequestError(err.Error(), false, nil, nil, nil)
	}

	passwordHash, err := validation.HashPassword(payload.Password)
	if err != nil {
		logger.Error().Err(err).Msg("failed to hash password")
		return nil, fmt.Errorf("failed to hash password")
	}

	payload.Password = passwordHash

	createdUser, err := auth.userRepo.CreateUser(ctx, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create user")
		return nil, err
	}

	logger.Info().
		Str("event", "user_registered").
		Str("user_id", createdUser.ID.String()).
		Msg("user registered successfully")

	return createdUser, nil
}

func (a *AuthService) Login(ctx *gin.Context, payload *authmodel.LoginPayload) (*authmodel.LoginResponsePayload, error) {
	logger := middleware.GetLogger(ctx)

	foundUser, err := a.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		logger.Warn().Err(err).Msg("login failed: user lookup failed")
		return nil, errs.NewUnauthorizedError("invalid email or password", false)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(payload.Password)); err != nil {
		logger.Warn().Err(err).Msg("login failed: password missmatch")
		return nil, errs.NewUnauthorizedError("invalid email or password", false)
	}

	sessionID := uuid.NewString()
	ttl := time.Duration(a.server.Config.Auth.TTLHours) * time.Hour

	session := &session.Session{
		ID:        sessionID,
		UserID:    foundUser.ID.String(),
		CreatedAt: time.Now(),
	}

	if err = a.sessionRepo.Create(ctx, session, ttl); err != nil {
		logger.Error().Err(err).Msg("failed to create session")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	token, err := auth.GenerateToken(foundUser.ID, sessionID, foundUser.Role, []byte(a.server.Config.Auth.SecretKey), ttl)
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate token")
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "user_login").
		Str("user_id", foundUser.ID.String()).
		Msg("User logged in successfully")

	resp := &authmodel.LoginResponsePayload{
		User:  foundUser,
		Token: token,
	}

	return resp, nil
}

func (a *AuthService) Logout(ctx *gin.Context) error {
	logger := middleware.GetLogger(ctx)

	sessionID := middleware.GetSessionID(ctx)
	if sessionID == "" {
		logger.Warn().Msg("logout called without a session id in context")
		return errs.NewUnauthorizedError("not authenticated", false)
	}

	if err := a.sessionRepo.Delete(ctx, sessionID); err != nil {
		logger.Error().Err(err).Msg("failed to delete session")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	logger.Info().
		Str("event", "user_logged_out").
		Str("session_id", sessionID).
		Msg("User logged out successfully")

	return nil
}
