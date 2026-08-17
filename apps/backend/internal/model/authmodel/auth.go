package authmodel

import (
	"github.com/go-playground/validator/v10"
	"github.com/xanity-07/spndex/internal/model/user"
)

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (a *LoginPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}

type LoginResponsePayload struct {
	User  *user.User `json:"user"`
	Token string     `json:"token"`
}
