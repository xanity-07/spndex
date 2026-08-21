package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/errs"
	"golang.org/x/crypto/bcrypt"
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

func BindAndValidate(c *gin.Context, payload Validatable, source enums.BindingSource) error {
	var err error

	// // Clasic console.log debugging dont delete LOL
	// fmt.Printf("binding source: %#v\n", source)
	// fmt.Printf("payload type: %T\n", payload)
	// fmt.Printf("query params: %#v\n", c.Request.URL.Query())
	// fmt.Printf("payload before binding: %+v\n", payload)
	// fmt.Printf("uri id: %q\n", c.Param("id"))

	switch source {
	case enums.BindingJSON:
		err = c.ShouldBind(payload)
	case enums.BindingQuery:
		err = c.ShouldBindQuery(payload)
	case enums.BindingURI:
		err = c.ShouldBindUri(payload)
	default:
		return errs.NewBadRequestError("invalid binding source", true, nil, nil, nil)
	}

	if err != nil {
		return errs.NewBadRequestError(fmt.Sprintf("binding failed: %s", err.Error()), false, nil, nil, nil)
	}

	if msg, fieldErrors := validateStruct(payload); fieldErrors != nil {
		code := "VALIDATION_FAILED"
		return errs.NewBadRequestError(msg, true, &code, fieldErrors, nil)
	}
	return nil
}

func BindAndValidateQuery(c *gin.Context, query Validatable) error {
	if err := c.ShouldBind(query); err != nil {
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

	// Check if the error is from validator package
	validationErrors, ok := err.(validator.ValidationErrors)
	if ok {
		// Check for validator struct tags errors
		for _, validationErr := range validationErrors {
			field := strings.ToLower(validationErr.Field())
			var msg string

			switch validationErr.Tag() {
			case "required":
				msg = "is required"

			case "min":
				if validationErr.Type().Kind() == reflect.String {
					msg = fmt.Sprintf("must be at least %s characters", validationErr.Param())
				} else {
					msg = fmt.Sprintf("must be at least %s", validationErr.Param())
				}

			case "max":
				if validationErr.Type().Kind() == reflect.String {
					msg = fmt.Sprintf("must not exceed %s characters", validationErr.Param())
				} else {
					msg = fmt.Sprintf("must not exceed %s", validationErr.Param())
				}

			case "oneof":
				msg = fmt.Sprintf("must be one of: %s", validationErr.Param())

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
				if validationErr.Param() != "" {
					msg = fmt.Sprintf("%s: %s:%s", field, validationErr.Tag(), validationErr.Param())
				} else {
					msg = fmt.Sprintf("%s: %s", field, validationErr.Tag())
				}
			}

			fieldErrors = append(fieldErrors, errs.FieldError{
				Field: field,
				Error: msg,
			})
		}

		return "Validation failed", fieldErrors
	}

	// Check if the type is CustomValidationErrors
	// loop through all the validation errors and append it into the field error slice
	customValidationErrors, ok := err.(CustomValidationErrors)
	if ok {
		for _, validationErr := range customValidationErrors {
			fieldErrors = append(fieldErrors, errs.FieldError{
				Field: validationErr.Field,
				Error: validationErr.Message,
			})
		}

		return "Validation failed", fieldErrors
	}

	return "Validation failed", []errs.FieldError{
		{
			Error: err.Error(),
		},
	}
}

func ValidateName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return errors.New("name cannot be empty")
	}

	if len(name) < 2 {
		return errors.New("name must be at least 2 characters")
	}

	if len(name) > 100 {
		return errors.New("name cannot exceed 100 characters")
	}

	for _, c := range name {
		if unicode.IsLetter(c) {
			continue
		}
		if c == '\'' {
			continue
		}
		return errors.New("name can only contain letters, hyphens, and apostrophes")
	}
	return nil
}

func ValidatePasswordStrength(password string) error {
	password = strings.TrimSpace(password)

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var (
		hasUpper  bool
		hasLower  bool
		hasSymbol bool
		hasNumber bool
	)

	for _, c := range password {
		switch {
		case unicode.IsDigit(c):
			hasNumber = true
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSymbol = true

		}
	}

	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if !hasSymbol {
		return errors.New("password must contain at least one symbol")
	}

	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	return nil
}

func HashPassword(password string) (string, error) {
	passwordBytes := []byte(password)

	hashedBytes, err := bcrypt.GenerateFromPassword(passwordBytes, 12)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

func ComparePassword(current string, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(current), []byte(password))
	if err == nil {
		return true, nil
	}

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	return false, nil
}
