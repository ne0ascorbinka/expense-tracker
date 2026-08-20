package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ne0ascorbinka/expense-tracker/cmd"
)

func TestRootCmdHelp(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected --help to execute cleanly, got: %v", err)
	}
}

func TestRootCmdPersistentPreRun(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "expenses.json")

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"--file", filePath})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected execute with custom file to succeed, got: %v", err)
	}

	if cmd.AppStore == nil {
		t.Fatalf("expected AppStore to be initialized")
	}
	if cmd.AppStore.FilePath != filePath {
		t.Fatalf("expected store path %s, got %s", filePath, cmd.AppStore.FilePath)
	}

	// Verify the file was created on disk
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("expected data file to be created on disk: %v", err)
	}
}
