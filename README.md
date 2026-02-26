# ExpenseVault — Personal Finance CLI + TUI

ExpenseVault is a Go CLI application for managing personal finances with an interactive TUI, MySQL/SQLite storage, import/export, bcrypt authentication, and comprehensive coverage of Go programming concepts from Units 1–4.

---

## Quick Start

```bash
# Build (Linux / macOS)
go build -o expensevault .
./expensevault tui

# Build (Windows)
go build -o expensevault.exe .
.\expensevault.exe tui

# CLI commands
./expensevault add -t expense -a 250 -d "Lunch" -c Food
./expensevault list
```

## Running Tests

```bash
# Run all tests
go test ./... -v

# Run with benchmarks
go test ./models/ ./services/ -bench=. -v

# Run with coverage
go test ./models/ ./services/ -cover
```

---

## Project Structure

```
expenseVault-revised/
├── main.go                  # Entry point
├── go.mod / go.sum          # Module dependencies
├── api/
│   ├── auth.go              # JWT token generation
│   └── server.go            # HTTP server + sync endpoint
├── cmd/
│   ├── root.go              # Cobra root command
│   ├── add.go               # Add transaction (factory, pointer receiver)
│   ├── edit.go              # Edit transaction (pointer-based mutation)
│   ├── list.go              # List transactions (value receiver String())
│   ├── delete.go            # Delete transaction
│   ├── report.go            # Reports (monthly/category/yearly)
│   ├── exportcmd.go         # Export to CSV/JSON
│   ├── importcmd.go         # Import from CSV/JSON
│   ├── backup.go / restore.go
│   ├── signup.go / login.go # bcrypt auth
│   ├── server.go / sync.go
│   ├── tui.go               # Launch interactive TUI
│   └── demo.go
├── db/
│   ├── sqlite.go            # SQLite store (defer, recover, pointer receivers)
│   └── mysql.go             # MySQL store (cross-platform)
├── export/
│   ├── exporter.go          # Exporter/Importer interfaces (polymorphism)
│   ├── csv.go               # CSV import/export (defer, type conversion)
│   ├── json_export.go       # JSON import/export (marshal/unmarshal)
│   └── helpers.go           # Parsing utilities
├── models/
│   ├── types.go             # Custom types, arrays, slices, maps, structs
│   ├── transaction.go       # Transaction struct, receivers, factory, JSON
│   ├── errors.go            # Error types, wrapping, errors.Is/As, variadic
│   ├── sync.go              # SyncPayload struct
│   └── transaction_test.go  # 40+ tests + 6 benchmarks
├── services/
│   ├── reporter.go          # Reporter interface, variadic, closures, recursion
│   ├── categorizer.go       # Auto-categorizer (factory, pointer receiver)
│   └── reporter_test.go     # 20+ tests + 5 benchmarks
├── tui/
│   ├── app.go               # BubbleTea TUI (closures, anonymous functions)
│   └── styles.go            # Lipgloss styles
└── utils/
    ├── config.go            # Config struct, LoadConfig factory
    └── helpers.go           # Logging, panic/recover, timed logger
```

---

## Go Concepts Implemented (Units 1–4)

### Unit 1 — Basics

| Concept | Where |
|---------|-------|
| `var` keyword | `models/types.go` — `ValidCategories`, `ZeroValueDemo`; `db/sqlite.go` — `var t Transaction` |
| Short declaration `:=` | Used throughout all files |
| Zero values | `models/types.go` — `ZeroValueDemo`, `ShowZeroValues()` |
| Custom types | `models/types.go` — `Rupees`, `Category`, `TransactionType` |
| Type conversion (not casting) | `models/types.go` — `ToFloat64()`, `ToInt()`; `export/csv.go` — `Rupees(amount)` |
| Constants / iota | `models/types.go` — category & type consts; `utils/helpers.go` — `LogLevel` with iota |
| `fmt` package | Used extensively for `Sprintf`, `Errorf`, `Printf` |
| Control flow (if/switch/for) | `db/sqlite.go` — dynamic query builder; `services/categorizer.go` — switch |

### Unit 2 — Composite Types

| Concept | Where |
|---------|-------|
| **Array** (fixed-size) | `models/types.go` — `ValidCategories = [11]Category{...}` |
| Slice: composite literal | `db/sqlite.go` — `[]string{...}`, `[]any{}`; `export/csv.go` — header row |
| Slice: `for range` | Every file iterating over transactions |
| Slice: slicing `[a:b]` | `models/types.go` — `SliceFirstN()`, `DeleteTransactionFromSlice()` |
| Slice: `append` | `db/sqlite.go` — dynamic query args; `models/types.go` — `FilterTransactions` |
| Slice: `delete` from slice | `models/types.go` — `DeleteTransactionFromSlice()` using `append(s[:i], s[i+1:]...)` |
| Slice: `make` | `models/types.go` — `FilterTransactions`; `services/reporter.go` — `make([]string, 0, len)` |
| Multi-dimensional slice | `models/types.go` — `BuildCategorySummary() [][]string` |
| Map: create, add, range | `models/errors.go` — `QuickSummary`; `services/reporter.go` — all reporters |
| Map: `delete` | `models/types.go` — `PurgeCategoryFromMap()` |
| Struct | `models/transaction.go` — `Transaction`, `User`, `ReportEntry`, `MonthlySummary` |
| **Embedded struct** | `models/types.go` — `TransactionWithMeta` embeds `Transaction` + `Metadata` |
| **Anonymous struct** | `models/types.go` — `ParsedDateRange()` returns inline struct |

### Unit 3 — Functions & Error Handling

| Concept | Where |
|---------|-------|
| **Variadic parameter** | `services/reporter.go` — `CombineReports(...Reporter)`; `models/errors.go` — `ValidateAll(...*Transaction)`; `utils/helpers.go` — `LogMessage(...interface{})` |
| **Unfurling a slice** | `services/reporter.go` — `allReporters...`; `db/sqlite.go` — `args...`; `models/types.go` — `txs[i+1:]...` |
| Defer | `db/sqlite.go` — `defer rows.Close()`; `export/csv.go` — `defer file.Close()` |
| **Panic** | `utils/helpers.go` — `MustParseDate()` panics on invalid date |
| **Recover** | `utils/helpers.go` — `SafeExecute()`; `db/sqlite.go` — `BulkInsert()` with defer+recover |
| Methods (value receiver) | `models/transaction.go` — `String()`, `Summary()`, `IsExpense()`, `IsIncome()` |
| Methods (pointer receiver) | `models/transaction.go` — `SetAmount()`, `SetCategory()`, `ApplyDiscount()` |
| **Interfaces & polymorphism** | `services/reporter.go` — `Reporter` interface; `export/exporter.go` — `Exporter`/`Importer` |
| **Anonymous function** | `services/reporter.go` — `sort.Slice(txs, func(i, j int) bool { ... })` |
| **Function expression** | `services/reporter.go` — `TransactionFilter` type; `export/exporter.go` — `TransformFunc` type |
| **Returning a function** | `services/reporter.go` — `MakeAmountFilter()`, `MakeCategoryFilter()`, `MakeTypeFilter()` |
| **Callback** | `models/types.go` — `FilterTransactions(txs, predicate)`; `services/reporter.go` — `ApplyFilter(txs, filter)` |
| **Closure** | `services/reporter.go` — `MakeAmountFilter` closes over `min`; `utils/helpers.go` — `MakeTimedLogger` closes over `start` |
| **Recursion** | `services/reporter.go` — `SumTransactionsRecursive()` |
| Error handling (errors with info) | `models/errors.go` — `DatabaseError`, `AuthError`, `ValidationError` with contextual fields |
| Error wrapping / `errors.Is` / `errors.As` | `models/errors.go` — `Unwrap()`, `WrapDBError()`, `IsNotFound()`, `AsValidationError()` |
| **Printing and logging** | `utils/helpers.go` — `LogMessage()` with `log.Printf`, `LogLevel` |

### Unit 4 — Pointers, JSON, Auth, Testing

| Concept | Where |
|---------|-------|
| Pointers / method sets | `models/transaction.go` — pointer receivers for mutation, value receivers for reads |
| Factory function returning `*T` | `models/transaction.go` — `NewTransaction()`; `db/sqlite.go` — `NewStore()`; `services/categorizer.go` — `NewCategorizer()` |
| Pass-by-value vs pass-by-pointer | `models/transaction.go` — `ModifyByValue()` vs `ModifyByPointer()` |
| JSON `Marshal` / `Unmarshal` | `models/transaction.go` — `MarshalTransaction()`, `UnmarshalTransactions()`; `export/json_export.go` |
| JSON struct tags | `models/transaction.go` — `json:"id"`, `json:"-"` on `PasswordHash` |
| bcrypt | `cmd/signup.go`, `cmd/login.go`, `tui/app.go` — password hashing & verification |
| Table-driven tests | `models/transaction_test.go` — 8 marshal/unmarshal cases, 8 validation cases, 5 discount cases |
| Benchmarks | `models/transaction_test.go` — 6 benchmarks; `services/reporter_test.go` — 5 benchmarks |

---

## TUI Auth Flow

Running `./expensevault tui` starts an in-TUI auth flow:

1. Sign up (bcrypt hash stored in DB)
2. Log in (bcrypt verification)
3. Dashboard unlocks after login

---

## MySQL Workbench Setup (multi-user)

### 1) Configure .env

```env
DB_TYPE=mysql
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=expensevault_app
MYSQL_PASSWORD=strong_password
MYSQL_DATABASE=expensevault
```

### 2) Create database + user (admin runs once)

```sql
CREATE DATABASE IF NOT EXISTS expensevault;

CREATE USER 'expensevault_app'@'%' IDENTIFIED BY 'strong_password';
GRANT ALL PRIVILEGES ON expensevault.* TO 'expensevault_app'@'%';
FLUSH PRIVILEGES;
```

### 3) Start app to auto-create tables

```bash
./expensevault tui
```

The app creates `transactions` and `users` tables automatically.

### 4) MySQL Workbench connection for another user

In MySQL Workbench:

- Hostname: MYSQL_HOST (public IP or DNS)
- Port: MYSQL_PORT
- Username: expensevault_app
- Password: strong_password
- Default Schema: expensevault

If MySQL is remote, open port 3306 and allow that host in `CREATE USER` (use `%` or a specific IP).

---

## Commands

| Command | Description |
|---------|-------------|
| `add` | Add a transaction |
| `list` | List / filter transactions |
| `edit` | Edit an existing transaction |
| `delete` | Delete a transaction |
| `report` | Generate report (monthly / category / yearly) |
| `export` | Export to CSV or JSON |
| `import` | Import from CSV or JSON |
| `backup` | Backup all data to JSON |
| `restore` | Restore from backup file |
| `tui` | Interactive terminal dashboard |
| `signup` | Create user account (bcrypt) |
| `login` | Authenticate (bcrypt + JWT) |
| `server` | Start HTTP API server |
| `sync` | Sync transactions via HTTP |

---

## Tech Stack

- **Go 1.25** — Module: `expenseVault`
- **Cobra** — CLI framework
- **BubbleTea + Lipgloss** — Interactive TUI
- **SQLite** (`modernc.org/sqlite`) — Default local storage (pure Go, no CGO)
- **MySQL** (`go-sql-driver/mysql`) — Optional multi-user storage
- **bcrypt** (`golang.org/x/crypto`) — Password hashing
- **JWT** (`golang-jwt/jwt/v5`) — Token-based auth
- **godotenv** — `.env` file configuration

## Cross-Platform

The project runs on **Linux, macOS, and Windows** without modification:
- All file paths use `filepath.Join()` (no hardcoded separators)
- SQLite driver is pure Go (no C compiler needed)
- BubbleTea has native Windows terminal support
