# Complete go-invoice Workflow Example

This example demonstrates a complete workflow from client setup to invoice generation using go-invoice.

## Scenario

You're a freelance developer who has just completed a month of work for "TechStartup Inc" and need to generate an invoice. You've been tracking your time in a CSV file and want to automate the invoice creation process.

## Prerequisites

- go-invoice installed and configured
- Basic business configuration completed (`go-invoice config setup`)

## Step 1: Set Up the Client

First, let's add the client to our system:

```bash
# Add the client with basic information
go-invoice client add \
  --name "TechStartup Inc" \
  --email "accounting@techstartup.com" \
  --address "456 Innovation Drive, Tech Valley, CA 94000" \
  --phone "+1-555-TECH-123"

# Verify the client was added
go-invoice client show --name "TechStartup Inc"
```

## Step 2: Prepare Time Tracking Data

Create a CSV file with your time entries. Save this as `january-2024-timesheet.csv`:

```csv
date,description,hours,rate
2024-01-15,React component development for user dashboard,8.0,125.00
2024-01-16,Backend API integration and authentication,7.5,125.00
2024-01-17,Database optimization and query performance,6.0,135.00
2024-01-18,Unit testing and test coverage improvements,5.5,110.00
2024-01-19,Code review and documentation updates,4.0,100.00
2024-01-22,Bug fixes and error handling improvements,6.5,120.00
2024-01-23,Mobile responsive design implementation,7.0,115.00
2024-01-24,Performance optimization and caching,5.0,140.00
2024-01-25,Security audit and vulnerability fixes,4.5,150.00
2024-01-26,Deployment and production configuration,3.5,130.00
```

## Step 3: Validate the Timesheet

Before importing, let's validate the CSV format:

```bash
# Validate without importing (dry run)
go-invoice import csv january-2024-timesheet.csv \
  --client "TechStartup Inc" \
  --validate \
  --dry-run
```

Expected output:
```
✅ CSV validation passed
📊 Found 10 work items
⏱️  Total hours: 57.5
💰 Total amount: $7,125.00
🔍 Validation complete - no errors found
```

## Step 4: Import the Timesheet

Now import the validated data:

```bash
# Import the timesheet data
go-invoice import csv january-2024-timesheet.csv \
  --client "TechStartup Inc" \
  --validate
```

Expected output:
```
✅ Successfully imported 10 work items
📊 Total hours: 57.5 hours
💰 Total amount: $7,125.00
🎯 Ready to create invoice
```

## Step 5: Create the Invoice

Generate an invoice from the imported data:

```bash
# Create invoice with description
go-invoice invoice create \
  --client "TechStartup Inc" \
  --description "January 2024 Development Services - React Dashboard & API Integration" \
  --output january-2024-invoice.html
```

Expected output:
```
✅ Invoice created successfully
📄 Invoice ID: INV-2024001
💰 Total: $7,695.00 (includes 8% tax)
📁 Saved to: january-2024-invoice.html
```

## Step 6: Review and Send the Invoice

Open the generated HTML file to review:

```bash
# Open in default browser (macOS)
open january-2024-invoice.html

# Open in default browser (Linux)
xdg-open january-2024-invoice.html

# Or open in specific browser
firefox january-2024-invoice.html
```

If everything looks good, mark the invoice as sent:

```bash
# Mark invoice as sent
go-invoice invoice send --invoice INV-2024001
```

## Step 7: Track Payment

When the client pays, update the invoice status:

```bash
# Mark as paid when payment received
go-invoice invoice mark-paid --invoice INV-2024001
```

## Advanced Options

### Custom Templates

Use a custom template for different branding:

```bash
# Generate with custom template
go-invoice invoice generate \
  --invoice INV-2024001 \
  --template examples/templates/modern-invoice.html \
  --output january-2024-branded.html
```

### Multiple Clients

For multiple clients, you can use automation scripts:

```bash
# Set up a new client with automation
examples/scripts/setup-client.sh "Another Client Ltd" "billing@anotherclient.com" "140.00"

# Run monthly billing for all clients
examples/scripts/monthly-billing.sh 01 2024
```

### Different CSV Formats

go-invoice supports various CSV formats:

```bash
# European date format (DD/MM/YYYY)
go-invoice import csv european-timesheet.csv \
  --client "EU Client GmbH" \
  --date-format "02/01/2006"

# Custom column names
go-invoice import csv custom-export.csv \
  --client "Custom Client" \
  --columns "work_date,task_description,time_spent,billing_rate"
```

## Project Organization

For better organization, consider this directory structure:

```
invoicing/
├── data/
│   ├── clients.txt
│   ├── timesheets/
│   │   └── 2024/
│   │       └── 01-January/
│   │           ├── techstartup-inc.csv
│   │           └── another-client-ltd.csv
│   └── invoices/
│       └── 2024/
│           └── 01-January/
│               ├── techstartup-inc.html
│               └── another-client-ltd.html
├── templates/
│   ├── modern-invoice.html
│   └── minimal-invoice.html
└── scripts/
    ├── monthly-billing.sh
    └── backup-data.sh
```

## Monthly Workflow Automation

Create a monthly workflow:

1. **Week 4 of each month**: Prepare timesheets for all clients
2. **Month end**: Run validation for all timesheets
3. **1st of next month**: Generate and send all invoices
4. **Follow up**: Track payments and send reminders

```bash
# Create monthly workflow script
cat > monthly-workflow.sh << 'EOF'
#!/bin/bash
MONTH=$(date +%m)
YEAR=$(date +%Y)

echo "📅 Processing month: $MONTH/$YEAR"

# Validate all timesheets
echo "🔍 Validating timesheets..."
examples/scripts/monthly-billing.sh $MONTH $YEAR

# Generate summary report
echo "📊 Generating reports..."
go-invoice report summary --month $YEAR-$MONTH --output monthly-summary.html

echo "✅ Monthly workflow completed!"
EOF

chmod +x monthly-workflow.sh
```

## Troubleshooting

Common issues and solutions:

### CSV Import Fails
```bash
# Check CSV format
head -5 your-timesheet.csv

# Validate specific issues
go-invoice import csv your-timesheet.csv --client "Client" --validate --dry-run --verbose
```

### Invoice Generation Errors
```bash
# Check client exists
go-invoice client list

# Verify work items were imported
go-invoice workitem list --client "Your Client"
```

### Template Issues
```bash
# Test with default template first
go-invoice invoice generate --invoice INV-001 --output test.html

# Validate custom template
go-invoice template validate custom-template.html
```

## Integration with Other Tools

### Time Tracking Apps

**Toggl Export**:
1. Export CSV from Toggl
2. Rename columns: `Start date` → `date`, `Description` → `description`, `Duration` → `hours`
3. Add `rate` column manually
4. Import with go-invoice

**Harvest Export**:
1. Export detailed time report
2. Format should work directly with go-invoice
3. May need to adjust date format

### Accounting Software

**QuickBooks**:
```bash
# Export for QuickBooks import
go-invoice export quickbooks --month 2024-01 --output january-qb.iif
```

**Xero**:
```bash
# Export for Xero import
go-invoice export xero --month 2024-01 --output january-xero.csv
```

## Best Practices

1. **Consistent Naming**: Use consistent client names across all systems
2. **Regular Backups**: Back up your invoice data regularly
3. **Template Testing**: Test custom templates with sample data first
4. **Automation**: Use scripts for repetitive tasks
5. **Validation**: Always validate CSV files before importing
6. **Version Control**: Keep invoice templates and scripts in version control

## Complete Example Summary

This example showed you how to:
- ✅ Set up a client in go-invoice
- ✅ Prepare and validate CSV timesheet data
- ✅ Import work items from CSV
- ✅ Generate professional HTML invoices
- ✅ Track invoice status through the payment lifecycle
- ✅ Use automation for scaling to multiple clients
- ✅ Integrate with other tools and workflows

The entire process from timesheet to invoice typically takes less than 5 minutes per client, making it perfect for freelancers and small businesses who need efficient invoice management.