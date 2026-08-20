package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ne0ascorbinka/expense-tracker/internal/store"
)

var (
	filePath string
	AppStore *store.Store
)

// NewRootCmd creates and returns a new instance of the root cobra command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "expense-tracker",
		Short:         "A simple command-line expense tracker",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			fPath := filePath
			if fPath == "" {
				defPath, err := store.DefaultPath()
				if err != nil {
					return err
				}
				fPath = defPath
			}
			AppStore = store.New(fPath)
			if err := AppStore.Load(); err != nil {
				return fmt.Errorf("failed to load data: %w", err)
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			return c.Help()
		},
	}

	defaultPath, _ := store.DefaultPath()
	cmd.PersistentFlags().StringVar(&filePath, "file", defaultPath, "Override data file path")

	cmd.AddCommand(
		NewAddCmd(),
		NewUpdateCmd(),
		NewDeleteCmd(),
		NewListCmd(),
		NewSummaryCmd(),
		NewExportCmd(),
		NewBudgetCmd(),
	)

	return cmd
}

// RootCmd is the default root command.
var RootCmd = NewRootCmd()

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}

// SetStore allows overriding the AppStore for testing or programmatic usage.
func SetStore(s *store.Store) {
	AppStore = s
}
