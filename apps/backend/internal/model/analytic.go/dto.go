package analytic

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type GetExpensesByCategoryPayload struct {
	UserID string `validate:"required,uuid"`
}

func (p *GetExpensesByCategoryPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type GetMonthlyExpensesPayload struct {
	Year *int `form:"year" validate:"omitempty,gte=1000,lte=9999"`
}

func (p *GetMonthlyExpensesPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type GetDashboardQuery struct {
	Year  *int `form:"year" validate:"omitempty,min=2000"`
	Month *int `form:"month" validate:"omitempty,min=1,max=12"`
	Range *int `form:"range" validate:"omitempty,min=1,max=12"`
}

func (q *GetDashboardQuery) Validate() error {
	validate := validator.New()
	if err := validate.Struct(q); err != nil {
		return err
	}

	if q.Year == nil {
		defaultYear := time.Now().Year()
		q.Year = &defaultYear
	}

	if q.Month == nil {
		defaultMonth := int(time.Now().Month())
		q.Month = &defaultMonth
	}

	if q.Range == nil {
		defaultRange := 1
		q.Range = &defaultRange
	}

	return nil
}
