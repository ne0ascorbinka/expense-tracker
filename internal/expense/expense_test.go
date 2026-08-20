package expense_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ne0ascorbinka/expense-tracker/internal/expense"
)

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		wantErr error
	}{
		{"valid description", "Lunch with team", nil},
		{"empty description", "", expense.ErrEmptyDescription},
		{"spaces only description", "   \t\n ", expense.ErrEmptyDescription},
		{"max length 255 chars", strings.Repeat("a", 255), nil},
		{"too long description 256 chars", strings.Repeat("a", 256), expense.ErrDescriptionTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := expense.ValidateDescription(tt.desc)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  int64
		wantErr error
	}{
		{"positive amount", 100, nil},
		{"zero amount", 0, expense.ErrAmountMustBePositive},
		{"negative amount", -50, expense.ErrAmountMustBePositive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := expense.ValidateAmount(tt.amount)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr error
	}{
		{"valid date", "2026-08-20", nil},
		{"invalid format short", "2026-8-20", expense.ErrInvalidDateFormat},
		{"invalid format slash", "2026/08/20", expense.ErrInvalidDateFormat},
		{"invalid calendar date", "2024-02-30", expense.ErrInvalidDateFormat},
		{"invalid calendar month", "2024-13-01", expense.ErrInvalidDateFormat},
		{"random string", "invalid-date", expense.ErrInvalidDateFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := expense.ValidateDate(tt.date)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateMonthAndYear(t *testing.T) {
	for m := 1; m <= 12; m++ {
		if err := expense.ValidateMonth(m); err != nil {
			t.Fatalf("expected month %d to be valid, got %v", m, err)
		}
	}
	if err := expense.ValidateMonth(0); err != expense.ErrInvalidMonth {
		t.Fatalf("expected month 0 to be invalid, got %v", err)
	}
	if err := expense.ValidateMonth(13); err != expense.ErrInvalidMonth {
		t.Fatalf("expected month 13 to be invalid, got %v", err)
	}

	if err := expense.ValidateYear(2026); err != nil {
		t.Fatalf("expected year 2026 to be valid, got %v", err)
	}
	if err := expense.ValidateYear(0); err != expense.ErrInvalidYear {
		t.Fatalf("expected year 0 to be invalid, got %v", err)
	}
	if err := expense.ValidateYear(-1); err != expense.ErrInvalidYear {
		t.Fatalf("expected year -1 to be invalid, got %v", err)
	}
}

func TestAdd(t *testing.T) {
	expenses := []expense.Expense{}
	nextID := int64(1)

	// Add first expense
	e1, err := expense.Add(&expenses, &nextID, "Lunch", 2000, "Food", "2024-08-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e1.ID != 1 || e1.Description != "Lunch" || e1.Amount != 2000 || e1.Category != "Food" || e1.Date != "2024-08-06" {
		t.Fatalf("unexpected expense fields: %+v", e1)
	}
	if nextID != 2 {
		t.Fatalf("expected nextID to be 2, got %d", nextID)
	}

	// Add second expense with default category and default date
	today := time.Now().Format("2006-01-02")
	e2, err := expense.Add(&expenses, &nextID, "Coffee", 450, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e2.ID != 2 || e2.Category != "uncategorized" || e2.Date != today {
		t.Fatalf("unexpected expense fields: %+v", e2)
	}
	if nextID != 3 {
		t.Fatalf("expected nextID to be 3, got %d", nextID)
	}
	if len(expenses) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(expenses))
	}

	// Validation failures
	_, err = expense.Add(&expenses, &nextID, "", 100, "", "")
	if err != expense.ErrEmptyDescription {
		t.Fatalf("expected empty description error, got %v", err)
	}
	_, err = expense.Add(&expenses, &nextID, "Test", 0, "", "")
	if err != expense.ErrAmountMustBePositive {
		t.Fatalf("expected amount error, got %v", err)
	}
	_, err = expense.Add(&expenses, &nextID, "Test", 100, "", "bad-date")
	if err != expense.ErrInvalidDateFormat {
		t.Fatalf("expected date format error, got %v", err)
	}
}

func TestUpdate(t *testing.T) {
	expenses := []expense.Expense{
		{ID: 1, Date: "2024-08-06", Description: "Lunch", Amount: 2000, Category: "Food"},
	}

	// Error when no fields provided
	_, err := expense.Update(expenses, 1, expense.UpdateParams{})
	if err != expense.ErrNoUpdateFieldsProvided {
		t.Fatalf("expected ErrNoUpdateFieldsProvided, got %v", err)
	}

	// Error when ID not found
	newDesc := "Dinner"
	_, err = expense.Update(expenses, 999, expense.UpdateParams{Description: &newDesc})
	if err == nil || !strings.Contains(err.Error(), "expense with ID 999 not found") {
		t.Fatalf("expected ID not found error, got %v", err)
	}

	// Validation errors on update
	emptyDesc := ""
	_, err = expense.Update(expenses, 1, expense.UpdateParams{Description: &emptyDesc})
	if err != expense.ErrEmptyDescription {
		t.Fatalf("expected ErrEmptyDescription, got %v", err)
	}
	badAmount := int64(0)
	_, err = expense.Update(expenses, 1, expense.UpdateParams{Amount: &badAmount})
	if err != expense.ErrAmountMustBePositive {
		t.Fatalf("expected ErrAmountMustBePositive, got %v", err)
	}
	badDate := "2024-15-01"
	_, err = expense.Update(expenses, 1, expense.UpdateParams{Date: &badDate})
	if err != expense.ErrInvalidDateFormat {
		t.Fatalf("expected ErrInvalidDateFormat, got %v", err)
	}

	// Successful partial update
	newAmount := int64(2500)
	newCategory := "Dining"
	newDate := "2024-08-07"
	updated, err := expense.Update(expenses, 1, expense.UpdateParams{
		Description: &newDesc,
		Amount:      &newAmount,
		Category:    &newCategory,
		Date:        &newDate,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Description != "Dinner" || updated.Amount != 2500 || updated.Category != "Dining" || updated.Date != "2024-08-07" {
		t.Fatalf("unexpected updated fields: %+v", updated)
	}

	// Category updated to empty string defaults to uncategorized
	emptyCat := ""
	updated, err = expense.Update(expenses, 1, expense.UpdateParams{Category: &emptyCat})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Category != "uncategorized" {
		t.Fatalf("expected uncategorized, got %s", updated.Category)
	}
}

func TestDelete(t *testing.T) {
	expenses := []expense.Expense{
		{ID: 1, Description: "Item 1", Amount: 100},
		{ID: 2, Description: "Item 2", Amount: 200},
		{ID: 3, Description: "Item 3", Amount: 300},
	}

	err := expense.Delete(&expenses, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expenses) != 2 {
		t.Fatalf("expected 2 expenses left, got %d", len(expenses))
	}
	if expenses[0].ID != 1 || expenses[1].ID != 3 {
		t.Fatalf("unexpected remaining expenses: %+v", expenses)
	}

	// Delete non-existent ID
	err = expense.Delete(&expenses, 999)
	if err == nil || !strings.Contains(err.Error(), "expense with ID 999 not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestFilterAndSort(t *testing.T) {
	expenses := []expense.Expense{
		{ID: 3, Date: "2026-08-15", Description: "Grocery", Amount: 5000, Category: "FOOD"},
		{ID: 1, Date: "2026-08-01", Description: "Lunch", Amount: 2000, Category: "food"},
		{ID: 2, Date: "2026-08-01", Description: "Coffee", Amount: 400, Category: "Drinks"},
		{ID: 4, Date: "2026-07-20", Description: "Taxi", Amount: 1500, Category: "Transport"},
		{ID: 5, Date: "2025-08-10", Description: "Old Lunch", Amount: 1200, Category: "Food"},
	}

	// Case-insensitive category filter
	foodList := expense.Filter(expenses, expense.FilterOption{Category: "food"})
	if len(foodList) != 3 {
		t.Fatalf("expected 3 food expenses, got %d", len(foodList))
	}
	// Verify sorting order: 2025-08-10 (ID 5), 2026-08-01 (ID 1), 2026-08-15 (ID 3)
	if foodList[0].ID != 5 || foodList[1].ID != 1 || foodList[2].ID != 3 {
		t.Fatalf("unexpected sorting order: %+v", foodList)
	}

	// Filter by year
	y2026List := expense.Filter(expenses, expense.FilterOption{Year: 2026})
	if len(y2026List) != 4 {
		t.Fatalf("expected 4 expenses in 2026, got %d", len(y2026List))
	}
	// Check sorting: 2026-07-20 (ID 4), 2026-08-01 (ID 1), 2026-08-01 (ID 2), 2026-08-15 (ID 3)
	if y2026List[0].ID != 4 || y2026List[1].ID != 1 || y2026List[2].ID != 2 || y2026List[3].ID != 3 {
		t.Fatalf("unexpected sorting order for 2026: %+v", y2026List)
	}

	// Filter by month & year
	aug2026List := expense.Filter(expenses, expense.FilterOption{Year: 2026, Month: 8})
	if len(aug2026List) != 3 {
		t.Fatalf("expected 3 expenses for August 2026, got %d", len(aug2026List))
	}

	// Filter by month, year, and category
	aug2026Food := expense.Filter(expenses, expense.FilterOption{Year: 2026, Month: 8, Category: "Food"})
	if len(aug2026Food) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(aug2026Food))
	}
	if aug2026Food[0].ID != 1 || aug2026Food[1].ID != 3 {
		t.Fatalf("unexpected expenses: %+v", aug2026Food)
	}

	// Filter with no match
	noMatch := expense.Filter(expenses, expense.FilterOption{Category: "NonExistent"})
	if len(noMatch) != 0 {
		t.Fatalf("expected 0 expenses, got %d", len(noMatch))
	}
}

func TestTotal(t *testing.T) {
	expenses := []expense.Expense{
		{Amount: 1000},
		{Amount: 2500},
		{Amount: 50},
	}
	total := expense.Total(expenses)
	if total != 3550 {
		t.Fatalf("expected total 3550, got %d", total)
	}

	emptyTotal := expense.Total([]expense.Expense{})
	if emptyTotal != 0 {
		t.Fatalf("expected total 0 for empty list, got %d", emptyTotal)
	}
}
