package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"expense-tracker/internal/expense"
	"expense-tracker/internal/format"
)

// NewListCmd creates and returns the list subcommand.
func NewListCmd() *cobra.Command {
	var (
		category string
		month    int
		year     int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display all expenses in a formatted table",
		RunE: func(c *cobra.Command, args []string) error {
			if c.Flags().Changed("month") {
				if err := expense.ValidateMonth(month); err != nil {
					return err
				}
			}

			if c.Flags().Changed("year") {
				if err := expense.ValidateYear(year); err != nil {
					return err
				}
			}

			if c.Flags().Changed("month") && !c.Flags().Changed("year") {
				year = time.Now().Year()
			}

			filtered := expense.Filter(AppStore.Data.Expenses, expense.FilterOption{
				Category: category,
				Month:    month,
				Year:     year,
			})

			return format.FormatTable(c.OutOrStdout(), filtered)
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "Filter to expenses whose category matches (case-insensitive)")
	cmd.Flags().IntVar(&month, "month", 0, "Filter to a specific month (1–12)")
	cmd.Flags().IntVar(&year, "year", 0, "Filter to a specific year")

	return cmd
}
