package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
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

// ──────────────────────────────────────────────────────────
// UNIT 3 — Printing and Logging
// ──────────────────────────────────────────────────────────

// LogLevel represents log severity.
// UNIT 1: Custom type wrapping int; UNIT 1: const with iota.
type LogLevel int

const (
	LogInfo  LogLevel = iota // 0
	LogWarn                  // 1
	LogError                 // 2
)

// UNIT 1: var keyword — package-level variable with zero value.
// UNIT 3: Closure — logger functions close over this.
var currentLogLevel LogLevel

// SetLogLevel sets the minimum severity for logging.
func SetLogLevel(level LogLevel) {
	currentLogLevel = level
}

// LogMessage logs a message if the level is at or above the current threshold.
// UNIT 3: Printing and logging — uses log.Printf for structured output.
// UNIT 3: Variadic parameter — args ...interface{} forwards to Printf.
func LogMessage(level LogLevel, format string, args ...interface{}) {
	if level < currentLogLevel {
		return
	}
	// UNIT 1: Control flow — switch.
	var prefix string
	switch level {
	case LogInfo:
		prefix = "[INFO]"
	case LogWarn:
		prefix = "[WARN]"
	case LogError:
		prefix = "[ERROR]"
	}
	// UNIT 3: Unfurling — args... spreads the variadic into Printf.
	log.Printf("%s "+format, append([]interface{}{prefix}, args...)...)
}

// ──────────────────────────────────────────────────────────
// UNIT 3 — Panic and Recover
// ──────────────────────────────────────────────────────────

// SafeExecute runs fn and recovers from any panic, returning an error instead.
// UNIT 3: Panic/Recover — defer + recover pattern to convert panics into errors.
func SafeExecute(fn func() error) (err error) {
	// UNIT 3: Defer — this anonymous function runs after fn() returns or panics.
	defer func() {
		// UNIT 3: Recover — catches panics and converts them to errors.
		if r := recover(); r != nil {
			// UNIT 3: Printing — log the panic for observability.
			log.Printf("[PANIC RECOVERED] %v", r)
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()
	return fn()
}

// MustParseDate parses a date string or panics if invalid.
// UNIT 3: Panic — used when the program is in an unrecoverable state.
func MustParseDate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// UNIT 3: Panic — signals a programming error (unrecoverable).
		panic(fmt.Sprintf("invalid date format %q: %v", dateStr, err))
	}
	return t
}

// ──────────────────────────────────────────────────────────
// UNIT 3 — Returning a function + Closure (logging helpers)
// ──────────────────────────────────────────────────────────

// MakeTimedLogger returns a function that logs with elapsed time.
// UNIT 3: Returning a function + Closure — captures `start` and `label`.
func MakeTimedLogger(label string) func(msg string) {
	start := time.Now()
	// UNIT 3: Closure — returned function closes over start and label.
	return func(msg string) {
		elapsed := time.Since(start)
		log.Printf("[%s] %s (elapsed: %v)", label, msg, elapsed)
	}
}
