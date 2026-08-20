package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ne0ascorbinka/expense-tracker/internal/expense"
	"github.com/ne0ascorbinka/expense-tracker/internal/store"
)

func TestStoreFirstRun(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "nested", "sub", "expenses.json")

	s := store.New(filePath)
	if err := s.Load(); err != nil {
		t.Fatalf("expected Load to succeed on first run, got: %v", err)
	}

	if s.Data.NextID != 1 {
		t.Fatalf("expected NextID 1, got %d", s.Data.NextID)
	}
	if len(s.Data.Expenses) != 0 {
		t.Fatalf("expected 0 expenses, got %d", len(s.Data.Expenses))
	}
	if len(s.Data.Budgets) != 0 {
		t.Fatalf("expected 0 budgets, got %d", len(s.Data.Budgets))
	}

	// Verify file actually exists on disk
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("expected file %s to exist on disk: %v", filePath, err)
	}
}

func TestStoreSaveAndReload(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "expenses.json")

	s := store.New(filePath)
	if err := s.Load(); err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}

	s.Data.NextID = 4
	s.Data.Budgets["2026-08"] = 50000
	s.Data.Expenses = append(s.Data.Expenses, expense.Expense{
		ID:          1,
		Date:        "2026-08-20",
		Description: "Lunch",
		Amount:      2000,
		Category:    "Food",
	})

	if err := s.Save(); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}

	// Reload using a new store instance
	s2 := store.New(filePath)
	if err := s2.Load(); err != nil {
		t.Fatalf("unexpected reload error: %v", err)
	}

	if s2.Data.NextID != 4 {
		t.Fatalf("expected NextID 4, got %d", s2.Data.NextID)
	}
	if s2.Data.Budgets["2026-08"] != 50000 {
		t.Fatalf("expected budget 50000, got %d", s2.Data.Budgets["2026-08"])
	}
	if len(s2.Data.Expenses) != 1 {
		t.Fatalf("expected 1 expense, got %d", len(s2.Data.Expenses))
	}
	e := s2.Data.Expenses[0]
	if e.ID != 1 || e.Description != "Lunch" || e.Amount != 2000 || e.Category != "Food" || e.Date != "2026-08-20" {
		t.Fatalf("unexpected expense content: %+v", e)
	}
}

func TestStoreCorruptFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "corrupt.json")

	// Write invalid JSON content
	if err := os.WriteFile(filePath, []byte("{invalid-json:"), 0644); err != nil {
		t.Fatalf("failed to write test corrupt file: %v", err)
	}

	s := store.New(filePath)
	err := s.Load()
	if err == nil {
		t.Fatalf("expected error on corrupt file, got nil")
	}
}

func TestDefaultPath(t *testing.T) {
	p, err := store.DefaultPath()
	if err != nil {
		t.Fatalf("unexpected error getting default path: %v", err)
	}
	if p == "" {
		t.Fatalf("expected non-empty default path")
	}
	if !filepath.IsAbs(p) {
		t.Fatalf("expected absolute path, got: %s", p)
	}
}
