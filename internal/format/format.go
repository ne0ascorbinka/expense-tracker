package format

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	"expense-tracker/internal/expense"
)

var (
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
	ErrAmountTooManyDecimals = errors.New("amount must have at most 2 decimal places")
	ErrInvalidAmountNumber  = errors.New("amount must be a valid number")
)

// ParseAmount parses an amount string (e.g. "20", "20.5", "20.50") into integer cents.
// Enforces ≤ 2 decimal places and positive value.
func ParseAmount(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidAmountNumber
	}

	isNegative := false
	if strings.HasPrefix(s, "-") {
		isNegative = true
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}

	if s == "" {
		return 0, ErrInvalidAmountNumber
	}

	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, ErrInvalidAmountNumber
	}

	intStr := parts[0]
	if intStr == "" {
		intStr = "0"
	}
	intVal, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return 0, ErrInvalidAmountNumber
	}

	var decVal int64
	if len(parts) == 2 {
		decStr := parts[1]
		if len(decStr) > 2 {
			return 0, ErrAmountTooManyDecimals
		}
		if len(decStr) == 1 {
			decStr += "0"
		} else if len(decStr) == 0 {
			decStr = "00"
		}
		var err error
		decVal, err = strconv.ParseInt(decStr, 10, 64)
		if err != nil {
			return 0, ErrInvalidAmountNumber
		}
	}

	totalCents := intVal*100 + decVal
	if isNegative {
		totalCents = -totalCents
	}

	if totalCents <= 0 {
		return 0, ErrAmountMustBePositive
	}

	return totalCents, nil
}

// CentsToDisplay formats integer cents as "$XX.XX" (always 2 decimal places with $ prefix).
func CentsToDisplay(cents int64) string {
	if cents < 0 {
		pos := -cents
		return fmt.Sprintf("-$%d.%02d", pos/100, pos%100)
	}
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}

// CentsToCSV formats integer cents as "XX.XX" (no $ prefix).
func CentsToCSV(cents int64) string {
	if cents < 0 {
		pos := -cents
		return fmt.Sprintf("-%d.%02d", pos/100, pos%100)
	}
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

// FormatTable outputs expenses formatted as a tabular text using tabwriter.
// If no expenses are provided, outputs "No expenses found.\n".
func FormatTable(w io.Writer, expenses []expense.Expense) error {
	if len(expenses) == 0 {
		_, err := fmt.Fprintln(w, "No expenses found.")
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tDate\tDescription\tAmount\tCategory"); err != nil {
		return err
	}

	for _, e := range expenses {
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n",
			e.ID,
			e.Date,
			e.Description,
			CentsToDisplay(e.Amount),
			e.Category,
		); err != nil {
			return err
		}
	}

	return tw.Flush()
}

// FormatCSV writes expenses as CSV with header: ID,Date,Description,Amount,Category.
func FormatCSV(w io.Writer, expenses []expense.Expense) error {
	cw := csv.NewWriter(w)

	if err := cw.Write([]string{"ID", "Date", "Description", "Amount", "Category"}); err != nil {
		return err
	}

	for _, e := range expenses {
		row := []string{
			strconv.FormatInt(e.ID, 10),
			e.Date,
			e.Description,
			CentsToCSV(e.Amount),
			e.Category,
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}
