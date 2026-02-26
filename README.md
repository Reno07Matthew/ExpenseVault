# ExpenseVault - Personal Finance CLI + TUI

ExpenseVault is a Go CLI application for managing personal finances with an interactive TUI, MySQL/SQLite storage, import/export, and bcrypt authentication.

## Quick Start

```bash
cd expenseVault

go build -o expensevault .

# Start TUI (includes signup -> login flow)
./expensevault tui

# Add a transaction
./expensevault add -t expense -a 250 -d "Lunch" -c Food

# List transactions
./expensevault list
```

## TUI Auth Flow

Running `./expensevault tui` starts an in-TUI auth flow:

1. Sign up (bcrypt hash stored in DB)
2. Log in (bcrypt verification)
3. Dashboard unlocks after login

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

The app creates `transactions` and `users` tables.

### 4) MySQL Workbench connection for another user

In MySQL Workbench:

- Hostname: MYSQL_HOST (public IP or DNS)
- Port: MYSQL_PORT
- Username: expensevault_app
- Password: strong_password
- Default Schema: expensevault

If MySQL is remote, open port 3306 and allow that host in `CREATE USER` (use `%` or a specific IP).

## Commands

- `add` / `list` / `edit` / `delete`
- `report` (monthly, category, yearly)
- `import` / `export` (csv, json)
- `backup` / `restore`
- `tui` (interactive dashboard)
- `signup` / `login`
- `server` / `sync` (basic HTTP endpoints)

## Tech Stack

- Go
- Cobra
- BubbleTea + Lipgloss
- SQLite (modernc.org/sqlite)
- MySQL (go-sql-driver/mysql)
- bcrypt
- JWT
