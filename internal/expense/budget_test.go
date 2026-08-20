package expense_test

import (
	"testing"

	"github.com/ne0ascorbinka/expense-tracker/internal/expense"
)

func TestMonthKey(t *testing.T) {
	tests := []struct {
		year     int
		month    int
		expected string
	}{
		{year: 2026, month: 8, expected: "2026-08"},
		{year: 2026, month: 12, expected: "2026-12"},
		{year: 2024, month: 1, expected: "2024-01"},
	}

	for _, tt := range tests {
		got := expense.MonthKey(tt.year, tt.month)
		if got != tt.expected {
			t.Errorf("MonthKey(%d, %d) = %q, want %q", tt.year, tt.month, got, tt.expected)
		}
	}
}

func TestSetAndClearBudget(t *testing.T) {
	budgets := make(map[string]int64)

	// Set budget
	expense.SetBudget(budgets, 2026, 8, 50000)
	if budgets["2026-08"] != 50000 {
		t.Fatalf("expected budget of 50000 cents, got %d", budgets["2026-08"])
	}

	// Update budget
	expense.SetBudget(budgets, 2026, 8, 75000)
	if budgets["2026-08"] != 75000 {
		t.Fatalf("expected updated budget of 75000 cents, got %d", budgets["2026-08"])
	}

	// Clear budget
	expense.ClearBudget(budgets, 2026, 8)
	if _, exists := budgets["2026-08"]; exists {
		t.Fatalf("expected 2026-08 budget to be cleared, but key still exists")
	}

	// Clear non-existent budget should not panic
	expense.ClearBudget(budgets, 2026, 9)
}

func TestCheckBudgetExceeded(t *testing.T) {
	budgets := map[string]int64{
		"2026-08": 50000, // $500.00
		"2026-09": 0,     // 0 budget -> treated as not set
	}

	expenses := []expense.Expense{
		{ID: 1, Date: "2026-08-01", Amount: 30000, Category: "food"},
		{ID: 2, Date: "2026-08-05", Amount: 15000, Category: "games"},
		{ID: 3, Date: "2026-09-01", Amount: 10000, Category: "bills"},
	}

	// Total for 2026-08 is 45000 cents ($450) <= 50000 cents ($500)
	budget, exceeded := expense.CheckBudgetExceeded(expenses, budgets, 2026, 8)
	if budget != 50000 || exceeded {
		t.Fatalf("expected (50000, false), got (%d, %v)", budget, exceeded)
	}

	// Add expense crossing the budget: 45000 + 10000 = 55000 cents ($550) > 50000 cents
	expenses = append(expenses, expense.Expense{ID: 4, Date: "2026-08-10", Amount: 10000, Category: "food"})
	budget, exceeded = expense.CheckBudgetExceeded(expenses, budgets, 2026, 8)
	if budget != 50000 || !exceeded {
		t.Fatalf("expected (50000, true), got (%d, %v)", budget, exceeded)
	}

	// Unset budget month (2026-07)
	budget, exceeded = expense.CheckBudgetExceeded(expenses, budgets, 2026, 7)
	if budget != 0 || exceeded {
		t.Fatalf("expected (0, false) for unset month, got (%d, %v)", budget, exceeded)
	}

	// Budget <= 0 (2026-09)
	budget, exceeded = expense.CheckBudgetExceeded(expenses, budgets, 2026, 9)
	if budget != 0 || exceeded {
		t.Fatalf("expected (0, false) for zero budget, got (%d, %v)", budget, exceeded)
	}
}
