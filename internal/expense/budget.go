package expense

import (
	"fmt"
)

// MonthKey formats year and month into the standard budget key "YYYY-MM".
func MonthKey(year, month int) string {
	return fmt.Sprintf("%04d-%02d", year, month)
}

// CheckBudgetExceeded checks if total expenses for a specific month and year exceed the set budget.
// Returns the budget in cents and whether it was exceeded.
func CheckBudgetExceeded(expenses []Expense, budgets map[string]int64, year, month int) (budget int64, exceeded bool) {
	key := MonthKey(year, month)
	b, exists := budgets[key]
	if !exists || b <= 0 {
		return 0, false
	}

	monthExpenses := Filter(expenses, FilterOption{
		Month: month,
		Year:  year,
	})
	total := Total(monthExpenses)
	return b, total > b
}

// SetBudget sets or updates the budget in cents for a given year and month.
func SetBudget(budgets map[string]int64, year, month int, cents int64) {
	key := MonthKey(year, month)
	budgets[key] = cents
}

// ClearBudget clears the budget for a given year and month.
func ClearBudget(budgets map[string]int64, year, month int) {
	key := MonthKey(year, month)
	delete(budgets, key)
}
