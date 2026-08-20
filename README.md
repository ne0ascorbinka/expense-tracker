# expense-tracker
A simple expense tracker application to manage your finances.

## Installation

Install the CLI using `go install`:

```bash
go install github.com/ne0ascorbinka/expense-tracker@latest
```

Ensure your `GOPATH/bin` or `GOBIN` directory is included in your `PATH`.

## Quick Start

```bash
# Add expenses
expense-tracker add --description "Lunch" --amount 20
expense-tracker add --description "Groceries" --amount 45.50 --category Food

# List expenses
expense-tracker list

# View summary
expense-tracker summary

# Set a monthly budget
expense-tracker budget set --amount 500 --month 8
```

