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
