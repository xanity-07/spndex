package middleware

import (
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/auth"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/repositories"
)

type ContextKey string

func RequireAuth(sessionRepo *repositories.SessionRepository, jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := GetLogger(c)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn().Msg("missing authorization header")
			errs.WriteHTTPError(c, errs.NewUnauthorizedError("missing authorization header", false))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn().Msg("malformed authorization header")
			errs.WriteHTTPError(c, errs.NewUnauthorizedError("malformed authorization header", false))
			return
		}

		claims, err := auth.ParseToken(parts[1], jwtSecret)
		if err != nil {
			logger.Warn().Err(err).Msg("invalid token")
			errs.WriteHTTPError(c, errs.NewUnauthorizedError("invalid or expired token", false))
			return
		}

		if _, err := sessionRepo.Get(c, claims.SessionID); err != nil {
			logger.Warn().Err(err).Msg("session not found or expired")
			errs.WriteHTTPError(c, errs.NewUnauthorizedError("session not found or expired", false))
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(SessionIDKey, claims.SessionID)
		c.Set(CTXKeyRole, claims.Role)

		c.Next()
	}
}

func RequireRole(allowed ...enums.UserRoles) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := GetLogger(c)

		value, ok := c.Get(CTXKeyRole)
		if !ok {
			errs.WriteHTTPError(c, errs.NewForbiddenError("forbidden", false))
			return
		}

		role, ok := value.(enums.UserRoles)
		if !ok {
			errs.WriteHTTPError(c, errs.NewForbiddenError("forbidden", false))
			return
		}

		if slices.Contains(allowed, role) {
			c.Next()
			return
		}

		logger.Warn().Str("role", string(role)).Msg("insufficient permissions")
		errs.WriteHTTPError(c, errs.NewForbiddenError("insufficient permissions", false))
	}
}
