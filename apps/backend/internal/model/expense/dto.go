package expense

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/xanity-07/spndex/internal/enums"
)

type CreateExpensePayload struct {
	Date        time.Time             `json:"date" validate:"required,datetime=2006-01-02"`
	Description *string               `json:"description" validate:"omitempty,min=1,max=255"`
	Category    enums.ExpenseCategory `json:"category" validate:"required,oneof=food transport utilities entertainment healthcare shopping education other"`
	Amount      float64               `json:"amount" validate:"required,gte=0.01"`
	UserID      uuid.UUID             `json:"user_id" validate:"required,uuid"`
}

func (p *CreateExpensePayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type GetExpensesQuery struct {
	Page     *int                   `form:"page" validate:"omitempty,min=1"`
	Limit    *int                   `form:"limit" validate:"omitempty, min=5,max=20"`
	Sort     *string                `form:"sort" validate:"omitempty,oneof=asc desc"`
	Order    *string                `form:"order" validate:"omitempty,oneof=amount created_at category date"`
	Search   *string                `form:"search" validate:"omitempty,min=1"`
	Category *enums.ExpenseCategory `form:"category" validate:"omitempty,oneof=food transport utilities entertainment healthcare shopping education other"`
}

func (q *GetExpensesQuery) Validate() error {
	validate := validator.New()
	if err := validate.Struct(q); err != nil {
		return err
	}

	if q.Page == nil {
		defaultPage := 1
		q.Page = &defaultPage
	}

	if q.Limit == nil {
		defaultLimit := 10
		q.Limit = &defaultLimit
	}

	return nil
}

type GetExpenseByID struct {
	ID uuid.UUID `uri:"id" validate:"required,uuid"`
}

func (p *GetExpenseByID) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type UpdateExpense struct {
	ID uuid.UUID `uri:"id" validate:"required,uuid"`
}

func (p *UpdateExpense) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type DeleteExpense struct {
	ID uuid.UUID `uri:"id" validate:"required,uuid"`
}

func (p *DeleteExpense) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
