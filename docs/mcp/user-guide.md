# go-invoice MCP User Guide

Welcome to the complete user guide for go-invoice MCP integration. This guide provides practical examples, workflows, and best practices for using natural language to manage your invoices through Claude Desktop and Claude Code.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic Operations](#basic-operations)
3. [Advanced Workflows](#advanced-workflows)
4. [Tool Reference](#tool-reference)
5. [Business Scenarios](#business-scenarios)
6. [Tips and Best Practices](#tips-and-best-practices)

## Getting Started

### Your First Invoice

Let's create your first invoice using natural language:

```
"Create an invoice for Acme Corporation for 40 hours of web development work at $85 per hour, due in 30 days"
```

Claude will:
1. Check if Acme Corporation exists as a client
2. Create the client if needed
3. Create the invoice with calculated totals
4. Set the due date automatically
5. Generate a unique invoice number

Expected response:
```
I've created invoice INV-2025-001 for Acme Corporation:

• Client: Acme Corporation
• Work: 40 hours of web development
• Rate: $85.00/hour
• Subtotal: $3,400.00
• Total: $3,400.00
• Due Date: February 15, 2025
• Status: Draft

The invoice is ready. Would you like me to generate an HTML version for sending?
```

### Import Your First Timesheet

If you have existing timesheet data in CSV format:

```
"Import the timesheet data from ~/Downloads/january-hours.csv and create invoices for each client"
```

Expected CSV format:
```csv
Date,Client,Project,Hours,Rate,Description
2024-01-15,Acme Corp,Website,8,85,Frontend development
2024-01-16,Beta Inc,Mobile App,6,95,UI design
2024-01-17,Acme Corp,Website,4,85,Backend API
```

Claude will:
1. Parse the CSV file
2. Group hours by client
3. Create separate invoices for each client
4. Calculate totals automatically
5. Apply your default settings

## Basic Operations

### Client Management

#### Creating Clients
```
"Add a new client called 'Tech Startup Inc' with email contact@techstartup.com and address '123 Innovation Drive, Silicon Valley, CA 94025'"
```

#### Finding Clients
```
"Show me all my clients"
"Find clients with 'tech' in their name"
"List clients who haven't been invoiced in the last 60 days"
```

#### Updating Client Information
```
"Update Tech Startup Inc's email to billing@techstartup.com"
"Change Acme Corp's address to their new office location"
```

### Invoice Management

#### Creating Invoices
```
# Simple invoice
"Create an invoice for Beta Inc for $2,500 due next month"

# Detailed invoice with items
"Create an invoice for Gamma Corp with these items:
- 20 hours of consulting at $150/hour
- 10 hours of implementation at $120/hour
- $500 for expenses
Due in 15 days"

# Project-based invoice
"Create an invoice for the mobile app project with Delta Company, including 60 hours at $100/hour plus $1,200 in licensing fees"
```

#### Finding and Viewing Invoices
```
"Show me all unpaid invoices"
"List invoices created this month"
"Find the invoice for Project Alpha"
"Show invoice INV-2025-001 details"
```

#### Updating Invoices
```
"Mark invoice INV-2025-001 as paid"
"Add a 5% late fee to invoice INV-2025-002"
"Update the due date for invoice INV-2025-003 to next Friday"
"Add 8 more hours to the Acme Corp invoice"
```

### Data Import and Export

#### Importing Timesheet Data
```
"Import hours from ~/Desktop/weekly-timesheet.csv"
"Import the Excel file and preview the data before creating invoices"
"Load timesheet data and group by project instead of client"
```

#### Generating Documents
```
"Generate an HTML invoice for INV-2025-001"
"Create a PDF invoice for Acme Corp's latest invoice"
"Export all unpaid invoices to ~/Documents/Invoices/"
"Generate a professional invoice using the corporate template"
```

#### Data Export
```
"Export all client data to CSV format"
"Create a revenue report for Q1 2024"
"Export all invoices from last month for QuickBooks"
```

## Advanced Workflows

### Monthly Billing Process

Automate your monthly billing with this workflow:

```
# Step 1: Review unbilled time
"Show me all unbilled hours grouped by client"

# Step 2: Create invoices for clients with significant hours
"Create invoices for all clients with more than 20 unbilled hours"

# Step 3: Review and adjust
"Show me all draft invoices and their totals"

# Step 4: Generate documents
"Generate HTML invoices for all draft invoices and save to ~/Documents/Invoices/March-2024/"

# Step 5: Mark as sent
"Mark all March invoices as sent"
```

### Project-Based Invoicing

For project-based work with milestones:

```
# Track project progress
"Create an invoice for Mobile App Project Phase 1: $15,000 for completion of UI design and wireframes, due in 30 days"

# Milestone billing
"Create milestone invoices for the Website Redesign project:
- Phase 1 (Discovery): $5,000 due immediately
- Phase 2 (Design): $8,000 due in 30 days
- Phase 3 (Development): $12,000 due in 60 days"

# Project completion
"Create final invoice for Website Redesign project: remaining balance of $3,500 plus $500 hosting setup"
```

### Recurring Invoice Management

Set up recurring billing patterns:

```
# Monthly retainer
"Create a template for monthly retainer invoices for Acme Corp: $5,000/month for ongoing support, due on the 1st of each month"

# Quarterly billing
"Set up quarterly invoices for Beta Inc: $15,000 per quarter for consulting services"

# Annual subscriptions
"Create annual license invoice for Gamma Corp: $12,000 for software licensing, due January 1st each year"
```

### Multi-Currency Invoicing

Handle international clients:

```
# Set currency for specific invoice
"Create an invoice for London Client Ltd in GBP: £2,500 for consulting services"

# Convert existing invoice
"Convert invoice INV-2025-001 from USD to EUR using current exchange rates"

# Currency reporting
"Show me revenue by currency for the last quarter"
```

## Tool Reference

### Invoice Management Tools (7 tools)

#### invoice_create
Create new invoices with comprehensive options.

**Common Usage:**
```
"Create an invoice for [client] for [amount/hours] due [date]"
"Invoice [client] for [description] at [rate]"
```

**Advanced Parameters:**
- Custom tax rates
- Multiple work items
- Project references
- Payment terms
- Discounts and adjustments

#### invoice_list
Find and filter invoices.

**Usage:**
```
"List invoices [status] [timeframe] [client]"
"Show me [criteria] invoices"
```

**Filters:**
- Status: draft, sent, paid, overdue
- Date ranges: this month, last quarter, 2024
- Clients: specific client names
- Amount ranges: over $1000, under $500

#### invoice_show
Display detailed invoice information.

**Usage:**
```
"Show invoice [ID]"
"Display details for [client]'s latest invoice"
```

#### invoice_update
Modify existing invoices.

**Usage:**
```
"Update invoice [ID] [changes]"
"Mark invoice [ID] as [status]"
"Change [field] for invoice [ID]"
```

#### invoice_delete
Remove invoices from the system.

**Usage:**
```
"Delete invoice [ID]"
"Remove the test invoice"
```

#### invoice_add_item
Add work items to existing invoices.

**Usage:**
```
"Add [hours] hours to invoice [ID]"
"Add [description] for [amount] to [client]'s invoice"
```

#### invoice_remove_item
Remove work items from invoices.

**Usage:**
```
"Remove [item] from invoice [ID]"
"Delete the last item from [client]'s invoice"
```

### Data Import Tools (3 tools)

#### import_csv
Import timesheet data from CSV files.

**Usage:**
```
"Import [file path]"
"Load timesheet data from [file]"
```

**Supported Formats:**
- Standard CSV with headers
- Excel CSV exports
- Custom delimiter files
- Multiple sheet imports

#### import_validate
Validate import data before processing.

**Usage:**
```
"Validate the import file [path]"
"Check [file] for import errors"
```

#### import_preview
Preview changes before importing.

**Usage:**
```
"Preview import of [file]"
"Show what would be imported from [file]"
```

### Data Export Tools (3 tools)

#### generate_html
Generate professional HTML invoices.

**Usage:**
```
"Generate HTML for invoice [ID]"
"Create HTML invoice for [client]"
```

**Templates:**
- Professional template
- Minimal template
- Corporate template
- Custom templates

#### generate_summary
Create business summaries and reports.

**Usage:**
```
"Generate revenue summary for [period]"
"Create client report for [timeframe]"
```

#### export_data
Export data in various formats.

**Usage:**
```
"Export [data type] to [format]"
"Create [report type] for [period]"
```

**Formats:**
- CSV for spreadsheets
- JSON for integrations
- PDF for documents
- XML for accounting systems

### Client Management Tools (5 tools)

#### client_create
Add new clients to the system.

**Usage:**
```
"Create client [name] with email [email]"
"Add new client [name] at [address]"
```

#### client_list
List and search clients.

**Usage:**
```
"List all clients"
"Show clients [criteria]"
```

#### client_show
Display detailed client information.

**Usage:**
```
"Show client [name/ID]"
"Display [client] details"
```

#### client_update
Modify client information.

**Usage:**
```
"Update [client] [field] to [value]"
"Change [client]'s [contact info]"
```

#### client_delete
Remove clients from the system.

**Usage:**
```
"Delete client [name/ID]"
"Remove [client] from system"
```

### Configuration Tools (3 tools)

#### config_show
Display current configuration.

**Usage:**
```
"Show my configuration"
"Display business settings"
```

#### config_validate
Validate system setup.

**Usage:**
```
"Validate my configuration"
"Check system setup"
```

#### config_init
Initialize new configuration.

**Usage:**
```
"Initialize configuration"
"Set up new business profile"
```

## Business Scenarios

### Freelancer Workflow

**Daily Usage:**
```
# Morning: Check status
"Show me unpaid invoices and total outstanding amount"

# Log work
"Add 6 hours to Acme Corp invoice for API development work"

# End of day: Create invoices
"Create an invoice for today's work at Beta Inc: 8 hours of consulting at $120/hour"
```

**Weekly Process:**
```
# Import timesheet
"Import ~/Downloads/weekly-timesheet.csv and create invoices for clients with more than 10 hours"

# Generate documents
"Generate HTML invoices for all invoices created this week"

# Follow up
"Show me invoices due this week that haven't been paid"
```

### Consulting Firm Workflow

**Project Setup:**
```
# New project
"Create client 'Enterprise Corp' for the Digital Transformation project"

# Milestone invoicing
"Create milestone invoices for Enterprise Corp:
- Phase 1 (Analysis): $25,000 due immediately
- Phase 2 (Design): $35,000 due in 30 days
- Phase 3 (Implementation): $50,000 due in 60 days"
```

**Monthly Reporting:**
```
# Revenue analysis
"Generate revenue summary by client for last month"

# Team utilization
"Show billable hours by project for Q1"

# Outstanding receivables
"List all overdue invoices with aging analysis"
```

### Agency Workflow

**Client Onboarding:**
```
# New client setup
"Create client 'Marketing Agency XYZ' with retainer agreement: $10,000/month for ongoing services"

# Retainer invoicing
"Create monthly retainer invoice for Marketing Agency XYZ for March 2024"
```

**Project Billing:**
```
# Additional work
"Add project work to Marketing Agency XYZ: Website redesign $15,000, Social media strategy $5,000"

# Expense billing
"Add expenses to Startup Inc invoice: $500 for stock photos, $200 for domain registration"
```

### Enterprise Integration

**Accounting System Integration:**
```
# Export for QuickBooks
"Export all invoices from Q1 2024 in QuickBooks format"

# Export for SAP
"Generate XML export of all client data for SAP import"

# Generate reports
"Create detailed revenue report with tax breakdown for accounting"
```

**Compliance and Auditing:**
```
# Audit trail
"Show all changes made to invoice INV-2025-001"

# Compliance reporting
"Generate tax compliance report for fiscal year 2024"

# Data backup
"Create backup of all invoice data for the last year"
```

## Tips and Best Practices

### Natural Language Tips

**Be Specific:**
```
# Good
"Create an invoice for Acme Corp for 40 hours of web development at $85/hour due March 15"

# Less specific
"Make an invoice for Acme"
```

**Use Context:**
```
# Reference existing data
"Add 8 more hours to Acme Corp's current invoice"
"Update the latest Beta Inc invoice with a 5% discount"
```

**Combine Operations:**
```
# Multiple actions in one request
"Create an invoice for Gamma Corp for $5,000, generate the HTML version, and save it to ~/Documents/Invoices/"
```

### Workflow Optimization

**Batch Operations:**
```
# Process multiple items
"Create invoices for all clients with more than 20 unbilled hours"
"Generate HTML documents for all unpaid invoices"
"Mark all January invoices as sent"
```

**Template Usage:**
```
# Consistent formatting
"Use the professional template for all Enterprise Corp invoices"
"Apply the consulting template to this invoice"
```

**Data Validation:**
```
# Check before processing
"Validate the timesheet import before creating invoices"
"Preview changes before updating client information"
```

### Error Prevention

**Common Mistakes to Avoid:**
```
# Don't forget required information
"Create invoice for Acme" # Missing amount/hours
# Better: "Create invoice for Acme Corp for 40 hours at $85/hour"

# Use clear date formats
"Due next month" # Ambiguous
# Better: "Due March 15, 2024" or "Due in 30 days"

# Be specific with updates
"Update the invoice" # Which invoice?
# Better: "Update invoice INV-2025-001 status to paid"
```

**Data Consistency:**
```
# Use consistent client names
"Acme Corporation" vs "Acme Corp" vs "ACME"
# Pick one format and stick with it

# Standardize rate formats
"$85/hour" vs "85 dollars per hour" vs "$85.00/hr"
# Use consistent formatting
```

### Performance Optimization

**Large Dataset Handling:**
```
# Use date ranges for large queries
"Show invoices from January 2024" instead of "Show all invoices"

# Filter early
"List unpaid invoices over $1000" instead of "List all invoices"

# Batch operations
"Generate HTML for invoices INV-001 through INV-010"
```

**Resource Management:**
```
# Close unused features
"Disable caching if not needed for large imports"

# Monitor performance
"Show system status and performance metrics"
```

### Security Best Practices

**Data Protection:**
```
# Regular backups
"Create backup of all data before making bulk changes"

# Validate imports
"Always preview import data before processing"

# Access control
"Ensure sensitive client data is only accessible to authorized users"
```

**File Management:**
```
# Organize exports
"Save invoices to ~/Documents/Invoices/2024/March/"

# Clean up temporary files
"Remove draft invoices that are no longer needed"
```

### Integration Tips

**External Tools:**
```
# Accounting software
"Export in format compatible with [your accounting software]"

# Email integration
"Generate invoice with email-friendly formatting"

# CRM integration
"Format client data for import into [your CRM]"
```

**Automation:**
```
# Scheduled processes
"Set up monthly invoicing for retainer clients"

# Triggered actions
"Automatically generate HTML when invoice is created"

# Workflow chains
"Import timesheet, create invoices, generate documents, and export for accounting"
```

This comprehensive user guide provides the foundation for effective invoice management using natural language. Start with basic operations and gradually incorporate advanced workflows as you become more comfortable with the system.

## Next Steps

1. **Start Simple**: Begin with basic invoice creation and client management
2. **Import Your Data**: Load existing client and timesheet information
3. **Establish Workflows**: Create consistent processes for your business
4. **Customize Templates**: Design professional invoice layouts
5. **Integrate Systems**: Connect with your existing business tools
6. **Monitor and Optimize**: Track usage and improve efficiency

Ready to transform your invoice management? Start with your first natural language command and experience the power of conversational business automation!
