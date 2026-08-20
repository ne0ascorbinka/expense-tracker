package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ne0ascorbinka/expense-tracker/internal/expense"
	"github.com/ne0ascorbinka/expense-tracker/internal/format"
)

// NewExportCmd creates and returns the export subcommand.
func NewExportCmd() *cobra.Command {
	var (
		outputFile string
		month      int
		year       int
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export expenses to a CSV file",
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
				Month: month,
				Year:  year,
			})

			f, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			defer f.Close()

			if err := format.FormatCSV(f, filtered); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			fmt.Fprintf(c.OutOrStdout(), "Expenses exported to %s successfully\n", outputFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&outputFile, "output", "", "Destination file path (created/overwritten)")
	cmd.Flags().IntVar(&month, "month", 0, "Filter to a specific month (1–12)")
	cmd.Flags().IntVar(&year, "year", 0, "Filter to a specific year")

	_ = cmd.MarkFlagRequired("output")

	return cmd
}
