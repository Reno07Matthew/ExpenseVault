# ExpenseVault — Personal Finance CLI + Pane-Based TUI

ExpenseVault is a production-grade Go CLI application and interactive Terminal User Interface (TUI) for managing personal finances. It supports multi-database backends (Supabase/PostgreSQL, MySQL, SQLite), AI-powered natural language queries, predictive budgeting, and custom query language (CQL) filtering. 

Built to demonstrate comprehensive coverage of Go programming concepts including concurrency, middleware patterns, cryptography, and compiler-level type safety.

---

## Key Advanced Features

### Intelligence & Analytics
* **AI Financial Assistant:** Powered by Google Gemini 2.5 Flash, query your finances using natural language (`ask "What did I spend on food last month?"`).
* **Predictive Budgeting:** Forecasts end-of-month savings based on your daily burn rate with automated confidence scoring.
* **Structural Anomaly Detection:** Calculates Z-scores (σ > 2) to flag unusual spikes in category spending.
* **Spending Trend Analysis:** Compares historical vs recent transaction velocity to detect up/down patterns.

### TUI & Navigation
* **Pane-Based Dashboard:** A fully redesigned TUI featuring a fixed sidebar, command/status bar, and dynamic main content pane.
* **Fuzzy Search & CQL:** Press `/` to quickly fuzzy search descriptions, or `:` to use Custom Query Language (`cat:food amt:>500 date:last-week`) with live TUI filtering.
* **Real-time KPI Tracking:** Visual progress bars for income, expenses, and savings targets.

### Architecture & Security
* **HTTP Middleware Chain:** Production-grade `net/http` server with Structured Logging (slog), Token-bucket Rate Limiting, and JWT Authentication.
* **Concurrent Worker Pool:** Generic fan-out/fan-in goroutine pool for processing heavy analytical tasks safely.
* **Recurring transaction Scheduler:** Background `time.Ticker` goroutine that auto-creates periodic transactions (daily/weekly/monthly).
* **Encrypted Backups:** AES-256-GCM authenticated encryption with scrypt key derivation.
* **Multi-DB Storage Layer:** Seamless switching between SQLite (local), MySQL, and Supabase PostgreSQL.

---

## Quick Start

### 1. Configure Environment
```bash
cp .env.example .env
# Edit .env with your Supabase DSN, MySQL credentials, or SQLite path
# Add your GEMINI_API_KEY to enable 'Ask AI' features
```

### 2. Build and Run
```bash
# Build the binary
go build -o expensevault .

# Launch the interactive pane-based TUI
./expensevault tui

# Or use CLI commands directly
./expensevault add -t expense -a 250 -d "Lunch" -c Food
```

---

## TUI Hotkeys

| Key | Action |
|-----|--------|
| `1`-`5` | Switch Views (Dashboard, Txns, Add, Reports, AI) |
| `/` | Activate fuzzy search mode |
| `:` | Activate CQL query mode |
| `Tab` | Switch focus across form fields |
| `Esc` | Cancel search/query or drop focus |
| `q` | Quit application |

---

## CLI Commands

| Command | Description |
|---------|-------------|
| `tui` | Launch interactive terminal dashboard |
| `add` | Add a new income/expense transaction |
| `list` | List all transactions |
| `edit` | Edit an existing transaction by ID |
| `delete` | Delete a transaction by ID |
| `budget` | Set monthly spending targets per category |
| `report` | Generate reports (monthly / category / yearly) |
| `export` | Export data to CSV or JSON |
| `import` | Import data from CSV or JSON |
| `backup` | Backup to JSON (use `--encrypt` for AES-256-GCM) |
| `restore` | Restore from JSON (use `--decrypt` for AES-256-GCM) |
| `ask` | Query finances with AI (requires Gemini API Key) |
| `signup` | Create user account (bcrypt) |
| `login` | Authenticate (bcrypt + JWT) |
| `server` | Start HTTP API server with rate limiting |
| `sync` | Sync transactions to remote HTTP server |

---

## Project Structure

```
expenseVault/
├── main.go                  # Entry point
├── api/                     # HTTP server & middleware chain (JWT, Rate limit, slog)
├── cmd/                     # CLI commands (Cobra framework)
├── db/                      # Multi-database storage (SQLite, MySQL, Supabase)
├── export/                  # CSV/JSON import/export engine
├── models/                  # Domain types, validation, errors
├── services/                # Business logic (Worker pool, AI, Crypto, Scheduler, CQL)
├── tui/                     # Interactive terminal UI (BubbleTea, Lipgloss)
├── diagrams/                # Mermaid architecture diagrams
└── utils/                   # Configuration & helpers
```

---

## Technical Documentation

For an in-depth dive into the codebase implementation, see our companion documentation:

* [**FEATURES.md**](./FEATURES.md): A comprehensive list of all 22 core and advanced features.
* [**CODE_SNIPPETS.md**](./CODE_SNIPPETS.md): A tour of the project's key functions, demonstrating goroutines, cryptography, and TUI pane logic.
* [**ARCHITECTURE.md**](./ARCHITECTURE.md): Mermaid diagrams covering the system architecture, middleware sequence, worker pool, and data pipelines.

---

## Testing and Benchmarks

The project includes over 60 unit tests and 11 benchmarks.

```bash
# Run all tests
go test ./... -v

# Run with benchmarks
go test ./models/ ./services/ -bench=. -v

# Run with coverage
go test ./models/ ./services/ -cover
```

## 💻 Cross-Platform Compatibility

The project runs on **Linux, macOS, and Windows** natively:
- Dependencies strictly bound to pure Go drivers (e.g., `modernc.org/sqlite` instead of `mattn/go-sqlite3`) to eliminate CGO compiler requirements.
- Uses `filepath` for OS-agnostic path resolving. 
- The BubbleTea TUI library provides out-of-the-box Windows Terminal support.
