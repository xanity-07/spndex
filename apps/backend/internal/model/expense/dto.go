package expense

import (
	"github.com/go-playground/validator/v10"
	"github.com/xanity-07/spndex/internal/enums"
)

type CreateExpensePayload struct {
	Date         string                `json:"date" validate:"required" time_format:"2006-01-02"`
	Description  *string               `json:"description" validate:"omitempty,min=1,max=255"`
	Category     enums.ExpenseCategory `json:"category" validate:"required,oneof=food transport utilities entertainment healthcare shopping education other"`
	CurrencyCode enums.CurrencyCode    `json:"currencyCode" validate:"required,len=3,uppercase,oneof=USD EUR GBP CAD AUD"`
	Amount       float64               `json:"amount" validate:"required,gt=0"`
}

func (p *CreateExpensePayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type GetExpensesQuery struct {
	Page         *int                   `form:"page" validate:"omitempty,min=1"`
	Limit        *int                   `form:"limit" validate:"omitempty,min=5,max=20"`
	Sort         *string                `form:"sort" validate:"omitempty,oneof=asc desc"`
	Order        *string                `form:"order" validate:"omitempty,oneof=amount created_at category date"`
	Search       *string                `form:"search" validate:"omitempty,min=1"`
	Category     *enums.ExpenseCategory `form:"category" validate:"omitempty,oneof=food transport utilities entertainment healthcare shopping education other"`
	CurrencyCode *enums.CurrencyCode    `form:"currencyCode" validate:"omitempty,len=3,uppercase,oneof=USD EUR GBP CAD AUD"`
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
	ID string `uri:"id" validate:"required,uuid"`
}

func (p *GetExpenseByID) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type UpdateExpense struct {
	Date         *string                `json:"date" validate:"omitempty"`
	Description  *string                `json:"description" validate:"omitempty,min=1,max=255"`
	Category     *enums.ExpenseCategory `json:"category" validate:"omitempty,oneof=food transport utilities entertainment healthcare shopping education other"`
	CurrencyCode *enums.CurrencyCode    `json:"currencyCode" validate:"omitempty,len=3,uppercase,oneof=USD EUR GBP CAD AUD"`
	Amount       *int64                 `json:"amount" validate:"omitempty,gt=0"`
}

func (p *UpdateExpense) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type DeleteExpense struct {
	ID string `uri:"id" validate:"required,uuid"`
}

func (p *DeleteExpense) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
