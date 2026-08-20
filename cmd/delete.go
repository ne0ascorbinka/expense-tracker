package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"expense-tracker/internal/expense"
)

// NewDeleteCmd creates and returns the delete subcommand.
func NewDeleteCmd() *cobra.Command {
	var id int64

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an expense by ID",
		RunE: func(c *cobra.Command, args []string) error {
			if err := expense.Delete(&AppStore.Data.Expenses, id); err != nil {
				return err
			}

			if err := AppStore.Save(); err != nil {
				return err
			}

			fmt.Fprintln(c.OutOrStdout(), "Expense deleted successfully")
			return nil
		},
	}

	cmd.Flags().Int64Var(&id, "id", 0, "ID of the expense to delete")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
