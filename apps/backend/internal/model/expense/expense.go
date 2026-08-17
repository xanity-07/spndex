package expense

import (
	"time"

	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/model"
)

type Expense struct {
	Date        time.Time             `json:"date" db:"date"`
	Description *string               `json:"description" db:"description"`
	UserID      string                `json:"userId" db:"user_id"`
	Category    enums.ExpenseCategory `json:"category" db:"category"`
	model.Base
	Amount float64 `json:"amount" db:"amount"`
}
