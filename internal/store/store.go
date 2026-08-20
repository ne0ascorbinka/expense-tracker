package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ne0ascorbinka/expense-tracker/internal/expense"
)

// Data represents the JSON schema stored on disk.
type Data struct {
	NextID   int64              `json:"next_id"`
	Budgets  map[string]int64   `json:"budgets"`
	Expenses []expense.Expense  `json:"expenses"`
}

// Store handles persisting and loading expenses and budgets from disk.
type Store struct {
	FilePath string
	Data     Data
}

// DefaultPath returns the default file path: $HOME/.expense-tracker/expenses.json
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".expense-tracker", "expenses.json"), nil
}

// New creates a new Store instance with the given file path.
func New(filePath string) *Store {
	return &Store{
		FilePath: filePath,
		Data: Data{
			NextID:   1,
			Budgets:  make(map[string]int64),
			Expenses: []expense.Expense{},
		},
	}
}

// Load reads the JSON store file from disk. If the file does not exist,
// it initializes a new empty store and creates the file and parent directories.
// Returns an error if the file is corrupted.
func (s *Store) Load() error {
	if _, err := os.Stat(s.FilePath); os.IsNotExist(err) {
		s.Data = Data{
			NextID:   1,
			Budgets:  make(map[string]int64),
			Expenses: []expense.Expense{},
		}
		return s.Save()
	} else if err != nil {
		return fmt.Errorf("failed to access data file %q: %w", s.FilePath, err)
	}

	raw, err := os.ReadFile(s.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read data file %q: %w", s.FilePath, err)
	}

	var data Data
	if err := json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("corrupt data file %q: %w", s.FilePath, err)
	}

	if data.Budgets == nil {
		data.Budgets = make(map[string]int64)
	}
	if data.Expenses == nil {
		data.Expenses = []expense.Expense{}
	}
	if data.NextID == 0 {
		data.NextID = 1
	}

	s.Data = data
	return nil
}

// Save marshals and writes the store data to disk.
func (s *Store) Save() error {
	dir := filepath.Dir(s.FilePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", dir, err)
		}
	}

	if s.Data.Budgets == nil {
		s.Data.Budgets = make(map[string]int64)
	}
	if s.Data.Expenses == nil {
		s.Data.Expenses = []expense.Expense{}
	}
	if s.Data.NextID == 0 {
		s.Data.NextID = 1
	}

	encoded, err := json.MarshalIndent(s.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}
	encoded = append(encoded, '\n')

	// Write file atomically using temp file in the same directory
	tempFile, err := os.CreateTemp(dir, "expenses-*.json.tmp")
	if err != nil {
		// Fallback to direct write if temp file cannot be created
		return os.WriteFile(s.FilePath, encoded, 0644)
	}
	tempPath := tempFile.Name()

	if _, err := tempFile.Write(encoded); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to write data: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tempPath, s.FilePath); err != nil {
		// On Windows, os.Rename fails if destination exists; remove destination first
		_ = os.Remove(s.FilePath)
		if renameErr := os.Rename(tempPath, s.FilePath); renameErr != nil {
			// Fallback: direct write
			os.Remove(tempPath)
			return os.WriteFile(s.FilePath, encoded, 0644)
		}
	}

	return nil
}
