# CSV Import Examples

This directory contains example CSV files demonstrating the supported formats for importing timesheet data into go-invoice.

## Supported Formats

### Standard CSV (RFC 4180)
**File**: `timesheet-standard.csv`
- Uses comma delimiters
- Standard header format
- ISO date format (YYYY-MM-DD)

```csv
date,hours,rate,description
2024-01-15,8.0,100.00,Backend API development
2024-01-16,6.5,100.00,Database optimization and queries
```

### Excel CSV Export
**File**: `timesheet-excel.csv`
- Uses comma delimiters with quoted fields
- US date format (MM/DD/YYYY)
- Descriptive header names

```csv
Date,Hours Worked,Hourly Rate,Work Description
01/15/2024,8,100,"Backend API development"
01/16/2024,6.5,100,"Database optimization, query performance"
```

### Tab-Separated Values (TSV)
**File**: `timesheet-tabs.tsv`
- Uses tab delimiters
- Alternative field names
- ISO date format

```
work_date	duration	billing_rate	task
2024-01-15	8.0	100.00	Backend API development
2024-01-16	6.5	100.00	Database optimization and queries
```

## Required Fields

All CSV files must contain these fields (header names can vary):

| Field | Alternative Names | Format | Description |
|-------|------------------|--------|-------------|
| date | work_date, day | YYYY-MM-DD, MM/DD/YYYY | Work date |
| hours | time, duration, hours_worked | Decimal number | Hours worked |
| rate | hourly_rate, billing_rate | Decimal number | Hourly rate in dollars |
| description | desc, task, work_description | Text | Work description |

## Usage Examples

### Validate CSV Format
```bash
go-invoice import validate examples/timesheet-standard.csv
```

### Create New Invoice from CSV
```bash
go-invoice import create examples/timesheet-standard.csv --client CLIENT_001
```

### Append to Existing Invoice
```bash
go-invoice import append examples/timesheet-standard.csv --invoice INV-001
```

### Dry Run (Validation Only)
```bash
go-invoice import create examples/timesheet-standard.csv --client CLIENT_001 --dry-run
```

### Force Specific Format
```bash
go-invoice import create examples/timesheet-excel.csv --client CLIENT_001 --format excel
```

## Format Detection

The import system automatically detects CSV format based on:
- Delimiter analysis (comma, tab, semicolon)
- Header field recognition
- Date format patterns

You can override automatic detection using the `--format` flag with values:
- `standard` - RFC 4180 CSV
- `excel` - Excel CSV export format
- `tab` - Tab-separated values
- `semicolon` - Semicolon-separated values

## Error Handling

The import system provides detailed error messages for:
- Invalid date formats
- Non-numeric hours or rates
- Missing required fields
- Duplicate work items
- Business rule violations

Use `--debug` flag for detailed logging during import operations.