# 💸 Expense Tracker CLI

A fast, simple, and reliable command-line application to track your daily expenses, set monthly budgets, and export financial data directly from your terminal.

Built with [Go](https://go.dev/) and [Cobra](https://github.com/spf13/cobra), this project is an implementation of the [roadmap.sh Expense Tracker project](https://roadmap.sh/projects/expense-tracker).

---

## ✨ Features

- **Expense Management**: Add, update, view, and delete expenses quickly.
- **Categorization**: Tag expenses with custom categories (e.g. `food`, `bills`, `entertainment`) with case-insensitive filtering.
- **Monthly Budgets**: Set spending limits per month and receive automatic warnings when you exceed your budget upon adding expenses or viewing summaries.
- **Summaries & Statistics**: View total expenses all-time, by year, by month, or broken down by category.
- **CSV Export**: Export filtered or all-time expenses to CSV for spreadsheets or backups.
- **Reliable Local Storage**: Automatically saves data to a local JSON file (`~/.expense-tracker/expenses.json`) using integer cent precision to avoid floating-point rounding errors.
- **Zero Heavy Dependencies**: Lightweight, single-binary CLI.

---

## 📋 Prerequisites

- **Go**: Version `1.22` or higher ([Download Go](https://go.dev/dl/))
- **Git** (optional, for cloning the repository)

---

## 🚀 Installation

### Option 1: Quick Install via `go install` (Recommended)

Install the CLI binary directly to your `$GOBIN` (or `$GOPATH/bin`):

```bash
go install github.com/ne0ascorbinka/expense-tracker@latest
```

> **Note**: Make sure your Go binary directory (`$GOBIN` or `$GOPATH/bin` on macOS/Linux, `%USERPROFILE%\go\bin` on Windows) is added to your system `PATH`.

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/ne0ascorbinka/expense-tracker.git
cd expense-tracker

# Build the executable
go build -o expense-tracker .
```

---

## 💡 Quick Start & Usage Examples

### 1. Adding Expenses

Add an expense with a description and amount. Optionally specify a category and date (`YYYY-MM-DD` format, defaults to today):

```bash
# Add a basic expense (defaults to "uncategorized" and today's date)
expense-tracker add --description "Lunch" --amount 12.50

# Add an expense with a custom category
expense-tracker add --description "Groceries" --amount 45.00 --category "food"

# Add an expense with a specific past date
expense-tracker add --description "Gym Membership" --amount 30.00 --category "fitness" --date 2026-08-01
```

### 2. Listing Expenses

Display expenses in an aligned table format. Filter by category, month, or year:

```bash
# List all expenses
expense-tracker list

# Filter by category (case-insensitive)
expense-tracker list --category food

# Filter by month (1–12) for the current year
expense-tracker list --month 8

# Filter by specific month and year
expense-tracker list --month 8 --year 2026
```

### 3. Viewing Expense Summaries

View aggregated totals across all expenses or filter by time and category:

```bash
# Total expenses (all-time)
expense-tracker summary

# Total expenses for a specific month (current year)
expense-tracker summary --month 8

# Total expenses for a specific year
expense-tracker summary --year 2026

# Total expenses for a specific category
expense-tracker summary --category food

# Combined month and category
expense-tracker summary --month 8 --year 2026 --category food
```

### 4. Managing Monthly Budgets

Set spending limits for a calendar month. If an expense pushes your total over the limit, or when checking the summary, you will see a budget warning:

```bash
# Set budget for the current month
expense-tracker budget set --amount 500

# Set budget for a specific month and year
expense-tracker budget set --amount 750.00 --month 8 --year 2026

# Clear/disable budget for a specific month
expense-tracker budget clear --month 8 --year 2026
```

### 5. Updating an Expense

Update the description, amount, category, or date of an existing expense by ID:

```bash
# Update amount
expense-tracker update --id 1 --amount 15.00

# Update description and category
expense-tracker update --id 2 --description "Weekly Groceries" --category "groceries"
```

### 6. Deleting an Expense

Delete an expense by its ID:

```bash
expense-tracker delete --id 1
```

### 7. Exporting to CSV

Export expenses into a standard CSV file:

```bash
# Export all expenses
expense-tracker export --output expenses.csv

# Export expenses filtered by month and year
expense-tracker export --output august_expenses.csv --month 8 --year 2026
```

---

## ⚙️ Global Flags

You can customize the storage file location across all commands:

| Flag | Default | Description |
|---|---|---|
| `--file` | `~/.expense-tracker/expenses.json` | Custom path to the JSON storage file |

**Example:**
```bash
expense-tracker --file ./custom_expenses.json list
```

---

## 🧪 Running Tests

Execute the test suite using standard Go tooling:

```bash
go test -v ./...
```

---

## 🔗 Reference

- Project Specification & Ideas: [roadmap.sh Expense Tracker](https://roadmap.sh/projects/expense-tracker)
