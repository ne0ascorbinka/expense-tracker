package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ne0ascorbinka/expense-tracker/internal/expense"
	"github.com/ne0ascorbinka/expense-tracker/internal/format"
)

// NewSummaryCmd creates and returns the summary subcommand.
func NewSummaryCmd() *cobra.Command {
	var (
		month    int
		year     int
		category string
	)

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show the total of all matching expenses",
		RunE: func(c *cobra.Command, args []string) error {
			hasMonth := c.Flags().Changed("month")
			hasYear := c.Flags().Changed("year")
			hasCategory := c.Flags().Changed("category")

			if hasCategory && category == "" {
				return expense.ErrEmptyCategory
			}

			if hasMonth {
				if err := expense.ValidateMonth(month); err != nil {
					return err
				}
			}

			if hasYear {
				if err := expense.ValidateYear(year); err != nil {
					return err
				}
			}

			if hasMonth && !hasYear {
				year = time.Now().Year()
			}

			filtered := expense.Filter(AppStore.Data.Expenses, expense.FilterOption{
				Category: category,
				Month:    month,
				Year:     year,
			})
			totalCents := expense.Total(filtered)
			formattedTotal := format.CentsToDisplay(totalCents)

			if hasMonth {
				monthName := time.Month(month).String()
				if hasCategory {
					fmt.Fprintf(c.OutOrStdout(), "Total expenses for %s %d (%s): %s\n", monthName, year, category, formattedTotal)
				} else {
					fmt.Fprintf(c.OutOrStdout(), "Total expenses for %s %d: %s\n", monthName, year, formattedTotal)
				}
			} else if hasYear {
				if hasCategory {
					fmt.Fprintf(c.OutOrStdout(), "Total expenses for %d (%s): %s\n", year, category, formattedTotal)
				} else {
					fmt.Fprintf(c.OutOrStdout(), "Total expenses for %d: %s\n", year, formattedTotal)
				}
			} else if hasCategory {
				fmt.Fprintf(c.OutOrStdout(), "Total expenses for %s: %s\n", category, formattedTotal)
			} else {
				fmt.Fprintf(c.OutOrStdout(), "Total expenses: %s\n", formattedTotal)
			}

			if hasMonth {
				if budget, exceeded := expense.CheckBudgetExceeded(AppStore.Data.Expenses, AppStore.Data.Budgets, year, month); exceeded {
					monthName := time.Month(month).String()
					fmt.Fprintf(c.OutOrStdout(), "Warning: you have exceeded your budget of %s for %s %d.\n",
						format.CentsToDisplay(budget),
						monthName,
						year,
					)
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&month, "month", 0, "Filter to a specific month (1–12)")
	cmd.Flags().IntVar(&year, "year", 0, "Filter to a specific year")
	cmd.Flags().StringVar(&category, "category", "", "Filter by category (case-insensitive)")

	return cmd
}
