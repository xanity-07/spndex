package errs

import "net/http"

func NewUnauthorizedError(message string, override bool) *AppError {
	return &AppError{
		Code:     MakeUpperCaseWithUnderscores(http.StatusText(http.StatusUnauthorized)),
		Message:  message,
		Status:   http.StatusUnauthorized,
		Override: override,
	}
}

func NewBadRequestError(message string, override bool, code *string, errors []FieldError, action *Action) *AppError {
	formattedCode := MakeUpperCaseWithUnderscores(http.StatusText(http.StatusBadRequest))

	if code != nil {
		formattedCode = *code
	}

	return &AppError{
		Action:   action,
		Code:     formattedCode,
		Message:  message,
		Errors:   errors,
		Status:   http.StatusBadRequest,
		Override: override,
	}
}

func NewNotFoundError(message string, override bool, code *string) *AppError {
	formattedCode := MakeUpperCaseWithUnderscores(http.StatusText(http.StatusNotFound))

	if code != nil {
		formattedCode = *code
	}

	return &AppError{
		Code:     formattedCode,
		Message:  message,
		Status:   http.StatusNotFound,
		Override: override,
	}
}

func NewForbiddenError(message string, override bool) *AppError {
	return &AppError{
		Code:     MakeUpperCaseWithUnderscores(http.StatusText(http.StatusForbidden)),
		Message:  message,
		Status:   http.StatusForbidden,
		Override: override,
	}
}

func NewInternalServerError() *AppError {
	return &AppError{
		Code:     MakeUpperCaseWithUnderscores(http.StatusText(http.StatusInternalServerError)),
		Message:  http.StatusText(http.StatusInternalServerError),
		Status:   http.StatusInternalServerError,
		Override: false,
	}
}

func ValidateError(err error) *AppError {
	return NewBadRequestError("Validation failed"+err.Error(), false, nil, nil, nil)
}
