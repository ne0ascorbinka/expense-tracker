package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ne0ascorbinka/expense-tracker/internal/expense"
	"github.com/ne0ascorbinka/expense-tracker/internal/format"
)

// NewUpdateCmd creates and returns the update subcommand.
func NewUpdateCmd() *cobra.Command {
	var (
		id       int64
		desc     string
		amount   string
		category string
		date     string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing expense",
		RunE: func(c *cobra.Command, args []string) error {
			if !c.Flags().Changed("description") &&
				!c.Flags().Changed("amount") &&
				!c.Flags().Changed("category") &&
				!c.Flags().Changed("date") {
				return expense.ErrNoUpdateFieldsProvided
			}

			var params expense.UpdateParams

			if c.Flags().Changed("description") {
				params.Description = &desc
			}

			if c.Flags().Changed("amount") {
				amountCents, err := format.ParseAmount(amount)
				if err != nil {
					return err
				}
				params.Amount = &amountCents
			}

			if c.Flags().Changed("category") {
				params.Category = &category
			}

			if c.Flags().Changed("date") {
				params.Date = &date
			}

			if _, err := expense.Update(AppStore.Data.Expenses, id, params); err != nil {
				return err
			}

			if err := AppStore.Save(); err != nil {
				return err
			}

			fmt.Fprintln(c.OutOrStdout(), "Expense updated successfully")
			return nil
		},
	}

	cmd.Flags().Int64Var(&id, "id", 0, "ID of the expense to update")
	cmd.Flags().StringVar(&desc, "description", "", "New description")
	cmd.Flags().StringVar(&amount, "amount", "", "New amount")
	cmd.Flags().StringVar(&category, "category", "", "New category")
	cmd.Flags().StringVar(&date, "date", "", "New date (YYYY-MM-DD)")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}
