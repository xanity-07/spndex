package expense

import (
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/model"
)

type Expense struct {
	Date         string                `json:"date" db:"date"`
	Description  *string               `json:"description" db:"description"`
	UserID       string                `json:"userId" db:"user_id"`
	Category     enums.ExpenseCategory `json:"category" db:"category"`
	CurrencyCode enums.CurrencyCode    `json:"currencyCode" db:"currency"`
	model.Base
	AmountCents int64 `json:"amountCents" db:"amount"`
}
