package user

import (
	"github.com/go-playground/validator/v10"
)

type CreateUserPayload struct {
	FirstName string `json:"firstName" validate:"required,min=2,max=100"`
	LastName  string `json:"lastName" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

func (u *CreateUserPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type GetUsersQuery struct {
	Page   *int    `form:"page" validate:"omitempty,min=1"`
	Limit  *int    `form:"limit" validate:"omitempty,min=5,max=20"`
	Sort   *string `form:"sort" validate:"omitempty,oneof=asc desc"`
	Order  *string `form:"order" validate:"omitempty,oneof=created_at first_name last_name"`
	Search *string `form:"search" validate:"omitempty,min=1"`
}

func (q *GetUsersQuery) Validate() error {
	validate := validator.New()
	if err := validate.Struct(q); err != nil {
		return err
	}

	if q.Page == nil {
		defaultPage := 1
		q.Page = &defaultPage
	}

	if q.Limit == nil {
		defaultLimit := 5
		q.Limit = &defaultLimit
	}

	return nil
}

type GetUserByIDPayload struct {
	ID string `uri:"id" validate:"required,uuid"`
}

func (p *GetUserByIDPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type UpdateUserPayload struct {
	FirstName *string `json:"firstName" validate:"omitempty,min=2,max=100"`
	LastName  *string `json:"lastName" validate:"omitempty,min=2,max=100"`
	Email     *string `json:"email" validate:"omitempty,email"`
	Password  *string `json:"password" validate:"omitempty,min=8"`
}

func (p *UpdateUserPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type DeleteUserPayload struct {
	ID string `uri:"id" validate:"required,uuid"`
}

func (p *DeleteUserPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
