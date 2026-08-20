package cmd_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ne0ascorbinka/expense-tracker/cmd"
	"github.com/ne0ascorbinka/expense-tracker/internal/expense"
	"github.com/ne0ascorbinka/expense-tracker/internal/store"
)

func executeCmd(file string, args ...string) (string, error) {
	rootCmd := cmd.NewRootCmd()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	fullArgs := append([]string{"--file", file}, args...)
	rootCmd.SetArgs(fullArgs)
	err := rootCmd.Execute()
	return out.String(), err
}

func TestAddCommand(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "expenses.json")

	// 1. Basic Add
	out, err := executeCmd(filePath, "add", "--description", "Lunch", "--amount", "20.50", "--category", "Food")
	if err != nil {
		t.Fatalf("unexpected error on add: %v", err)
	}
	if !strings.Contains(out, "Expense added successfully (ID: 1)") {
		t.Fatalf("unexpected output: %s", out)
	}

	// Verify store state
	s := store.New(filePath)
	if err := s.Load(); err != nil {
		t.Fatalf("failed to load store: %v", err)
	}
	if len(s.Data.Expenses) != 1 {
		t.Fatalf("expected 1 expense, got %d", len(s.Data.Expenses))
	}
	e := s.Data.Expenses[0]
	if e.ID != 1 || e.Description != "Lunch" || e.Amount != 2050 || e.Category != "Food" {
		t.Fatalf("unexpected expense content: %+v", e)
	}
	if e.Date != time.Now().Format("2006-01-02") {
		t.Fatalf("expected today's date %s, got %s", time.Now().Format("2006-01-02"), e.Date)
	}

	// 2. Add with default category and custom date
	out, err = executeCmd(filePath, "add", "--description", "Coffee", "--amount", "4.00", "--date", "2026-08-15")
	if err != nil {
		t.Fatalf("unexpected error on second add: %v", err)
	}
	if !strings.Contains(out, "Expense added successfully (ID: 2)") {
		t.Fatalf("unexpected output: %s", out)
	}

	if err := s.Load(); err != nil {
		t.Fatalf("failed to load store: %v", err)
	}
	if len(s.Data.Expenses) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(s.Data.Expenses))
	}
	if s.Data.Expenses[1].Category != "uncategorized" || s.Data.Expenses[1].Date != "2026-08-15" {
		t.Fatalf("unexpected expense content: %+v", s.Data.Expenses[1])
	}

	// 3. Validation errors
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "empty description",
			args:    []string{"add", "--description", "   ", "--amount", "10"},
			wantErr: expense.ErrEmptyDescription.Error(),
		},
		{
			name:    "description too long",
			args:    []string{"add", "--description", strings.Repeat("a", 256), "--amount", "10"},
			wantErr: expense.ErrDescriptionTooLong.Error(),
		},
		{
			name:    "negative amount",
			args:    []string{"add", "--description", "test", "--amount", "-5.00"},
			wantErr: "amount must be greater than zero",
		},
		{
			name:    "zero amount",
			args:    []string{"add", "--description", "test", "--amount", "0"},
			wantErr: "amount must be greater than zero",
		},
		{
			name:    "amount > 2 decimal places",
			args:    []string{"add", "--description", "test", "--amount", "10.999"},
			wantErr: "amount must have at most 2 decimal places",
		},
		{
			name:    "amount not a number",
			args:    []string{"add", "--description", "test", "--amount", "abc"},
			wantErr: "amount must be a valid number",
		},
		{
			name:    "invalid date format",
			args:    []string{"add", "--description", "test", "--amount", "10", "--date", "2026-8-1"},
			wantErr: expense.ErrInvalidDateFormat.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeCmd(filePath, tt.args...)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestAddBudgetWarning(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "expenses.json")

	// Set a budget of $500.00 for August 2026 in store
	s := store.New(filePath)
	s.Data.Budgets["2026-08"] = 50000 // 500.00
	if err := s.Save(); err != nil {
		t.Fatalf("failed to save store: %v", err)
	}

	// Add an expense under budget ($400.00)
	out, err := executeCmd(filePath, "add", "--description", "Rent", "--amount", "400.00", "--date", "2026-08-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "Warning: you have exceeded your budget") {
		t.Fatalf("did not expect budget warning, got: %s", out)
	}

	// Add an expense that crosses budget ($150.00 -> total $550.00)
	out, err = executeCmd(filePath, "add", "--description", "Groceries", "--amount", "150.00", "--date", "2026-08-05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedSuccess := "Expense added successfully (ID: 2)"
	expectedWarning := "Warning: you have exceeded your budget of $500.00 for August 2026."
	if !strings.Contains(out, expectedSuccess) {
		t.Fatalf("expected output to contain %q, got: %s", expectedSuccess, out)
	}
	if !strings.Contains(out, expectedWarning) {
		t.Fatalf("expected output to contain %q, got: %s", expectedWarning, out)
	}
}

func TestUpdateCommand(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "expenses.json")

	// Add initial expense
	_, err := executeCmd(filePath, "add", "--description", "Lunch", "--amount", "20.00", "--category", "food", "--date", "2026-08-01")
	if err != nil {
		t.Fatalf("failed to add expense: %v", err)
	}

	// 1. Update description & amount
	out, err := executeCmd(filePath, "update", "--id", "1", "--description", "Fancy Lunch", "--amount", "35.50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Expense updated successfully") {
		t.Fatalf("unexpected output: %s", out)
	}

	s := store.New(filePath)
	if err := s.Load(); err != nil {
		t.Fatalf("failed to load store: %v", err)
	}
	if s.Data.Expenses[0].Description != "Fancy Lunch" || s.Data.Expenses[0].Amount != 3550 || s.Data.Expenses[0].Category != "food" {
		t.Fatalf("unexpected expense state: %+v", s.Data.Expenses[0])
	}

	// 2. Update category and date
	out, err = executeCmd(filePath, "update", "--id", "1", "--category", "Dining", "--date", "2026-08-02")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Expense updated successfully") {
		t.Fatalf("unexpected output: %s", out)
	}

	if err := s.Load(); err != nil {
		t.Fatalf("failed to load store: %v", err)
	}
	if s.Data.Expenses[0].Category != "Dining" || s.Data.Expenses[0].Date != "2026-08-02" {
		t.Fatalf("unexpected expense state: %+v", s.Data.Expenses[0])
	}

	// 3. Update errors
	// No flags provided
	_, err = executeCmd(filePath, "update", "--id", "1")
	if err == nil || !strings.Contains(err.Error(), expense.ErrNoUpdateFieldsProvided.Error()) {
		t.Fatalf("expected ErrNoUpdateFieldsProvided, got: %v", err)
	}

	// ID not found
	_, err = executeCmd(filePath, "update", "--id", "99", "--description", "None")
	if err == nil || !strings.Contains(err.Error(), "expense with ID 99 not found") {
		t.Fatalf("expected not found error, got: %v", err)
	}

	// Invalid amount
	_, err = executeCmd(filePath, "update", "--id", "1", "--amount", "-10")
	if err == nil || !strings.Contains(err.Error(), "amount must be greater than zero") {
		t.Fatalf("expected positive amount error, got: %v", err)
	}
}

func TestDeleteCommand(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "expenses.json")

	// Add two expenses
	_, _ = executeCmd(filePath, "add", "--description", "Lunch", "--amount", "20.00")
	_, _ = executeCmd(filePath, "add", "--description", "Dinner", "--amount", "30.00")

	// Delete ID 1
	out, err := executeCmd(filePath, "delete", "--id", "1")
	if err != nil {
		t.Fatalf("unexpected error on delete: %v", err)
	}
	if !strings.Contains(out, "Expense deleted successfully") {
		t.Fatalf("unexpected output: %s", out)
	}

	s := store.New(filePath)
	if err := s.Load(); err != nil {
		t.Fatalf("failed to load store: %v", err)
	}
	if len(s.Data.Expenses) != 1 || s.Data.Expenses[0].ID != 2 {
		t.Fatalf("expected only ID 2 to remain, got: %+v", s.Data.Expenses)
	}
	if s.Data.NextID != 3 {
		t.Fatalf("next_id should not be modified on delete, expected 3 got %d", s.Data.NextID)
	}

	// Delete non-existent ID
	_, err = executeCmd(filePath, "delete", "--id", "99")
	if err == nil || !strings.Contains(err.Error(), "expense with ID 99 not found") {
		t.Fatalf("expected ID 99 not found error, got: %v", err)
	}
}

func TestListCommand(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "expenses.json")

	// 1. Empty list
	out, err := executeCmd(filePath, "list")
	if err != nil {
		t.Fatalf("unexpected error on empty list: %v", err)
	}
	if strings.TrimSpace(out) != "No expenses found." {
		t.Fatalf("expected 'No expenses found.', got: %q", out)
	}

	// Add expenses with different dates and categories
	_, _ = executeCmd(filePath, "add", "--description", "Lunch", "--amount", "20.00", "--category", "Food", "--date", "2024-08-06")
	_, _ = executeCmd(filePath, "add", "--description", "Dinner", "--amount", "10.00", "--date", "2024-08-06")
	_, _ = executeCmd(filePath, "add", "--description", "Book", "--amount", "15.00", "--category", "Books", "--date", "2024-09-01")
	_, _ = executeCmd(filePath, "add", "--description", "Game", "--amount", "60.00", "--category", "Gaming", "--date", "2025-01-15")

	// 2. List all
	out, err = executeCmd(filePath, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "Date") || !strings.Contains(out, "Description") || !strings.Contains(out, "Amount") || !strings.Contains(out, "Category") {
		t.Fatalf("missing table headers in output:\n%s", out)
	}
	if !strings.Contains(out, "1") || !strings.Contains(out, "2024-08-06") || !strings.Contains(out, "Lunch") || !strings.Contains(out, "$20.00") || !strings.Contains(out, "Food") {
		t.Fatalf("missing row 1 in output:\n%s", out)
	}
	if !strings.Contains(out, "2") || !strings.Contains(out, "Dinner") || !strings.Contains(out, "$10.00") || !strings.Contains(out, "uncategorized") {
		t.Fatalf("missing row 2 in output:\n%s", out)
	}

	// 3. Filter by category (case-insensitive)
	out, err = executeCmd(filePath, "list", "--category", "food")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Lunch") || strings.Contains(out, "Dinner") || strings.Contains(out, "Book") {
		t.Fatalf("unexpected filtered output for category food:\n%s", out)
	}

	// 4. Filter by year
	out, err = executeCmd(filePath, "list", "--year", "2025")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Game") || strings.Contains(out, "Lunch") {
		t.Fatalf("unexpected filtered output for year 2025:\n%s", out)
	}

	// 5. Filter by month and year
	out, err = executeCmd(filePath, "list", "--month", "8", "--year", "2024")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Lunch") || !strings.Contains(out, "Dinner") || strings.Contains(out, "Book") {
		t.Fatalf("unexpected filtered output for month 8 and year 2024:\n%s", out)
	}

	// 6. Filter by month only (defaults to current year)
	curYear := time.Now().Year()
	out, err = executeCmd(filePath, "list", "--month", "8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Unless current year is 2024, it shouldn't match 2024-08-06 items
	if curYear != 2024 {
		if strings.TrimSpace(out) != "No expenses found." {
			t.Fatalf("expected 'No expenses found.' for month 8 in %d, got: %s", curYear, out)
		}
	}

	// 7. Validation errors
	_, err = executeCmd(filePath, "list", "--month", "13")
	if err == nil || !strings.Contains(err.Error(), expense.ErrInvalidMonth.Error()) {
		t.Fatalf("expected invalid month error, got: %v", err)
	}

	_, err = executeCmd(filePath, "list", "--year", "0")
	if err == nil || !strings.Contains(err.Error(), expense.ErrInvalidYear.Error()) {
		t.Fatalf("expected invalid year error, got: %v", err)
	}

	_, err = executeCmd(filePath, "list", "--category", "")
	if err == nil || !strings.Contains(err.Error(), expense.ErrEmptyCategory.Error()) {
		t.Fatalf("expected ErrEmptyCategory, got: %v", err)
	}
}
