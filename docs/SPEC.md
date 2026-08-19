# SPEC.md — Expense Tracker CLI

> **Status**: Locked  
> **Language**: Go  
> **Last updated**: 2026-08-19

---

## 1. Overview

A single-binary command-line expense tracker. Expenses are persisted to a JSON file on disk. All monetary values are stored as **integer cents** internally and displayed as `$XX.XX` to the user.

---

## 2. Data Model

### 2.1 Storage File

| Setting | Value |
|---|---|
| Default path | `$HOME/.expense-tracker/expenses.json` |
| Override | `--file <path>` global flag (see §4.1) |
| On first run | Create file and parent directories silently |
| On corrupt file | Print error to stderr, exit 1 |

### 2.2 JSON Schema

```json
{
  "next_id": 4,
  "budgets": {
    "2026-08": 50000
  },
  "expenses": [
    {
      "id": 1,
      "date": "2024-08-06",
      "description": "Lunch",
      "amount": 2000,
      "category": "food"
    }
  ]
}
```

| Field | Type | Notes |
|---|---|---|
| `next_id` | `int64` | Monotonically increasing; **never reused** after delete |
| `budgets` | `map[string]int64` (cents) | Keyed by `"YYYY-MM"`; `0` or missing key means budget is not set for that month |
| `expenses[].id` | `int64` | Assigned from `next_id`, then `next_id` is incremented |
| `expenses[].date` | `string` | ISO 8601 `YYYY-MM-DD`; local time |
| `expenses[].description` | `string` | 1–255 characters |
| `expenses[].amount` | `int64` (cents) | Must be > 0 |
| `expenses[].category` | `string` | Stored **exactly as the user typed** (see §2.3); default `"uncategorized"` |

### 2.3 Category Semantics

- Stored verbatim (user's casing is preserved: `"Food"`, `"FOOD"`, `"food"` are stored as typed).
- **Matching and filtering are case-insensitive** (e.g. `--category food` matches `"Food"`, `"FOOD"`, `"food"`).
- No separate category entity; categories are an attribute of an expense.
- No category listing command.

### 2.4 Amount Parsing & Display

| Direction | Rule |
|---|---|
| **Input** | Accept up to 2 decimal places (e.g. `20`, `20.5`, `20.50`). Reject `20.999`. |
| **Storage** | Multiply by 100, round to nearest integer, store as `int64` cents. |
| **Display** | `cents / 100` formatted as `$XX.XX` (always 2 dp, always `$` prefix). |
| **CSV export** | `cents / 100` formatted as `XX.XX` (no `$`). |

---

## 3. Global Flags

These flags are available on every subcommand via the root command.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--file` | `string` | `$HOME/.expense-tracker/expenses.json` | Override data file path |

---

## 4. Commands

### 4.1 `add`

Add a new expense.

```
expense-tracker add --description <string> --amount <decimal> [--category <string>] [--date <YYYY-MM-DD>]
```

| Flag | Required | Default | Description |
|---|---|---|---|
| `--description` | ✅ | — | Expense description (1–255 chars) |
| `--amount` | ✅ | — | Expense amount (positive, ≤ 2 dp) |
| `--category` | ❌ | `"uncategorized"` | Category label (stored verbatim) |
| `--date` | ❌ | today (local) | Override expense date (`YYYY-MM-DD`) |

**Stdout on success:**
```
Expense added successfully (ID: 3)
```

**Budget warning** (printed to **stdout**, after the success line, if a budget is set for the newly added expense's month/year and the month total now exceeds it):
```
Warning: you have exceeded your budget of $500.00 for August 2026.
```

**Validation errors → stderr, exit 1:**

| Condition | Message |
|---|---|
| Missing `--description` | cobra required-flag error |
| Missing `--amount` | cobra required-flag error |
| Empty description | `description must not be empty` |
| Description > 255 chars | `description must not exceed 255 characters` |
| `--amount` ≤ 0 | `amount must be greater than zero` |
| `--amount` > 2 dp | `amount must have at most 2 decimal places` |
| `--amount` not a number | cobra type error |
| Invalid `--date` format | `date must be in YYYY-MM-DD format` |

---

### 4.2 `update`

Update one or more fields of an existing expense. At least one optional flag must be provided. Unspecified flags leave the corresponding field unchanged.

```
expense-tracker update --id <int> [--description <string>] [--amount <decimal>] [--category <string>] [--date <YYYY-MM-DD>]
```

| Flag | Required | Description |
|---|---|---|
| `--id` | ✅ | ID of the expense to update |
| `--description` | ❌ | New description |
| `--amount` | ❌ | New amount |
| `--category` | ❌ | New category |
| `--date` | ❌ | New date (`YYYY-MM-DD`) — allows correcting date or backdating |

**Stdout on success:**
```
Expense updated successfully
```

**Validation errors → stderr, exit 1:**

| Condition | Message |
|---|---|
| `--id` not found | `expense with ID 5 not found` |
| No optional flags provided | `at least one of --description, --amount, --category, --date must be provided` |
| Same amount/description/date rules as `add` | (same messages as §4.1) |

---

### 4.3 `delete`

Delete an expense by ID.

```
expense-tracker delete --id <int>
```

**Stdout on success:**
```
Expense deleted successfully
```

**Errors → stderr, exit 1:**

| Condition | Message |
|---|---|
| `--id` not found | `expense with ID 5 not found` |

---

### 4.4 `list`

Display all expenses in a formatted table, optionally filtered by category, month, and/or year. All filter flags compose (AND logic).

```
expense-tracker list [--category <string>] [--month <1-12>] [--year <int>]
```

| Flag | Required | Default | Description |
|---|---|---|---|
| `--category` | ❌ | (none) | Filter to expenses whose category matches (case-insensitive) |
| `--month` | ❌ | (none) | Filter to a specific month (1–12) |
| `--year` | ❌ | current year (if `--month` is set) | Filter to a specific year |

**Rules:**
- `--year` without `--month` filters to that full year.
- `--month` without `--year` defaults to **current year**.
- Neither `--month` nor `--year` → all-time expenses.

**Stdout — table output:**
```
ID   Date        Description   Amount    Category
1    2024-08-06  Lunch         $20.00    food
2    2024-08-06  Dinner        $10.00    uncategorized
```

- Columns are padded using `text/tabwriter` for alignment.
- Rows are sorted by date ascending, then by ID ascending as a tiebreaker.
- **Empty result** (no expenses, or no expenses match the filter):
  ```
  No expenses found.
  ```

**Validation errors → stderr, exit 1:**

| Condition | Message |
|---|---|
| `--month` not in 1–12 | `month must be between 1 and 12` |
| `--year` ≤ 0 | `year must be a positive integer` |

---

### 4.5 `summary`

Show the total of all matching expenses. Optionally filter by month, year, and/or category. All filter flags compose (AND logic).

```
expense-tracker summary [--month <1-12>] [--year <int>] [--category <string>]
```

| Flag | Required | Default | Description |
|---|---|---|---|
| `--month` | ❌ | (none) | Filter to a specific month (1–12) |
| `--year` | ❌ | current year (if `--month` is set) | Filter to a specific year |
| `--category` | ❌ | (none) | Filter by category (case-insensitive) |

**Rules:**
- `--year` without `--month` is valid (shows full-year total).
- `--month` without `--year` defaults to **current year**.
- Neither `--month` nor `--year` → all-time total.
- A future month with zero matching expenses returns `$0.00` — no warning.

**Stdout examples:**

```
# No filters
Total expenses: $30.00

# Month + year
Total expenses for August 2026: $20.00

# Year only
Total expenses for 2026: $50.00

# Category filter
Total expenses for food: $20.00

# Month + category
Total expenses for August 2026 (food): $15.00
```

**Budget warning** (printed after total line, if a budget is set for that specific month (`YYYY-MM`) and `--month` is specified and total exceeds budget):
```
Warning: you have exceeded your budget of $500.00 for August 2026.
```

**Validation errors → stderr, exit 1:**

| Condition | Message |
|---|---|
| `--month` not in 1–12 | `month must be between 1 and 12` |
| `--year` ≤ 0 | `year must be a positive integer` |

---

### 4.6 `export`

Export expenses to a CSV file. Optionally filter by month and/or year. No category filter for export.

```
expense-tracker export --output <file> [--month <1-12>] [--year <int>]
```

| Flag | Required | Default | Description |
|---|---|---|---|
| `--output` | ✅ | — | Destination file path (created/overwritten) |
| `--month` | ❌ | (none) | Filter to a specific month |
| `--year` | ❌ | current year (if `--month` set) | Filter to a specific year |

**CSV format:**
```csv
ID,Date,Description,Amount,Category
1,2024-08-06,Lunch,20.00,food
2,2024-08-06,Dinner,10.00,uncategorized
```

- Header row always present.
- `Amount` column: decimal with exactly 2 dp, **no `$`**.
- Description is quoted if it contains a comma.
- Rows sorted same as `list` (date asc, ID asc).

**Stdout on success:**
```
Expenses exported to expenses.csv successfully
```

**Errors → stderr, exit 1:**

| Condition | Message |
|---|---|
| Missing `--output` | cobra required-flag error |
| Cannot write file | `failed to write file: <os error>` |

---

### 4.7 `budget set`

Set (or update) the budget for a specific calendar month.

```
expense-tracker budget set --amount <decimal> [--month <1-12>] [--year <int>]
```

| Flag | Required | Default | Description |
|---|---|---|---|
| `--amount` | ✅ | — | Monthly budget limit (positive, ≤ 2 dp) |
| `--month` | ❌ | current month (local) | Target month (1–12) |
| `--year` | ❌ | current year (local) | Target year |

**Stdout on success:**
```
Budget for August 2026 set to $500.00
```

To **clear** the budget for a specific month (disable warnings):

```
expense-tracker budget clear [--month <1-12>] [--year <int>]
```

| Flag | Required | Default | Description |
|---|---|---|---|
| `--month` | ❌ | current month (local) | Target month (1–12) |
| `--year` | ❌ | current year (local) | Target year |

**Stdout:**
```
Budget for August 2026 cleared
```

**Validation errors → stderr, exit 1:**

| Condition | Message |
|---|---|
| Missing `--amount` (`budget set`) | cobra required-flag error |
| `--amount` ≤ 0 | `amount must be greater than zero` |
| `--amount` > 2 dp | `amount must have at most 2 decimal places` |
| `--month` not in 1–12 | `month must be between 1 and 12` |
| `--year` ≤ 0 | `year must be a positive integer` |

---

## 5. Budget Warning Logic

A budget warning is shown when:
1. A budget has been set for the relevant calendar month (`YYYY-MM`), **and**
2. The sum of expenses for that calendar month exceeds its set budget.

Warning is triggered on:
- `add` — for the month of the newly added expense, after calculating the new monthly total.
- `summary --month <n>` — for the specified month and year, after printing the total.

Warning format (stdout, on its own line after the primary output):
```
Warning: you have exceeded your budget of $500.00 for August 2026.
```

---

## 6. Error & Exit Code Contract

| Outcome | Stream | Exit code |
|---|---|---|
| Success | stdout | `0` |
| User / validation error | stderr | `1` |
| I/O / system error | stderr | `1` |

Cobra's built-in flag errors also write to stderr and exit `1`.

---

## 7. Suggested Package Structure

```
expense-tracker/
├── main.go                  # entry point; calls cmd.Execute()
├── cmd/
│   ├── root.go              # root command, --file flag, PersistentPreRun loads store
│   ├── add.go
│   ├── update.go
│   ├── delete.go
│   ├── list.go
│   ├── summary.go
│   ├── export.go
│   └── budget.go            # budget set + budget clear subcommands
└── internal/
    ├── store/
    │   ├── store.go         # Load / Save JSON; Store struct
    │   └── store_test.go
    ├── expense/
    │   ├── expense.go       # Expense struct, Add/Update/Delete/Filter logic
    │   └── expense_test.go
    └── format/
        ├── format.go        # CentsToDisplay, ParseAmount, FormatTable, FormatCSV
        └── format_test.go
```

**Key design rules:**
- `cmd/` layer: flag parsing, calling `internal/` functions, printing results. No business logic here.
- `internal/store/`: only JSON I/O. Knows nothing about formatting.
- `internal/expense/`: pure business logic (filter, aggregate, validate). No I/O.
- `internal/format/`: all display/serialisation concerns (tabwriter, CSV, `$` formatting).

---

## 8. Out of Scope for this Implementation

- Multi-currency support (always `$`).
- Per-category budgets.
- Category listing command.
- Shell completion (cobra has it built-in; enable if desired later).
- Machine-readable output (`--json` flag).
