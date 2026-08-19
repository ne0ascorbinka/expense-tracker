package format_test

import (
	"bytes"
	"strings"
	"testing"

	"expense-tracker/internal/expense"
	"expense-tracker/internal/format"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCents int64
		wantErr   error
	}{
		{"integer amount", "20", 2000, nil},
		{"single decimal", "20.5", 2050, nil},
		{"two decimals", "20.50", 2050, nil},
		{"small decimal", "0.05", 5, nil},
		{"large amount", "1234.56", 123456, nil},
		{"three decimals rejected", "20.999", 0, format.ErrAmountTooManyDecimals},
		{"zero amount", "0", 0, format.ErrAmountMustBePositive},
		{"zero decimal", "0.00", 0, format.ErrAmountMustBePositive},
		{"negative integer", "-10", 0, format.ErrAmountMustBePositive},
		{"negative decimal", "-0.50", 0, format.ErrAmountMustBePositive},
		{"non-numeric string", "abc", 0, format.ErrInvalidAmountNumber},
		{"empty string", "", 0, format.ErrInvalidAmountNumber},
		{"multiple dots", "1.2.3", 0, format.ErrInvalidAmountNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := format.ParseAmount(tt.input)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.wantCents {
					t.Fatalf("expected %d cents, got %d", tt.wantCents, got)
				}
			}
		})
	}
}

func TestCentsToDisplay(t *testing.T) {
	tests := []struct {
		cents int64
		want  string
	}{
		{2000, "$20.00"},
		{50000, "$500.00"},
		{5, "$0.05"},
		{50, "$0.50"},
		{0, "$0.00"},
		{123456, "$1234.56"},
		{-500, "-$5.00"},
	}

	for _, tt := range tests {
		got := format.CentsToDisplay(tt.cents)
		if got != tt.want {
			t.Errorf("CentsToDisplay(%d) = %q, want %q", tt.cents, got, tt.want)
		}
	}
}

func TestCentsToCSV(t *testing.T) {
	tests := []struct {
		cents int64
		want  string
	}{
		{2000, "20.00"},
		{50000, "500.00"},
		{5, "0.05"},
		{0, "0.00"},
		{123456, "1234.56"},
	}

	for _, tt := range tests {
		got := format.CentsToCSV(tt.cents)
		if got != tt.want {
			t.Errorf("CentsToCSV(%d) = %q, want %q", tt.cents, got, tt.want)
		}
	}
}

func TestFormatTable(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		var buf bytes.Buffer
		err := format.FormatTable(&buf, []expense.Expense{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.TrimSpace(buf.String()) != "No expenses found." {
			t.Fatalf("expected 'No expenses found.', got %q", buf.String())
		}
	})

	t.Run("formatted table", func(t *testing.T) {
		expenses := []expense.Expense{
			{ID: 1, Date: "2024-08-06", Description: "Lunch", Amount: 2000, Category: "food"},
			{ID: 2, Date: "2024-08-06", Description: "Dinner", Amount: 1000, Category: "uncategorized"},
		}

		var buf bytes.Buffer
		err := format.FormatTable(&buf, expenses)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "ID") || !strings.Contains(out, "Date") || !strings.Contains(out, "Description") || !strings.Contains(out, "Amount") || !strings.Contains(out, "Category") {
			t.Fatalf("table missing header columns: %s", out)
		}
		if !strings.Contains(out, "$20.00") || !strings.Contains(out, "$10.00") {
			t.Fatalf("table missing formatted amounts: %s", out)
		}
		if !strings.Contains(out, "Lunch") || !strings.Contains(out, "uncategorized") {
			t.Fatalf("table missing expense data: %s", out)
		}
	})
}

func TestFormatCSV(t *testing.T) {
	expenses := []expense.Expense{
		{ID: 1, Date: "2024-08-06", Description: "Lunch, with team", Amount: 2000, Category: "food"},
		{ID: 2, Date: "2024-08-06", Description: "Dinner", Amount: 1000, Category: "uncategorized"},
	}

	var buf bytes.Buffer
	err := format.FormatCSV(&buf, expenses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), buf.String())
	}
	if strings.TrimSpace(lines[0]) != "ID,Date,Description,Amount,Category" {
		t.Fatalf("unexpected CSV header: %q", lines[0])
	}
	if !strings.Contains(lines[1], `"Lunch, with team"`) {
		t.Fatalf("expected comma-containing description to be quoted, got: %s", lines[1])
	}
	if !strings.Contains(lines[1], "20.00") || !strings.Contains(lines[2], "10.00") {
		t.Fatalf("expected raw decimal amounts without $, got:\n%s", buf.String())
	}
}
