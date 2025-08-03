# MCP Tool Usage Examples

This document provides practical examples of using go-invoice MCP tools through natural language conversations with Claude Desktop.

## Basic Workflows

### Creating Your First Invoice

**You**: "I need help creating an invoice for my client"

**Claude**: I'd be happy to help you create an invoice! Let me walk you through the process. First, let me check if you have any existing clients set up.

```json
{
  "tool": "client_list",
  "parameters": {
    "limit": 5
  }
}
```

I see you don't have any clients yet. Let's start by creating a client record. What's the name of the client you want to invoice?

**You**: "TechCorp Solutions, their email is billing@techcorp.com"

**Claude**: Perfect! Let me create TechCorp Solutions as a client first.

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "TechCorp Solutions",
    "email": "billing@techcorp.com"
  }
}
```

Great! TechCorp Solutions has been added to your system. Now let's create an invoice for them. What kind of work did you do and what are your rates?

**You**: "I did 25 hours of software development at $120 per hour"

**Claude**: I'll create an invoice for your software development work:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "TechCorp Solutions",
    "description": "Software development services",
    "work_items": [
      {
        "description": "Software development",
        "hours": 25,
        "rate": 120
      }
    ],
    "due_days": 30
  }
}
```

Excellent! I've created invoice INV-2025-001 for TechCorp Solutions:
- 25 hours of software development at $120/hour
- Total: $3,000
- Due: 30 days from today

Would you like me to generate an HTML version you can send to the client?

### Monthly Timesheet Import

**You**: "I have my monthly timesheet in a CSV file. Can you help me import it and create invoices?"

**Claude**: Absolutely! I'll help you import your timesheet data and create invoices. First, let me take a look at your CSV file to understand its structure.

```json
{
  "tool": "import_preview",
  "parameters": {
    "file_path": "/Users/username/Documents/august-timesheet.csv",
    "show_sample": true
  }
}
```

Perfect! I can see your CSV has columns for Date, Client, Description, Hours, and Rate. The data looks clean and ready to import. Now let me import it and create invoices grouped by client:

```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/Users/username/Documents/august-timesheet.csv",
    "import_mode": "group_by_client",
    "create_invoices": true,
    "due_days": 30,
    "description": "Professional services for August 2025"
  }
}
```

Great! I've successfully imported your timesheet and created invoices for 3 clients:

1. **TechCorp Solutions**: $4,800 (40 hours)
2. **StartupXYZ**: $2,250 (18 hours) 
3. **LocalBusiness**: $1,500 (20 hours)

Total billable amount: $8,550

Would you like me to generate HTML invoices for all of them?

### Checking Invoice Status

**You**: "Can you show me which invoices are overdue?"

**Claude**: I'll check for overdue invoices in your system.

```json
{
  "tool": "invoice_list",
  "parameters": {
    "status_filter": "overdue",
    "sort_by": "due_date",
    "include_details": true
  }
}
```

You have 2 overdue invoices that need attention:

1. **INV-2025-008** - WebCorp ($2,400)
   - Due: July 15, 2025 (18 days overdue)
   - Client: webcorp@email.com

2. **INV-2025-012** - LocalBiz ($850)
   - Due: July 22, 2025 (11 days overdue)
   - Client: accounts@localbiz.com

Total overdue amount: $3,250

Would you like me to help you send follow-up emails for these overdue invoices?

## Advanced Workflows

### Multi-Project Client Management

**You**: "I have a large client with multiple ongoing projects. How should I organize the billing?"

**Claude**: For large clients with multiple projects, I recommend using project-specific descriptions and clear organization. Let me show you how to set this up effectively.

First, let me check if this client already exists:

```json
{
  "tool": "client_list",
  "parameters": {
    "name_filter": "Enterprise Corp",
    "include_stats": true
  }
}
```

Now let me create a project-based invoice structure for them:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "Enterprise Corp",
    "description": "Project Alpha - Mobile App Development (Phase 1)",
    "project_name": "Project Alpha",
    "work_items": [
      {
        "description": "UI/UX Design - Mobile App",
        "hours": 32,
        "rate": 125
      },
      {
        "description": "Backend API Development",
        "hours": 28,
        "rate": 135
      },
      {
        "description": "Project Management & Coordination",
        "hours": 8,
        "rate": 150
      }
    ],
    "due_days": 30,
    "notes": "Phase 1 of 3-phase mobile app development project. Next phase: Core functionality development."
  }
}
```

This approach gives you:
- Clear project identification in invoice descriptions
- Detailed breakdown of different types of work
- Project phase tracking for long-term engagements
- Professional presentation for the client

For Project Beta, you'd create a separate invoice with similar structure but different project naming.

### Batch Document Generation

**You**: "I need to generate professional invoices for all my pending invoices this month"

**Claude**: I'll help you generate professional HTML invoices for all your pending invoices. Let me first get the list of pending invoices, then generate documents for all of them.

```json
{
  "tool": "invoice_list",
  "parameters": {
    "status_filter": "pending",
    "date_from": "2025-08-01",
    "sort_by": "client"
  }
}
```

Perfect! You have 5 pending invoices for August. Now let me generate professional HTML documents for all of them:

```json
{
  "tool": "generate_html",
  "parameters": {
    "batch_invoices": ["INV-2025-015", "INV-2025-016", "INV-2025-017", "INV-2025-018", "INV-2025-019"],
    "template": "professional",
    "output_dir": "./august-invoices/",
    "include_logo": true,
    "auto_name": true
  }
}
```

Excellent! I've generated professional HTML invoices for all 5 pending invoices:

1. INV-2025-015-TechCorp.html ($4,800)
2. INV-2025-016-StartupXYZ.html ($2,250)
3. INV-2025-017-LocalBusiness.html ($1,500)
4. INV-2025-018-Enterprise-Corp.html ($6,750)
5. INV-2025-019-WebAgency.html ($3,200)

All files are saved in your `august-invoices/` directory with professional styling and your company logo. They're ready to send to clients!

### Configuration and System Health

**You**: "I'm getting some errors when generating invoices. Can you check if my system is configured correctly?"

**Claude**: Let me run a comprehensive system validation to identify any configuration issues that might be causing problems.

```json
{
  "tool": "config_validate",
  "parameters": {
    "comprehensive_check": true,
    "check_templates": true,
    "check_data_integrity": true,
    "verbose": true
  }
}
```

I found a few issues that are likely causing your invoice generation errors:

**🔧 Configuration Issues:**
1. **Template path error**: The custom template path is pointing to a missing file
2. **Missing default rate**: No default rate set for new work items
3. **Date format inconsistency**: Mixed date formats in some imported data

Let me help fix these issues:

```json
{
  "tool": "config_init",
  "parameters": {
    "repair_mode": true,
    "fix_template_paths": true,
    "set_safe_defaults": true,
    "backup_before_changes": true
  }
}
```

**✅ Issues Fixed:**
1. Reset template paths to working defaults
2. Set default hourly rate to $100 (you can adjust this)
3. Standardized date formats
4. Created configuration backup

Your system should now work properly! Let me test invoice generation to confirm:

```json
{
  "tool": "generate_html",
  "parameters": {
    "invoice_identifier": "INV-2025-015",
    "template": "default",
    "validate_before_generation": true
  }
}
```

Perfect! Invoice generation is now working correctly. The system validation fixed the configuration issues.

## Natural Language Tips

### Effective Communication Patterns

**✅ Good Examples:**
- "Create an invoice for Acme Corp with 40 hours at $85/hour"
- "Show me all overdue invoices sorted by amount"
- "Import my timesheet CSV and create invoices grouped by client"
- "Generate professional HTML invoices for all my August invoices"

**❌ Less Effective:**
- "Use the invoice_create tool with these exact parameters..."
- "Execute the command with ID 12345..."

### Context Preservation

**✅ Maintaining Context:**
- "Show me the invoice I just created"
- "Add 5 more hours to that TechCorp invoice"
- "Generate an HTML version of the invoice we were just working on"

**✅ Being Specific:**
- "Show me invoice INV-2025-015"
- "Update the TechCorp invoice from last week"
- "Generate HTML for the $4,800 invoice"

### Flexible Identification

The tools support various ways to identify invoices and clients:

- **By Name**: "Show the TechCorp invoice"
- **By ID**: "Display invoice INV-2025-015"
- **By Description**: "Find the mobile app development invoice"
- **By Amount**: "Show the $4,800 invoice"
- **By Date**: "Display last week's invoices"

## Common Workflows

### Weekly Review Process

**You**: "I want to do my weekly invoice review. Can you help me see what needs attention?"

**Claude**: I'll help you with a comprehensive weekly review. Let me check several key areas:

```json
{
  "tool": "invoice_list",
  "parameters": {
    "date_from": "2025-08-01",
    "sort_by": "status",
    "include_summary": true
  }
}
```

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

**📊 Weekly Invoice Summary:**

**Recent Invoices (This Week):**
- 3 new invoices created ($8,550 total)
- 2 invoices paid ($3,200)
- 1 invoice sent, awaiting payment ($2,400)

**⚠️ Attention Needed:**
- 2 overdue invoices ($3,250 total)
- 1 invoice due this week ($1,500)

**💰 Financial Snapshot:**
- Week's billings: $8,550
- Outstanding receivables: $7,150
- Average payment time: 18 days

Would you like me to help you follow up on the overdue invoices or generate any specific reports?

### End-of-Month Processing

**You**: "It's the end of the month. Help me wrap up all my billing and generate reports"

**Claude**: I'll help you complete your end-of-month billing process. Let me handle this systematically:

```json
{
  "tool": "invoice_list",
  "parameters": {
    "date_from": "2025-08-01",
    "date_to": "2025-08-31",
    "include_summary": true,
    "group_by": "status"
  }
}
```

```json
{
  "tool": "export_data",
  "parameters": {
    "export_type": "monthly_summary",
    "period": "August 2025",
    "include_fields": [
      "invoice_number",
      "client_name",
      "amount",
      "status",
      "payment_date"
    ],
    "output_path": "./reports/August-2025-Summary.csv"
  }
}
```

```json
{
  "tool": "generate_summary",
  "parameters": {
    "report_type": "monthly_financial",
    "period": "August 2025",
    "include_charts": true,
    "output_format": "pdf",
    "output_path": "./reports/August-2025-Financial-Report.pdf"
  }
}
```

**📈 August 2025 Summary:**

**Revenue Metrics:**
- Total invoiced: $24,750
- Total collected: $18,200
- Outstanding: $6,550
- Collection rate: 73.5%

**Client Activity:**
- Active clients: 8
- New clients: 2
- Average invoice: $3,094

**Files Generated:**
- CSV export: `August-2025-Summary.csv`
- Financial report: `August-2025-Financial-Report.pdf`

Your month-end processing is complete! All reports are saved in your reports directory.

## Troubleshooting Common Issues

### Import Problems

**You**: "My CSV import is failing with date format errors"

**Claude**: Let me help you troubleshoot the CSV import issue. First, let me preview your file to see what's causing the date format problems:

```json
{
  "tool": "import_preview",
  "parameters": {
    "file_path": "/path/to/your/timesheet.csv",
    "detect_issues": true,
    "analyze_date_formats": true
  }
}
```

I can see the issue! Your CSV has mixed date formats - some rows use MM/DD/YYYY and others use DD/MM/YYYY. Let me fix this:

```json
{
  "tool": "import_validate",
  "parameters": {
    "file_path": "/path/to/your/timesheet.csv",
    "fix_date_formats": true,
    "create_cleaned_file": true,
    "cleaned_file_path": "/path/to/cleaned-timesheet.csv"
  }
}
```

Now let me import the cleaned file:

```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/path/to/cleaned-timesheet.csv",
    "date_format": "auto_detect",
    "validate_before_import": true
  }
}
```

Perfect! The import completed successfully after fixing the date format issues. In the future, try to export your timesheet with consistent date formatting to avoid this issue.

This comprehensive set of examples shows how to effectively use the go-invoice MCP tools through natural language conversations with Claude Desktop.