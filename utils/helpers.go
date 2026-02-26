package utils

import (
	"os"
	"path/filepath"
)

// GetDBPath returns the default SQLite database path.
func GetDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "expense.db"
	}
	dbDir := filepath.Join(homeDir, ".expensevault")
	_ = os.MkdirAll(dbDir, 0755)
	return filepath.Join(dbDir, "expense.db")
}
