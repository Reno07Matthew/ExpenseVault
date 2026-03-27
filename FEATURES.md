# ExpenseVault — Complete Feature Reference

A comprehensive Go-based personal finance CLI & TUI application with multi-database support, AI-powered querying, and production-grade security features.

---

## Architecture Overview

```
expenseVault/
├── main.go              → Entry point
├── cmd/                 → 19 Cobra CLI commands
├── models/              → Domain types, validation, errors
├── db/                  → Multi-database storage layer
├── services/            → Business logic & advanced features
├── tui/                 → Interactive terminal UI (Bubble Tea)
├── api/                 → HTTP server & middleware chain
├── export/              → CSV/JSON import/export engine
└── utils/               → Configuration & helpers
```

---

## 🗂️ Core Features (Pre-existing)

### 1. Transaction Management (CRUD)

| Command | Description | File |
|---------|-------------|------|
| `add` | Add a new income/expense transaction | `cmd/add.go` |
| `list` | List all transactions for the current user | `cmd/list.go` |
| `edit` | Edit an existing transaction by ID | `cmd/edit.go` |
| `delete` | Delete a transaction by ID | `cmd/delete.go` |

- **Transaction Model** (`models/transaction.go`): Tracks `ID`, `UserID`, `Type` (income/expense), `Amount` (custom `Rupees` type), `Category`, `Description`, `Date`, `Notes`, `CreatedAt`, `UpdatedAt`
- **11 Categories**: Food, Travel, Shopping, Bills, Health, Education, Entertainment, Salary, Freelance, Investment, Other
- **Factory pattern**: `NewTransaction()` returns `*Transaction` (pointer semantics)
- **Mutation methods**: `SetAmount()`, `SetCategory()`, `SetDescription()`, `SetDate()`, `SetNotes()`, `ApplyDiscount()` — all pointer receivers
- **Read methods**: `String()`, `Summary()`, `IsExpense()`, `IsIncome()` — all value receivers

---

### 2. User Authentication

| Command | Description | File |
|---------|-------------|------|
| `signup` | Register a new user | `cmd/signup.go` |
| `login` | Authenticate and save session token | `cmd/login.go` |

- **Bcrypt password hashing** — industry-standard password security
- **Session tokens** stored at `~/.expensevault/token` for CLI persistence
- **TUI auth flow** — Sign Up / Log In / Exit menu with password masking (`EchoPassword`)
- Auth utilities in `cmd/auth_utils.go`

---

### 3. Multi-Database Support

| Database | Driver | Config Key | File |
|----------|--------|------------|------|
| **SQLite** | `modernc.org/sqlite` | `DB_TYPE=sqlite` | `db/sqlite.go` |
| **MySQL** | `go-sql-driver/mysql` | `DB_TYPE=mysql` | `db/mysql.go` |
| **Supabase (PostgreSQL)** | `lib/pq` | `DB_TYPE=supabase` | `db/supabase.go` |

- Unified `Store` struct with `isPostgres` flag for query rebinding (`?` → `$1`)
- `convertDSN()` helper handles Supabase pooler's dotted usernames
- Auto-creates tables (`users`, `transactions`, `budgets`) on startup
- Environment-based configuration via `.env` file

---

### 4. Budget Management

| Command | Description | File |
|---------|-------------|------|
| `budget` | Set monthly spending targets per category | `cmd/budget.go` |

- **Per-category monthly budgets** stored in DB
- **Dashboard integration**: Budget vs actual spending comparison
- **Smart insights**: Automatic alerts when budget is exceeded

---

### 5. Reporting System

| Command | Description | File |
|---------|-------------|------|
| `report` | Generate financial reports | `cmd/report.go` |

**3 Report Types** (`services/reporter.go`):

| Type | Description |
|------|-------------|
| **MonthlyReporter** | Groups transactions by month (YYYY-MM) |
| **CategoryReporter** | Groups by category with totals |
| **YearlyReporter** | Aggregates by year |

- All reporters implement the `Reporter` interface (polymorphism)
- Thread-safe concurrent report generation using `sync.Mutex` and `runtime.NumCPU()`
- **Transaction filters**: `TransactionFilter` function type with `CombineFilters()` for composable filtering
- Tab-switchable in TUI Reports view

---

### 6. Data Import/Export

| Command | Description | File |
|---------|-------------|------|
| `export` | Export transactions to CSV | `cmd/exportcmd.go` |
| `import` | Import transactions from CSV | `cmd/importcmd.go` |
| `backup` | Backup all transactions to JSON | `cmd/backup.go` |
| `restore` | Restore transactions from JSON backup | `cmd/restore.go` |

**Export Engine** (`export/`):
- `Exporter` interface with CSV and JSON implementations
- `CSVExporter` — full CSV serialization with headers
- `JSONExporter` / `JSONImporter` — JSON backup/restore
- `BulkInsert()` for efficient batch database writes

---

### 7. AI-Powered Natural Language Queries

| Command | Description | File |
|---------|-------------|------|
| `ask` | Query finances using natural language | `cmd/ask.go` |

- **Gemini API integration** (`services/llm.go`) using the `gemini-2.5-flash` model
- **Two-stage pipeline**:
  1. `GenerateSQL()` — Translates natural language → SQL query
  2. `SummarizeData()` — Formats raw DB results → conversational answer
- **Security**: User-scoped queries (`WHERE user_id = <USER_ID>`)
- Available in both CLI (`ask` command) and TUI (Ask AI view)
- Requires `GEMINI_API_KEY` environment variable

---

### 8. Smart Dashboard & Insights

- **Dashboard metrics**: Monthly income, total expenses, savings, expense/savings ratios
- **Visual progress bars**: Color-coded (green/yellow/red based on thresholds)
- **Category breakdown**: Per-category spending with budget targets
- **Smart insights** (`services/insights.go`):
  - Budget exceeded warnings
  - Shopping/food overspend detection (>30%/>25% of income)
  - Savings achievement recognition
  - Daily budget tips (random rotation)

---

### 9. Error Handling System

**Custom error types** (`models/errors.go`):

| Error Type | Purpose |
|------------|---------|
| `DatabaseError` | Wraps DB errors with operation context |
| `AuthError` | Authentication failure with reason |
| `ValidationError` | Field-level validation failures |

- **Sentinel errors**: `ErrNotFound`, `ErrDuplicateUser`, `ErrUnauthorized`
- `errors.Is()` / `errors.As()` for error chain inspection
- `ValidateAll()` — variadic batch validation with `fmt.Errorf %w` wrapping

---

### 10. Type System & Data Operations

**Custom types** (`models/types.go`):
- `Rupees` — currency type wrapping `float64`
- `Category` — string-based enum with validation
- `TransactionType` — `income` / `expense`
- `ZeroValueDemo` — demonstrates Go's zero value semantics

**Collection operations**:
- `FilterTransactions()` — callback-based filtering
- `DeleteTransactionFromSlice()` — slice deletion
- `GroupByCategory()` — map-based grouping
- `BuildCategorySummary()` — 2D slice generation
- `SliceFirstN()` — slice windowing

---

### 11. HTTP API & Data Sync

| Endpoint | Description | File |
|----------|-------------|------|
| `/health` | Health check (public) | `api/server.go` |
| `/sync` | Sync transactions (protected) | `api/server.go` |

- `sync` CLI command (`cmd/sync.go`) pushes local transactions to server
- JSON-based REST API with `SyncPayload` model (`models/sync.go`)

---

### 12. Configuration System

**Environment-based config** (`utils/config.go`):
- `.env` file support via `godotenv`
- Supports `DB_TYPE`, `SUPABASE_DSN`, `MYSQL_*`, `GEMINI_API_KEY`
- Loaded in `cmd/root.go` via Cobra's `PersistentPreRunE`

---

### 13. Interactive TUI (Bubble Tea)

**6 Views** (`tui/app.go`):

| View | Description |
|------|-------------|
| `ViewAuthMenu` | Sign up / Log in / Exit |
| `ViewSignup` | User registration form |
| `ViewLogin` | Authentication form |
| `ViewDashboard` | KPIs, progress bars, insights, anomalies |
| `ViewTransactions` | Full transaction list with filtering |
| `ViewAddForm` | 6-field transaction input form |
| `ViewReports` | Monthly/Category/Yearly reports |
| `ViewAsk` | AI-powered natural language queries |

- Keyboard navigation: `1-5` for views, `↑/↓` for selection, `Tab` for fields

---

## 🚀 Advanced Features (Newly Added)

### 14. HTTP Middleware Chain

**File:** `api/middleware.go`

| Layer | Description | Pattern |
|-------|-------------|---------|
| **Structured Logger** | JSON request logging with `log/slog` | `responseRecorder` captures status codes |
| **Rate Limiter** | Token-bucket per-IP rate limiting | `sync.Mutex` + background goroutine cleanup |
| **JWT Auth** | Bearer token validation | `context.WithValue` for user injection |

- Routes: `/health` (logging only), `/sync` (all 3 layers)
- Token-bucket: 10 req/sec, burst of 20, auto-cleanup of stale buckets every minute

---

### 15. Concurrent Worker Pool

**File:** `services/workerpool.go`

- Generic `Job` / `Result` types with task functions
- Configurable worker count with fan-out/fan-in via channels
- `context.Context` for graceful shutdown
- `ProcessBatch()` convenience function for full pipeline
- Structured logging with `slog` for job tracking

---

### 16. Anomaly Detection

**File:** `services/anomaly.go`

- **Z-Score Analysis**: Flags transactions > 2σ above category mean
- Requires minimum 3 data points per category
- Sorted by severity (highest z-score first)
- Displays up to 3 anomaly alerts on dashboard with `⚠️` warnings

---

### 17. Spending Trend Analysis

**File:** `services/anomaly.go`

- Compares recent vs older transaction averages (split at midpoint)
- Detects spending increases (📈 UP) or decreases (📉 DOWN) beyond ±15%
- Displays trend indicator on dashboard

---

### 18. Predictive Budgeting (End-of-Month Forecast)

**File:** `services/anomaly.go`

- `PredictEndOfMonth()` — linear projection based on daily burn rate
- Calculates: daily burn rate, projected total expenses, projected savings
- **Confidence scoring**:
  - 🟢 **High** — 15+ transactions, 15+ days elapsed
  - 🟡 **Medium** — 7+ transactions, 7+ days elapsed
  - 🔴 **Low** — insufficient data
- Color-coded display: green (healthy), yellow (tight), red (deficit)
- Displayed in `PredictionBoxStyle` on dashboard

---

### 19. Recurring Transaction Scheduler

**File:** `services/scheduler.go`

- **Background goroutine** with `time.Ticker` (hourly checks)
- Supports **daily**, **weekly**, **monthly** frequencies
- Auto-creates past-due transactions on startup
- `TransactionAdder` interface — decoupled from `db.Store`
- Graceful stop via channel signaling

---

### 20. Encrypted Backup/Restore

**File:** `services/crypto.go`

| Component | Detail |
|-----------|--------|
| **Algorithm** | AES-256-GCM (authenticated encryption) |
| **Key Derivation** | scrypt (N=32768, r=8, p=1) |
| **Format** | `[32-byte salt][12-byte nonce][ciphertext+tag]` |
| **Salt/Nonce** | Random per encryption (`crypto/rand`) |

**CLI Usage:**
```bash
# Encrypted backup
go run main.go backup --encrypt -o backup.enc

# Decrypted restore
go run main.go restore --decrypt -f backup.enc
```
- Password prompted securely via `golang.org/x/term` (no echo)

---

### 21. Custom Query Language (CQL) & Fuzzy Search

**File:** `services/queryparser.go`

| Operator | Example | Description |
|----------|---------|-------------|
| `cat:` | `cat:food` | Filter by category |
| `amt:>` | `amt:>500` | Amount greater than |
| `amt:<` | `amt:<100` | Amount less than |
| `date:` | `date:last-week` | Date range filter |
| Free text | `amazon` | Fuzzy match on description/category/notes |

**Supported date ranges:** `today`, `last-week`, `last-month`, `this-month`, or exact `YYYY-MM-DD`

**TUI Integration:**
- Press `/` to activate **fuzzy search** mode
- Press `:` to activate **CQL query** mode (e.g., `cat:food amt:>500 date:last-week`)
- Press `Esc` to deactivate search/query mode
- **Live filtering** — results update on every keystroke

---

### 22. Pane-Based TUI Redesign

**Files:** `tui/styles.go`, `tui/app.go`

```
┌──────────────────┐ ┌─────────────────────────────────────┐
│ 📂 NAVIGATION    │ │ 💰 Income  💸 Expenses  📊 Savings │
│                  │ │                                     │
│ ▸ 📊 Dashboard   │ │ BUDGET OVERVIEW                     │
│   📋 Transactions│ │ Income   [████████████████████] 100% │
│   ➕ Add         │ │ Expenses [██░░░░░░░░░░░░░░░░░]  10% │
│   📈 Reports     │ │                                     │
│   🤖 Ask AI      │ │ 🔍 ANOMALY DETECTION                │
│   🚪 Exit        │ │ ⚠️ Unusual Food expense...          │
│                  │ │                                     │
│ ─────────────    │ │ 📊 Spending is stable               │
│  📝 42 txns      │ │ 🔮 Predicted EOM Savings: ₹45,000   │
└──────────────────┘ └─────────────────────────────────────┘
 [1]Dash [2]Txns [3]Add [4]Reports [5]AI [/]Search [:]Query [q]Quit
```

**Key UI Components:**

| Component | Description | Style |
|-----------|-------------|-------|
| **Sidebar** | Fixed left navigation with icons + active highlighting | `SidebarStyle`, `SidebarActiveStyle` |
| **Main Pane** | Right content area with rounded border | `MainPaneStyle` |
| **Status Bar** | Bottom bar with hotkeys; transforms into search/query input | `StatusBarStyle`, `SearchBarStyle` |
| **KPI Boxes** | Horizontal Income/Expenses/Savings row | `IncomeBoxStyle`, `ExpenseBoxStyle`, `BalanceBoxStyle` |
| **Anomaly Box** | Orange-bordered warning alerts | `AnomalyBoxStyle` |
| **Prediction Box** | Purple-bordered EOM forecast | `PredictionBoxStyle` |
| **Progress Bars** | Color-coded budget usage bars | Dynamic styling (green/yellow/red) |

- Auth views render fullscreen (no sidebar)
- 10+ reusable lipgloss styles in `tui/styles.go`
- `lipgloss.JoinHorizontal` for sidebar + main content composition

---

## 📦 Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | TUI styling |
| `github.com/charmbracelet/bubbles` | TUI text input components |
| `github.com/joho/godotenv` | `.env` file loading |
| `golang.org/x/crypto/bcrypt` | Password hashing |
| `golang.org/x/crypto/scrypt` | Key derivation for encryption |
| `golang.org/x/term` | Secure password input |
| `github.com/golang-jwt/jwt/v5` | JWT authentication |
| `github.com/lib/pq` | PostgreSQL driver (Supabase) |
| `github.com/go-sql-driver/mysql` | MySQL driver |
| `modernc.org/sqlite` | Pure Go SQLite driver |

---

## 🧪 Testing

- **Unit tests**: `models/transaction_test.go` (31KB, comprehensive)
- **Service tests**: `services/reporter_test.go` (test filtering & reporting)
- Run: `go test ./...`

---

## 🏃 Quick Start

```bash
# Configure database
cp .env.example .env
# Edit .env with your database settings

# Run TUI
go run main.go tui

# Run HTTP server
go run main.go server

# CLI commands
go run main.go signup
go run main.go login
go run main.go add
go run main.go list
go run main.go report
go run main.go backup --encrypt
go run main.go ask "What did I spend the most on?"
```
