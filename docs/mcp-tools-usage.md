# go-invoice MCP Tools Usage Guide

This guide provides practical guidance for using the go-invoice MCP tools effectively with Claude Desktop. Learn best practices, common workflows, and optimization techniques for natural language invoice management.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic Workflows](#basic-workflows)
3. [Advanced Usage Patterns](#advanced-usage-patterns)
4. [Natural Language Best Practices](#natural-language-best-practices)
5. [Performance Optimization](#performance-optimization)
6. [Error Handling](#error-handling)
7. [Integration Patterns](#integration-patterns)

## Getting Started

### Prerequisites

Before using the MCP tools, ensure you have:

1. **go-invoice MCP server running**: The MCP server must be active and connected to Claude Desktop
2. **Valid configuration**: Run `config_validate` to verify system setup
3. **Proper file permissions**: Ensure Claude can access your file system for imports/exports
4. **Base data**: At least one client in the system for invoice creation

### Initial Setup Validation

Start by validating your system configuration:

```json
{
  "tool": "config_validate",
  "parameters": {
    "verbose": true
  }
}
```

### First Steps

1. **Create your first client**:
```json
{
  "tool": "client_create",
  "parameters": {
    "name": "Test Client",
    "email": "test@example.com",
    "default_rate": 75
  }
}
```

2. **Create a simple invoice**:
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_identifier": "Test Client",
    "due_date": "30 days"
  }
}
```

3. **Generate the invoice document**:
```json
{
  "tool": "invoice_generate_html",
  "parameters": {
    "invoice_identifier": "INV-2024-001"
  }
}
```

## Basic Workflows

### 1. Manual Invoice Creation

**Scenario**: Creating invoices manually with work items

**Steps**:
1. Ensure client exists or create new client
2. Create invoice with work items
3. Generate document for delivery

**Example Conversation**:

**User**: "I need to bill Acme Corp for 40 hours of web development at $85/hour, due in 30 days"

**Process**:
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_identifier": "Acme Corp",
    "work_items": [
      {
        "description": "Web development",
        "hours": 40,
        "rate": 85
      }
    ],
    "due_date": "30 days",
    "project_name": "Website Project"
  }
}
```

### 2. Timesheet Import Workflow

**Scenario**: Importing hours from CSV and creating invoices

**Steps**:
1. Prepare CSV timesheet file
2. Import timesheet data
3. Create invoices from imported data
4. Generate documents

**Example Process**:

```json
// Step 1: Import timesheet
{
  "tool": "timesheet_import_csv",
  "parameters": {
    "file_path": "/Users/username/timesheets/january.csv",
    "date_format": "MM/dd/yyyy"
  }
}

// Step 2: Create invoice for specific client
{
  "tool": "invoice_create",
  "parameters": {
    "client_identifier": "Acme Corp",
    "due_date": "February 15, 2024"
  }
}

// Step 3: Generate HTML document
{
  "tool": "invoice_generate_html",
  "parameters": {
    "invoice_identifier": "INV-2024-001",
    "template": "modern"
  }
}
```

### 3. Client Management Workflow

**Scenario**: Setting up and managing client relationships

**Steps**:
1. Create client with complete information
2. Set default billing preferences
3. Create ongoing invoices
4. Update client information as needed

**Example**:

```json
// Create comprehensive client profile
{
  "tool": "client_create",
  "parameters": {
    "name": "Tech Solutions Inc",
    "email": "billing@techsolutions.com",
    "phone": "+1-555-0123",
    "address": "123 Tech Street, Silicon Valley, CA 94101",
    "billing_contact": "Jane Smith",
    "default_rate": 120,
    "payment_terms": "Net 30"
  }
}
```

### 4. Bulk Operations Workflow

**Scenario**: Processing multiple invoices or timesheet files

**Example**:
```json
// Bulk import multiple timesheet files
{
  "tool": "bulk_import_timesheets",
  "parameters": {
    "directory_path": "/Users/username/timesheets/january/",
    "file_pattern": "*.csv",
    "default_rate": 85
  }
}
```

## Advanced Usage Patterns

### 1. Multi-Project Client Management

For clients with multiple projects, use descriptive project names and work item categorization:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_identifier": "Acme Corp",
    "project_name": "Website Redesign - Phase 2",
    "work_items": [
      {"description": "Frontend development", "hours": 30, "rate": 85},
      {"description": "Backend API work", "hours": 20, "rate": 95},
      {"description": "Database optimization", "hours": 10, "rate": 110}
    ],
    "due_date": "30 days"
  }
}
```

### 2. Template-Based Document Generation

Use different templates for different client types or invoice purposes:

```json
// Professional corporate template
{
  "tool": "invoice_generate_html",
  "parameters": {
    "invoice_identifier": "INV-2024-001",
    "template": "modern",
    "company_logo": "/path/to/company-logo.png"
  }
}

// Minimal template for simple invoices
{
  "tool": "invoice_generate_html",
  "parameters": {
    "invoice_identifier": "INV-2024-002",
    "template": "minimal"
  }
}
```

### 3. Data Export for Analysis

Regular export of data for business analysis:

```json
// Export monthly invoice summary
{
  "tool": "export_data_csv",
  "parameters": {
    "data_type": "invoices",
    "date_range": "2024-01-01,2024-01-31",
    "columns": ["invoice_id", "client", "total", "status", "due_date"],
    "output_path": "/path/to/reports/january_summary.csv"
  }
}

// Export client statistics
{
  "tool": "export_data_csv",
  "parameters": {
    "data_type": "summary",
    "columns": ["client", "total_invoiced", "avg_payment_time", "status"]
  }
}
```

### 4. Automated Workflows with External Systems

Integration with external time tracking systems:

```json
{
  "tool": "external_data_sync",
  "parameters": {
    "source_system": "toggl",
    "api_key": "your_toggl_api_key",
    "date_range": "last_week",
    "sync_mode": "incremental"
  }
}
```

## Natural Language Best Practices

### 1. Conversational Tool Usage

**Effective**: "Create an invoice for Acme Corp with 40 hours of development work at $85/hour"
**Less Effective**: "Use invoice_create tool with parameters..."

**Why**: Claude can better understand natural language and choose appropriate parameters.

### 2. Context Preservation

When working with multiple invoices or clients, provide context:

**Good**: "Show me the invoice I just created for Acme Corp"
**Better**: "Show me invoice INV-2024-001 for the web development project"

### 3. Flexible Identification

Use various identification methods based on what you remember:

- "Show me the latest invoice" (implicit identification)
- "Display invoice INV-2024-001" (explicit ID)
- "Find the invoice for the website project" (description-based)
- "Show the Acme Corp invoice from last week" (context-based)

### 4. Natural Date Expressions

The system supports various date formats:

- "Due in 30 days"
- "Due February 15th"
- "Due next Friday"
- "Due 2024-02-15"

### 5. Batch Operations Language

**Single Operation**: "Create an invoice for Acme Corp"
**Batch Operation**: "Import all timesheets from my January folder"

## Performance Optimization

### 1. Efficient Data Import

**For Large CSV Files**:
- Use `parallel_processing: true` for bulk imports
- Break large files into smaller chunks
- Use specific file patterns to avoid processing unnecessary files

```json
{
  "tool": "bulk_import_timesheets",
  "parameters": {
    "directory_path": "/path/to/large_dataset/",
    "file_pattern": "2024_*.csv",
    "parallel_processing": true,
    "default_rate": 85
  }
}
```

### 2. Targeted Queries

**Efficient**: Filter and limit results
```json
{
  "tool": "invoice_list",
  "parameters": {
    "client_filter": "Acme",
    "status_filter": "overdue",
    "limit": 10
  }
}
```

**Inefficient**: Request all data then filter in Claude
```json
{
  "tool": "invoice_list",
  "parameters": {}
}
```

### 3. Minimal Data Requests

Request only necessary information:

```json
// For quick status check
{
  "tool": "invoice_list",
  "parameters": {
    "columns": ["invoice_id", "client", "status"],
    "limit": 20
  }
}

// For detailed analysis (when needed)
{
  "tool": "invoice_show",
  "parameters": {
    "invoice_identifier": "INV-2024-001"
  }
}
```

### 4. Template Optimization

- Use appropriate templates for the context (minimal for internal use, modern for client delivery)
- Pre-configure templates for different client types
- Use embedded CSS for email delivery, separate CSS for web viewing

## Error Handling

### 1. Common Error Scenarios

**File Not Found**:
```
Error: CSV file not found at /path/to/file.csv
Solution: Verify file path and permissions
```

**Client Not Found**:
```
Error: Client "Acme Corp" not found
Solution: Check client name or create client first
```

**Invalid Date Format**:
```
Error: Unable to parse date "Feb 30, 2024"
Solution: Use valid date or supported format
```

### 2. Validation Before Operations

Always validate critical operations:

```json
// Before bulk import
{
  "tool": "config_validate",
  "parameters": {
    "check_data": true,
    "verbose": true
  }
}

// Before client operations
{
  "tool": "client_list",
  "parameters": {
    "name_filter": "Acme",
    "limit": 5
  }
}
```

### 3. Graceful Recovery

**For Import Errors**:
1. Check individual file formats
2. Use smaller batch sizes
3. Validate data before importing

**For Generation Errors**:
1. Verify invoice exists and is complete
2. Check template availability
3. Ensure output directory permissions

### 4. Error Reporting Best Practices

When reporting errors to Claude:

**Good**: "The CSV import failed for january_hours.csv with date format error"
**Better**: "The CSV import failed because the date column contains '2/30/2024' which is invalid"

## Integration Patterns

### 1. Accounting Software Integration

**Export to QuickBooks**:
```json
{
  "tool": "export_data_csv",
  "parameters": {
    "data_type": "invoices",
    "columns": ["invoice_id", "client", "total", "tax_amount", "due_date"],
    "format": "quickbooks"
  }
}
```

### 2. Email Integration

**Generate for Email Delivery**:
```json
{
  "tool": "invoice_generate_html",
  "parameters": {
    "invoice_identifier": "INV-2024-001",
    "template": "modern",
    "include_css": true,
    "optimize_for_email": true
  }
}
```

### 3. Backup and Sync Workflows

**Regular Backup**:
```json
// Export all data for backup
{
  "tool": "export_data_csv",
  "parameters": {
    "data_type": "invoices",
    "output_path": "/backups/invoices_backup.csv"
  }
}

{
  "tool": "export_data_csv",
  "parameters": {
    "data_type": "clients",
    "output_path": "/backups/clients_backup.csv"
  }
}
```

### 4. Reporting Workflows

**Monthly Business Reports**:
```json
// Revenue summary
{
  "tool": "export_data_csv",
  "parameters": {
    "data_type": "summary",
    "date_range": "2024-01-01,2024-01-31",
    "group_by": "client"
  }
}

// Overdue analysis
{
  "tool": "invoice_list",
  "parameters": {
    "status_filter": "overdue",
    "sort_by": "amount",
    "include_age": true
  }
}
```

## Best Practices Summary

### Do's

1. **Use natural language** - Let Claude interpret your requests
2. **Provide context** - Give Claude enough information to make good decisions
3. **Validate regularly** - Run system checks to prevent issues
4. **Use appropriate templates** - Match templates to use cases
5. **Leverage flexible identification** - Use whatever identifier you remember
6. **Plan for errors** - Have recovery strategies for common issues

### Don'ts

1. **Don't hardcode parameters** - Let Claude choose appropriate defaults
2. **Don't ignore validation errors** - Address configuration issues promptly
3. **Don't request unnecessary data** - Use filtering and limits
4. **Don't skip backups** - Export data regularly
5. **Don't use overly complex workflows** - Break complex tasks into steps

### Performance Tips

1. **Use filters and limits** to reduce data transfer
2. **Enable parallel processing** for bulk operations
3. **Choose appropriate templates** for the context
4. **Validate configuration** before large operations
5. **Monitor system resources** during heavy usage

This usage guide provides the foundation for effective natural language interaction with the go-invoice MCP tools through Claude Desktop. Focus on clear communication, proper error handling, and efficient workflows for the best experience.
