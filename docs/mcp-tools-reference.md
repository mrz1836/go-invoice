# go-invoice MCP Tools Reference

This comprehensive reference documents all 21 MCP tools available for Claude Desktop integration with the go-invoice system. These tools enable natural language invoice management, data import/export, client management, and system configuration.

## Quick Reference

### Tool Categories

| Category                                  | Tools   | Description                         |
|-------------------------------------------|---------|-------------------------------------|
| [Invoice Management](#invoice-management) | 7 tools | Create, update, and manage invoices |
| [Data Import](#data-import)               | 3 tools | Import timesheet and external data  |
| [Data Export](#data-export)               | 3 tools | Generate and export documents       |
| [Client Management](#client-management)   | 5 tools | Manage client information           |
| [Configuration](#configuration)           | 3 tools | System setup and validation         |

### Most Common Tools

- `invoice_create` - Create new invoices
- `import_csv` - Import work hours from CSV
- `generate_html` - Generate invoice documents
- `client_create` - Add new clients
- `config_validate` - Validate system setup

## Invoice Management

Tools for creating, updating, and managing invoices throughout their lifecycle.

### invoice_create

Create a new invoice with optional work items and automatic calculations.

**Description**: Creates a new invoice in the system with comprehensive metadata, automatic ID generation, and optional work items. Supports flexible client identification and automatic calculation of totals.

**Parameters**:
- `client_identifier` (required): Client name, email, or ID
- `due_date` (optional): Invoice due date (formats: "2024-01-15", "January 15, 2024", "15 days")
- `work_items` (optional): Array of work items with description, hours, rate
- `notes` (optional): Additional notes or comments
- `tax_rate` (optional): Tax percentage (e.g., 8.5 for 8.5%)
- `project_name` (optional): Project or reference name

**Examples**:

1. **Simple invoice**:
```json
{
  "client_identifier": "Acme Corp",
  "due_date": "2024-02-15"
}
```

2. **Invoice with work items**:
```json
{
  "client_identifier": "john@example.com",
  "due_date": "30 days",
  "work_items": [
    {"description": "Web development", "hours": 40, "rate": 75},
    {"description": "Code review", "hours": 8, "rate": 85}
  ],
  "project_name": "Website Redesign",
  "tax_rate": 8.5
}
```

**Claude Conversation Examples**:
- "Create an invoice for Acme Corp due in 30 days"
- "I need to bill John Smith for 40 hours at $75/hour for web development work"
- "Create an invoice for project Alpha with 20 hours of consulting at $100/hour, due February 15th"

### invoice_list

List and filter invoices with flexible search criteria and sorting options.

**Description**: Retrieves invoices with powerful filtering capabilities including client search, date ranges, status filtering, and flexible sorting. Supports natural language date expressions and partial text matching.

**Parameters**:
- `client_filter` (optional): Filter by client name or email (partial match)
- `date_from` (optional): Start date for filtering (various formats supported)
- `date_to` (optional): End date for filtering
- `status_filter` (optional): Filter by status ("draft", "sent", "paid", "overdue")
- `limit` (optional): Maximum number of results (default: 50)
- `sort_by` (optional): Sort field ("date", "amount", "client", "status")
- `sort_order` (optional): Sort direction ("asc", "desc")

**Examples**:

1. **List recent invoices**:
```json
{
  "limit": 10,
  "sort_by": "date",
  "sort_order": "desc"
}
```

2. **Filter by client and status**:
```json
{
  "client_filter": "Acme",
  "status_filter": "overdue",
  "sort_by": "date"
}
```

3. **Date range query**:
```json
{
  "date_from": "2024-01-01",
  "date_to": "2024-01-31",
  "sort_by": "amount",
  "sort_order": "desc"
}
```

**Claude Conversation Examples**:
- "Show me all overdue invoices"
- "List invoices for Acme Corp from last month"
- "What are my top 5 invoices by amount this year?"

### invoice_show

Display detailed information about a specific invoice including all metadata and work items.

**Description**: Retrieves comprehensive invoice details including client information, work items, calculations, and status. Supports multiple identification methods for flexible access.

**Parameters**:
- `invoice_identifier` (required): Invoice ID, number, or description search

**Examples**:

1. **By invoice ID**:
```json
{
  "invoice_identifier": "INV-2024-001"
}
```

2. **By partial description**:
```json
{
  "invoice_identifier": "web development"
}
```

**Claude Conversation Examples**:
- "Show me invoice INV-2024-001"
- "Display the invoice for the web development project"
- "What's the status of the Acme Corp invoice?"

### invoice_update

Modify invoice metadata, settings, and status with validation and history tracking.

**Description**: Updates invoice information while preserving audit trails and validating changes. Supports partial updates and flexible identification methods.

**Parameters**:
- `invoice_identifier` (required): Invoice ID, number, or search term
- `due_date` (optional): New due date
- `status` (optional): New status ("draft", "sent", "paid", "cancelled")
- `notes` (optional): Updated notes
- `tax_rate` (optional): Updated tax rate
- `project_name` (optional): Updated project name

**Examples**:

1. **Update due date**:
```json
{
  "invoice_identifier": "INV-2024-001",
  "due_date": "2024-03-15"
}
```

2. **Mark as paid**:
```json
{
  "invoice_identifier": "Acme Corp web project",
  "status": "paid"
}
```

**Claude Conversation Examples**:
- "Mark invoice INV-2024-001 as paid"
- "Extend the due date for the Acme Corp invoice to March 15th"
- "Update the tax rate for my latest invoice to 10%"

### invoice_delete

Remove invoices from the system with safety confirmations and audit logging.

**Description**: Safely removes invoices with validation checks and audit trail maintenance. Includes safeguards against accidental deletion of important invoices.

**Parameters**:
- `invoice_identifier` (required): Invoice ID, number, or search term
- `confirm` (optional): Confirmation flag for safety (default: false)

**Examples**:

1. **Delete with confirmation**:
```json
{
  "invoice_identifier": "INV-2024-001",
  "confirm": true
}
```

**Claude Conversation Examples**:
- "Delete invoice INV-2024-001" (will prompt for confirmation)
- "Remove the draft invoice for test client"

### invoice_add_item

Add work items to existing invoices with automatic recalculation of totals.

**Description**: Adds new work items to invoices while maintaining calculation accuracy and audit trails. Supports various rate and hour formats.

**Parameters**:
- `invoice_identifier` (required): Invoice ID, number, or search term
- `description` (required): Work item description
- `hours` (required): Hours worked (supports decimal values)
- `rate` (required): Hourly rate or total amount
- `date` (optional): Date of work performed

**Examples**:

1. **Add consulting hours**:
```json
{
  "invoice_identifier": "INV-2024-001",
  "description": "Code review and optimization",
  "hours": 4.5,
  "rate": 85,
  "date": "2024-01-20"
}
```

**Claude Conversation Examples**:
- "Add 4 hours of testing work at $75/hour to invoice INV-2024-001"
- "I need to add a code review session to the Acme Corp invoice"

### invoice_remove_item

Remove specific work items from invoices with automatic total recalculation.

**Description**: Removes work items from invoices while maintaining data integrity and calculation accuracy. Supports flexible item identification.

**Parameters**:
- `invoice_identifier` (required): Invoice ID, number, or search term
- `item_identifier` (required): Item description or index number

**Examples**:

1. **Remove by description**:
```json
{
  "invoice_identifier": "INV-2024-001",
  "item_identifier": "code review"
}
```

2. **Remove by index**:
```json
{
  "invoice_identifier": "INV-2024-001",
  "item_identifier": "2"
}
```

**Claude Conversation Examples**:
- "Remove the code review item from invoice INV-2024-001"
- "Delete the second work item from my latest invoice"

## Data Import

Tools for importing timesheet data, client information, and external data into the invoice system.

### import_csv

Import timesheet data from CSV files with intelligent parsing and validation.

**Description**: Processes CSV timesheet files with flexible column mapping, data validation, and error reporting. Supports various CSV formats and automatically handles common variations.

**Parameters**:
- `file_path` (required): Path to CSV file
- `client_identifier` (optional): Default client for all entries
- `rate` (optional): Default hourly rate
- `date_format` (optional): Date format hint ("MM/dd/yyyy", "dd-MM-yyyy", etc.)
- `skip_header` (optional): Skip first row as header (default: true)
- `delimiter` (optional): CSV delimiter (default: auto-detect)

**Examples**:

1. **Basic import**:
```json
{
  "file_path": "/path/to/timesheet.csv"
}
```

2. **Import with defaults**:
```json
{
  "file_path": "/path/to/hours.csv",
  "client_identifier": "Acme Corp",
  "rate": 75,
  "date_format": "MM/dd/yyyy"
}
```

**Expected CSV Format**:
```csv
Date,Client,Description,Hours,Rate
2024-01-15,Acme Corp,Web development,8,75
2024-01-16,Acme Corp,Code review,4,85
```

**Claude Conversation Examples**:
- "Import the timesheet from my Downloads folder"
- "Load hours from timesheet.csv with a default rate of $85/hour"
- "Process the CSV file with European date format"

### import_validate

Validate CSV structure and data before import execution.

**Description**: Imports complex data structures from JSON files including invoices, clients, and work items. Provides comprehensive validation and error reporting.

**Parameters**:
- `file_path` (required): Path to JSON file
- `data_type` (required): Type of data ("invoices", "clients", "work_items")
- `merge_strategy` (optional): How to handle duplicates ("skip", "update", "error")
- `validate_only` (optional): Only validate without importing (default: false)

**Examples**:

1. **Import invoices**:
```json
{
  "file_path": "/path/to/invoices.json",
  "data_type": "invoices",
  "merge_strategy": "update"
}
```

**Claude Conversation Examples**:
- "Import invoice data from the JSON backup file"
- "Load client information from clients.json"

### import_preview

Preview import results without making any changes.

**Description**: Processes multiple timesheet files simultaneously with progress reporting and comprehensive error handling. Supports various file formats and naming patterns.

**Parameters**:
- `directory_path` (required): Directory containing timesheet files
- `file_pattern` (optional): File name pattern (e.g., "*.csv", "timesheet_*.xlsx")
- `default_rate` (optional): Default rate for entries without rates
- `parallel_processing` (optional): Enable parallel processing (default: true)

**Examples**:

1. **Import all CSV files**:
```json
{
  "directory_path": "/path/to/timesheets/",
  "file_pattern": "*.csv",
  "default_rate": 80
}
```

**Claude Conversation Examples**:
- "Import all timesheet files from the timesheets folder"
- "Batch process all CSV files in my Documents directory"

## Document Generation

Tools for generating and exporting invoice documents, reports, and data.

**Description**: Integrates with external time tracking, accounting, or project management systems to import data automatically.

**Parameters**:
- `source_system` (required): External system type ("toggl", "harvest", "clockwise")
- `api_key` (optional): API key for authentication
- `date_range` (optional): Date range for sync
- `sync_mode` (optional): Sync mode ("incremental", "full")

**Examples**:

1. **Sync from Toggl**:
```json
{
  "source_system": "toggl",
  "api_key": "your_api_key",
  "date_range": "last_week"
}
```

**Claude Conversation Examples**:
- "Sync my time entries from Toggl for last week"
- "Import data from Harvest for the current month"

## Data Export

Tools for generating and exporting invoice documents, reports, and data.

### generate_html

Generate professional HTML invoice documents ready for email or printing.

**Description**: Creates beautifully formatted HTML invoices using customizable templates. Supports various layouts, branding options, and responsive design.

**Parameters**:
- `invoice_identifier` (required): Invoice ID, number, or search term
- `template` (optional): Template name ("default", "modern", "minimal")
- `output_path` (optional): Where to save the file
- `include_css` (optional): Embed CSS in HTML (default: true)
- `company_logo` (optional): Path to company logo image

**Examples**:

1. **Generate with default template**:
```json
{
  "invoice_identifier": "INV-2024-001",
  "output_path": "/path/to/invoice.html"
}
```

2. **Modern template with branding**:
```json
{
  "invoice_identifier": "INV-2024-001",
  "template": "modern",
  "company_logo": "/path/to/logo.png"
}
```

**Claude Conversation Examples**:
- "Generate an HTML invoice for INV-2024-001"
- "Create a modern-styled invoice document for the Acme Corp project"
- "Export my latest invoice as HTML with the company logo"

### generate_summary

Create invoice summaries and reports.

**Description**: Generates high-quality PDF invoices from HTML templates with print optimization and digital signature support.

**Parameters**:
- `invoice_identifier` (required): Invoice ID, number, or search term
- `template` (optional): Template name
- `output_path` (optional): Where to save the PDF
- `paper_size` (optional): Paper size ("A4", "Letter", "Legal")
- `include_attachments` (optional): Include supporting documents

**Examples**:

1. **Generate PDF**:
```json
{
  "invoice_identifier": "INV-2024-001",
  "paper_size": "A4",
  "output_path": "/path/to/invoice.pdf"
}
```

**Claude Conversation Examples**:
- "Convert invoice INV-2024-001 to PDF"
- "Generate a PDF invoice for printing on A4 paper"

### export_data

Export invoice and client data to various formats for analysis or backup.

**Description**: Exports structured data to CSV format with customizable columns and filtering options. Perfect for accounting software integration or data analysis.

**Parameters**:
- `data_type` (required): Type of data ("invoices", "clients", "work_items", "summary")
- `date_range` (optional): Date range for export
- `columns` (optional): Specific columns to include
- `output_path` (optional): Where to save the CSV file

**Examples**:

1. **Export invoice summary**:
```json
{
  "data_type": "invoices",
  "date_range": "2024-01-01,2024-01-31",
  "columns": ["invoice_id", "client", "total", "status"]
}
```

**Claude Conversation Examples**:
- "Export all invoices from January to CSV"
- "Create a CSV file with client contact information"

## Client Management

Tools for managing client information, contacts, and relationships.

### client_create

Add new clients to the system with comprehensive contact information.

**Description**: Creates new client records with complete contact details, billing preferences, and relationship metadata. Supports flexible data entry and validation.

**Parameters**:
- `name` (required): Client or company name
- `email` (optional): Primary email address
- `phone` (optional): Phone number
- `address` (optional): Mailing address
- `billing_contact` (optional): Billing contact person
- `default_rate` (optional): Default hourly rate for this client
- `payment_terms` (optional): Payment terms ("Net 30", "Net 15", etc.)

**Examples**:

1. **Basic client**:
```json
{
  "name": "Acme Corporation",
  "email": "billing@acme.com",
  "default_rate": 85
}
```

2. **Complete client profile**:
```json
{
  "name": "Tech Solutions Inc",
  "email": "accounts@techsolutions.com",
  "phone": "+1-555-0123",
  "address": "123 Tech Street, Silicon Valley, CA 94101",
  "billing_contact": "Jane Smith",
  "default_rate": 95,
  "payment_terms": "Net 30"
}
```

**Claude Conversation Examples**:
- "Add a new client called Acme Corporation"
- "Create a client profile for Tech Solutions with email billing@techsolutions.com"
- "I need to add a client with a default rate of $120/hour"

### client_list

List and search clients with filtering and sorting capabilities.

**Description**: Retrieves client information with powerful search and filtering options. Supports partial name matching and various sorting criteria.

**Parameters**:
- `name_filter` (optional): Filter by client name (partial match)
- `email_filter` (optional): Filter by email address
- `limit` (optional): Maximum number of results
- `sort_by` (optional): Sort field ("name", "email", "created_date")
- `include_stats` (optional): Include invoice statistics (default: false)

**Examples**:

1. **List all clients**:
```json
{
  "limit": 50,
  "sort_by": "name"
}
```

2. **Search for specific client**:
```json
{
  "name_filter": "Acme",
  "include_stats": true
}
```

**Claude Conversation Examples**:
- "Show me all my clients"
- "Find clients with 'Tech' in the name"
- "List clients with their invoice statistics"

### client_update

Update client information and billing preferences.

**Description**: Modifies existing client records with validation and audit trail maintenance. Supports partial updates and flexible identification.

**Parameters**:
- `client_identifier` (required): Client name, email, or ID
- `name` (optional): Updated name
- `email` (optional): Updated email
- `phone` (optional): Updated phone
- `address` (optional): Updated address
- `default_rate` (optional): Updated default rate
- `payment_terms` (optional): Updated payment terms

**Examples**:

1. **Update contact info**:
```json
{
  "client_identifier": "Acme Corp",
  "email": "newbilling@acme.com",
  "phone": "+1-555-0199"
}
```

**Claude Conversation Examples**:
- "Update Acme Corp's email to newbilling@acme.com"
- "Change the default rate for Tech Solutions to $110/hour"

### client_delete

Remove clients from the system with safety checks.

**Description**: Safely removes client records after checking for associated invoices and providing appropriate warnings or transfer options.

**Parameters**:
- `client_identifier` (required): Client name, email, or ID
- `force` (optional): Force deletion even with existing invoices
- `transfer_to` (optional): Transfer existing invoices to another client

**Examples**:

1. **Safe deletion**:
```json
{
  "client_identifier": "Old Client LLC"
}
```

**Claude Conversation Examples**:
- "Delete the client 'Test Company'"
- "Remove the old client but transfer their invoices to the new one"

## Configuration

Tools for system configuration, settings management, and validation.

### config_validate

Validate system configuration and identify potential issues.

**Description**: Performs comprehensive system validation including configuration files, templates, data integrity, and dependency checks.

**Parameters**:
- `config_path` (optional): Path to configuration file
- `check_templates` (optional): Validate invoice templates (default: true)
- `check_data` (optional): Validate data integrity (default: true)
- `verbose` (optional): Detailed validation output (default: false)

**Examples**:

1. **Full validation**:
```json
{
  "verbose": true
}
```

2. **Quick config check**:
```json
{
  "config_path": "/path/to/config.json",
  "check_templates": false,
  "check_data": false
}
```

**Claude Conversation Examples**:
- "Validate my system configuration"
- "Check if everything is set up correctly"
- "Run a detailed system validation"

### config_show

Display current configuration with formatting options.

**Description**: Modifies system configuration with validation and backup creation. Supports various setting categories and types.

**Parameters**:
- `setting_path` (required): Configuration setting path (e.g., "invoice.default_template")
- `value` (required): New setting value
- `backup` (optional): Create backup before change (default: true)

**Examples**:

1. **Set default template**:
```json
{
  "setting_path": "invoice.default_template",
  "value": "modern"
}
```

2. **Update tax rate**:
```json
{
  "setting_path": "billing.default_tax_rate",
  "value": 8.5
}
```

**Claude Conversation Examples**:
- "Set the default invoice template to modern"
- "Change the tax rate to 10%"

### config_init

Initialize new configuration with guided setup.

**Description**: Retrieves configuration values with optional filtering and formatting. Useful for reviewing current settings and troubleshooting.

**Parameters**:
- `setting_path` (optional): Specific setting to retrieve
- `category` (optional): Configuration category ("invoice", "billing", "templates")
- `format` (optional): Output format ("json", "yaml", "table")

**Examples**:

1. **Get all settings**:
```json
{
  "format": "table"
}
```

2. **Get specific setting**:
```json
{
  "setting_path": "invoice.default_template"
}
```

**Claude Conversation Examples**:
- "Show me all configuration settings"
- "What's the current default tax rate?"
- "Display billing configuration"

## Error Handling and Troubleshooting

### Common Error Types

1. **Validation Errors**: Invalid parameters or missing required fields
2. **Not Found Errors**: Requested invoices, clients, or files don't exist
3. **Permission Errors**: Insufficient access to files or directories
4. **Format Errors**: Invalid date formats, CSV structure, or data types

### Troubleshooting Guide

1. **Tool Not Found**: Check tool name spelling and availability
2. **Parameter Validation Failed**: Review required parameters and formats
3. **File Access Issues**: Verify file paths and permissions
4. **Data Integrity Problems**: Run `config_validate` for system check

### Best Practices

1. **Use descriptive identifiers**: Prefer names over IDs when possible
2. **Validate before bulk operations**: Test with small datasets first
3. **Keep backups**: System automatically creates backups for destructive operations
4. **Monitor validation**: Regularly run system validation checks

## Integration Examples

### Complete Workflow Example

```json
// 1. Create a client
{
  "tool": "client_create",
  "parameters": {
    "name": "New Client Corp",
    "email": "billing@newclient.com",
    "default_rate": 95
  }
}

// 2. Import timesheet data
{
  "tool": "timesheet_import_csv",
  "parameters": {
    "file_path": "/path/to/january_hours.csv",
    "client_identifier": "New Client Corp"
  }
}

// 3. Create invoice from imported data
{
  "tool": "invoice_create",
  "parameters": {
    "client_identifier": "New Client Corp",
    "due_date": "30 days"
  }
}

// 4. Generate HTML invoice
{
  "tool": "invoice_generate_html",
  "parameters": {
    "invoice_identifier": "INV-2024-001",
    "template": "modern"
  }
}
```

### Claude Natural Language Examples

**User**: "I need to create an invoice for 40 hours of development work for Acme Corp at $85/hour"

**Claude Response**: I'll create an invoice for Acme Corp with your development work. Let me set that up for you.

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_identifier": "Acme Corp",
    "work_items": [
      {
        "description": "Development work",
        "hours": 40,
        "rate": 85
      }
    ],
    "due_date": "30 days"
  }
}
```

**User**: "Show me all overdue invoices sorted by amount"

**Claude Response**: I'll get your overdue invoices sorted by amount from highest to lowest.

```json
{
  "tool": "invoice_list",
  "parameters": {
    "status_filter": "overdue",
    "sort_by": "amount",
    "sort_order": "desc"
  }
}
```

This reference provides comprehensive documentation for all 21 MCP tools, enabling effective natural language interaction with the go-invoice system through Claude Desktop.
