# 📚 go-invoice Examples

This directory contains comprehensive examples to help you get started with go-invoice quickly and efficiently. Whether you're a freelancer, consultant, or small business owner, these examples will guide you through common workflows and advanced scenarios.

## 📁 Directory Structure

```
examples/
├── README.md                           # This file
├── timesheet-standard.csv              # Basic CSV format example
├── timesheet-excel.csv                 # Excel export format example  
├── timesheet-tabs.tsv                  # Tab-separated format example
├── templates/                          # Custom invoice templates
│   ├── modern-invoice.html             # Modern gradient design template
│   └── minimal-invoice.html            # Clean minimal template
├── scripts/                            # Automation scripts
│   ├── monthly-billing.sh              # Monthly billing automation
│   └── setup-client.sh                 # New client setup helper
├── advanced/                           # Advanced usage examples
│   ├── multi-rate-timesheet.csv        # Different rates per task
│   ├── european-format.csv             # European date format (DD/MM/YYYY)
│   └── project-phases.csv              # Project-based billing
├── config/                             # Configuration examples
│   └── sample.env                      # Environment variables template
└── workflows/                          # Complete workflow examples
    └── complete-example.md             # Step-by-step tutorial
```

## 🚀 Quick Start Examples

### 1. Basic Invoice Creation

```bash
# Add a client
go-invoice client add --name "TechCorp" --email "billing@techcorp.com"

# Import timesheet and create invoice
go-invoice import csv examples/timesheet-standard.csv --client "TechCorp"
go-invoice invoice create --client "TechCorp" --output invoice.html
```

### 2. Using Custom Templates

```bash
# Generate invoice with modern template
go-invoice invoice generate \
  --invoice INV-001 \
  --template examples/templates/modern-invoice.html \
  --output modern-invoice.html

# Generate with minimal template  
go-invoice invoice generate \
  --invoice INV-001 \
  --template examples/templates/minimal-invoice.html \
  --output minimal-invoice.html
```

### 3. Automated Client Setup

```bash
# Set up new client with directory structure and sample files
examples/scripts/setup-client.sh "New Client LLC" "billing@newclient.com" "150.00"

# Run monthly billing for all clients
examples/scripts/monthly-billing.sh 01 2024
```

## 📄 CSV Format Examples

### Standard Format (`timesheet-standard.csv`)
```csv
date,description,hours,rate
2024-01-15,Backend API development,8.0,100.00
2024-01-16,Database optimization and queries,6.5,100.00
```

### Excel Export Format (`timesheet-excel.csv`)
```csv
Date,Hours Worked,Hourly Rate,Work Description
01/15/2024,8,100,"Backend API development"
01/16/2024,6.5,100,"Database optimization, query performance"
```

### Multi-Rate Format (`advanced/multi-rate-timesheet.csv`)
```csv
date,description,hours,rate
2024-01-15,Senior consultation,4.0,150.00
2024-01-15,Regular development,4.0,125.00
2024-01-16,Junior mentoring,3.0,75.00
```

## 🎨 Template Examples

### Modern Template Features
- Gradient background design
- Professional typography
- Print-friendly layout
- Status badges
- Responsive design

### Minimal Template Features
- Clean, simple design
- Black and white aesthetic
- Focused on content
- Lightweight HTML
- Classic professional look

## 🤖 Automation Scripts

### Monthly Billing Script (`scripts/monthly-billing.sh`)

Automates the complete monthly billing process:

```bash
# Process current month for all clients
./examples/scripts/monthly-billing.sh

# Process specific month/year
./examples/scripts/monthly-billing.sh 12 2023
```

Features:
- ✅ Validates all timesheets
- ✅ Imports data for each client
- ✅ Creates and sends invoices
- ✅ Generates summary reports
- ✅ Creates backups
- ✅ Detailed logging

### Client Setup Script (`scripts/setup-client.sh`)

Sets up a new client with complete workflow:

```bash
# Basic setup
./examples/scripts/setup-client.sh "Client Name"

# Full setup with email and rate
./examples/scripts/setup-client.sh "Client Name" "email@client.com" "125.00"
```

Creates:
- ✅ Client record in go-invoice
- ✅ Directory structure for timesheets/invoices
- ✅ Sample timesheet template
- ✅ Quick command scripts
- ✅ Client list for automation

## 🏗️ Advanced Examples

### Multi-Rate Billing
Handle different rates for different types of work:
- Senior consultation: $150/hour
- Regular development: $125/hour
- Code review: $100/hour
- Documentation: $75/hour

### Project Phase Billing
Organize work by project phases:
- Phase 1: Analysis & Design
- Phase 2: Core Development  
- Phase 3: Frontend & UI
- Phase 4: Testing & QA
- Phase 5: Deployment & Launch

### European Date Formats
Support for DD/MM/YYYY date formats:
```bash
go-invoice import csv examples/advanced/european-format.csv \
  --client "EU Client" \
  --date-format "02/01/2006"
```

## ⚙️ Configuration Examples

### Environment Variables (`config/sample.env`)

Complete configuration template with:
- Business information
- Banking details
- Invoice settings
- Payment terms
- Integration settings
- Security options

Copy and customize:
```bash
cp examples/config/sample.env .env
# Edit .env with your business details
```

## 📋 Complete Workflow Example

See `workflows/complete-example.md` for a comprehensive step-by-step tutorial covering:

1. **Client Setup** - Adding clients to the system
2. **Time Tracking** - Preparing CSV data
3. **Validation** - Checking data before import
4. **Import Process** - Loading timesheet data
5. **Invoice Creation** - Generating professional invoices
6. **Status Management** - Tracking sent/paid status
7. **Automation** - Scaling to multiple clients
8. **Integration** - Working with other tools

## 🛠️ Usage Patterns

### For Freelancers
- Use `timesheet-standard.csv` format
- Customize `modern-invoice.html` template
- Set up monthly automation with `monthly-billing.sh`

### For Consultants  
- Use `multi-rate-timesheet.csv` for different service rates
- Organize by `project-phases.csv` for complex projects
- Use `minimal-invoice.html` for corporate clients

### For Agencies
- Use automation scripts for multiple clients
- Set up team-specific rate structures
- Implement project-based billing workflows

## 🔧 Customization Tips

### Templates
1. Copy default template: `cp templates/invoice/default.html my-template.html`
2. Modify CSS styles and layout
3. Test with sample data: `go-invoice invoice generate --template my-template.html`
4. Add custom branding and colors

### Scripts
1. Copy automation scripts to your project
2. Modify paths and configuration variables
3. Add custom logic for your workflow
4. Set up cron jobs for automatic execution

### CSV Formats
1. Export from your time tracking tool
2. Map columns to go-invoice format
3. Test with `--dry-run` flag first
4. Create templates for consistency

## 📖 Additional Resources

- **Tutorial**: `docs/tutorial.md` - Comprehensive user guide
- **Documentation**: `README.md` - Full feature documentation
- **CLI Help**: `go-invoice help` - Built-in command help
- **Templates**: `templates/` - Default invoice templates

## 🤝 Contributing Examples

Have a useful example or workflow? Contributions are welcome!

1. Fork the repository
2. Add your example to the appropriate directory
3. Include documentation and comments
4. Test with sample data
5. Submit a pull request

## 💡 Tips for Success

1. **Start Simple**: Begin with basic CSV format and default template
2. **Validate Early**: Always use `--dry-run` and `--validate` flags
3. **Automate Gradually**: Start manual, then add automation as you scale
4. **Backup Regularly**: Use the backup features in automation scripts
5. **Test Templates**: Verify custom templates with sample data first
6. **Stay Organized**: Use consistent naming and directory structures

---

**Happy Invoicing!** 🎉

For more help, see the main documentation or run `go-invoice help` for built-in assistance.