package analytic

import (
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/model/expense"
)

type CategoryTotals struct {
	Category enums.ExpenseCategory `json:"category" db:"category"`
	CategoryStats
	Percentage float64 `json:"percentage" db:"-"`
}

type CategoryStats struct {
	TotalCents int64 `json:"-" db:"total"`
	Count      int64 `json:"count" db:"count"`
}

type MonthlyTotals struct {
	Month      string `json:"month" db:"month"`
	TotalCents int64  `json:"totalCents" db:"total"`
	Count      int64  `json:"count" db:"count"`
}

type DashboardStats struct {
	HighestExpense            expense.Expense `json:"highestExpense"`
	LowestExpense             expense.Expense `json:"lowestExpense"`
	TotalExpensesCents        int64           `json:"totalExpenses"`
	ExpenseCount              int             `json:"expenseCount"`
	AverageExpenseAmountCents int64           `json:"averageExpenseAmountCents"`
	CurrentMonthTotalCents    int64           `json:"currentMonthTotalCents"`
	LastMonthTotalCents       int64           `json:"lastMonthTotalCents"`
	MonthlyNetChangeCents     int64           `json:"monthlyNetChangeCents"`
}
