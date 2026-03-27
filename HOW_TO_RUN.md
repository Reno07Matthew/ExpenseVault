# How to Run ExpenseVault

## Build

```bash
cd expenseVault
go build -o expensevault .
```

## TUI (recommended)

```bash
./expensevault tui
```

The TUI starts with signup and login screens. After login you can access the dashboard.

## CLI Examples

```bash
./expensevault signup -u alice
./expensevault login -u alice

./expensevault add -t expense -a 250 -d "Lunch" -c Food
./expensevault list
./expensevault report --type monthly
```

## Supabase (PostgreSQL)

To use Supabase, set the following environment variables (or use a `.env` file):

```bash
export DB_TYPE=supabase
export SUPABASE_DSN="postgresql://postgres:[PASSWORD]@db.[PROJECT-ID].supabase.co:5432/postgres"
```

Then run the application as usual.
