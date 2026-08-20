package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"expense-tracker/internal/expense"
	"expense-tracker/internal/format"
)

// NewAddCmd creates and returns the add subcommand.
func NewAddCmd() *cobra.Command {
	var (
		desc     string
		amount   string
		category string
		date     string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new expense",
		RunE: func(c *cobra.Command, args []string) error {
			amountCents, err := format.ParseAmount(amount)
			if err != nil {
				return err
			}

			newExp, err := expense.Add(&AppStore.Data.Expenses, &AppStore.Data.NextID, desc, amountCents, category, date)
			if err != nil {
				return err
			}

			if err := AppStore.Save(); err != nil {
				return err
			}

			fmt.Fprintf(c.OutOrStdout(), "Expense added successfully (ID: %d)\n", newExp.ID)

			t, err := time.Parse(expense.DateFormat, newExp.Date)
			if err == nil {
				if budget, exceeded := expense.CheckBudgetExceeded(AppStore.Data.Expenses, AppStore.Data.Budgets, t.Year(), int(t.Month())); exceeded {
					fmt.Fprintf(c.OutOrStdout(), "Warning: you have exceeded your budget of %s for %s.\n",
						format.CentsToDisplay(budget),
						t.Format("January 2006"),
					)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&desc, "description", "", "Expense description (1-255 chars)")
	cmd.Flags().StringVar(&amount, "amount", "", "Expense amount (positive, <= 2 dp)")
	cmd.Flags().StringVar(&category, "category", expense.DefaultCategory, "Category label (stored verbatim)")
	cmd.Flags().StringVar(&date, "date", "", "Expense date (YYYY-MM-DD)")

	_ = cmd.MarkFlagRequired("description")
	_ = cmd.MarkFlagRequired("amount")

	return cmd
}
