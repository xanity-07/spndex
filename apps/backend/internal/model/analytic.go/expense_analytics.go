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
	TotalCents int64   `json:"-" db:"total"`
	Total      float64 `json:"total" db:"-"`
	Count      int64   `json:"count" db:"count"`
}

type MonthlyTotals struct {
	Month string  `json:"month" db:"month"`
	Total float64 `json:"total" db:"total"`
	Count int64   `json:"count" db:"count"`
}

type DashboardStats struct {
	HighestExpense       expense.Expense `json:"highestExpense"`
	LowestExpense        expense.Expense `json:"lowestExpense"`
	TotalExpenses        int64           `json:"totalExpenses"`
	ExpenseCount         int             `json:"expenseCount"`
	AverageExpenseAmount int64           `json:"averageExpenseAmount"`
	CurrentMonthTotal    int64           `json:"currentMonthTotal"`
	LastMonthTotal       int64           `json:"lastMonthTotal"`
	MonthlyNetChange     int64           `json:"monthlyNetChange"`
}
