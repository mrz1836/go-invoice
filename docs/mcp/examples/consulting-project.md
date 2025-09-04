# Consulting Project - From Timesheet to Invoice

This example demonstrates a complete consulting project workflow where a business consultant imports timesheet data, validates it, creates invoices, and manages the entire billing process. This scenario showcases how **Michael**, a management consultant, handles a multi-week project for a corporate client.

## Scenario Overview

**Michael** is a management consultant who just completed a 3-week organizational restructuring project for "GlobalTech Solutions." He needs to:
1. Import his detailed timesheet data from his time tracking app
2. Validate the data before processing
3. Create a professional invoice with all work breakdown
4. Handle client feedback and invoice adjustments
5. Track payment and generate final reports

This workflow demonstrates **12 different MCP tools** in a realistic consulting context.

---

## The Complete Conversation

**Michael**: "Hi Claude! I just finished a 3-week consulting project for GlobalTech Solutions. I have my timesheet data in a CSV file that I exported from my time tracking app. Can you help me turn this into a professional invoice? The file is at `/Users/michael/timesheets/globaltech-july-2025.csv`."

**Claude**: I'd be happy to help you create a professional invoice from your timesheet data! Let me start by previewing the CSV file to understand the structure and validate the data before we import it.

*Using import_preview tool:*
```json
{
  "tool": "import_preview",
  "parameters": {
    "file_path": "/Users/michael/timesheets/globaltech-july-2025.csv",
    "preview_rows": 10,
    "show_column_mapping": true
  }
}
```

Great! I can see your timesheet has excellent detail with columns for date, hours, rate, description, and project phase. Let me validate the data structure to ensure everything will import correctly.

*Using import_validate tool:*
```json
{
  "tool": "import_validate",
  "parameters": {
    "file_path": "/Users/michael/timesheets/globaltech-july-2025.csv",
    "check_data_integrity": true,
    "validate_rates": true,
    "check_date_ranges": true
  }
}
```

Excellent! Your timesheet data is well-structured:
- **21 work entries** covering July 8-26, 2025
- **Total hours**: 168.5 hours
- **Rate validation**: Consistent $200/hour rate
- **No data errors** detected

Before importing, let me check if GlobalTech Solutions is already in your client database.

*Using client_list tool:*
```json
{
  "tool": "client_list",
  "parameters": {
    "search_term": "GlobalTech",
    "output_format": "table"
  }
}
```

I don't see GlobalTech Solutions in your client database yet. Let me create their client record first, then import your timesheet data.

**Michael**: "Yes, they're a new client. Their billing contact is accounts.payable@globaltech.com, and they're located at 500 Corporate Plaza, Dallas, TX 75201. They have NET-30 payment terms."

**Claude**: Perfect! Let me set up GlobalTech Solutions as a new client, then import your timesheet data to create the invoice.

*Using client_create tool:*
```json
{
  "tool": "client_create",
  "parameters": {
    "name": "GlobalTech Solutions",
    "email": "accounts.payable@globaltech.com",
    "address": "500 Corporate Plaza, Dallas, TX 75201"
  }
}
```

Excellent! GlobalTech Solutions is now in your system with client ID `client_gts_001`. Now let me import your timesheet data and create the invoice automatically.

*Using import_csv tool:*
```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/Users/michael/timesheets/globaltech-july-2025.csv",
    "client_name": "GlobalTech Solutions",
    "description": "Organizational Restructuring Consulting - July 2025",
    "import_mode": "new_invoice",
    "due_days": 30,
    "field_mapping": {
      "date": "work_date",
      "hours": "hours_worked",
      "rate": "hourly_rate",
      "description": "task_description"
    }
  }
}
```

Fantastic! Invoice INV-002 has been created successfully:
- **Client**: GlobalTech Solutions
- **Project**: Organizational Restructuring Consulting - July 2025
- **Total Hours**: 168.5 hours @ $200/hour
- **Total Amount**: $33,700.00
- **Due Date**: September 2, 2025 (NET-30 terms)

Let me show you the complete invoice breakdown:

*Using invoice_show tool:*
```json
{
  "tool": "invoice_show",
  "parameters": {
    "invoice_number": "INV-002",
    "show_work_items": true,
    "show_client_details": true,
    "output_format": "text"
  }
}
```

**Invoice INV-002 - GlobalTech Solutions**

**Project**: Organizational Restructuring Consulting - July 2025
**Total**: $33,700.00 (168.5 hours @ $200/hour)

**Work Breakdown by Week:**

**Week 1 (July 8-12):**
- Current state analysis and stakeholder interviews (48 hours)
- Document review and process mapping (12 hours)
- **Week 1 Subtotal**: $12,000.00

**Week 2 (July 15-19):**
- Organizational design workshops (40 hours)
- Department restructuring planning (16 hours)
- **Week 2 Subtotal**: $11,200.00

**Week 3 (July 22-26):**
- Implementation roadmap development (32 hours)
- Change management planning (20.5 hours)
- **Week 3 Subtotal**: $10,500.00

This looks comprehensive! Would you like me to generate a professional HTML version for client presentation?

**Michael**: "Yes, please create a professional HTML invoice. Also, I realized I need to add one more entry - I had a follow-up call with their CEO yesterday for 2 hours at my standard rate. Can you add that to the invoice?"

**Claude**: Absolutely! Let me first add that follow-up call to the invoice, then generate the professional HTML version.

*Using invoice_add_item tool:*
```json
{
  "tool": "invoice_add_item",
  "parameters": {
    "invoice_number": "INV-002",
    "work_items": [
      {
        "date": "2025-08-02",
        "hours": 2.0,
        "rate": 200.0,
        "description": "Executive follow-up call with CEO - implementation discussion"
      }
    ]
  }
}
```

Perfect! The follow-up call has been added. Your invoice total is now $34,100.00 (170.5 hours). Now let me generate the professional HTML version:

*Using generate_html tool:*
```json
{
  "tool": "generate_html",
  "parameters": {
    "invoice_number": "INV-002",
    "template": "professional",
    "include_logo": true,
    "company_name": "Michael Chen Consulting",
    "output_path": "./invoices/GlobalTech-Consulting-INV-002.html"
  }
}
```

Excellent! I've generated a professional HTML invoice at `./invoices/GlobalTech-Consulting-INV-002.html`. The invoice includes:
- Your professional branding
- Detailed work breakdown by week
- Clear project phases and deliverables
- Professional formatting suitable for corporate clients

**Michael**: "Perfect! I'm sending this to their accounts payable team now. Can you mark it as sent and set up tracking so I know when to follow up?"

**Claude**: Absolutely! Let me mark the invoice as sent and show you the tracking setup.

*Using invoice_update tool:*
```json
{
  "tool": "invoice_update",
  "parameters": {
    "invoice_number": "INV-002",
    "status": "sent"
  }
}
```

Great! INV-002 is now marked as sent. Let me also show you your current invoice status and receivables overview:

*Using invoice_list tool:*
```json
{
  "tool": "invoice_list",
  "parameters": {
    "status": "sent",
    "include_summary": true,
    "sort_by": "due_date",
    "sort_order": "asc"
  }
}
```

**Current Outstanding Invoices:**
- INV-002: GlobalTech Solutions - $34,100.00 - Due: Sep 2, 2025 (29 days remaining)

**Tracking Recommendations:**
- **First follow-up**: August 25, 2025 (1 week before due date)
- **Payment reminder**: September 1, 2025 (1 day before due date)
- **Collections follow-up**: September 5, 2025 (if payment not received)

I'll help you track this automatically. Would you like me to set reminders for these dates?

**Michael**: "Three weeks later... GlobalTech's accounting department is asking for the invoice to be split into two parts - they want Phase 1 (first two weeks) billed separately from Phase 2 (final week + follow-up). Can you help me restructure this?"

**Claude**: I can definitely help you restructure the invoice to match their accounting requirements! Let me create two separate invoices from the existing work items. First, let me get the detailed breakdown of the current invoice:

*Using invoice_show tool:*
```json
{
  "tool": "invoice_show",
  "parameters": {
    "invoice_number": "INV-002",
    "show_work_items": true,
    "output_format": "json"
  }
}
```

I can see all the work items. Since the original invoice is already sent, I'll create two new invoices and then void the original one. Let me create Phase 1 first:

*Using invoice_create tool:*
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "GlobalTech Solutions",
    "description": "Organizational Restructuring Consulting - Phase 1 (Analysis & Design)",
    "invoice_date": "2025-08-03",
    "due_days": 30,
    "work_items": [
      {
        "date": "2025-07-08",
        "hours": 8.0,
        "rate": 200.0,
        "description": "Current state analysis - stakeholder interviews"
      },
      {
        "date": "2025-07-09",
        "hours": 8.0,
        "rate": 200.0,
        "description": "Department workflow documentation"
      },
      {
        "date": "2025-07-10",
        "hours": 8.0,
        "rate": 200.0,
        "description": "Process mapping and gap analysis"
      },
      {
        "date": "2025-07-15",
        "hours": 8.0,
        "rate": 200.0,
        "description": "Organizational design workshop - day 1"
      },
      {
        "date": "2025-07-16",
        "hours": 8.0,
        "rate": 200.0,
        "description": "Organizational design workshop - day 2"
      },
      {
        "date": "2025-07-17",
        "hours": 8.0,
        "rate": 200.0,
        "description": "Department restructuring planning"
      }
    ]
  }
}
```

Great! Phase 1 invoice (INV-003) created for $9,600.00. Now let me create Phase 2:

*Using invoice_create tool:*
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "GlobalTech Solutions",
    "description": "Organizational Restructuring Consulting - Phase 2 (Implementation & Follow-up)",
    "invoice_date": "2025-08-03",
    "due_days": 30,
    "work_items": [
      {
        "date": "2025-07-22",
        "hours": 8.0,
        "rate": 200.0,
        "description": "Implementation roadmap development"
      },
      {
        "date": "2025-07-23",
        "hours": 8.0,
        "rate": 200.0,
        "description": "Change management planning"
      },
      {
        "date": "2025-07-24",
        "hours": 8.0,
        "rate": 200.0,
        "description": "Training program design"
      },
      {
        "date": "2025-08-02",
        "hours": 2.0,
        "rate": 200.0,
        "description": "Executive follow-up call with CEO"
      }
    ]
  }
}
```

Perfect! Phase 2 invoice (INV-004) created for $5,200.00. Now let me update the original invoice to show it was superseded:

*Using invoice_update tool:*
```json
{
  "tool": "invoice_update",
  "parameters": {
    "invoice_number": "INV-002",
    "status": "cancelled",
    "description": "Organizational Restructuring Consulting - SUPERSEDED by INV-003 and INV-004 per client request"
  }
}
```

Excellent! Here's your new invoice structure:

**Original Invoice**: INV-002 - $34,100.00 - CANCELLED (superseded)

**New Split Invoices**:
- **INV-003**: Phase 1 (Analysis & Design) - $9,600.00
- **INV-004**: Phase 2 (Implementation & Follow-up) - $5,200.00
- **Total**: $14,800.00 (matches the restructured scope)

Wait, I notice the totals don't match. Let me check the original work breakdown again and make sure we've captured all hours properly...

**Michael**: "Actually, that's correct - they only want to pay for the core consulting work, not all the additional analysis I did. The split amounts are what we agreed on. Can you generate professional HTML versions of both invoices?"

**Claude**: Understood! The split reflects the agreed-upon scope adjustment. Let me generate professional HTML versions of both Phase invoices:

*Using generate_html tool:*
```json
{
  "tool": "generate_html",
  "parameters": {
    "invoice_number": "INV-003",
    "template": "professional",
    "include_logo": true,
    "company_name": "Michael Chen Consulting",
    "output_path": "./invoices/GlobalTech-Phase1-INV-003.html"
  }
}
```

*Using generate_html tool:*
```json
{
  "tool": "generate_html",
  "parameters": {
    "invoice_number": "INV-004",
    "template": "professional",
    "include_logo": true,
    "company_name": "Michael Chen Consulting",
    "output_path": "./invoices/GlobalTech-Phase2-INV-004.html"
  }
}
```

Perfect! Both invoices are ready:
- `./invoices/GlobalTech-Phase1-INV-003.html` - Analysis & Design phase
- `./invoices/GlobalTech-Phase2-INV-004.html` - Implementation & Follow-up phase

Both invoices maintain professional formatting and clearly show the phase-based work breakdown that GlobalTech requested.

**Michael**: "Excellent! Can you also create a summary report for this entire project that I can use for my project portfolio?"

**Claude**: Absolutely! Let me generate a comprehensive project summary that showcases your consulting work and results.

*Using generate_summary tool:*
```json
{
  "tool": "generate_summary",
  "parameters": {
    "from_date": "2025-07-01",
    "to_date": "2025-08-31",
    "client_filter": "GlobalTech Solutions",
    "include_project_breakdown": true,
    "include_timeline": true,
    "output_format": "detailed",
    "export_path": "./reports/GlobalTech-Project-Summary-2025.pdf"
  }
}
```

Perfect! I've created a comprehensive project summary at `./reports/GlobalTech-Project-Summary-2025.pdf` that includes:

**Project Overview:**
- **Client**: GlobalTech Solutions
- **Project**: Organizational Restructuring Consulting
- **Duration**: July 8 - August 2, 2025 (3+ weeks)
- **Total Value**: $14,800.00 (adjusted scope)

**Phase Breakdown:**
- **Phase 1**: Analysis & Design - $9,600.00
- **Phase 2**: Implementation & Follow-up - $5,200.00

**Key Deliverables:**
- Current state analysis and stakeholder assessment
- Organizational design recommendations
- Implementation roadmap and change management plan
- Executive consultation and follow-up

This report is perfect for your portfolio and demonstrates your structured approach to complex organizational consulting projects.

---

## Tools Demonstrated

This workflow showcased **12 different MCP tools**:

1. **import_preview** - Previewing CSV structure before import
2. **import_validate** - Validating timesheet data integrity
3. **client_list** - Checking for existing client records
4. **client_create** - Setting up new corporate client
5. **import_csv** - Importing timesheet data to create invoice
6. **invoice_show** - Reviewing detailed invoice breakdown
7. **invoice_add_item** - Adding additional work entries
8. **generate_html** - Creating professional client-ready invoices
9. **invoice_update** - Managing invoice status changes and modifications
10. **invoice_list** - Tracking outstanding receivables
11. **invoice_create** - Creating split invoices for restructured billing
12. **generate_summary** - Producing comprehensive project reports

## Key Business Benefits

- **Efficient Data Import**: Seamlessly convert time tracking data into professional invoices
- **Data Validation**: Catch errors before they become billing problems
- **Flexible Billing**: Adapt to client requirements with invoice restructuring
- **Professional Presentation**: Generate polished invoices that reflect expertise
- **Project Documentation**: Create portfolio-ready project summaries
- **Client Relationship Management**: Handle complex client requests smoothly
- **Cash Flow Tracking**: Monitor receivables and payment timelines

## Consulting-Specific Advantages

- **Time-to-Invoice Reduction**: From hours of manual work to minutes of conversation
- **Project Phase Management**: Handle complex multi-phase billing requirements
- **Scope Adjustments**: Easily modify invoices to match changed project scope
- **Professional Image**: Consistent, branded invoice presentation
- **Audit Trail**: Complete documentation of project work and billing changes
- **Portfolio Building**: Automated generation of project summaries for business development

This workflow demonstrates how consultants can leverage go-invoice's MCP integration to handle complex, multi-phase projects with changing requirements while maintaining professionalism and efficiency throughout the billing process.
