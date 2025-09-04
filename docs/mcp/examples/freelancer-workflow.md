# Freelancer Workflow - Complete Invoicing Process

This example demonstrates a complete freelancer workflow from client setup through payment tracking. This scenario showcases how a freelance web developer manages their invoicing process using natural language interaction with Claude.

## Scenario Overview

**Sarah** is a freelance web developer who just completed a project for a new client, "TechStart Inc." She needs to:
1. Set up the new client in her system
2. Create an invoice for the completed work
3. Generate a professional HTML invoice to send
4. Track the invoice status through payment

This workflow demonstrates **8 different MCP tools** in a realistic business context.

---

## The Complete Conversation

**Sarah**: "Hi Claude! I just finished a website project for a new client called TechStart Inc. I need to set them up in my system and create an invoice for $4,500. Their billing contact is finance@techstart.com and they're located at 123 Innovation Drive, Austin, TX 78701."

**Claude**: I'll help you set up TechStart Inc as a new client and create an invoice for your website project! Let me start by adding them to your client database with their contact information.

*Using client_create tool:*
```json
{
  "tool": "client_create",
  "parameters": {
    "name": "TechStart Inc",
    "email": "finance@techstart.com",
    "address": "123 Innovation Drive, Austin, TX 78701"
  }
}
```

Perfect! TechStart Inc has been added to your client database with ID `client_ts_001`. Now let's create the invoice for your website development work. What was the project about specifically, and when should payment be due?

**Sarah**: "It was a complete website redesign and development project. I worked on it from July 15th through August 2nd, putting in about 30 hours total at my standard rate of $150/hour. They should pay within 30 days."

**Claude**: Excellent! Let me create a professional invoice for your website development project. I'll set it up with a 30-day payment term.

*Using invoice_create tool:*
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "TechStart Inc",
    "description": "Website Redesign and Development Project",
    "invoice_date": "2025-08-03",
    "due_days": 30,
    "work_items": [
      {
        "date": "2025-07-15",
        "hours": 8.0,
        "rate": 150.0,
        "description": "Initial consultation and wireframe development"
      },
      {
        "date": "2025-07-18",
        "hours": 8.0,
        "rate": 150.0,
        "description": "Homepage design and user experience planning"
      },
      {
        "date": "2025-07-22",
        "hours": 6.0,
        "rate": 150.0,
        "description": "Frontend development - responsive layout"
      },
      {
        "date": "2025-07-25",
        "hours": 8.0,
        "rate": 150.0,
        "description": "Backend integration and CMS setup"
      }
    ]
  }
}
```

Great! Invoice INV-001 has been created for $4,500.00 with a due date of September 2nd, 2025. The invoice includes all your work items totaling 30 hours at $150/hour. Now let me generate a professional HTML version that you can send to the client.

*Using generate_html tool:*
```json
{
  "tool": "generate_html",
  "parameters": {
    "invoice_number": "INV-001",
    "template": "professional",
    "include_logo": true,
    "company_name": "Sarah Johnson Web Development",
    "output_path": "./invoices/TechStart-INV-001.html"
  }
}
```

Perfect! I've generated a professional HTML invoice at `./invoices/TechStart-INV-001.html`. The invoice looks polished and includes your branding. You can now email this to finance@techstart.com.

**Sarah**: "That's great! Before I send it, can you show me exactly what the invoice looks like? I want to make sure all the details are correct."

**Claude**: Absolutely! Let me display the complete invoice details so you can review everything before sending.

*Using invoice_show tool:*
```json
{
  "tool": "invoice_show",
  "parameters": {
    "invoice_number": "INV-001",
    "show_work_items": true,
    "show_client_details": true,
    "output_format": "text"
  }
}
```

Here's your complete invoice:

**Invoice INV-001**
- **Client**: TechStart Inc (finance@techstart.com)
- **Project**: Website Redesign and Development Project
- **Invoice Date**: August 3, 2025
- **Due Date**: September 2, 2025
- **Status**: Draft

**Work Items:**
1. July 15, 2025 - 8.0 hours @ $150/hr - Initial consultation and wireframe development ($1,200.00)
2. July 18, 2025 - 8.0 hours @ $150/hr - Homepage design and user experience planning ($1,200.00)
3. July 22, 2025 - 6.0 hours @ $150/hr - Frontend development - responsive layout ($900.00)
4. July 25, 2025 - 8.0 hours @ $150/hr - Backend integration and CMS setup ($1,200.00)

**Total: $4,500.00**

Everything looks perfect! Would you like me to mark it as sent once you email it to the client?

**Sarah**: "Yes, this looks perfect! I'm sending it now. Please mark it as sent."

**Claude**: I'll mark the invoice as sent now that you've delivered it to the client.

*Using invoice_update tool:*
```json
{
  "tool": "invoice_update",
  "parameters": {
    "invoice_number": "INV-001",
    "status": "sent"
  }
}
```

Excellent! Invoice INV-001 is now marked as "sent" with today's timestamp. The system will now track it as an outstanding receivable until payment is received.

**Sarah**: "Two weeks later... Good news! TechStart just paid the invoice. Can you help me update the system and show me my current financial status?"

**Claude**: That's wonderful news! Congratulations on getting paid! Let me mark the invoice as paid and then show you an overview of your current invoicing situation.

*Using invoice_update tool:*
```json
{
  "tool": "invoice_update",
  "parameters": {
    "invoice_number": "INV-001",
    "status": "paid"
  }
}
```

Perfect! INV-001 is now marked as paid. Let me give you a summary of your current invoice status.

*Using invoice_list tool:*
```json
{
  "tool": "invoice_list",
  "parameters": {
    "include_summary": true,
    "output_format": "table",
    "sort_by": "invoice_date",
    "sort_order": "desc"
  }
}
```

Here's your current invoice overview:

**Recent Invoices:**
- INV-001: TechStart Inc - $4,500.00 - PAID ✅

**Financial Summary:**
- Total Outstanding: $0.00
- Total Paid This Month: $4,500.00
- Average Payment Time: 14 days

Excellent work! You have no outstanding invoices and just earned $4,500 this month. TechStart paid in just 14 days, which is great for cash flow.

**Sarah**: "This is fantastic! Can you also help me generate a summary report that I can save for my records and tax preparation?"

**Claude**: Absolutely! Let me generate a comprehensive summary report that you can use for your records and tax preparation.

*Using generate_summary tool:*
```json
{
  "tool": "generate_summary",
  "parameters": {
    "from_date": "2025-08-01",
    "to_date": "2025-08-31",
    "include_paid_invoices": true,
    "include_client_breakdown": true,
    "output_format": "detailed",
    "export_path": "./reports/August-2025-Summary.pdf"
  }
}
```

Perfect! I've generated a comprehensive August 2025 summary report saved as `./reports/August-2025-Summary.pdf`. The report includes:

- **Monthly Revenue**: $4,500.00
- **Client Breakdown**: TechStart Inc (100% of revenue)
- **Project Details**: Website development services
- **Payment Timeline**: 14-day average collection period
- **Tax Information**: All details formatted for easy tax preparation

This report is perfect for your records and will make tax time much easier!

---

## Tools Demonstrated

This workflow showcased **8 different MCP tools**:

1. **client_create** - Adding new client with contact information
2. **invoice_create** - Creating detailed invoice with work items
3. **generate_html** - Producing professional invoice for client delivery
4. **invoice_show** - Reviewing invoice details before sending
5. **invoice_update** - Marking invoice as sent, then as paid
6. **invoice_list** - Getting overview of invoice status and financial summary
7. **generate_summary** - Creating comprehensive report for records
8. **export_data** (implied through generate_summary) - Exporting data for tax preparation

## Key Business Benefits

- **Streamlined Client Onboarding**: New clients added in seconds with all contact details
- **Professional Invoice Generation**: Automated creation of polished, branded invoices
- **Real-time Status Tracking**: Always know which invoices are outstanding vs. paid
- **Automated Financial Reporting**: Monthly summaries generated automatically for tax preparation
- **Time Savings**: What used to take hours of manual work now takes minutes of conversation
- **Cash Flow Visibility**: Clear overview of outstanding receivables and payment timelines

## Next Steps

From here, Sarah could:
- Set up recurring invoices for retainer clients
- Import time tracking data from CSV files
- Generate year-end tax reports
- Manage multiple client relationships
- Track project profitability

This workflow demonstrates how go-invoice's MCP integration transforms routine invoicing from a tedious administrative task into a smooth, conversational process that freelancers can complete in minutes.
