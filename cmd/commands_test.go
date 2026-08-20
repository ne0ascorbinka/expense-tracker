package cmd_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ne0ascorbinka/expense-tracker/internal/expense"
	"github.com/ne0ascorbinka/expense-tracker/internal/store"
)

func TestSummaryCommand(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "expenses.json")

	// 1. Summary on empty store
	out, err := executeCmd(filePath, "summary")
	if err != nil {
		t.Fatalf("unexpected error on empty summary: %v", err)
	}
	if strings.TrimSpace(out) != "Total expenses: $0.00" {
		t.Fatalf("expected 'Total expenses: $0.00', got: %q", out)
	}

	// Add expenses across different months, years, categories
	_, _ = executeCmd(filePath, "add", "--description", "Lunch", "--amount", "20.00", "--category", "Food", "--date", "2026-08-01")
	_, _ = executeCmd(filePath, "add", "--description", "Dinner", "--amount", "10.00", "--category", "Food", "--date", "2026-08-05")
	_, _ = executeCmd(filePath, "add", "--description", "Book", "--amount", "15.00", "--category", "Books", "--date", "2026-08-10")
	_, _ = executeCmd(filePath, "add", "--description", "Game", "--amount", "60.00", "--category", "Gaming", "--date", "2026-09-01")
	_, _ = executeCmd(filePath, "add", "--description", "Old Phone", "--amount", "100.00", "--category", "Tech", "--date", "2025-05-15")

	// 2. All-time summary (no filters) -> Total: 20 + 10 + 15 + 60 + 100 = $205.00
	out, err = executeCmd(filePath, "summary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "Total expenses: $205.00" {
		t.Fatalf("expected 'Total expenses: $205.00', got: %q", out)
	}

	// 3. Year only -> Total 2026: 20 + 10 + 15 + 60 = $105.00
	out, err = executeCmd(filePath, "summary", "--year", "2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "Total expenses for 2026: $105.00" {
		t.Fatalf("expected 'Total expenses for 2026: $105.00', got: %q", out)
	}

	// 4. Month + Year -> August 2026: 20 + 10 + 15 = $45.00
	out, err = executeCmd(filePath, "summary", "--month", "8", "--year", "2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "Total expenses for August 2026: $45.00" {
		t.Fatalf("expected 'Total expenses for August 2026: $45.00', got: %q", out)
	}

	// 5. Category only -> Food: 20 + 10 = $30.00 (case-insensitive check)
	out, err = executeCmd(filePath, "summary", "--category", "food")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "Total expenses for food: $30.00" {
		t.Fatalf("expected 'Total expenses for food: $30.00', got: %q", out)
	}

	// 6. Month + Year + Category -> August 2026 (food): 20 + 10 = $30.00
	out, err = executeCmd(filePath, "summary", "--month", "8", "--year", "2026", "--category", "food")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "Total expenses for August 2026 (food): $30.00" {
		t.Fatalf("expected 'Total expenses for August 2026 (food): $30.00', got: %q", out)
	}

	// 7. Year + Category -> 2026 (Gaming): $60.00
	out, err = executeCmd(filePath, "summary", "--year", "2026", "--category", "Gaming")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "Total expenses for 2026 (Gaming): $60.00" {
		t.Fatalf("expected 'Total expenses for 2026 (Gaming): $60.00', got: %q", out)
	}

	// 8. Month without year defaults to current year
	curYear := time.Now().Year()
	out, err = executeCmd(filePath, "summary", "--month", "12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := fmt.Sprintf("Total expenses for December %d: $0.00", curYear)
	if !strings.Contains(out, expected) {
		t.Fatalf("expected output to contain %q, got: %q", expected, out)
	}

	// 9. Validation errors
	_, err = executeCmd(filePath, "summary", "--month", "0")
	if err == nil || !strings.Contains(err.Error(), expense.ErrInvalidMonth.Error()) {
		t.Fatalf("expected invalid month error, got: %v", err)
	}

	_, err = executeCmd(filePath, "summary", "--year", "-2024")
	if err == nil || !strings.Contains(err.Error(), expense.ErrInvalidYear.Error()) {
		t.Fatalf("expected invalid year error, got: %v", err)
	}

	_, err = executeCmd(filePath, "summary", "--category", "")
	if err == nil || !strings.Contains(err.Error(), expense.ErrEmptyCategory.Error()) {
		t.Fatalf("expected ErrEmptyCategory, got: %v", err)
	}
}

func TestSummaryBudgetWarning(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "expenses.json")

	// Set budget of $40.00 for August 2026
	s := store.New(filePath)
	s.Data.Budgets["2026-08"] = 4000
	if err := s.Save(); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Add expenses totaling $45.00 in August 2026 ($30 food, $15 books)
	_, _ = executeCmd(filePath, "add", "--description", "Lunch", "--amount", "30.00", "--category", "Food", "--date", "2026-08-01")
	_, _ = executeCmd(filePath, "add", "--description", "Book", "--amount", "15.00", "--category", "Books", "--date", "2026-08-05")

	// 1. summary --month 8 --year 2026 should show total and budget warning
	out, err := executeCmd(filePath, "summary", "--month", "8", "--year", "2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Total expenses for August 2026: $45.00") {
		t.Fatalf("expected total line, got: %s", out)
	}
	expectedWarn := "Warning: you have exceeded your budget of $40.00 for August 2026."
	if !strings.Contains(out, expectedWarn) {
		t.Fatalf("expected budget warning %q, got: %s", expectedWarn, out)
	}

	// 2. summary --month 8 --year 2026 --category food ($30 food < $40 budget, but overall month total $45 > $40 budget)
	out, err = executeCmd(filePath, "summary", "--month", "8", "--year", "2026", "--category", "food")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Total expenses for August 2026 (food): $30.00") {
		t.Fatalf("expected total line, got: %s", out)
	}
	if !strings.Contains(out, expectedWarn) {
		t.Fatalf("expected budget warning %q, got: %s", expectedWarn, out)
	}

	// 3. summary --year 2026 (without --month) should NOT trigger budget warning per SPEC §5
	out, err = executeCmd(filePath, "summary", "--year", "2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "Warning: you have exceeded your budget") {
		t.Fatalf("did not expect budget warning without --month flag, got: %s", out)
	}
}

func TestExportCommand(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "expenses.json")
	csvPath := filepath.Join(tempDir, "export.csv")

	// Add expenses with special chars (commas in description) and different dates
	_, _ = executeCmd(filePath, "add", "--description", "Lunch, with coffee", "--amount", "20.00", "--category", "food", "--date", "2024-08-06")
	_, _ = executeCmd(filePath, "add", "--description", "Dinner", "--amount", "10.00", "--category", "uncategorized", "--date", "2024-08-06")
	_, _ = executeCmd(filePath, "add", "--description", "Book", "--amount", "15.50", "--category", "reading", "--date", "2025-01-10")

	// 1. Export all
	out, err := executeCmd(filePath, "export", "--output", csvPath)
	if err != nil {
		t.Fatalf("unexpected error on export: %v", err)
	}
	if !strings.Contains(out, "Expenses exported to "+csvPath+" successfully") {
		t.Fatalf("unexpected stdout: %s", out)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("failed to read exported csv: %v", err)
	}

	csvStr := string(content)
	expectedLines := []string{
		"ID,Date,Description,Amount,Category",
		`1,2024-08-06,"Lunch, with coffee",20.00,food`,
		"2,2024-08-06,Dinner,10.00,uncategorized",
		"3,2025-01-10,Book,15.50,reading",
	}
	for _, line := range expectedLines {
		if !strings.Contains(csvStr, line) {
			t.Fatalf("missing line %q in CSV output:\n%s", line, csvStr)
		}
	}

	// 2. Export with filter (month 8, year 2024)
	csvFilteredPath := filepath.Join(tempDir, "export_aug.csv")
	out, err = executeCmd(filePath, "export", "--output", csvFilteredPath, "--month", "8", "--year", "2024")
	if err != nil {
		t.Fatalf("unexpected error on export filtered: %v", err)
	}

	filteredContent, err := os.ReadFile(csvFilteredPath)
	if err != nil {
		t.Fatalf("failed to read filtered csv: %v", err)
	}
	filteredStr := string(filteredContent)
	if !strings.Contains(filteredStr, "1,2024-08-06") || !strings.Contains(filteredStr, "2,2024-08-06") {
		t.Fatalf("missing august entries: %s", filteredStr)
	}
	if strings.Contains(filteredStr, "2025-01-10") {
		t.Fatalf("did not expect 2025 entry in filtered output: %s", filteredStr)
	}

	// 3. Validation errors
	_, err = executeCmd(filePath, "export")
	if err == nil {
		t.Fatalf("expected error when --output flag is missing")
	}

	_, err = executeCmd(filePath, "export", "--output", csvPath, "--month", "15")
	if err == nil || !strings.Contains(err.Error(), expense.ErrInvalidMonth.Error()) {
		t.Fatalf("expected invalid month error, got: %v", err)
	}
}

func TestBudgetCommand(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "expenses.json")

	// 1. Set budget with explicit month and year
	out, err := executeCmd(filePath, "budget", "set", "--amount", "500.00", "--month", "8", "--year", "2026")
	if err != nil {
		t.Fatalf("unexpected error on budget set: %v", err)
	}
	if !strings.Contains(out, "Budget for August 2026 set to $500.00") {
		t.Fatalf("unexpected output: %s", out)
	}

	// Verify store data
	s := store.New(filePath)
	if err := s.Load(); err != nil {
		t.Fatalf("failed to load store: %v", err)
	}
	if s.Data.Budgets["2026-08"] != 50000 {
		t.Fatalf("expected budget 50000 cents, got %d", s.Data.Budgets["2026-08"])
	}

	// 2. Set budget with default (current) month and year
	now := time.Now()
	out, err = executeCmd(filePath, "budget", "set", "--amount", "250.75")
	if err != nil {
		t.Fatalf("unexpected error on default budget set: %v", err)
	}
	expectedMsg := "Budget for " + now.Format("January 2006") + " set to $250.75"
	if !strings.Contains(out, expectedMsg) {
		t.Fatalf("expected %q, got %q", expectedMsg, out)
	}

	if err := s.Load(); err != nil {
		t.Fatalf("failed to load store: %v", err)
	}
	curKey := now.Format("2006-01")
	if s.Data.Budgets[curKey] != 25075 {
		t.Fatalf("expected budget 25075 cents, got %d", s.Data.Budgets[curKey])
	}

	// 3. Clear budget for August 2026
	out, err = executeCmd(filePath, "budget", "clear", "--month", "8", "--year", "2026")
	if err != nil {
		t.Fatalf("unexpected error on budget clear: %v", err)
	}
	if !strings.Contains(out, "Budget for August 2026 cleared") {
		t.Fatalf("unexpected output: %s", out)
	}

	if err := s.Load(); err != nil {
		t.Fatalf("failed to load store: %v", err)
	}
	if _, exists := s.Data.Budgets["2026-08"]; exists {
		t.Fatalf("expected 2026-08 budget to be cleared, still present")
	}

	// 4. Clear current month budget (default)
	out, err = executeCmd(filePath, "budget", "clear")
	if err != nil {
		t.Fatalf("unexpected error on default budget clear: %v", err)
	}
	expectedClear := "Budget for " + now.Format("January 2006") + " cleared"
	if !strings.Contains(out, expectedClear) {
		t.Fatalf("expected %q, got %q", expectedClear, out)
	}

	if err := s.Load(); err != nil {
		t.Fatalf("failed to load store: %v", err)
	}
	if _, exists := s.Data.Budgets[curKey]; exists {
		t.Fatalf("expected %s budget to be cleared, still present", curKey)
	}

	// 5. Validation errors
	// Missing --amount on set
	_, err = executeCmd(filePath, "budget", "set")
	if err == nil {
		t.Fatalf("expected error on missing --amount")
	}

	// Negative amount
	_, err = executeCmd(filePath, "budget", "set", "--amount", "-100")
	if err == nil || !strings.Contains(err.Error(), "amount must be greater than zero") {
		t.Fatalf("expected positive amount error, got: %v", err)
	}

	// > 2 decimal places
	_, err = executeCmd(filePath, "budget", "set", "--amount", "100.123")
	if err == nil || !strings.Contains(err.Error(), "amount must have at most 2 decimal places") {
		t.Fatalf("expected decimal places error, got: %v", err)
	}

	// Invalid month
	_, err = executeCmd(filePath, "budget", "set", "--amount", "100", "--month", "13")
	if err == nil || !strings.Contains(err.Error(), expense.ErrInvalidMonth.Error()) {
		t.Fatalf("expected invalid month error, got: %v", err)
	}

	// Invalid year
	_, err = executeCmd(filePath, "budget", "set", "--amount", "100", "--year", "0")
	if err == nil || !strings.Contains(err.Error(), expense.ErrInvalidYear.Error()) {
		t.Fatalf("expected invalid year error, got: %v", err)
	}
}
