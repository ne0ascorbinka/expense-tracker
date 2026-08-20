package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ne0ascorbinka/expense-tracker/internal/expense"
	"github.com/ne0ascorbinka/expense-tracker/internal/format"
)

// NewBudgetCmd creates and returns the budget parent command with set and clear subcommands.
func NewBudgetCmd() *cobra.Command {
	budgetCmd := &cobra.Command{
		Use:   "budget",
		Short: "Manage monthly budgets",
		RunE: func(c *cobra.Command, args []string) error {
			return c.Help()
		},
	}

	budgetCmd.AddCommand(
		newBudgetSetCmd(),
		newBudgetClearCmd(),
	)

	return budgetCmd
}

func newBudgetSetCmd() *cobra.Command {
	var (
		amount string
		month  int
		year   int
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set or update the budget for a specific calendar month",
		RunE: func(c *cobra.Command, args []string) error {
			amountCents, err := format.ParseAmount(amount)
			if err != nil {
				return err
			}
			if err := expense.ValidateAmount(amountCents); err != nil {
				return err
			}

			now := time.Now()
			if !c.Flags().Changed("month") {
				month = int(now.Month())
			} else {
				if err := expense.ValidateMonth(month); err != nil {
					return err
				}
			}

			if !c.Flags().Changed("year") {
				year = now.Year()
			} else {
				if err := expense.ValidateYear(year); err != nil {
					return err
				}
			}

			expense.SetBudget(AppStore.Data.Budgets, year, month, amountCents)

			if err := AppStore.Save(); err != nil {
				return err
			}

			monthName := time.Month(month).String()
			fmt.Fprintf(c.OutOrStdout(), "Budget for %s %d set to %s\n",
				monthName,
				year,
				format.CentsToDisplay(amountCents),
			)
			return nil
		},
	}

	cmd.Flags().StringVar(&amount, "amount", "", "Monthly budget limit (positive, <= 2 dp)")
	cmd.Flags().IntVar(&month, "month", 0, "Target month (1–12)")
	cmd.Flags().IntVar(&year, "year", 0, "Target year")

	_ = cmd.MarkFlagRequired("amount")

	return cmd
}

func newBudgetClearCmd() *cobra.Command {
	var (
		month int
		year  int
	)

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear the budget for a specific calendar month",
		RunE: func(c *cobra.Command, args []string) error {
			now := time.Now()
			if !c.Flags().Changed("month") {
				month = int(now.Month())
			} else {
				if err := expense.ValidateMonth(month); err != nil {
					return err
				}
			}

			if !c.Flags().Changed("year") {
				year = now.Year()
			} else {
				if err := expense.ValidateYear(year); err != nil {
					return err
				}
			}

			expense.ClearBudget(AppStore.Data.Budgets, year, month)

			if err := AppStore.Save(); err != nil {
				return err
			}

			monthName := time.Month(month).String()
			fmt.Fprintf(c.OutOrStdout(), "Budget for %s %d cleared\n", monthName, year)
			return nil
		},
	}

	cmd.Flags().IntVar(&month, "month", 0, "Target month (1–12)")
	cmd.Flags().IntVar(&year, "year", 0, "Target year")

	return cmd
}
