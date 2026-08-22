package errs

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Form fields related errors
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type ActionType string

const (
	ActionTypeRedirect ActionType = "redirect"
)

type Action struct {
	Type    ActionType `json:"type"`
	Message string     `json:"message"`
	Value   string     `json:"value"`
}

type AppError struct {
	Action   *Action      `json:"action,omitempty"` // action to be taken
	Code     string       `json:"code"`
	Message  string       `json:"message"`
	Errors   []FieldError `json:"errors,omitempty"` // field level errors
	Status   int          `json:"status"`
	Override bool         `json:"override,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Is(target error) bool {
	_, ok := target.(*AppError)
	return ok
}

// WithMessage returns a AppError struct with a custum message
func (e *AppError) WithMessage(message string) *AppError {
	return &AppError{
		Action:   e.Action,
		Code:     e.Code,
		Message:  message,
		Errors:   e.Errors,
		Status:   e.Status,
		Override: e.Override,
	}
}

// MakeUpperCaseWithUnderscores returns a HTTP status with format example: "BAD_REQUEST"
func MakeUpperCaseWithUnderscores(status string) string {
	return strings.ToUpper(strings.ReplaceAll(status, " ", "_"))
}

func WriteHTTPError(c *gin.Context, err error) {
	var appErr *AppError

	if errors.As(err, &appErr) {
		c.AbortWithStatusJSON(appErr.Status, appErr)
		return
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalServerError())
}
