# 📚 go-invoice Tutorial: Complete Step-by-Step Guide

Welcome to the comprehensive tutorial for go-invoice! This guide will walk you through every aspect of using go-invoice, from initial setup to advanced workflows. Whether you're a freelancer, consultant, or small business owner, this tutorial will help you master invoice management and time tracking.

## 🎯 Table of Contents

1. [Getting Started](#getting-started)
2. [First-Time Setup](#first-time-setup)
3. [Client Management](#client-management)
4. [Time Tracking & CSV Import](#time-tracking--csv-import)
5. [Invoice Creation & Management](#invoice-creation--management)
6. [Template Customization](#template-customization)
7. [Advanced Workflows](#advanced-workflows)
8. [Automation & Scripting](#automation--scripting)
9. [Troubleshooting](#troubleshooting)
10. [Best Practices](#best-practices)

---

## Getting Started

### What is go-invoice?

go-invoice is a powerful command-line application that helps you:
- Manage client information
- Import time tracking data from CSV files
- Generate professional HTML invoices
- Track invoice status (draft, sent, paid, etc.)
- Automate repetitive billing tasks

### Prerequisites

Before we begin, make sure you have:
- **Go 1.24 or later** installed on your system
- Basic familiarity with command-line interfaces
- Time tracking data (CSV format) or manual entry capability

### Installation

**Option 1: Install from source (recommended for development)**
```bash
git clone https://github.com/mrz/go-invoice.git
cd go-invoice
make build-go
sudo make install
```

**Option 2: Install via Go**
```bash
go install github.com/mrz/go-invoice@latest
```

**Verify installation:**
```bash
go-invoice --version
```

---

## First-Time Setup

### Configuration Wizard

The easiest way to get started is with the interactive setup wizard:

```bash
go-invoice config setup
```

This will guide you through configuring:

1. **Business Information**
   - Business name
   - Address (supports multi-line)
   - Email and phone
   - Website (optional)

2. **Financial Settings**
   - Default currency (USD, EUR, GBP, etc.)
   - Tax/VAT rate (as decimal, e.g., 0.10 for 10%)
   - Payment terms (e.g., "Net 30", "Due on receipt")

3. **Invoice Settings**
   - Invoice number prefix (e.g., "INV", "2024")
   - Starting number (e.g., 1000)
   - Data storage location

### Manual Configuration

If you prefer manual setup, create a `.env` file in your working directory:

```bash
# Business Information
BUSINESS_NAME=Acme Consulting LLC
BUSINESS_ADDRESS=123 Main Street\nSuite 100\nAnytown, ST 12345
BUSINESS_EMAIL=billing@acmeconsulting.com
BUSINESS_PHONE=+1-555-123-4567
BUSINESS_WEBSITE=https://acmeconsulting.com
PAYMENT_TERMS=Net 30

# Banking Information (optional)
BANK_NAME=First National Bank
BANK_ACCOUNT=****1234
BANK_ROUTING=021000021
PAYMENT_INSTRUCTIONS=Please remit payment within 30 days

# Invoice Settings
INVOICE_PREFIX=AC
INVOICE_START_NUMBER=2024001
CURRENCY=USD
VAT_RATE=0.08

# Storage Settings
DATA_DIR=./invoice-data
AUTO_BACKUP=true
RETENTION_DAYS=2555  # 7 years
```

### Verify Configuration

Check your configuration:

```bash
go-invoice config show
```

---

## Client Management

### Adding Your First Client

```bash
go-invoice client add \
  --name "TechStartup Inc" \
  --email "accounts@techstartup.com" \
  --address "456 Innovation Drive, Tech City, TC 67890" \
  --phone "+1-555-987-6543"
```

### Managing Multiple Clients

**List all clients:**
```bash
go-invoice client list
```

**View specific client details:**
```bash
go-invoice client show --name "TechStartup Inc"
```

**Update client information:**
```bash
go-invoice client update \
  --name "TechStartup Inc" \
  --email "billing@techstartup.com" \
  --phone "+1-555-987-6544"
```

**Deactivate a client (soft delete):**
```bash
go-invoice client deactivate --name "Old Client Corp"
```

### Client Organization Tips

1. **Use consistent naming**: "Company Name Inc" vs "Company Name, Inc."
2. **Keep contact info current**: Update emails and addresses promptly
3. **Use descriptive names**: Avoid abbreviations that might confuse you later

---

## Time Tracking & CSV Import

### Understanding CSV Format

go-invoice supports CSV files with this structure:

```csv
date,description,hours,rate
2024-01-15,Frontend development and UI improvements,8.5,125.00
2024-01-16,Backend API development and testing,7.25,135.00
2024-01-17,Code review and documentation updates,3.0,100.00
```

### Preparing Your CSV Data

**From popular time tracking tools:**

1. **Toggl**: Export as CSV, ensure columns match expected format
2. **Harvest**: Export detailed time report as CSV
3. **RescueTime**: Export productivity data (may need column adjustment)
4. **Manual tracking**: Create CSV in Excel, Google Sheets, or text editor

### CSV Import Process

**Step 1: Validate your CSV first**
```bash
go-invoice import csv timesheet-january.csv \
  --client "TechStartup Inc" \
  --validate \
  --dry-run
```

**Step 2: Import the data**
```bash
go-invoice import csv timesheet-january.csv \
  --client "TechStartup Inc" \
  --validate
```

**Step 3: Verify the import**
```bash
go-invoice workitem list --client "TechStartup Inc"
```

### Advanced CSV Options

**Custom date formats:**
```bash
go-invoice import csv european-format.csv \
  --client "EU Client" \
  --date-format "02/01/2006"
```

**Import with immediate invoice creation:**
```bash
go-invoice import csv monthly-hours.csv \
  --client "TechStartup Inc" \
  --create-invoice \
  --description "January 2024 Development Services" \
  --output january-invoice.html
```

### CSV Validation Rules

The system enforces these rules:
- **Hours**: 0.1 to 24.0 hours per entry
- **Rate**: $1.00 to $1,000.00 per hour
- **Description**: 3-500 characters, must be specific
- **Date**: Within last 2 years, not in future

**Common validation errors and fixes:**

❌ **"Description too generic"**
```csv
2024-01-15,Development,8,125.00  # Too generic
```

✅ **Fixed:**
```csv
2024-01-15,React component development for user dashboard,8,125.00
```

❌ **"Hours out of range"**
```csv
2024-01-15,Bug fixes,25,125.00  # Over 24 hours
```

✅ **Fixed:**
```csv
2024-01-15,Critical bug fixes and testing,8,125.00
2024-01-16,Additional bug fixes and deployment,8,125.00
```

---

## Invoice Creation & Management

### Creating Your First Invoice

**Method 1: From imported timesheet data**
```bash
# After importing CSV data
go-invoice invoice create \
  --client "TechStartup Inc" \
  --description "January 2024 Development Services" \
  --output january-2024-invoice.html
```

**Method 2: Manual work item entry**
```bash
# Create invoice with manual entries
go-invoice invoice create \
  --client "TechStartup Inc" \
  --description "Custom software development" \
  --output custom-invoice.html

# Add work items individually
go-invoice invoice add-work \
  --invoice INV-2024001 \
  --description "Database schema design and implementation" \
  --hours 12 \
  --rate 150.00 \
  --date 2024-01-15

go-invoice invoice add-work \
  --invoice INV-2024001 \
  --description "API endpoint development and testing" \
  --hours 16 \
  --rate 135.00 \
  --date 2024-01-16
```

### Invoice Lifecycle Management

**1. Draft Status (default)**
```bash
# List draft invoices
go-invoice invoice list --status draft
```

**2. Send Invoice**
```bash
go-invoice invoice send --invoice INV-2024001
```

**3. Mark as Paid**
```bash
go-invoice invoice mark-paid --invoice INV-2024001
```

**4. Handle Overdue Invoices**
```bash
# List overdue invoices
go-invoice invoice list --status overdue

# Generate overdue report
go-invoice report overdue --output overdue-january.html
```

### Invoice Management Commands

**List and filter invoices:**
```bash
# All invoices
go-invoice invoice list

# By status
go-invoice invoice list --status sent
go-invoice invoice list --status paid

# By client
go-invoice invoice list --client "TechStartup Inc"

# By date range
go-invoice invoice list --from 2024-01-01 --to 2024-01-31
```

**View invoice details:**
```bash
go-invoice invoice show --invoice INV-2024001
```

**Update invoice information:**
```bash
go-invoice invoice update \
  --invoice INV-2024001 \
  --description "Updated: January 2024 Development Services"
```

---

## Template Customization

### Understanding Templates

go-invoice uses Go's `html/template` system for generating invoices. Templates are HTML files with special placeholders for dynamic content.

### Default Template Features

The included template provides:
- Clean, professional layout
- Print-friendly design
- Automatic calculations
- Tax/VAT display
- Company branding area

### Creating Custom Templates

**1. Copy the default template:**
```bash
cp templates/invoice/default.html my-custom-invoice.html
```

**2. Customize the design:**

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Invoice {{.Number}} - {{.Config.Business.Name}}</title>
    <style>
        /* Custom CSS styles */
        body {
            font-family: 'Georgia', serif;
            color: #2c3e50;
            line-height: 1.6;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 2rem;
            margin-bottom: 2rem;
        }
        .company-name {
            font-size: 2.5rem;
            font-weight: bold;
            margin-bottom: 0.5rem;
        }
        .invoice-details {
            background: #f8f9fa;
            padding: 1.5rem;
            border-radius: 8px;
            margin-bottom: 2rem;
        }
        /* Add your custom styles here */
    </style>
</head>
<body>
    <div class="header">
        <div class="company-name">{{.Config.Business.Name}}</div>
        <div class="company-address">
            {{range $line := split .Config.Business.Address "\n"}}
                {{$line}}<br>
            {{end}}
        </div>
        <div class="company-contact">
            {{.Config.Business.Email}} | {{.Config.Business.Phone}}
        </div>
    </div>

    <div class="invoice-details">
        <h1>Invoice {{.Number}}</h1>
        <div class="invoice-meta">
            <div><strong>Date:</strong> {{.Date.Format "January 2, 2006"}}</div>
            <div><strong>Due Date:</strong> {{.DueDate.Format "January 2, 2006"}}</div>
            <div><strong>Status:</strong> <span class="status-{{.Status}}">{{.Status | title}}</span></div>
        </div>
    </div>

    <div class="client-info">
        <h2>Bill To:</h2>
        <div class="client-details">
            <div class="client-name">{{.Client.Name}}</div>
            <div class="client-address">
                {{if .Client.Address}}
                    {{range $line := split .Client.Address "\n"}}
                        {{$line}}<br>
                    {{end}}
                {{end}}
            </div>
            <div class="client-email">{{.Client.Email}}</div>
        </div>
    </div>

    <table class="work-items">
        <thead>
            <tr>
                <th>Date</th>
                <th>Description</th>
                <th>Hours</th>
                <th>Rate</th>
                <th>Amount</th>
            </tr>
        </thead>
        <tbody>
            {{range .WorkItems}}
            <tr>
                <td>{{.Date.Format "01/02"}}</td>
                <td>{{.Description}}</td>
                <td>{{printf "%.1f" .Hours}}</td>
                <td>${{printf "%.2f" .Rate}}</td>
                <td>${{printf "%.2f" .Total}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>

    <div class="totals">
        <div class="subtotal">
            <span>Subtotal:</span>
            <span>${{printf "%.2f" .Subtotal}}</span>
        </div>
        {{if gt .TaxAmount 0}}
        <div class="tax">
            <span>Tax ({{printf "%.1f" (mul .TaxRate 100)}}%):</span>
            <span>${{printf "%.2f" .TaxAmount}}</span>
        </div>
        {{end}}
        <div class="total">
            <span><strong>Total:</strong></span>
            <span><strong>${{printf "%.2f" .Total}}</strong></span>
        </div>
    </div>

    <div class="payment-terms">
        <h3>Payment Terms</h3>
        <p>{{.Config.Business.PaymentTerms}}</p>
        {{if .Config.Business.PaymentInstructions}}
        <p>{{.Config.Business.PaymentInstructions}}</p>
        {{end}}
    </div>
</body>
</html>
```

**3. Use your custom template:**
```bash
go-invoice invoice generate \
  --invoice INV-2024001 \
  --template my-custom-invoice.html \
  --output styled-invoice.html
```

### Template Variables Reference

Available variables in templates:

- **`.Number`** - Invoice number (e.g., "INV-2024001")
- **`.Date`** - Invoice date
- **`.DueDate`** - Payment due date
- **`.Status`** - Invoice status ("draft", "sent", "paid", etc.)
- **`.Client.Name`** - Client company name
- **`.Client.Email`** - Client email address
- **`.Client.Address`** - Client address (multi-line)
- **`.WorkItems`** - Array of work items with `.Date`, `.Description`, `.Hours`, `.Rate`, `.Total`
- **`.Subtotal`** - Invoice subtotal before tax
- **`.TaxRate`** - Tax rate as decimal (0.10 = 10%)
- **`.TaxAmount`** - Calculated tax amount
- **`.Total`** - Final total including tax
- **`.Config.Business.*`** - Business configuration values

### Template Functions

Built-in template functions:

- **`printf "%.2f" .Amount`** - Format numbers with decimals
- **`title .Status`** - Capitalize first letter
- **`split .Address "\n"`** - Split multi-line strings
- **`mul .TaxRate 100`** - Mathematical operations

---

## Advanced Workflows

### Monthly Billing Workflow

**Step 1: Organize your time tracking**
```bash
# Create monthly directory structure
mkdir -p invoices/2024/{01-January,02-February,03-March}
mkdir -p timesheets/2024
```

**Step 2: Import monthly timesheet**
```bash
go-invoice import csv timesheets/2024/january-hours.csv \
  --client "TechStartup Inc" \
  --validate
```

**Step 3: Generate monthly invoice**
```bash
go-invoice invoice create \
  --client "TechStartup Inc" \
  --description "January 2024 Development Services" \
  --output invoices/2024/01-January/techstartup-january.html
```

**Step 4: Send and track**
```bash
# Get the invoice ID from the output
INVOICE_ID="INV-2024001"

# Send the invoice
go-invoice invoice send --invoice $INVOICE_ID

# Set up reminder for follow-up
echo "Follow up on $INVOICE_ID in 15 days" >> calendar-reminders.txt
```

### Multi-Client Management

**Batch operations for multiple clients:**

```bash
#!/bin/bash
# monthly-invoicing.sh

MONTH="January-2024"
CLIENTS=("TechStartup Inc" "Marketing Agency LLC" "E-commerce Solutions")

for client in "${CLIENTS[@]}"; do
    echo "Processing $client..."

    # Import timesheet
    timesheet_file="timesheets/${client// /-}-january.csv"
    if [[ -f "$timesheet_file" ]]; then
        go-invoice import csv "$timesheet_file" --client "$client"

        # Create invoice
        invoice_file="invoices/${client// /-}-$MONTH.html"
        go-invoice invoice create \
            --client "$client" \
            --description "$MONTH Development Services" \
            --output "$invoice_file"

        echo "✅ Created invoice for $client"
    else
        echo "⚠️  No timesheet found for $client"
    fi
done
```

### Project-Based Billing

**Organize work by project:**

```bash
# Create project-specific invoices
go-invoice invoice create \
  --client "TechStartup Inc" \
  --description "Mobile App Development - Phase 1" \
  --project "mobile-app-v1" \
  --output project-invoices/mobile-app-phase1.html

# Filter timesheets by project
go-invoice import csv mobile-app-hours.csv \
  --client "TechStartup Inc" \
  --project "mobile-app-v1" \
  --description-filter "mobile"
```

---

## Automation & Scripting

### Automated Monthly Billing

Create a comprehensive automation script:

```bash
#!/bin/bash
# automated-billing.sh

set -e  # Exit on any error

# Configuration
CURRENT_MONTH=$(date +%B-%Y)
INVOICE_DIR="invoices/$(date +%Y/%m-%B)"
TIMESHEET_DIR="timesheets/$(date +%Y)"

# Create directories
mkdir -p "$INVOICE_DIR"
mkdir -p "$TIMESHEET_DIR"

# Function to process client
process_client() {
    local client="$1"
    local client_slug="${client// /-}"
    local timesheet="$TIMESHEET_DIR/${client_slug}-$(date +%m).csv"
    local invoice_file="$INVOICE_DIR/${client_slug}.html"

    echo "🔄 Processing $client..."

    if [[ ! -f "$timesheet" ]]; then
        echo "❌ Timesheet not found: $timesheet"
        return 1
    fi

    # Import timesheet
    if go-invoice import csv "$timesheet" --client "$client" --validate; then
        echo "✅ Imported timesheet for $client"
    else
        echo "❌ Failed to import timesheet for $client"
        return 1
    fi

    # Create invoice
    if go-invoice invoice create \
        --client "$client" \
        --description "$CURRENT_MONTH Development Services" \
        --output "$invoice_file"; then
        echo "✅ Created invoice: $invoice_file"
    else
        echo "❌ Failed to create invoice for $client"
        return 1
    fi

    # Get the latest invoice ID for this client
    local invoice_id
    invoice_id=$(go-invoice invoice list --client "$client" --status draft --format json | \
                jq -r '.[0].id // empty')

    if [[ -n "$invoice_id" ]]; then
        # Send invoice
        if go-invoice invoice send --invoice "$invoice_id"; then
            echo "📧 Sent invoice $invoice_id for $client"
        else
            echo "❌ Failed to send invoice for $client"
        fi
    else
        echo "⚠️  Could not find invoice ID for $client"
    fi
}

# Main execution
echo "🚀 Starting automated billing for $CURRENT_MONTH"

# Read client list from file
if [[ -f "clients.txt" ]]; then
    while IFS= read -r client; do
        [[ -n "$client" && ! "$client" =~ ^# ]] && process_client "$client"
    done < clients.txt
else
    echo "❌ clients.txt not found. Create it with one client name per line."
    exit 1
fi

echo "✅ Automated billing completed for $CURRENT_MONTH"

# Generate summary report
echo "📊 Generating summary report..."
go-invoice report summary --month "$(date +%Y-%m)" --output "$INVOICE_DIR/monthly-summary.html"
```

**Create clients.txt:**
```
TechStartup Inc
Marketing Agency LLC
E-commerce Solutions
# Add more clients here
```

### Scheduled Automation with Cron

Set up monthly automated billing:

```bash
# Edit crontab
crontab -e

# Add this line for first day of each month at 9 AM
0 9 1 * * /path/to/automated-billing.sh >> /var/log/go-invoice-automation.log 2>&1
```

### Integration with Git for Backup

```bash
#!/bin/bash
# backup-invoices.sh

# Backup invoice data to git repository
cd /path/to/invoice-data

# Add all new invoices and data
git add invoices/ data/

# Commit with timestamp
git commit -m "Monthly backup: $(date +%Y-%m-%d)"

# Push to remote repository
git push origin main

echo "✅ Invoice data backed up to git"
```

---

## Troubleshooting

### Common Issues and Solutions

#### Issue: "Configuration not found"

**Symptoms:**
```
Error: configuration not found. Run 'go-invoice config setup' first.
```

**Solutions:**
1. Run the setup wizard:
   ```bash
   go-invoice config setup
   ```

2. Check current configuration:
   ```bash
   go-invoice config show
   ```

3. Manual configuration via environment variables:
   ```bash
   export BUSINESS_NAME="Your Business"
   export BUSINESS_EMAIL="billing@business.com"
   # ... other required variables
   ```

#### Issue: CSV Import Failures

**Symptoms:**
```
Error: validation failed at line 5: description too generic
Error: invalid date format at line 3
```

**Solutions:**

1. **Generic descriptions:**
   ```bash
   # Bad
   2024-01-15,Development,8,125.00

   # Good
   2024-01-15,React component development for user dashboard,8,125.00
   ```

2. **Date format issues:**
   ```bash
   # Check supported formats
   go-invoice help import csv

   # Specify custom format
   go-invoice import csv file.csv --date-format "01/02/2006"
   ```

3. **Validate before importing:**
   ```bash
   go-invoice import csv timesheet.csv --validate --dry-run
   ```

#### Issue: Template Rendering Problems

**Symptoms:**
```
Error: template execution failed
Error: function "formatCurrency" not defined
```

**Solutions:**

1. **Test with default template first:**
   ```bash
   go-invoice invoice generate --invoice INV-001 --output test.html
   ```

2. **Validate custom template:**
   ```bash
   go-invoice template validate custom-template.html
   ```

3. **Check template syntax:**
   - Use `{{.Variable}}` not `{Variable}`
   - Use built-in functions: `printf`, `title`, `split`
   - Avoid undefined custom functions

#### Issue: Permission Errors

**Symptoms:**
```
Error: permission denied writing to ./data/invoices/
Error: failed to create directory
```

**Solutions:**

1. **Check directory permissions:**
   ```bash
   ls -la data/
   chmod 755 data/
   ```

2. **Use different data directory:**
   ```bash
   export DATA_DIR=$HOME/Documents/invoices
   go-invoice config show
   ```

3. **Run with proper permissions:**
   ```bash
   # Create data directory first
   mkdir -p ~/invoice-data
   export DATA_DIR=~/invoice-data
   ```

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
# Set debug environment variable
export DEBUG=true

# Run commands with verbose output
go-invoice invoice create --client "Test" --verbose

# Check log files
tail -f ~/.go-invoice/logs/debug.log
```

### Getting Help

1. **Built-in help:**
   ```bash
   go-invoice help
   go-invoice help invoice
   go-invoice help import csv
   ```

2. **Show current configuration:**
   ```bash
   go-invoice config show
   go-invoice status  # System status
   ```

3. **Validate setup:**
   ```bash
   go-invoice doctor  # Check system health
   ```

---

## Best Practices

### File Organization

**Recommended directory structure:**
```
~/invoicing/
├── data/                    # go-invoice data directory
├── timesheets/
│   ├── 2024/
│   │   ├── 01-january/
│   │   ├── 02-february/
│   │   └── ...
├── invoices/
│   ├── 2024/
│   │   ├── 01-January/
│   │   ├── 02-February/
│   │   └── ...
├── templates/
│   ├── standard-invoice.html
│   ├── detailed-invoice.html
│   └── simple-invoice.html
├── scripts/
│   ├── monthly-billing.sh
│   ├── backup-data.sh
│   └── generate-reports.sh
└── config/
    ├── .env
    └── clients.txt
```

### Naming Conventions

**Invoice numbering:**
- Use consistent prefixes: `INV-`, `2024-`, `AC-`
- Include year: `INV-2024001`, `2024-001`
- Sequential numbering: avoid gaps

**File naming:**
- Client files: `client-name-YYYY-MM.html`
- Timesheets: `client-name-YYYY-MM.csv`
- Templates: `descriptive-name-template.html`

**Client names:**
- Consistent format: "Company Name LLC"
- Avoid special characters in automation scripts
- Keep a master list in `clients.txt`

### Data Backup Strategy

1. **Daily backups:**
   ```bash
   # Add to daily cron
   0 23 * * * rsync -av ~/invoicing/ ~/backups/invoicing-$(date +%Y%m%d)/
   ```

2. **Git version control:**
   ```bash
   cd ~/invoicing
   git init
   echo "*.log" >> .gitignore
   git add .
   git commit -m "Initial invoice data"
   git remote add origin git@github.com:yourname/invoicing-backup.git
   git push -u origin main
   ```

3. **Cloud backup:**
   ```bash
   # Sync to cloud storage
   rclone sync ~/invoicing/ dropbox:invoicing/
   ```

### Security Considerations

1. **File permissions:**
   ```bash
   chmod 700 ~/invoicing/data  # Only owner access
   chmod 600 ~/invoicing/.env  # Config file protection
   ```

2. **Sensitive data:**
   - Never commit `.env` files to public repositories
   - Use environment variables in scripts
   - Regularly rotate access credentials

3. **Client data protection:**
   - Encrypt backups when possible
   - Use secure file transfer for invoice delivery
   - Follow data retention policies

### Performance Optimization

1. **Large datasets:**
   ```bash
   # Import large CSV files in chunks
   split -l 1000 large-timesheet.csv chunk-
   for chunk in chunk-*; do
       go-invoice import csv "$chunk" --client "Big Client"
   done
   ```

2. **Batch operations:**
   ```bash
   # Process multiple invoices efficiently
   go-invoice invoice send --all-drafts --batch-size 10
   ```

3. **Template caching:**
   - Reuse templates across invoices
   - Pre-compile frequently used templates

### Quality Assurance

1. **Always validate imports:**
   ```bash
   go-invoice import csv file.csv --validate --dry-run
   ```

2. **Review invoices before sending:**
   ```bash
   # Generate draft first
   go-invoice invoice create --client "Client" --output draft.html
   # Review in browser, then send
   go-invoice invoice send --invoice INV-001
   ```

3. **Regular data verification:**
   ```bash
   # Monthly data health check
   go-invoice doctor
   go-invoice report summary --month $(date +%Y-%m)
   ```

### Integration with Other Tools

**Accounting software:**
```bash
# Export for QuickBooks
go-invoice export qb --month 2024-01 --output january-qb.iif

# Export for Xero
go-invoice export xero --month 2024-01 --output january-xero.csv
```

**Email automation:**
```bash
# Send via email using external tool
go-invoice invoice generate --invoice INV-001 --output invoice.html
mutt -s "Invoice INV-001" -a invoice.html client@company.com < email-template.txt
```

**Time tracking integration:**
```bash
# Automatic Toggl sync
toggl-cli export --format csv --output toggl-export.csv
go-invoice import csv toggl-export.csv --client "Client"
```

---

## Conclusion

Congratulations! You've completed the comprehensive go-invoice tutorial. You now have the knowledge to:

✅ Set up and configure go-invoice for your business
✅ Manage clients and their information effectively
✅ Import time tracking data from CSV files
✅ Create and manage professional invoices
✅ Customize invoice templates for your brand
✅ Automate repetitive billing workflows
✅ Troubleshoot common issues
✅ Follow best practices for data management

### Next Steps

1. **Practice with sample data**: Use the examples in this tutorial
2. **Customize templates**: Create branded invoices for your business
3. **Set up automation**: Implement monthly billing workflows
4. **Integrate with your tools**: Connect with time tracking and accounting software

### Getting Support

- **Documentation**: Check the [README](../README.md) for quick reference
- **Examples**: Browse the [examples/](../examples/) directory
- **Issues**: Report bugs and request features on GitHub
- **Community**: Join discussions and share tips with other users

Happy invoicing! 🎉

---

*This tutorial covers go-invoice v1.0. Check for updates and new features regularly.*
