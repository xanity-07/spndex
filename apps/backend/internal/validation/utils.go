package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/xanity-07/spndex/internal/errs"
)

// Whatever data type implements this Validate() error method will satisfy this interface
// this will be how we handle validation for all DTOs
type Validatable interface {
	Validate() error
}

type CustomValidationError struct {
	Field   string
	Message string
}

type CustomValidationErrors []CustomValidationError

// This method satisfies the Error interface and now CustomValidationErrors can be of error type
func (c CustomValidationErrors) Error() string {
	return "Validation failed"
}

func BindAndValidate(c *gin.Context, payload Validatable) error {
	if err := c.Bind(payload); err != nil {
		message := strings.Split(strings.Split(err.Error(), ",")[1], "message=")[1]
		return errs.NewBadRequestError(message, true, nil, nil, nil)
	}

	// After we have a native data type of Go we take this data type and
	// validate it against a set of rules
	if msg, fieldErrors := validateStruct(payload); fieldErrors != nil {
		return errs.NewBadRequestError(msg, true, nil, fieldErrors, nil)
	}
	return nil
}

func BindAndValidateQuery(c *gin.Context, query Validatable) error {
	if err := c.Bind(query); err != nil {
		return errs.NewBadRequestError("Invalid query parameters", false, nil, nil, nil)
	}

	if msg, fieldErrors := validateStruct(query); fieldErrors != nil {
		return errs.NewBadRequestError(msg, true, nil, fieldErrors, nil)
	}
	return nil
}

// validateStruct takes a data type that implements Validatable then we call Validate on top of that data type
// if we get any errors we call extractValidationErrors
func validateStruct(v Validatable) (string, []errs.FieldError) {
	if err := v.Validate(); err != nil {
		return extractValidationErrors(err)
	}
	return "", nil
}

// extractValidationErrors properly extracts the errors received from validateStruct
// so that we can send them to the client
func extractValidationErrors(err error) (string, []errs.FieldError) {
	var fieldErrors []errs.FieldError
	// Check if the error is from validator package if not its a CustumValidationErrors
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// Check if the type is CustomValidationErrors
		// loop through all the validation errors and append it into the field error slice
		customValidationErrors := err.(CustomValidationErrors)
		for _, err := range customValidationErrors {
			fieldErrors = append(fieldErrors,
				errs.FieldError{
					Field: err.Field,
					Error: err.Message,
				},
			)
		}
	}

	// Check for validator struct tags errors
	for _, err := range validationErrors {
		field := strings.ToLower(err.Field())
		var msg string

		switch err.Tag() {
		case "required":
			msg = "is required"
		case "min":
			if err.Type().Kind() == reflect.String {
				msg = fmt.Sprintf("must be at least %s characters", err.Param())
			} else {
				msg = fmt.Sprintf("must be at least %s", err.Param())
			}
		case "max":
			if err.Type().Kind() == reflect.String {
				msg = fmt.Sprintf("must not exceed %s characters", err.Param())
			} else {
				msg = fmt.Sprintf("must not exceed %s", err.Param())
			}
		case "oneof":
			msg = fmt.Sprintf("must be one of: %s", err.Param())
		case "email":
			msg = "must be a valid email address"
		case "e164":
			msg = "must be a valid phone number with country code"
		case "uuid":
			msg = "must be a valid UUID"
		case "uuidList":
			msg = "must be a comma-separated list of valid UUIDs"
		case "dive":
			msg = "some items are invalid"
		default:
			if err.Param() != "" {
				msg = fmt.Sprintf("%s: %s:%s", field, err.Tag(), err.Param())
			} else {
				msg = fmt.Sprintf("%s: %s", field, err.Tag())
			}
		}

		fieldErrors = append(fieldErrors, errs.FieldError{
			Field: strings.ToLower(err.Field()),
			Error: msg,
		})
	}

	return "Validation failed", fieldErrors

}
