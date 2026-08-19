package expense

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// Expense represents an individual expense entry.
type Expense struct {
	ID          int64  `json:"id"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
	Category    string `json:"category"`
}

const (
	DefaultCategory = "uncategorized"
	DateFormat      = "2006-01-02"
)

// Validation errors per SPEC §4.1–4.5, §4.7.
var (
	ErrEmptyDescription       = errors.New("description must not be empty")
	ErrDescriptionTooLong     = errors.New("description must not exceed 255 characters")
	ErrAmountMustBePositive   = errors.New("amount must be greater than zero")
	ErrInvalidDateFormat      = errors.New("date must be in YYYY-MM-DD format")
	ErrInvalidMonth           = errors.New("month must be between 1 and 12")
	ErrInvalidYear            = errors.New("year must be a positive integer")
	ErrNoUpdateFieldsProvided = errors.New("at least one of --description, --amount, --category, --date must be provided")
)

// ValidateDescription validates the expense description.
func ValidateDescription(desc string) error {
	if strings.TrimSpace(desc) == "" {
		return ErrEmptyDescription
	}
	if utf8.RuneCountInString(desc) > 255 {
		return ErrDescriptionTooLong
	}
	return nil
}

// ValidateAmount validates that the amount in cents is strictly positive.
func ValidateAmount(amount int64) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}
	return nil
}

// ValidateDate validates that the date string matches YYYY-MM-DD and is a valid calendar date.
func ValidateDate(date string) error {
	if len(date) != 10 {
		return ErrInvalidDateFormat
	}
	t, err := time.Parse(DateFormat, date)
	if err != nil {
		return ErrInvalidDateFormat
	}
	// Verify exact formatted round-trip (catches non-standard inputs like 2024-02-30 or 2024-1-1)
	if t.Format(DateFormat) != date {
		return ErrInvalidDateFormat
	}
	return nil
}

// ValidateMonth validates that month is between 1 and 12.
func ValidateMonth(month int) error {
	if month < 1 || month > 12 {
		return ErrInvalidMonth
	}
	return nil
}

// ValidateYear validates that year is a positive integer.
func ValidateYear(year int) error {
	if year <= 0 {
		return ErrInvalidYear
	}
	return nil
}

// Add creates a new Expense, assigns nextID, increments nextID, and appends it to expenses.
func Add(expenses *[]Expense, nextID *int64, desc string, amount int64, category string, date string) (*Expense, error) {
	if err := ValidateDescription(desc); err != nil {
		return nil, err
	}
	if err := ValidateAmount(amount); err != nil {
		return nil, err
	}

	if date == "" {
		date = time.Now().Format(DateFormat)
	} else if err := ValidateDate(date); err != nil {
		return nil, err
	}

	if category == "" {
		category = DefaultCategory
	}

	newExpense := Expense{
		ID:          *nextID,
		Date:        date,
		Description: desc,
		Amount:      amount,
		Category:    category,
	}

	*nextID++
	*expenses = append(*expenses, newExpense)
	return &newExpense, nil
}

// UpdateParams contains optional fields for updating an expense.
type UpdateParams struct {
	Description *string
	Amount      *int64
	Category    *string
	Date        *string
}

// Update updates an existing expense identified by id.
func Update(expenses []Expense, id int64, params UpdateParams) (*Expense, error) {
	if params.Description == nil && params.Amount == nil && params.Category == nil && params.Date == nil {
		return nil, ErrNoUpdateFieldsProvided
	}

	idx := -1
	for i, e := range expenses {
		if e.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, fmt.Errorf("expense with ID %d not found", id)
	}

	if params.Description != nil {
		if err := ValidateDescription(*params.Description); err != nil {
			return nil, err
		}
	}
	if params.Amount != nil {
		if err := ValidateAmount(*params.Amount); err != nil {
			return nil, err
		}
	}
	if params.Date != nil {
		if err := ValidateDate(*params.Date); err != nil {
			return nil, err
		}
	}

	if params.Description != nil {
		expenses[idx].Description = *params.Description
	}
	if params.Amount != nil {
		expenses[idx].Amount = *params.Amount
	}
	if params.Category != nil {
		cat := *params.Category
		if cat == "" {
			cat = DefaultCategory
		}
		expenses[idx].Category = cat
	}
	if params.Date != nil {
		expenses[idx].Date = *params.Date
	}

	return &expenses[idx], nil
}

// Delete removes the expense with ID id from expenses.
func Delete(expenses *[]Expense, id int64) error {
	idx := -1
	for i, e := range *expenses {
		if e.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("expense with ID %d not found", id)
	}

	*expenses = append((*expenses)[:idx], (*expenses)[idx+1:]...)
	return nil
}

// FilterOption configures filtering parameters.
type FilterOption struct {
	Category string
	Month    int
	Year     int
}

// Filter returns a filtered and sorted copy of expenses matching the given criteria.
// Rows are sorted by date ascending, then by ID ascending as a tiebreaker.
func Filter(expenses []Expense, opt FilterOption) []Expense {
	var result []Expense
	for _, e := range expenses {
		if opt.Category != "" && !strings.EqualFold(e.Category, opt.Category) {
			continue
		}

		if opt.Month > 0 || opt.Year > 0 {
			t, err := time.Parse(DateFormat, e.Date)
			if err != nil {
				continue
			}
			if opt.Year > 0 && t.Year() != opt.Year {
				continue
			}
			if opt.Month > 0 && int(t.Month()) != opt.Month {
				continue
			}
		}

		result = append(result, e)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Date != result[j].Date {
			return result[i].Date < result[j].Date
		}
		return result[i].ID < result[j].ID
	})

	return result
}

// Total calculates the sum of amounts for the given expenses slice.
func Total(expenses []Expense) int64 {
	var sum int64
	for _, e := range expenses {
		sum += e.Amount
	}
	return sum
}
