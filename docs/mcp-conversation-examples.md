# go-invoice MCP Conversation Examples

This document provides natural language interaction examples for the go-invoice MCP tools system. These examples demonstrate how to effectively communicate with Claude Desktop using conversational language to accomplish invoice management tasks.

## Table of Contents

1. [Getting Started Conversations](#getting-started-conversations)
2. [Invoice Management Workflows](#invoice-management-workflows)
3. [Client Management Scenarios](#client-management-scenarios)
4. [Data Import and Export Examples](#data-import-and-export-examples)
5. [Configuration and Setup](#configuration-and-setup)
6. [Complex Business Workflows](#complex-business-workflows)
7. [Error Handling and Troubleshooting](#error-handling-and-troubleshooting)
8. [Advanced Usage Patterns](#advanced-usage-patterns)

## Getting Started Conversations

### First Time Setup

**User**: "I'm new to go-invoice. Can you help me set up the system and create my first invoice?"

**Claude**: I'll help you get started with go-invoice! Let me first check your system configuration and then guide you through creating your first invoice.

First, let me validate your system setup:

```json
{
  "tool": "config_validate",
  "parameters": {
    "verbose": true,
    "show_recommendations": true
  }
}
```

Great! Your system is properly configured. Now let's create your first client and invoice. What's the name of the client you'd like to invoice?

**User**: "The client is Acme Corporation, their email is billing@acme.com"

**Claude**: Perfect! Let me create Acme Corporation as a client first, then we'll set up an invoice for them.

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "Acme Corporation",
    "email": "billing@acme.com"
  }
}
```

Now let's create an invoice for them. What kind of work did you do, and what are your rates?

**User**: "I did 20 hours of web development at $85 per hour last week"

**Claude**: I'll create an invoice for your web development work:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "Acme Corporation",
    "description": "Web development services",
    "work_items": [
      {
        "description": "Web development",
        "hours": 20,
        "rate": 85
      }
    ],
    "due_days": 30
  }
}
```

Your invoice has been created! The total is $1,700 (20 hours × $85). Would you like me to generate an HTML version you can send to the client?

### Quick Status Check

**User**: "Show me a quick overview of my current invoices"

**Claude**: I'll get you a summary of your current invoices:

```json
{
  "tool": "invoice_list",
  "parameters": {
    "limit": 10,
    "sort_by": "date",
    "sort_order": "desc",
    "include_summary": true
  }
}
```

Here's your invoice summary:
- Total invoices: 15
- Pending: $8,450
- Paid: $12,300
- Overdue: $1,200 (2 invoices)

Would you like me to show you the overdue invoices so you can follow up?

## Invoice Management Workflows

### Monthly Billing Workflow

**User**: "I need to do my monthly billing for all my clients. I have timesheet data in a CSV file from my time tracking app."

**Claude**: I'll help you process your monthly billing efficiently. Let's start by importing your timesheet data and then create invoices for each client.

First, let me preview your CSV file to understand its structure:

```json
{
  "tool": "import_preview",
  "parameters": {
    "file_path": "/path/to/monthly-timesheet.csv",
    "show_sample": true
  }
}
```

Perfect! I can see your CSV has columns for Date, Client, Description, Hours, and Rate. Now let me import this data and create invoices grouped by client:

```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/path/to/monthly-timesheet.csv",
    "import_mode": "group_by_client",
    "create_invoices": true,
    "due_days": 30,
    "description": "Professional services for {month} {year}"
  }
}
```

Great! I've created invoices for 5 clients based on your timesheet data:
- Acme Corp: $3,400 (40 hours)
- Tech Solutions: $2,850 (30 hours)
- StartupXYZ: $1,700 (20 hours)
- Local Business: $950 (10 hours)
- Remote Client: $2,275 (25 hours)

Would you like me to generate HTML invoices for all of them so you can email them to your clients?

**User**: "Yes, please generate HTML invoices for all of them with the professional template"

**Claude**: I'll generate professional HTML invoices for all your clients:

```json
{
  "tool": "generate_html",
  "parameters": {
    "batch_invoices": ["INV-2025-015", "INV-2025-016", "INV-2025-017", "INV-2025-018", "INV-2025-019"],
    "template": "professional",
    "output_dir": "./monthly-invoices/",
    "include_logo": true,
    "auto_name": true
  }
}
```

Perfect! I've generated professional HTML invoices for all 5 clients. The files are saved in your monthly-invoices directory with names like "INV-2025-015-Acme-Corp.html". Each invoice includes your company logo and professional styling.

### Project-Based Invoicing

**User**: "I'm working on a large project for MegaCorp that spans 3 months. I want to invoice them monthly but track all work under the same project. How should I handle this?"

**Claude**: Great question! For project-based billing across multiple months, I recommend creating invoices with clear project references and phases. Let me show you how to set this up.

First, let me check if MegaCorp is already a client:

```json
{
  "tool": "client_list",
  "parameters": {
    "name_filter": "MegaCorp",
    "include_stats": true
  }
}
```

Perfect! MegaCorp is already in your system. Now let's create your first monthly invoice for the project:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "MegaCorp",
    "description": "Enterprise Software Development - Phase 1 of 3",
    "project_name": "Enterprise Software Development",
    "work_items": [
      {
        "description": "Architecture and system design",
        "hours": 32,
        "rate": 125
      },
      {
        "description": "Database schema development",
        "hours": 24,
        "rate": 125
      },
      {
        "description": "Frontend framework setup",
        "hours": 16,
        "rate": 115
      }
    ],
    "due_days": 30,
    "notes": "Phase 1: Foundation and architecture (Month 1 of 3-month project)"
  }
}
```

For next month, you can create Phase 2 with:
- Similar project naming for consistency
- Phase-specific work items
- Reference to the overall project timeline

This approach gives you:
- Clear project tracking across invoices
- Monthly billing cycles
- Easy reporting on project profitability
- Professional presentation to the client

Would you like me to show you how to generate a project summary report once you have multiple phases invoiced?

### Handling Invoice Corrections

**User**: "I made a mistake on invoice INV-2025-012. I charged 40 hours but it should have been 35 hours for the web development work. How do I fix this?"

**Claude**: I'll help you correct that invoice. Let me first show you the current invoice details, then we'll fix the hours.

```json
{
  "tool": "invoice_show",
  "parameters": {
    "invoice_number": "INV-2025-012"
  }
}
```

I can see the invoice has "Web development" for 40 hours at $85/hour = $3,400. To correct this to 35 hours, I'll need to remove the current work item and add the corrected one:

First, let me remove the incorrect work item:

```json
{
  "tool": "invoice_remove_item",
  "parameters": {
    "invoice_number": "INV-2025-012",
    "item_description": "Web development"
  }
}
```

Now I'll add the corrected work item:

```json
{
  "tool": "invoice_add_item",
  "parameters": {
    "invoice_number": "INV-2025-012",
    "description": "Web development (corrected)",
    "hours": 35,
    "rate": 85
  }
}
```

Perfect! The invoice now shows 35 hours at $85/hour = $2,975, which is $425 less than before. The invoice total has been automatically recalculated.

If the client has already received the original invoice, you might want to:
1. Add a note explaining the correction
2. Generate a new HTML version
3. Send a brief email explaining the adjustment

Would you like me to add a note to the invoice or regenerate the HTML document?

## Client Management Scenarios

### Setting Up a New Corporate Client

**User**: "I just landed a big corporate client - TechGlobal Industries. They have specific requirements: NET 45 payment terms, they need separate billing for different departments, and they require detailed project tracking. How should I set this up?"

**Claude**: Excellent! Landing a big corporate client is great news. Let me help you set up TechGlobal Industries properly to meet their specific requirements.

First, let's create the main client record with their corporate information:

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "TechGlobal Industries",
    "email": "accounting@techglobal.com",
    "phone": "+1-555-TECH-GLO",
    "address": "789 Corporate Plaza, Suite 1500, Business City, BC 12345",
    "payment_terms": "NET 45",
    "default_rate": 135,
    "notes": "Large corporate client - requires department-specific billing and detailed project tracking"
  }
}
```

For department-specific billing, I recommend creating sub-clients or using detailed project names. Let me show you both approaches:

**Approach 1: Department-specific project naming**
When creating invoices, use descriptive project names like:
- "TechGlobal - IT Department - Server Migration"
- "TechGlobal - Marketing Department - Website Redesign"
- "TechGlobal - Finance Department - System Integration"

**Approach 2: Separate client entries for departments**
If they need completely separate billing:

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "TechGlobal Industries - IT Department",
    "email": "it-billing@techglobal.com",
    "phone": "+1-555-TECH-GLO",
    "address": "789 Corporate Plaza, Suite 1500, Business City, BC 12345",
    "payment_terms": "NET 45",
    "default_rate": 135
  }
}
```

For detailed project tracking, always include:
- Specific project names in invoice descriptions
- Detailed work item descriptions
- Date tracking for each work item
- Clear phase/milestone references

Which approach would work better for your client's billing structure?

### Managing Freelancer Relationships

**User**: "I work with several freelancers who help me on projects. Sometimes I need to include their work on client invoices but track costs separately. What's the best way to handle this?"

**Claude**: Great question! Managing subcontractor relationships requires careful tracking for both client billing and cost management. Let me show you a few approaches depending on your business model.

**Approach 1: Include subcontractor work in client invoices (markup model)**

First, let me create client entries for your freelancers to track their costs:

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "Sarah Johnson (Designer) - Subcontractor",
    "email": "sarah@freelancedesign.com",
    "phone": "+1-555-DESIGN1",
    "default_rate": 75,
    "notes": "Freelance designer - tracks costs for markup to end clients"
  }
}
```

When billing your end client, create work items that reflect your markup:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "Acme Corporation",
    "description": "Website Redesign Project",
    "work_items": [
      {
        "description": "Project management and coordination",
        "hours": 10,
        "rate": 125
      },
      {
        "description": "UI/UX design services",
        "hours": 20,
        "rate": 95,
        "notes": "Includes coordination with specialist designer"
      },
      {
        "description": "Development and integration",
        "hours": 30,
        "rate": 115
      }
    ]
  }
}
```

**Approach 2: Separate cost tracking with pass-through billing**

Create a separate cost-tracking invoice for the freelancer's work:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "Sarah Johnson (Designer) - Subcontractor",
    "description": "Design work for Acme Corp project (cost tracking)",
    "work_items": [
      {
        "description": "UI/UX design for Acme Corp website",
        "hours": 20,
        "rate": 75
      }
    ],
    "notes": "Cost tracking - to be passed through to Acme Corp"
  }
}
```

Then bill the end client:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "Acme Corporation",
    "description": "Website Redesign Project",
    "work_items": [
      {
        "description": "Project management and development",
        "hours": 40,
        "rate": 125
      },
      {
        "description": "Specialist design services (pass-through)",
        "hours": 20,
        "rate": 75,
        "notes": "Third-party design specialist"
      }
    ]
  }
}
```

This approach gives you:
- Clear cost tracking for each freelancer
- Transparent client billing
- Easy profit margin analysis
- Separate payment tracking

Would you like me to show you how to generate reports that help you track profitability across these different arrangements?

## Data Import and Export Examples

### Processing Timesheet Exports

**User**: "I use Toggl for time tracking and export my data monthly. The CSV has columns like 'User', 'Email', 'Client', 'Project', 'Start date', 'Start time', 'End date', 'End time', 'Duration', 'Billable', 'Description'. How do I import this efficiently?"

**Claude**: Perfect! Toggl exports contain rich data that we can use effectively. Let me help you set up a custom import that maps Toggl's format to go-invoice.

First, let me preview your Toggl export to understand the exact structure:

```json
{
  "tool": "import_preview",
  "parameters": {
    "file_path": "/path/to/toggl-export.csv",
    "show_sample": true,
    "detect_headers": true
  }
}
```

Great! I can see the Toggl format. Now let me import it with proper field mapping:

```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/path/to/toggl-export.csv",
    "field_mapping": {
      "date": "Start date",
      "client": "Client",
      "description": "Description",
      "duration_format": "hh:mm:ss",
      "duration_source": "Duration"
    },
    "filter_conditions": {
      "billable": "Yes"
    },
    "import_mode": "group_by_client",
    "create_invoices": true,
    "default_rate": 95,
    "due_days": 30,
    "description_template": "Professional services for {project} - {month} {year}"
  }
}
```

This import will:
- Only include billable entries (filter_conditions)
- Convert Toggl's time format to hours
- Group entries by client automatically
- Use project names in invoice descriptions
- Create separate invoices for each client

The import created invoices for 4 clients from your Toggl data:
- WebCorp: $2,375 (25 hours across Mobile App project)
- StartupABC: $1,425 (15 hours across API Development project)
- LocalBiz: $950 (10 hours across Website Updates project)
- ConsultingGig: $1,900 (20 hours across Strategy Planning project)

Would you like me to generate HTML invoices for all of these, or would you prefer to review and adjust any of them first?

### Exporting for Accounting Software

**User**: "My accountant needs quarterly reports in a specific format for QuickBooks. Can you export my invoice data in a way that's easy for them to import?"

**Claude**: Absolutely! I'll export your quarterly data in a QuickBooks-friendly format. Let me create a comprehensive export that includes all the information your accountant needs.

```json
{
  "tool": "export_data",
  "parameters": {
    "export_type": "invoices",
    "date_range": {
      "start": "2025-07-01",
      "end": "2025-09-30"
    },
    "format": "csv",
    "include_fields": [
      "invoice_number",
      "client_name",
      "client_email",
      "invoice_date",
      "due_date",
      "description",
      "subtotal",
      "tax_amount",
      "total_amount",
      "payment_status",
      "payment_date"
    ],
    "output_path": "./exports/Q3-2025-invoices-for-quickbooks.csv",
    "include_line_items": true,
    "group_by": "invoice"
  }
}
```

I'll also create a summary report for your accountant:

```json
{
  "tool": "generate_summary",
  "parameters": {
    "report_type": "quarterly_summary",
    "period": "Q3-2025",
    "include_sections": [
      "revenue_summary",
      "client_breakdown",
      "payment_status",
      "outstanding_receivables"
    ],
    "output_format": "pdf",
    "output_path": "./exports/Q3-2025-financial-summary.pdf"
  }
}
```

Your exports include:

**CSV Export (Q3-2025-invoices-for-quickbooks.csv):**
- All invoice details in QuickBooks-compatible format
- Individual line items for detailed tracking
- Tax amounts separately calculated
- Payment status and dates

**Summary Report (Q3-2025-financial-summary.pdf):**
- Total revenue: $48,750
- Outstanding receivables: $12,300
- Number of clients invoiced: 15
- Average invoice value: $3,250

The CSV file is formatted with standard QuickBooks column headers, so your accountant can import it directly. Would you like me to also export your client information separately for QuickBooks customer setup?

### Batch Processing Multiple Files

**User**: "I have a folder with 12 monthly timesheet CSV files that I need to process. Each file represents one month of work across multiple clients. How can I process all of them efficiently?"

**Claude**: Perfect! Batch processing multiple timesheet files is much more efficient than handling them one by one. Let me set up a workflow that processes all 12 files while maintaining proper organization.

First, let me validate all the files in your folder to ensure they have consistent formats:

```json
{
  "tool": "import_validate",
  "parameters": {
    "batch_mode": true,
    "directory_path": "/path/to/monthly-timesheets/",
    "file_pattern": "*.csv",
    "check_consistency": true,
    "generate_report": true
  }
}
```

Great! All 12 files have consistent formats. Now I'll process them in batch mode:

```json
{
  "tool": "import_csv",
  "parameters": {
    "batch_mode": true,
    "directory_path": "/path/to/monthly-timesheets/",
    "file_pattern": "timesheet-*.csv",
    "import_mode": "group_by_client_and_month",
    "create_invoices": true,
    "due_days": 30,
    "description_template": "Professional services - {filename_month} {filename_year}",
    "filename_date_pattern": "timesheet-YYYY-MM.csv",
    "progress_tracking": true
  }
}
```

Perfect! I've processed all 12 months of data. Here's what was created:

**Summary:**
- Files processed: 12 (Jan 2024 - Dec 2024)
- Total invoices created: 47
- Total billable hours: 1,847
- Total revenue: $184,750
- Clients invoiced: 8 unique clients

**By Month:**
- January: 6 invoices, $15,300
- February: 4 invoices, $12,800
- March: 5 invoices, $18,950
- [... and so on]

**By Client:**
- TechCorp: 12 invoices (monthly), $67,200
- StartupXYZ: 8 invoices, $28,500
- LocalBusiness: 6 invoices, $15,750
- [... and so on]

The system automatically:
- Extracted month/year from filenames
- Grouped work by client within each month
- Created properly dated invoices
- Maintained chronological order

Would you like me to generate HTML invoices for all of these, or create a comprehensive annual report showing the year's billing activity?

## Configuration and Setup

### Initial System Configuration

**User**: "I'm setting up go-invoice for my consulting business. I need to configure company information, default rates, invoice templates, and tax settings. Can you walk me through this?"

**Claude**: I'll help you set up go-invoice perfectly for your consulting business. Let's start with a guided configuration that covers all the essential settings.

First, let me initialize a new configuration file with prompts for your business details:

```json
{
  "tool": "config_init",
  "parameters": {
    "guided_setup": true,
    "business_type": "consulting",
    "include_templates": ["professional", "minimal", "consulting"],
    "setup_wizard": true
  }
}
```

The setup wizard will walk through these sections:

**Company Information:**
- Company name and legal business name
- Business address and contact information
- Tax ID/EIN number
- Professional license numbers (if applicable)

**Default Billing Settings:**
- Standard hourly rate ($125 for senior consulting)
- Payment terms (NET 30 is common for consulting)
- Late fee policies and percentages
- Currency and tax rate settings

**Invoice Preferences:**
- Invoice numbering format (e.g., "CONS-2025-001")
- Default template (professional for corporate clients)
- Company logo and branding elements
- Email templates for invoice delivery

Let me show you what a typical consulting business configuration looks like:

```json
{
  "tool": "config_show",
  "parameters": {
    "output_format": "yaml",
    "include_examples": true,
    "show_recommendations": true
  }
}
```

Here's a sample configuration structure for consulting:

```yaml
company:
  name: "Your Consulting LLC"
  address: "123 Professional Drive, Suite 200, City, State 12345"
  phone: "+1-555-CONSULT"
  email: "billing@yourconsulting.com"
  tax_id: "12-3456789"
  
billing:
  default_rate: 125.00
  payment_terms: "NET 30"
  late_fee_rate: 1.5
  currency: "USD"
  tax_rate: 8.5
  
invoices:
  number_format: "CONS-{YYYY}-{NNN}"
  default_template: "professional"
  auto_send: false
  include_time_details: true
  
email:
  smtp_server: "your-email-provider.com"
  from_address: "billing@yourconsulting.com"
  signature: "Thank you for your business!"
```

Would you like me to help you customize any of these settings, or shall we proceed to set up your first client and invoice template?

### Validating System Health

**User**: "I've been using go-invoice for a few months and occasionally get weird errors. Can you check if everything is configured correctly?"

**Claude**: Absolutely! Let me run a comprehensive system validation to identify any configuration issues or potential problems.

```json
{
  "tool": "config_validate",
  "parameters": {
    "comprehensive_check": true,
    "check_data_integrity": true,
    "verify_templates": true,
    "test_integrations": true,
    "generate_detailed_report": true
  }
}
```

I found a few issues that might be causing your errors:

**⚠️ Configuration Issues Found:**

1. **Template Path Issue:**
   - Problem: Custom template path "/templates/custom.html" not found
   - Impact: Invoice generation fails for custom template
   - Fix: Update template path or restore missing file

2. **Email Configuration:**
   - Problem: SMTP password not set in environment
   - Impact: Email sending fails silently
   - Fix: Set EMAIL_PASSWORD environment variable

3. **Data Integrity:**
   - Problem: 3 invoices have orphaned work items (client deleted)
   - Impact: Reports may show incorrect totals
   - Fix: Clean up orphaned data or restore clients

**✅ Working Correctly:**
- Basic invoice creation and calculations
- Client management operations
- CSV import functionality
- Database connections and performance

Let me fix the most critical issues:

```json
{
  "tool": "config_init",
  "parameters": {
    "repair_mode": true,
    "fix_template_paths": true,
    "cleanup_orphaned_data": true,
    "backup_before_repair": true
  }
}
```

**Fixes Applied:**
1. ✅ Reset template paths to working defaults
2. ✅ Cleaned up 3 orphaned work items
3. ✅ Created configuration backup
4. ⚠️ Email settings still need manual configuration

**Recommendations:**
1. Set your email password: `export EMAIL_PASSWORD="your-app-password"`
2. Test email functionality: I can help you send a test invoice
3. Consider updating to latest templates for better compatibility
4. Schedule monthly validation checks to catch issues early

Your system should now work much more reliably! Would you like me to test invoice generation to make sure everything is working properly?

## Complex Business Workflows

### Multi-Client Project Management

**User**: "I'm managing a complex project that involves work for three different companies that are collaborating. Company A is the lead, Company B handles infrastructure, and Company C does design. I need to bill my coordination work to all three, but track different rates and responsibilities. How should I structure this?"

**Claude**: This is a sophisticated billing scenario! You're essentially acting as a project coordinator with different roles for each company. Let me help you structure this properly to maintain clear billing and avoid conflicts.

**Recommended Structure: Role-Based Billing Approach**

First, let me create client records that reflect their roles in the collaboration:

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "TechLead Corp (Project Alpha - Lead Client)",
    "email": "billing@techlead.com",
    "phone": "+1-555-TECH-001",
    "default_rate": 150,
    "payment_terms": "NET 30",
    "notes": "Lead client for Project Alpha collaboration - responsible for project management coordination"
  }
}
```

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "InfraCorp (Project Alpha - Infrastructure Partner)",
    "email": "accounts@infracorp.com",
    "phone": "+1-555-INFRA-001",
    "default_rate": 125,
    "payment_terms": "NET 30",
    "notes": "Infrastructure partner for Project Alpha - technical coordination and integration support"
  }
}
```

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "DesignStudio (Project Alpha - Design Partner)",
    "email": "billing@designstudio.com",
    "phone": "+1-555-DESIGN-001",
    "default_rate": 135,
    "payment_terms": "NET 30",
    "notes": "Design partner for Project Alpha - design review and brand coordination"
  }
}
```

**Monthly Invoice Structure Example:**

For TechLead Corp (Lead Client):
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "TechLead Corp (Project Alpha - Lead Client)",
    "description": "Project Alpha Coordination - Lead Client Services (Month 1)",
    "project_name": "Project Alpha Multi-Client Coordination",
    "work_items": [
      {
        "description": "Overall project management and stakeholder coordination",
        "hours": 20,
        "rate": 150
      },
      {
        "description": "Cross-company communication and alignment meetings",
        "hours": 8,
        "rate": 150
      },
      {
        "description": "Technical requirements coordination with infrastructure team",
        "hours": 6,
        "rate": 150
      },
      {
        "description": "Budget and timeline management",
        "hours": 4,
        "rate": 150
      }
    ],
    "notes": "Lead client billing includes full project oversight, stakeholder management, and coordination of all partner activities"
  }
}
```

For InfraCorp (Infrastructure Partner):
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "InfraCorp (Project Alpha - Infrastructure Partner)",
    "description": "Project Alpha Coordination - Infrastructure Integration (Month 1)",
    "project_name": "Project Alpha Multi-Client Coordination",
    "work_items": [
      {
        "description": "Infrastructure requirements analysis and coordination",
        "hours": 12,
        "rate": 125
      },
      {
        "description": "Technical integration planning with design team",
        "hours": 6,
        "rate": 125
      },
      {
        "description": "Performance requirements documentation and review",
        "hours": 4,
        "rate": 125
      }
    ],
    "notes": "Infrastructure coordination billing covers technical integration, requirements alignment, and infrastructure-specific project coordination"
  }
}
```

For DesignStudio (Design Partner):
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "DesignStudio (Project Alpha - Design Partner)",
    "description": "Project Alpha Coordination - Design Integration (Month 1)",
    "project_name": "Project Alpha Multi-Client Coordination",
    "work_items": [
      {
        "description": "Design requirements coordination and brand alignment",
        "hours": 10,
        "rate": 135
      },
      {
        "description": "Cross-team design review and feedback coordination",
        "hours": 4,
        "rate": 135
      },
      {
        "description": "Design-infrastructure integration planning",
        "hours": 3,
        "rate": 135
      }
    ],
    "notes": "Design coordination billing includes brand alignment, design review coordination, and design-technical integration"
  }
}
```

**Key Benefits of This Structure:**
1. **Clear Role Definition**: Each company pays for coordination specific to their role
2. **Transparent Billing**: No overlap in charged time across companies
3. **Rate Differentiation**: Higher rates for more complex coordination work
4. **Project Tracking**: All invoices clearly linked to Project Alpha
5. **Audit Trail**: Clear documentation of who paid for what services

**Alternative Approach: Lead Client + Subcontractor Model**

If the companies prefer, you could bill only the lead client and let them handle internal cost allocation:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "TechLead Corp",
    "description": "Project Alpha Multi-Company Coordination (Month 1)",
    "work_items": [
      {
        "description": "Project management and lead client coordination",
        "hours": 38,
        "rate": 150
      },
      {
        "description": "Infrastructure partner coordination and integration",
        "hours": 22,
        "rate": 125
      },
      {
        "description": "Design partner coordination and brand alignment",
        "hours": 17,
        "rate": 135
      }
    ],
    "notes": "Comprehensive multi-company project coordination. Lead client handles internal cost allocation with partners."
  }
}
```

Which approach would work better for your client relationships and their preferred billing structure?

### Annual Retainer Management

**User**: "I have several clients on annual retainers where they pay upfront for a certain number of hours, and I track usage throughout the year. How can I manage this in go-invoice?"

**Claude**: Excellent question! Retainer management requires careful tracking of prepaid hours versus usage, with clear reporting for both you and your clients. Let me show you a comprehensive approach.

**Setting Up Retainer Clients**

First, let's create a retainer client with special notation:

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "MegaCorp (Annual Retainer 2025)",
    "email": "billing@megacorp.com",
    "phone": "+1-555-MEGA-001",
    "default_rate": 175,
    "payment_terms": "Retainer - Prepaid",
    "notes": "Annual retainer: 200 hours prepaid at $175/hour = $35,000. Retainer period: Jan 1 - Dec 31, 2025. Overage billed monthly at $175/hour."
  }
}
```

**Initial Retainer Invoice (Paid Upfront)**

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "MegaCorp (Annual Retainer 2025)",
    "description": "2025 Annual Retainer - 200 Hours Prepaid",
    "invoice_date": "2025-01-01",
    "due_date": "2025-01-15",
    "work_items": [
      {
        "description": "Annual retainer - 200 hours prepaid consulting services",
        "hours": 200,
        "rate": 175,
        "notes": "Retainer period: January 1 - December 31, 2025"
      }
    ],
    "notes": "Annual retainer payment. Hours will be tracked and applied throughout 2025. Overage hours billed monthly at standard rate.",
    "invoice_type": "retainer_setup"
  }
}
```

**Monthly Retainer Usage Tracking (Internal)**

Create monthly "usage tracking" entries to monitor retainer consumption:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "MegaCorp (Annual Retainer 2025)",
    "description": "January 2025 - Retainer Hours Usage Tracking",
    "invoice_date": "2025-01-31",
    "work_items": [
      {
        "description": "Strategic planning consultation",
        "hours": 8,
        "rate": 0,
        "notes": "Applied against annual retainer (8/200 hours used)"
      },
      {
        "description": "Technical architecture review",
        "hours": 12,
        "rate": 0,
        "notes": "Applied against annual retainer (20/200 hours used total)"
      },
      {
        "description": "Team training and workshops",
        "hours": 6,
        "rate": 0,
        "notes": "Applied against annual retainer (26/200 hours used total)"
      }
    ],
    "notes": "Retainer usage for January 2025. Total hours used: 26. Remaining retainer balance: 174 hours.",
    "invoice_type": "retainer_tracking",
    "send_to_client": false
  }
}
```

**Monthly Client Report Generation**

```json
{
  "tool": "generate_summary",
  "parameters": {
    "report_type": "retainer_status",
    "client_name": "MegaCorp (Annual Retainer 2025)",
    "period": "January 2025",
    "include_sections": [
      "hours_used_this_month",
      "hours_used_year_to_date",
      "remaining_retainer_balance",
      "projected_usage",
      "detailed_work_log"
    ],
    "output_format": "pdf",
    "output_path": "./retainer-reports/MegaCorp-January-2025-Retainer-Status.pdf"
  }
}
```

**Overage Billing (When Retainer Exceeded)**

When client exceeds their retainer hours:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "MegaCorp (Annual Retainer 2025)",
    "description": "August 2025 - Retainer Overage Hours",
    "invoice_date": "2025-08-31",
    "due_date": "2025-09-30",
    "work_items": [
      {
        "description": "Emergency system consultation (overage)",
        "hours": 8,
        "rate": 175,
        "notes": "Retainer fully utilized. Billing at standard overage rate."
      },
      {
        "description": "Additional strategic planning (overage)",
        "hours": 4,
        "rate": 175,
        "notes": "Overage hours - beyond annual retainer allocation"
      }
    ],
    "notes": "Annual retainer (200 hours) fully utilized as of August 15, 2025. These are overage hours billed at standard rate.",
    "invoice_type": "retainer_overage"
  }
}
```

**Quarterly Retainer Status Reports**

```json
{
  "tool": "generate_summary",
  "parameters": {
    "report_type": "quarterly_retainer_summary",
    "client_name": "MegaCorp (Annual Retainer 2025)",
    "period": "Q2-2025",
    "include_sections": [
      "retainer_utilization_chart",
      "monthly_usage_breakdown",
      "projected_year_end_status",
      "service_category_analysis",
      "recommendations"
    ],
    "output_format": "pdf",
    "output_path": "./retainer-reports/MegaCorp-Q2-2025-Quarterly-Review.pdf",
    "include_charts": true
  }
}
```

**Year-End Retainer Reconciliation**

```json
{
  "tool": "generate_summary",
  "parameters": {
    "report_type": "annual_retainer_reconciliation",
    "client_name": "MegaCorp (Annual Retainer 2025)",
    "period": "2025",
    "include_sections": [
      "total_retainer_value",
      "total_hours_used",
      "unused_hours_analysis",
      "overage_billing_summary",
      "next_year_recommendations"
    ],
    "output_format": "pdf",
    "output_path": "./retainer-reports/MegaCorp-2025-Annual-Reconciliation.pdf"
  }
}
```

**Benefits of This Approach:**
1. **Clear Tracking**: Separate invoices for setup, usage tracking, and overages
2. **Client Transparency**: Regular reports show exactly how retainer is being used
3. **Financial Management**: Easy to see retainer utilization and cash flow impact
4. **Planning Tool**: Usage patterns help plan future retainer sizes
5. **Professional Presentation**: Detailed reports maintain client confidence

Would you like me to show you how to set up automated monthly retainer reports, or help you configure multiple retainer clients with different terms?

## Error Handling and Troubleshooting

### CSV Import Problems

**User**: "I'm trying to import a CSV file but getting errors. The file has some dates in MM/DD/YYYY format and others in DD/MM/YYYY format, and some rows have empty client names. How do I fix this?"

**Claude**: CSV import issues like mixed date formats and missing data are common! Let me help you diagnose and fix these problems step by step.

First, let me preview your CSV to see exactly what we're dealing with:

```json
{
  "tool": "import_preview",
  "parameters": {
    "file_path": "/path/to/problematic-timesheet.csv",
    "show_sample": true,
    "detect_issues": true,
    "analyze_date_formats": true,
    "show_data_quality": true
  }
}
```

I can see the issues! Here's what the preview found:

**🔍 Data Quality Issues Detected:**
1. **Mixed Date Formats**: 
   - Rows 1-15: MM/DD/YYYY format (e.g., "03/15/2025")
   - Rows 16-30: DD/MM/YYYY format (e.g., "15/03/2025")
2. **Missing Client Names**: 8 rows have empty client fields
3. **Inconsistent Hours Format**: Some show "8.5", others show "8:30"

Let me fix these issues with a comprehensive import strategy:

**Step 1: Validate and Clean the Data**

```json
{
  "tool": "import_validate",
  "parameters": {
    "file_path": "/path/to/problematic-timesheet.csv",
    "validation_rules": {
      "fix_date_formats": true,
      "handle_missing_clients": "prompt",
      "standardize_time_format": true,
      "remove_empty_rows": true
    },
    "date_format_detection": "smart",
    "create_cleaned_file": true,
    "cleaned_file_path": "/path/to/cleaned-timesheet.csv"
  }
}
```

**Step 2: Handle Missing Client Names**

For the 8 rows with missing clients, I found these patterns:
- 3 rows: Work description mentions "Acme project" → Likely Acme Corp
- 2 rows: Work description mentions "website for TechCo" → Likely TechCo
- 3 rows: Generic descriptions → Need your input

Let me import with smart client resolution:

```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/path/to/cleaned-timesheet.csv",
    "import_mode": "interactive_fix",
    "date_format": "smart_detection",
    "client_resolution": {
      "missing_client_strategy": "infer_from_description",
      "client_mapping": {
        "acme project": "Acme Corp",
        "website for techco": "TechCo",
        "generic work": "prompt_user"
      }
    },
    "time_format": "decimal_hours",
    "skip_invalid_rows": false,
    "generate_error_report": true
  }
}
```

**Step 3: Review Import Results**

The system found and fixed most issues:

**✅ Successfully Fixed:**
- ✅ 22 dates converted to standard format
- ✅ 5 client names inferred from descriptions
- ✅ All time formats standardized to decimal hours
- ✅ 2 duplicate rows removed

**⚠️ Needs Your Input:**
3 rows still need client assignment:
- Row 23: "Misc consulting work, 4 hours" → Which client?
- Row 27: "Documentation review, 2 hours" → Which client?
- Row 29: "Phone call follow-up, 0.5 hours" → Which client?

**❌ Could Not Import:**
- 1 row with invalid date "32/15/2025" → Please fix manually

For the rows needing client assignment, here are your options:

**Option 1: Manual Assignment**
```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/path/to/cleaned-timesheet.csv",
    "manual_client_assignment": {
      "row_23": "Default Client",
      "row_27": "Default Client", 
      "row_29": "Default Client"
    },
    "skip_rows": [31],
    "final_import": true
  }
}
```

**Option 2: Create Holding Invoice**
```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/path/to/cleaned-timesheet.csv",
    "unassigned_client_strategy": "create_holding_invoice",
    "holding_client": "Unassigned Work - Needs Review",
    "final_import": true
  }
}
```

**Preventing Future Issues:**

1. **Standardize Your Export**: Configure your time tracking app to always export in MM/DD/YYYY format
2. **Required Fields**: Make client name mandatory in your time tracking
3. **Regular Validation**: Run monthly validation checks before importing
4. **Backup Strategy**: Always validate before importing large datasets

Would you like me to:
1. Help you assign those 3 unassigned rows to specific clients?
2. Show you how to configure your time tracking app for consistent exports?
3. Set up automated validation rules for future imports?

### Invoice Generation Failures

**User**: "I'm trying to generate an HTML invoice but it keeps failing. The error mentions something about a template not found and missing client information. The invoice shows up when I list invoices, so I know it exists."

**Claude**: Template and data issues during HTML generation are frustrating! Let me help you diagnose and fix this problem systematically.

First, let me examine the specific invoice to understand what data might be missing:

```json
{
  "tool": "invoice_show",
  "parameters": {
    "invoice_identifier": "INV-2025-012",
    "include_validation": true,
    "check_completeness": true
  }
}
```

Now let me check your template configuration and availability:

```json
{
  "tool": "config_validate",
  "parameters": {
    "focus_areas": ["templates", "invoice_generation"],
    "check_template_paths": true,
    "verify_template_syntax": true,
    "test_sample_generation": true
  }
}
```

**🔍 Issues Found:**

**Template Issues:**
1. **Missing Template File**: `/templates/modern.html` not found
2. **Permission Problem**: Template directory not readable
3. **Syntax Error**: Line 45 in default.html has unclosed div tag

**Invoice Data Issues:**
1. **Missing Client Data**: Client "TechCorp" exists but missing address field
2. **Empty Work Items**: Invoice has 0 work items (total calculation fails)
3. **Invalid Date**: Due date is set to "Invalid Date"

Let me fix these issues step by step:

**Step 1: Fix Template Issues**

```json
{
  "tool": "config_init",
  "parameters": {
    "repair_mode": true,
    "fix_template_issues": true,
    "restore_default_templates": true,
    "template_directory": "./templates/",
    "backup_existing": true
  }
}
```

**Step 2: Fix Client Data**

```json
{
  "tool": "client_update",
  "parameters": {
    "client_identifier": "TechCorp",
    "address": "123 Business Ave, Tech City, TC 12345",
    "validation_fix": true
  }
}
```

**Step 3: Fix Invoice Data**

Let me check what's wrong with the invoice data:

```json
{
  "tool": "invoice_show",
  "parameters": {
    "invoice_identifier": "INV-2025-012",
    "show_raw_data": true
  }
}
```

I see the problems! The invoice has no work items and an invalid due date. Let me fix these:

```json
{
  "tool": "invoice_update",
  "parameters": {
    "invoice_identifier": "INV-2025-012",
    "due_date": "2025-09-15",
    "fix_data_issues": true
  }
}
```

The invoice appears to be missing work items. Let me add some based on the invoice description:

```json
{
  "tool": "invoice_add_item",
  "parameters": {
    "invoice_identifier": "INV-2025-012",
    "description": "Consulting services (recovered from description)",
    "hours": 10,
    "rate": 125,
    "notes": "Added during data recovery - please verify hours and rate"
  }
}
```

**Step 4: Test HTML Generation**

Now let's try generating the HTML with a working template:

```json
{
  "tool": "generate_html",
  "parameters": {
    "invoice_identifier": "INV-2025-012",
    "template": "default",
    "output_path": "./test-invoice.html",
    "validate_before_generation": true
  }
}
```

**✅ Success!** The HTML invoice generated properly. Here's what was fixed:

1. **Templates**: Restored default templates and fixed permissions
2. **Client Data**: Added missing address for TechCorp
3. **Invoice Data**: Fixed invalid due date and added missing work items
4. **Validation**: Added pre-generation validation to prevent future issues

**Preventing Future Issues:**

**1. Regular Validation Checks:**
```json
{
  "tool": "config_validate",
  "parameters": {
    "schedule": "weekly",
    "focus_areas": ["templates", "data_integrity"],
    "auto_fix_minor_issues": true
  }
}
```

**2. Template Backup Strategy:**
- Keep template backups in version control
- Test templates after any modifications
- Use template validation before making changes

**3. Data Quality Checks:**
- Validate invoices before generating documents
- Require complete client information
- Set up alerts for data consistency issues

**Common Template Issues to Watch For:**
- Missing template files
- Incorrect file permissions
- Syntax errors in HTML/CSS
- Missing placeholder variables
- Path configuration problems

**Common Data Issues:**
- Incomplete client information
- Empty work item lists
- Invalid dates
- Missing rate information
- Orphaned invoice records

Would you like me to:
1. Set up automated weekly validation to catch these issues early?
2. Show you how to create custom templates safely?
3. Help you validate all your existing invoices for similar issues?

## Advanced Usage Patterns

### Multi-Currency Client Management

**User**: "I have international clients who pay in different currencies. I have clients in the UK (GBP), Germany (EUR), and Japan (JPY). How should I handle this in go-invoice while maintaining proper accounting records?"

**Claude**: International multi-currency billing requires careful setup to handle exchange rates, accounting records, and client expectations. Let me show you a comprehensive approach that maintains accuracy and professionalism.

**Setting Up Currency-Specific Clients**

First, let's create clients with their preferred currencies clearly indicated:

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "London Tech Solutions Ltd (GBP)",
    "email": "accounts@londontechsolutions.co.uk",
    "phone": "+44-20-7123-4567",
    "address": "25 Canary Wharf, London E14 5AB, United Kingdom",
    "default_rate": 95,
    "payment_terms": "NET 30",
    "currency": "GBP",
    "tax_id": "GB123456789",
    "notes": "UK client - billing in GBP. Standard UK VAT applies (20%). Payment via SWIFT wire transfer."
  }
}
```

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "Berlin Innovation GmbH (EUR)",
    "email": "buchhaltung@berlininnovation.de",
    "phone": "+49-30-1234-5678",
    "address": "Potsdamer Platz 10, 10785 Berlin, Germany",
    "default_rate": 105,
    "payment_terms": "NET 30",
    "currency": "EUR",
    "tax_id": "DE123456789",
    "notes": "German client - billing in EUR. German VAT (19%) applies. SEPA payment preferred."
  }
}
```

```json
{
  "tool": "client_create",
  "parameters": {
    "name": "Tokyo Systems KK (JPY)",
    "email": "keiri@tokyosystems.co.jp",
    "phone": "+81-3-1234-5678",
    "address": "1-2-3 Shibuya, Shibuya-ku, Tokyo 150-0002, Japan",
    "default_rate": 11500,
    "payment_terms": "NET 30",
    "currency": "JPY",
    "tax_id": "JP123456789",
    "notes": "Japanese client - billing in JPY. Japanese consumption tax (10%) applies. Payment via international wire transfer."
  }
}
```

**Creating Currency-Specific Invoices**

For UK client (GBP):
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "London Tech Solutions Ltd (GBP)",
    "description": "Software Development Services - Q3 2025",
    "invoice_date": "2025-08-01",
    "due_date": "2025-08-31",
    "currency": "GBP",
    "work_items": [
      {
        "description": "Full-stack web application development",
        "hours": 40,
        "rate": 95,
        "currency": "GBP"
      },
      {
        "description": "API integration and testing",
        "hours": 20,
        "rate": 105,
        "currency": "GBP"
      }
    ],
    "tax_rate": 20.0,
    "tax_label": "UK VAT (20%)",
    "notes": "Payment in GBP via SWIFT transfer. VAT included as per UK regulations."
  }
}
```

For German client (EUR):
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "Berlin Innovation GmbH (EUR)",
    "description": "Beratungsdienstleistungen - Software Architektur",
    "invoice_date": "2025-08-01",
    "due_date": "2025-08-31",
    "currency": "EUR",
    "work_items": [
      {
        "description": "Software architecture consulting",
        "hours": 32,
        "rate": 105,
        "currency": "EUR"
      },
      {
        "description": "Technical documentation (German)",
        "hours": 16,
        "rate": 95,
        "currency": "EUR"
      }
    ],
    "tax_rate": 19.0,
    "tax_label": "Deutsche MwSt (19%)",
    "notes": "Zahlung in EUR via SEPA-Überweisung bevorzugt. MwSt gemäß deutschen Bestimmungen."
  }
}
```

For Japanese client (JPY):
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "Tokyo Systems KK (JPY)",
    "description": "システム開発コンサルティング (System Development Consulting)",
    "invoice_date": "2025-08-01",
    "due_date": "2025-08-31",
    "currency": "JPY",
    "work_items": [
      {
        "description": "Mobile application development consulting",
        "hours": 35,
        "rate": 11500,
        "currency": "JPY"
      },
      {
        "description": "Code review and optimization",
        "hours": 15,
        "rate": 12000,
        "currency": "JPY"
      }
    ],
    "tax_rate": 10.0,
    "tax_label": "消費税 (Consumption Tax 10%)",
    "notes": "Payment in JPY via international wire transfer. All amounts include Japanese consumption tax."
  }
}
```

**Currency-Specific HTML Generation**

Generate invoices with proper currency formatting:

```json
{
  "tool": "generate_html",
  "parameters": {
    "invoice_identifier": "INV-UK-2025-001",
    "template": "international",
    "currency_formatting": {
      "currency": "GBP",
      "locale": "en-GB",
      "symbol_position": "before",
      "decimal_places": 2
    },
    "language_options": {
      "primary_language": "en",
      "include_translations": false
    },
    "output_path": "./international-invoices/UK-invoice-GBP.html"
  }
}
```

**Multi-Currency Reporting and Analysis**

Monthly currency summary:
```json
{
  "tool": "generate_summary",
  "parameters": {
    "report_type": "multi_currency_summary",
    "period": "August 2025",
    "include_sections": [
      "revenue_by_currency",
      "exchange_rate_summary",
      "payment_status_by_currency",
      "conversion_to_base_currency"
    ],
    "base_currency": "USD",
    "exchange_rate_source": "xe.com",
    "output_format": "pdf",
    "output_path": "./reports/Multi-Currency-August-2025.pdf"
  }
}
```

**Currency Exchange Rate Tracking**

For accounting purposes, track exchange rates at invoice creation:
```json
{
  "tool": "invoice_update",
  "parameters": {
    "invoice_identifier": "INV-UK-2025-001",
    "add_metadata": {
      "exchange_rate_usd": 1.27,
      "exchange_rate_date": "2025-08-01",
      "exchange_rate_source": "xe.com",
      "usd_equivalent": 8128.00
    },
    "notes": "Exchange rate: 1 GBP = 1.27 USD as of 2025-08-01"
  }
}
```

**Annual Multi-Currency Tax Reporting**

Year-end export for tax purposes:
```json
{
  "tool": "export_data",
  "parameters": {
    "export_type": "multi_currency_tax_report",
    "date_range": {
      "start": "2025-01-01",
      "end": "2025-12-31"
    },
    "include_fields": [
      "invoice_number",
      "client_name",
      "currency",
      "foreign_amount",
      "exchange_rate",
      "usd_equivalent",
      "tax_amount_foreign",
      "tax_amount_usd",
      "payment_date"
    ],
    "group_by": "currency",
    "include_exchange_rates": true,
    "output_path": "./tax-reports/2025-Multi-Currency-Tax-Report.csv"
  }
}
```

**Best Practices for Multi-Currency Management:**

1. **Clear Currency Identification**: Always include currency in client names and invoice numbers
2. **Consistent Rate Documentation**: Record exchange rates at invoice creation, not payment
3. **Tax Compliance**: Understand VAT/tax requirements for each country
4. **Payment Instructions**: Provide clear, currency-specific payment instructions
5. **Regular Reconciliation**: Monthly currency reconciliation for accounting accuracy

**Template Considerations:**
- Use currency-appropriate number formatting
- Include appropriate tax labels and rates
- Provide payment instructions in local language when possible
- Consider cultural preferences for invoice layout

Would you like me to:
1. Show you how to set up automated exchange rate updates?
2. Create templates for specific country requirements?
3. Help you configure payment tracking across different currencies?

### Automated Workflow Integration

**User**: "I want to automate my entire billing workflow. When I finish a project phase, I want the system to automatically create an invoice, send it to the client, update my project management system, and create accounting entries. Is this possible with go-invoice?"

**Claude**: Absolutely! Creating an automated end-to-end billing workflow is one of the most valuable uses of go-invoice's MCP tools. Let me design a comprehensive automation system that handles your entire billing pipeline.

**Workflow Overview:**
Project Completion → Auto Invoice Creation → Client Notification → External System Updates → Accounting Integration → Follow-up Scheduling

**Step 1: Project Completion Trigger Setup**

First, let's create a standardized project completion process:

```json
{
  "tool": "config_init",
  "parameters": {
    "automation_setup": true,
    "workflow_triggers": [
      "project_phase_completion",
      "milestone_achievement",
      "time_threshold_reached"
    ],
    "integration_endpoints": {
      "project_management": "asana_api",
      "accounting": "quickbooks_api",
      "email": "smtp_configured"
    }
  }
}
```

**Step 2: Automated Invoice Creation**

When a project phase completes, automatically create and customize the invoice:

```json
{
  "tool": "invoice_create",
  "parameters": {
    "automation_mode": true,
    "client_name": "{{project.client_name}}",
    "description": "{{project.name}} - {{phase.name}} Completion",
    "project_name": "{{project.name}}",
    "work_items": "{{auto_generate_from_timesheet}}",
    "due_date": "{{client.payment_terms}}",
    "template_variables": {
      "project_milestone": "{{phase.milestone}}",
      "completion_date": "{{phase.completion_date}}",
      "next_phase": "{{phase.next_phase}}"
    },
    "auto_validate": true,
    "auto_calculate": true
  }
}
```

**Step 3: Automated Document Generation**

Generate professional invoice documents with project-specific customization:

```json
{
  "tool": "generate_html",
  "parameters": {
    "invoice_identifier": "{{generated_invoice.id}}",
    "template": "{{client.preferred_template}}",
    "automation_mode": true,
    "customizations": {
      "include_project_summary": true,
      "include_milestone_achievements": true,
      "include_next_phase_preview": true,
      "attach_deliverables_summary": true
    },
    "output_path": "./automated-invoices/{{invoice.number}}-{{client.safe_name}}.html",
    "generate_pdf_copy": true
  }
}
```

**Step 4: Automated Client Communication**

Send customized emails with project context:

```json
{
  "tool": "send_invoice_email",
  "parameters": {
    "invoice_identifier": "{{generated_invoice.id}}",
    "automation_mode": true,
    "email_template": "phase_completion_billing",
    "personalization": {
      "project_name": "{{project.name}}",
      "phase_completed": "{{phase.name}}",
      "completion_highlights": "{{phase.achievements}}",
      "next_steps": "{{phase.next_phase_preview}}"
    },
    "attachments": [
      "{{invoice.html_path}}",
      "{{invoice.pdf_path}}",
      "{{phase.deliverables_summary}}"
    ],
    "schedule_follow_up": {
      "reminder_days": [7, 14, 21],
      "escalation_enabled": true
    }
  }
}
```

**Step 5: Project Management System Integration**

Update external project management tools:

```json
{
  "tool": "external_integration",
  "parameters": {
    "integration_type": "project_management",
    "action": "update_project_status",
    "system": "asana",
    "updates": {
      "project_id": "{{project.external_id}}",
      "phase_status": "completed_and_billed",
      "invoice_number": "{{generated_invoice.number}}",
      "billing_amount": "{{generated_invoice.total}}",
      "billing_date": "{{generated_invoice.date}}",
      "next_phase_status": "ready_to_start"
    },
    "create_tasks": [
      {
        "name": "Follow up on invoice {{invoice.number}}",
        "assignee": "{{project.manager}}",
        "due_date": "{{invoice.due_date_minus_3_days}}"
      },
      {
        "name": "Prepare next phase materials",
        "assignee": "{{project.team_lead}}",
        "due_date": "{{next_phase.start_date_minus_1_week}}"
      }
    ]
  }
}
```

**Step 6: Accounting System Integration**

Automatically create accounting entries:

```json
{
  "tool": "accounting_integration",
  "parameters": {
    "system": "quickbooks",
    "action": "create_invoice_entry",
    "invoice_data": {
      "customer": "{{client.quickbooks_id}}",
      "invoice_number": "{{generated_invoice.number}}",
      "line_items": "{{invoice.work_items}}",
      "tax_rate": "{{client.tax_rate}}",
      "payment_terms": "{{client.payment_terms}}",
      "project_code": "{{project.accounting_code}}"
    },
    "create_receivable": true,
    "update_project_costs": true,
    "generate_profit_analysis": true
  }
}
```

**Step 7: Automated Reporting and Analytics**

Generate automated reports for stakeholders:

```json
{
  "tool": "generate_summary",
  "parameters": {
    "automation_trigger": "invoice_creation",
    "report_type": "project_billing_summary",
    "stakeholders": ["project_manager", "finance_team", "client_success"],
    "include_sections": [
      "phase_completion_summary",
      "billing_details",
      "project_progress",
      "financial_performance",
      "next_phase_preparation"
    ],
    "distribution": {
      "project_manager": ["project_progress", "next_phase_preparation"],
      "finance_team": ["billing_details", "financial_performance"],
      "client_success": ["phase_completion_summary", "client_communication_log"]
    },
    "output_format": "email_digest",
    "schedule": "immediate"
  }
}
```

**Complete Automation Script Example**

Here's how you might trigger the entire workflow:

```bash
#!/bin/bash
# Project Phase Completion Automation Script

# 1. Detect project completion (could be triggered by your project management tool)
PROJECT_ID="PROJ-2025-WEBAPP"
PHASE_NAME="Backend Development"
CLIENT_NAME="TechCorp Industries"

# 2. Import final timesheet data
echo "Importing final timesheet data..."
curl -X POST "http://localhost:3001/mcp/tools/import_csv" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "./timesheets/'$PROJECT_ID'-'$PHASE_NAME'-final.csv",
    "client_name": "'$CLIENT_NAME'",
    "import_mode": "append_to_project",
    "project_name": "'$PROJECT_ID'"
  }'

# 3. Auto-create invoice
echo "Creating invoice for completed phase..."
INVOICE_RESPONSE=$(curl -X POST "http://localhost:3001/mcp/tools/invoice_create" \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "'$CLIENT_NAME'",
    "description": "'$PROJECT_ID' - '$PHASE_NAME' Completion",
    "project_name": "'$PROJECT_ID'",
    "auto_import_timesheet": true,
    "due_days": 30
  }')

# Extract invoice ID from response
INVOICE_ID=$(echo $INVOICE_RESPONSE | jq -r '.invoice_id')

# 4. Generate invoice documents
echo "Generating invoice documents..."
curl -X POST "http://localhost:3001/mcp/tools/generate_html" \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_identifier": "'$INVOICE_ID'",
    "template": "professional",
    "generate_pdf": true,
    "include_project_summary": true
  }'

# 5. Send to client
echo "Sending invoice to client..."
curl -X POST "http://localhost:3001/mcp/tools/send_invoice" \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_identifier": "'$INVOICE_ID'",
    "email_template": "phase_completion",
    "schedule_reminders": true
  }'

# 6. Update project management system
echo "Updating project management system..."
curl -X POST "http://localhost:3001/mcp/tools/external_integration" \
  -H "Content-Type: application/json" \
  -d '{
    "system": "asana",
    "action": "mark_phase_complete_and_billed",
    "project_id": "'$PROJECT_ID'",
    "phase_name": "'$PHASE_NAME'",
    "invoice_id": "'$INVOICE_ID'"
  }'

# 7. Create accounting entries
echo "Creating accounting entries..."
curl -X POST "http://localhost:3001/mcp/tools/accounting_integration" \
  -H "Content-Type: application/json" \
  -d '{
    "system": "quickbooks",
    "action": "create_invoice",
    "invoice_id": "'$INVOICE_ID'"
  }'

echo "Automated billing workflow completed for '$PROJECT_ID' - '$PHASE_NAME'"
echo "Invoice ID: $INVOICE_ID"
echo "Next phase can now begin!"
```

**Monitoring and Error Handling**

Set up monitoring for the automated workflow:

```json
{
  "tool": "automation_monitoring",
  "parameters": {
    "workflow_name": "project_phase_billing",
    "error_handling": {
      "retry_attempts": 3,
      "fallback_to_manual": true,
      "notification_channels": ["email", "slack"],
      "escalation_delay": "30_minutes"
    },
    "success_metrics": [
      "invoice_created",
      "email_sent",
      "systems_updated",
      "accounting_synced"
    ],
    "reporting": {
      "daily_summary": true,
      "weekly_performance": true,
      "monthly_optimization": true
    }
  }
}
```

**Benefits of This Automated Workflow:**

1. **Time Savings**: Reduces manual billing time from hours to minutes
2. **Consistency**: Ensures every phase completion follows the same process
3. **Accuracy**: Eliminates manual data entry errors
4. **Client Experience**: Immediate, professional communication
5. **Cash Flow**: Faster invoicing improves payment timing
6. **Integration**: Keeps all systems synchronized automatically
7. **Scalability**: Handles multiple projects without additional effort

**Customization Options:**
- Different workflows for different client types
- Varying automation levels (full auto vs. review-before-send)
- Custom approval processes for large invoices
- Integration with additional tools (CRM, time tracking, etc.)

Would you like me to:
1. Help you set up specific integrations with your existing tools?
2. Create custom email templates for different project types?
3. Design approval workflows for high-value invoices?
4. Set up monitoring and reporting for the automated system?

This comprehensive documentation provides extensive examples of natural language interactions with the go-invoice MCP tools system, covering everything from basic setup to complex automated workflows.