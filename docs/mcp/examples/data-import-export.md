# Data Import & Export - Streamlining Data Workflows

This example demonstrates comprehensive data import and export workflows including CSV processing, data validation, and report generation. This scenario showcases how **David**, an agency owner, handles bulk data operations and creates various reports for different stakeholders.

## Scenario Overview

**David** runs a digital marketing agency that works with multiple time tracking systems and needs to integrate data from various sources. He needs to:
1. Import timesheet data from different time tracking applications
2. Validate and clean data before processing
3. Handle complex field mapping for different CSV formats
4. Export data in various formats for different purposes
5. Create automated reporting pipelines for clients and internal use

This workflow demonstrates **11 different MCP tools** focused on data operations and reporting.

---

## The Complete Conversation

### Part 1: Multi-Source Timesheet Import

**David**: "Hi Claude! I have a complex data situation. I'm working with three different time tracking systems that my team uses for different clients. I need to import timesheet data from all three systems and create invoices. Each system exports data in a different CSV format."

**Claude**: I can definitely help you handle multiple timesheet formats! Working with different time tracking systems is common in agencies. Let's start by examining each CSV format to understand what we're working with. What are the three systems and do you have sample files ready?

**David**: "Yes, I have three CSV files ready:
1. `/data/timesheets/toggl-export-client-alpha.csv` - from Toggl for Client Alpha
2. `/data/timesheets/harvest-export-client-beta.csv` - from Harvest for Client Beta
3. `/data/timesheets/clockify-export-client-gamma.csv` - from Clockify for Client Gamma

Each has different column names and date formats. Can you help me preview each one first?"

**Claude**: Absolutely! Let's preview each CSV file to understand their different structures and then create appropriate import strategies for each. I'll start with the Toggl export:

*Using import_preview tool:*
```json
{
  "tool": "import_preview",
  "parameters": {
    "file_path": "/data/timesheets/toggl-export-client-alpha.csv",
    "preview_rows": 5,
    "show_column_mapping": true,
    "analyze_data_types": true
  }
}
```

**Toggl Format Analysis:**
- **Columns**: Date, Start time, End time, Duration, Project, Description, User
- **Date Format**: YYYY-MM-DD
- **Duration Format**: Decimal hours (e.g., 2.5)
- **Rate Information**: Not included (will need to apply standard rates)

Now let's look at the Harvest export:

*Using import_preview tool:*
```json
{
  "tool": "import_preview",
  "parameters": {
    "file_path": "/data/timesheets/harvest-export-client-beta.csv",
    "preview_rows": 5,
    "show_column_mapping": true,
    "analyze_data_types": true
  }
}
```

**Harvest Format Analysis:**
- **Columns**: Date, Client, Project, Task, Hours, Notes, Billable Rate
- **Date Format**: MM/DD/YYYY
- **Duration Format**: Decimal hours with rate included
- **Rate Information**: Included per entry

Finally, the Clockify export:

*Using import_preview tool:*
```json
{
  "tool": "import_preview",
  "parameters": {
    "file_path": "/data/timesheets/clockify-export-client-gamma.csv",
    "preview_rows": 5,
    "show_column_mapping": true,
    "analyze_data_types": true
  }
}
```

**Clockify Format Analysis:**
- **Columns**: Start Date, Start Time, End Date, End Time, Duration (h), Project, Description, User
- **Date Format**: DD/MM/YYYY
- **Duration Format**: Hours:Minutes (e.g., 2:30)
- **Rate Information**: Not included

Perfect! I can see each system has a completely different format. Let me validate each file to ensure data integrity before we start importing:

*Using import_validate tool:*
```json
{
  "tool": "import_validate",
  "parameters": {
    "file_path": "/data/timesheets/toggl-export-client-alpha.csv",
    "check_data_integrity": true,
    "validate_date_formats": true,
    "check_for_duplicates": true
  }
}
```

*Using import_validate tool:*
```json
{
  "tool": "import_validate",
  "parameters": {
    "file_path": "/data/timesheets/harvest-export-client-beta.csv",
    "check_data_integrity": true,
    "validate_date_formats": true,
    "check_for_duplicates": true
  }
}
```

*Using import_validate tool:*
```json
{
  "tool": "import_validate",
  "parameters": {
    "file_path": "/data/timesheets/clockify-export-client-gamma.csv",
    "check_data_integrity": true,
    "validate_date_formats": true,
    "check_for_duplicates": true
  }
}
```

**Validation Results:**
- **Toggl**: 45 entries, all valid, no duplicates
- **Harvest**: 32 entries, all valid, no duplicates
- **Clockify**: 38 entries, 2 entries with missing descriptions (will need cleanup)

Great! The data looks good overall. Now let's import each file with appropriate field mapping. What are the billing rates for Client Alpha (Toggl) and Client Gamma (Clockify) since those files don't include rates?

**David**: "Client Alpha is billed at $175/hour and Client Gamma at $150/hour. The Harvest data already has the correct rates. Also, can you create separate invoices for each client?"

**Claude**: Perfect! Let me import each timesheet with the appropriate field mapping and create individual invoices. Starting with the Toggl data for Client Alpha:

*Using import_csv tool:*
```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/data/timesheets/toggl-export-client-alpha.csv",
    "client_name": "Client Alpha",
    "description": "Digital Marketing Services - August 2025",
    "import_mode": "new_invoice",
    "due_days": 30,
    "default_rate": 175.0,
    "field_mapping": {
      "date": "Date",
      "hours": "Duration",
      "description": "Description",
      "project": "Project"
    },
    "date_format": "YYYY-MM-DD"
  }
}
```

Excellent! Invoice INV-005 created for Client Alpha: $7,875.00 (45 hours @ $175/hour)

Now the Harvest data for Client Beta (this one includes rates):

*Using import_csv tool:*
```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/data/timesheets/harvest-export-client-beta.csv",
    "client_name": "Client Beta",
    "description": "Marketing Campaign Management - August 2025",
    "import_mode": "new_invoice",
    "due_days": 30,
    "field_mapping": {
      "date": "Date",
      "hours": "Hours",
      "rate": "Billable Rate",
      "description": "Notes",
      "project": "Project"
    },
    "date_format": "MM/DD/YYYY"
  }
}
```

Great! Invoice INV-006 created for Client Beta: $6,400.00 (32 hours at varying rates from $175-$225/hour)

Finally, the Clockify data for Client Gamma:

*Using import_csv tool:*
```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/data/timesheets/clockify-export-client-gamma.csv",
    "client_name": "Client Gamma",
    "description": "SEO and Content Marketing - August 2025",
    "import_mode": "new_invoice",
    "due_days": 30,
    "default_rate": 150.0,
    "field_mapping": {
      "date": "Start Date",
      "hours": "Duration (h)",
      "description": "Description",
      "project": "Project"
    },
    "date_format": "DD/MM/YYYY",
    "duration_format": "hours:minutes"
  }
}
```

Perfect! Invoice INV-007 created for Client Gamma: $5,700.00 (38 hours @ $150/hour)

**Summary of Imports:**
- **Client Alpha** (Toggl): INV-005 - $7,875.00
- **Client Beta** (Harvest): INV-006 - $6,400.00
- **Client Gamma** (Clockify): INV-007 - $5,700.00
- **Total Billing**: $19,975.00

All three timesheet imports completed successfully with appropriate field mapping for each system!

### Part 2: Data Export for Different Stakeholders

**David**: "This is fantastic! Now I need to export data in different formats for various purposes. I need:
1. A CSV export for my accountant with all invoice data
2. A JSON export for our project management system integration
3. A detailed PDF report for the agency's monthly review"

**Claude**: Perfect! Different stakeholders need different formats. Let me create each export with the appropriate level of detail and formatting. Starting with the CSV for your accountant:

*Using export_data tool:*
```json
{
  "tool": "export_data",
  "parameters": {
    "data_type": "invoices",
    "from_date": "2025-08-01",
    "to_date": "2025-08-31",
    "output_format": "csv",
    "include_financial_details": true,
    "include_tax_information": true,
    "export_path": "./exports/accounting/August-2025-Invoices.csv"
  }
}
```

Great! Accounting CSV exported to `./exports/accounting/August-2025-Invoices.csv` with all the financial details your accountant needs for bookkeeping.

Now the JSON export for your project management system:

*Using export_data tool:*
```json
{
  "tool": "export_data",
  "parameters": {
    "data_type": "time_entries",
    "from_date": "2025-08-01",
    "to_date": "2025-08-31",
    "output_format": "json",
    "include_project_breakdown": true,
    "include_user_details": true,
    "group_by": "client",
    "export_path": "./exports/project-management/August-2025-TimeEntries.json"
  }
}
```

Excellent! JSON export created at `./exports/project-management/August-2025-TimeEntries.json` with structured data perfect for API integration.

Finally, the comprehensive PDF report for your monthly agency review:

*Using generate_summary tool:*
```json
{
  "tool": "generate_summary",
  "parameters": {
    "from_date": "2025-08-01",
    "to_date": "2025-08-31",
    "include_client_breakdown": true,
    "include_project_analysis": true,
    "include_team_productivity": true,
    "include_financial_metrics": true,
    "output_format": "comprehensive",
    "export_path": "./reports/agency/August-2025-Agency-Review.pdf"
  }
}
```

Perfect! Your comprehensive agency review is ready at `./reports/agency/August-2025-Agency-Review.pdf`. The report includes:

**Executive Summary:**
- **Total Revenue**: $19,975.00
- **Billable Hours**: 115 hours across 3 clients
- **Average Rate**: $173.70/hour
- **Team Utilization**: 92% (excellent)

**Client Performance:**
- **Client Alpha**: $7,875.00 (39.4% of revenue) - Premium rate client
- **Client Beta**: $6,400.00 (32.1% of revenue) - Variable rate projects
- **Client Gamma**: $5,700.00 (28.5% of revenue) - Steady SEO work

**Project Insights:**
- **Digital Marketing**: Highest margins
- **Campaign Management**: Good efficiency
- **SEO/Content**: Steady recurring revenue

### Part 3: Advanced Data Analysis and Reporting

**David**: "This is incredibly helpful! I'm curious about our productivity patterns. Can you help me analyze which time tracking system gives us the most accurate data and which clients have the best profit margins?"

**Claude**: Great question! Let me analyze your data to provide insights about time tracking accuracy and client profitability. First, let me generate a detailed analysis of the different data sources:

*Using export_data tool:*
```json
{
  "tool": "export_data",
  "parameters": {
    "data_type": "time_entries",
    "from_date": "2025-08-01",
    "to_date": "2025-08-31",
    "output_format": "analytical",
    "include_data_source_analysis": true,
    "include_accuracy_metrics": true,
    "group_by": "data_source",
    "export_path": "./analysis/Time-Tracking-System-Comparison.json"
  }
}
```

**Time Tracking System Analysis:**

**Toggl (Client Alpha):**
- **Data Completeness**: 100% (all required fields present)
- **Description Quality**: High (detailed project descriptions)
- **Time Precision**: Good (decimal hours, easy to process)
- **Consistency**: Excellent (standardized format)

**Harvest (Client Beta):**
- **Data Completeness**: 100% (includes rates per entry)
- **Description Quality**: Medium (some generic descriptions)
- **Time Precision**: Excellent (includes rates and detailed breakdowns)
- **Consistency**: Good (occasional rate variations to track)

**Clockify (Client Gamma):**
- **Data Completeness**: 95% (2 missing descriptions)
- **Description Quality**: Medium (needs cleanup)
- **Time Precision**: Good (hours:minutes format requires conversion)
- **Consistency**: Good (standardized project categorization)

**Recommendation**: Harvest provides the most comprehensive data due to included billing rates, but Toggl has the most consistent format for bulk processing.

Now let me analyze client profitability:

*Using generate_summary tool:*
```json
{
  "tool": "generate_summary",
  "parameters": {
    "from_date": "2025-08-01",
    "to_date": "2025-08-31",
    "focus": "profitability_analysis",
    "include_margin_breakdown": true,
    "include_efficiency_metrics": true,
    "cost_analysis": true,
    "export_path": "./analysis/Client-Profitability-Analysis.pdf"
  }
}
```

**Client Profitability Analysis:**

**Client Alpha (Toggl/Premium Rate)**
- **Revenue**: $7,875.00
- **Estimated Costs**: $2,250.00 (team hourly costs)
- **Gross Margin**: $5,625.00 (71.4%)
- **Efficiency**: High (clear project definitions, minimal revisions)

**Client Beta (Harvest/Variable Rates)**
- **Revenue**: $6,400.00
- **Estimated Costs**: $1,920.00
- **Gross Margin**: $4,480.00 (70.0%)
- **Efficiency**: Medium (some scope creep, rate variations)

**Client Gamma (Clockify/Standard Rate)**
- **Revenue**: $5,700.00
- **Estimated Costs**: $1,900.00
- **Gross Margin**: $3,800.00 (66.7%)
- **Efficiency**: Medium (steady work, some communication gaps)

**Key Insights:**
- **Client Alpha** has the highest margin and efficiency
- **Harvest integration** provides best visibility into actual profitability
- **Premium rate clients** (Alpha) are most profitable per hour
- **Data quality** directly correlates with project efficiency

### Part 4: Automated Reporting Pipeline

**David**: "This analysis is fantastic! Can you help me set up a process where I can easily repeat this each month? I want to standardize my reporting pipeline."

**Claude**: Absolutely! Let's create a standardized monthly reporting process that you can easily repeat. I'll show you how to structure this as a repeatable workflow.

*Using config_show tool:*
```json
{
  "tool": "config_show",
  "parameters": {
    "section": "export_settings",
    "output_format": "yaml",
    "show_templates": true
  }
}
```

I can see your export configuration supports templates. Let me create a standardized monthly report template:

*Using export_data tool:*
```json
{
  "tool": "export_data",
  "parameters": {
    "data_type": "monthly_package",
    "create_template": true,
    "template_name": "agency_monthly_standard",
    "include_formats": ["csv", "json", "pdf"],
    "include_analysis": ["profitability", "productivity", "data_quality"],
    "stakeholder_versions": {
      "accounting": "csv_financial",
      "project_management": "json_detailed",
      "executive": "pdf_comprehensive"
    },
    "automation_schedule": "monthly",
    "export_path": "./templates/monthly-reporting-template.json"
  }
}
```

Perfect! I've created a standardized monthly reporting template. Here's your repeatable process:

**Monthly Reporting Workflow:**

1. **Data Collection Phase:**
   - Import timesheets from all three systems using validated field mappings
   - Validate data integrity before processing
   - Create client-specific invoices

2. **Export Generation Phase:**
   - Accounting CSV: Financial data for bookkeeping
   - Project Management JSON: Structured data for system integration
   - Executive PDF: Comprehensive business analysis

3. **Analysis Phase:**
   - Client profitability comparison
   - Time tracking system effectiveness review
   - Team productivity and efficiency metrics

4. **Distribution Phase:**
   - Accounting team gets CSV exports
   - Project managers get JSON data feeds
   - Executive team gets comprehensive PDF reports

**Next Month Usage:**
Simply tell me "Run the standard monthly report for [Month Year]" and I'll execute all these steps automatically using the template we just created!

**David**: "This is perfect! One last request - can you export a consolidated dataset that includes everything from this month? I want to use it for some custom analysis in Excel."

**Claude**: Absolutely! Let me create a comprehensive consolidated export that includes all your data in Excel-friendly formats.

*Using export_data tool:*
```json
{
  "tool": "export_data",
  "parameters": {
    "data_type": "consolidated",
    "from_date": "2025-08-01",
    "to_date": "2025-08-31",
    "output_format": "xlsx",
    "include_all_tables": true,
    "separate_worksheets": true,
    "worksheets": [
      "time_entries_raw",
      "invoices_summary",
      "client_analysis",
      "profitability_breakdown",
      "data_source_comparison"
    ],
    "include_formulas": true,
    "include_charts": true,
    "export_path": "./exports/consolidated/August-2025-Complete-Dataset.xlsx"
  }
}
```

Excellent! Your consolidated Excel workbook is ready at `./exports/consolidated/August-2025-Complete-Dataset.xlsx`.

**The workbook includes 5 worksheets:**

1. **Time Entries Raw**: All imported timesheet data with source system identification
2. **Invoices Summary**: Invoice details, totals, and status for all three clients
3. **Client Analysis**: Profitability, efficiency, and relationship metrics per client
4. **Profitability Breakdown**: Detailed margin analysis and cost breakdowns
5. **Data Source Comparison**: Quality metrics and recommendations for each time tracking system

**Excel Features Included:**
- **Pivot Tables**: Ready-made for further analysis
- **Charts**: Visual representations of key metrics
- **Formulas**: Automated calculations for margins and efficiency ratios
- **Filters**: Easy data exploration and drill-down capabilities

This gives you complete flexibility to perform any additional analysis in Excel while having all the processed and validated data at your fingertips!

---

## Tools Demonstrated

This workflow showcased **11 different MCP tools**:

1. **import_preview** - Analyzing CSV structure from different time tracking systems
2. **import_validate** - Ensuring data integrity before processing
3. **import_csv** - Importing timesheets with custom field mapping for each system
4. **export_data** - Creating exports in multiple formats (CSV, JSON, Excel)
5. **generate_summary** - Producing comprehensive analytical reports
6. **config_show** - Understanding system configuration for templates
7. **invoice_create** (via import) - Creating client-specific invoices from timesheet data
8. **invoice_list** (implied) - Managing multiple invoice statuses
9. **client_show** (implied) - Accessing client information for billing
10. **generate_html** (referenced) - Creating client-ready invoice presentations
11. **invoice_update** (referenced) - Managing invoice lifecycle

## Key Data Management Benefits

- **Multi-Source Integration**: Handle different CSV formats and time tracking systems seamlessly
- **Data Validation**: Catch errors and inconsistencies before they become problems
- **Flexible Export Options**: Generate appropriate formats for different stakeholders
- **Automated Analysis**: Transform raw timesheet data into business intelligence
- **Standardized Processes**: Create repeatable workflows for consistent results
- **Stakeholder Communication**: Provide each team with data in their preferred format

## Advanced Analytics Capabilities

- **System Comparison**: Evaluate different time tracking tools for accuracy and completeness
- **Profitability Analysis**: Understand which clients and projects drive the most value
- **Efficiency Metrics**: Identify patterns in team productivity and client work
- **Data Quality Assessment**: Monitor and improve data collection processes
- **Trend Analysis**: Track performance changes over time
- **Cost Allocation**: Understand true project costs and margins

## Business Process Optimization

This workflow demonstrates how agencies can transform scattered timesheet data into comprehensive business intelligence:

- **Operational Efficiency**: Reduce manual data entry and processing time
- **Financial Visibility**: Clear understanding of profitability by client and project
- **Quality Control**: Systematic validation prevents billing errors
- **Stakeholder Alignment**: Each team gets appropriate data in their preferred format
- **Decision Support**: Data-driven insights for pricing, client management, and tool selection
- **Scalability**: Standardized processes that work as the business grows

The go-invoice MCP integration turns complex data operations into simple conversations, enabling agencies to focus on analysis and strategy rather than data wrangling.
