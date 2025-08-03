# MCP Tools Technical Reference

**go-invoice MCP Server v1.0.0 - Complete API Reference**

---

## Overview

This document provides comprehensive technical reference for all 21 MCP tools in the go-invoice system. These tools enable Claude and other MCP clients to perform complete invoice management workflows through natural language interactions.

### Architecture Summary

- **Total Tools**: 21 tools across 5 categories
- **Protocol**: MCP (Model Context Protocol) 1.0
- **Schema Standard**: JSON Schema Draft 7
- **CLI Integration**: Direct go-invoice command execution
- **Context Support**: Full cancellation and timeout handling

### Categories and Tool Count

1. **Invoice Management** (7 tools): Complete invoice lifecycle management
2. **Client Management** (5 tools): Client relationship and contact management  
3. **Data Import** (3 tools): CSV import and validation workflows
4. **Data Export** (3 tools): Document generation and data export
5. **Configuration** (3 tools): System configuration and validation

---

## Tool Categories

### Category: Invoice Management

**Purpose**: Complete invoice lifecycle management from creation to payment tracking  
**Tools**: 7 tools for comprehensive invoice operations  
**Common Use Cases**: Creating invoices, tracking payments, managing work items, status updates

#### Tool: `invoice_create`

**Description**: Create new invoices for clients with optional work items and automatic client creation

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "client_name": {
      "type": "string",
      "description": "Client name or partial name for invoice",
      "minLength": 1,
      "maxLength": 200
    },
    "client_id": {
      "type": "string", 
      "description": "Exact client ID for invoice",
      "pattern": "^[A-Za-z0-9_-]+$"
    },
    "client_email": {
      "type": "string",
      "format": "email",
      "description": "Client email address to identify the client"
    },
    "invoice_date": {
      "type": "string",
      "format": "date",
      "description": "Invoice date in YYYY-MM-DD format (defaults to today)"
    },
    "due_date": {
      "type": "string",
      "format": "date", 
      "description": "Payment due date in YYYY-MM-DD format"
    },
    "description": {
      "type": "string",
      "maxLength": 500,
      "description": "Optional description for the invoice"
    },
    "create_client_if_missing": {
      "type": "boolean",
      "default": false,
      "description": "Whether to create a new client if not found"
    },
    "new_client_email": {
      "type": "string",
      "format": "email",
      "description": "Email for new client (required when creating client)"
    },
    "new_client_address": {
      "type": "string",
      "maxLength": 500,
      "description": "Address for new client"
    },
    "new_client_phone": {
      "type": "string",
      "pattern": "^[\\d\\s\\+\\-\\(\\)\\.\\/ext]+$",
      "description": "Phone number for new client"
    },
    "work_items": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "date": {"type": "string", "format": "date"},
          "hours": {"type": "number", "minimum": 0.01, "maximum": 24.0},
          "rate": {"type": "number", "minimum": 0.01},
          "description": {"type": "string", "minLength": 1, "maxLength": 500}
        },
        "required": ["date", "hours", "rate", "description"]
      }
    }
  },
  "anyOf": [
    {"required": ["client_name"]},
    {"required": ["client_id"]}, 
    {"required": ["client_email"]}
  ]
}
```

**Output Format**:
```json
{
  "success": true,
  "invoice_id": "INV-2025-001",
  "invoice_number": "INV-001",
  "client_id": "client_abc123",
  "total_amount": 1250.00,
  "status": "draft",
  "created_at": "2025-08-03T10:00:00Z"
}
```

**Usage Examples**:

*Simple Invoice Creation*:
```json
{
  "client_name": "Acme Corp",
  "description": "January 2025 consulting services"
}
```

*Invoice with Work Items*:
```json
{
  "client_name": "Tech Solutions Inc",
  "invoice_date": "2025-08-01",
  "due_date": "2025-08-31", 
  "description": "Website development - Phase 1",
  "work_items": [
    {
      "date": "2025-08-01",
      "hours": 8.0,
      "rate": 125.0,
      "description": "Frontend development and UI design"
    }
  ]
}
```

*Create Invoice with New Client*:
```json
{
  "client_name": "New Client LLC",
  "create_client_if_missing": true,
  "new_client_email": "contact@newclient.com",
  "new_client_address": "456 Business Ave, Suite 200",
  "description": "Initial consulting engagement"
}
```

**Error Conditions**:
- `CLIENT_NOT_FOUND` (404): Client not found and create_client_if_missing is false
- `VALIDATION_ERROR` (400): Invalid input parameters (date format, required fields)
- `DUPLICATE_CLIENT` (409): Client email already exists when creating new client
- `TIMEOUT_ERROR` (408): Operation exceeded 30 second timeout

**Security Considerations**:
- Client identification prevents unauthorized invoice creation
- New client creation requires email validation
- Rate limiting applies to prevent spam invoice creation
- Audit logging tracks all invoice creation activities

**Performance Notes**:
- Average execution time: 2-5 seconds
- Client lookup optimized with indexed searches
- Work item validation processed in parallel
- Memory usage scales with number of work items

**Related Tools**: `client_create`, `invoice_add_item`, `invoice_list`

---

#### Tool: `invoice_list`

**Description**: List and filter invoices with comprehensive search criteria and multiple output formats

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["draft", "sent", "paid", "overdue", "voided"],
      "description": "Filter invoices by status"
    },
    "client_name": {
      "type": "string",
      "description": "Filter by client name (partial matches supported)"
    },
    "client_id": {
      "type": "string",
      "description": "Filter by exact client ID"
    },
    "from_date": {
      "type": "string",
      "format": "date",
      "description": "Show invoices from this date onwards"
    },
    "to_date": {
      "type": "string", 
      "format": "date",
      "description": "Show invoices up to this date"
    },
    "sort_by": {
      "type": "string",
      "enum": ["date", "amount", "status", "client", "due_date"],
      "default": "date"
    },
    "sort_order": {
      "type": "string",
      "enum": ["asc", "desc"],
      "default": "desc"
    },
    "limit": {
      "type": "number",
      "minimum": 1,
      "maximum": 1000,
      "default": 50
    },
    "output_format": {
      "type": "string",
      "enum": ["table", "json", "csv"],
      "default": "table"
    },
    "include_summary": {
      "type": "boolean",
      "default": false,
      "description": "Include summary statistics"
    }
  }
}
```

**Output Format**:

*Table Format*:
```
┌─────────────┬─────────────────┬────────────┬────────────┬─────────────┬───────────┐
│ Number      │ Client          │ Date       │ Due Date   │ Status      │ Amount    │
├─────────────┼─────────────────┼────────────┼────────────┼─────────────┼───────────┤
│ INV-001     │ Acme Corp       │ 2025-08-01 │ 2025-08-31 │ sent        │ $1,250.00 │
│ INV-002     │ Tech Solutions  │ 2025-08-02 │ 2025-09-01 │ draft       │ $2,100.00 │
└─────────────┴─────────────────┴────────────┴────────────┴─────────────┴───────────┘
```

*JSON Format*:
```json
{
  "invoices": [
    {
      "id": "INV-001",
      "number": "INV-001", 
      "client_name": "Acme Corp",
      "client_id": "client_123",
      "date": "2025-08-01",
      "due_date": "2025-08-31",
      "status": "sent",
      "amount": 1250.00,
      "currency": "USD"
    }
  ],
  "summary": {
    "total_count": 15,
    "total_amount": 25780.00,
    "by_status": {
      "draft": 3,
      "sent": 7,
      "paid": 4, 
      "overdue": 1
    }
  }
}
```

**Usage Examples**:

*List All Unpaid Invoices*:
```json
{
  "status": "sent",
  "include_summary": true,
  "sort_by": "due_date",
  "sort_order": "asc"
}
```

*Monthly Client Report*:
```json
{
  "client_name": "Acme Corp",
  "from_date": "2025-08-01",
  "to_date": "2025-08-31",
  "output_format": "json",
  "include_summary": true
}
```

*Export for Accounting*:
```json
{
  "from_date": "2025-01-01",
  "to_date": "2025-12-31", 
  "output_format": "csv",
  "status": "paid"
}
```

**Error Conditions**:
- `INVALID_DATE_RANGE` (400): from_date is after to_date
- `INVALID_STATUS` (400): Unknown status value provided
- `LIMIT_EXCEEDED` (400): Limit parameter exceeds maximum of 1000
- `TIMEOUT_ERROR` (408): Query exceeded 20 second timeout

**Security Considerations**:
- Client filtering prevents unauthorized invoice access
- Rate limiting on large queries to prevent system overload
- Audit logging for all invoice access patterns
- Data export controls for sensitive financial information

**Performance Notes**:
- Database queries optimized with proper indexing
- Large result sets paginated automatically
- CSV export optimized for memory efficiency
- Summary calculations cached for performance

**Related Tools**: `invoice_show`, `export_data`, `generate_summary`

---

#### Tool: `invoice_show`

**Description**: Display comprehensive details for a specific invoice including client information, work items, and financial breakdown

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "invoice_id": {
      "type": "string",
      "description": "Invoice ID to display details for",
      "minLength": 1
    },
    "invoice_number": {
      "type": "string", 
      "description": "Invoice number to display details for",
      "minLength": 1
    },
    "output_format": {
      "type": "string",
      "enum": ["text", "json", "yaml"],
      "default": "text"
    },
    "show_work_items": {
      "type": "boolean",
      "default": true,
      "description": "Include detailed work items"
    },
    "show_client_details": {
      "type": "boolean", 
      "default": true,
      "description": "Include full client information"
    }
  },
  "anyOf": [
    {"required": ["invoice_id"]},
    {"required": ["invoice_number"]}
  ]
}
```

**Output Format**:

*Text Format*:
```
Invoice Details: INV-001
========================

Invoice Information:
  Number: INV-001
  Date: August 1, 2025
  Due Date: August 31, 2025
  Status: Sent
  Description: January 2025 consulting services

Client Information:
  Name: Acme Corporation
  Email: contact@acme.com
  Phone: +1-555-123-4567
  Address: 123 Business St, City, State 12345

Work Items:
  1. Aug 1, 2025 - Frontend development (8.0 hrs @ $125.00/hr) = $1,000.00
  2. Aug 2, 2025 - Bug fixes and testing (4.5 hrs @ $125.00/hr) = $562.50

Financial Summary:
  Subtotal: $1,562.50
  Tax (8.5%): $132.81
  Total: $1,695.31
```

*JSON Format*:
```json
{
  "invoice": {
    "id": "INV-001",
    "number": "INV-001",
    "date": "2025-08-01",
    "due_date": "2025-08-31", 
    "status": "sent",
    "description": "January 2025 consulting services",
    "client": {
      "id": "client_123",
      "name": "Acme Corporation",
      "email": "contact@acme.com",
      "phone": "+1-555-123-4567",
      "address": "123 Business St, City, State 12345"
    },
    "work_items": [
      {
        "id": "wi_001",
        "date": "2025-08-01",
        "hours": 8.0,
        "rate": 125.00,
        "description": "Frontend development",
        "amount": 1000.00
      }
    ],
    "totals": {
      "subtotal": 1562.50,
      "tax_rate": 0.085,
      "tax_amount": 132.81,
      "total": 1695.31
    },
    "metadata": {
      "created_at": "2025-08-01T09:00:00Z",
      "updated_at": "2025-08-01T14:30:00Z",
      "version": 2
    }
  }
}
```

**Usage Examples**:

*Complete Invoice Review*:
```json
{
  "invoice_number": "INV-001",
  "show_work_items": true,
  "show_client_details": true
}
```

*Quick Summary Check*:
```json
{
  "invoice_number": "INV-025",
  "show_work_items": false,
  "output_format": "text"
}
```

*Data Export Integration*:
```json
{
  "invoice_id": "invoice_abc123",
  "output_format": "json"
}
```

**Error Conditions**:
- `INVOICE_NOT_FOUND` (404): Invoice with specified ID/number not found
- `ACCESS_DENIED` (403): Insufficient permissions to view invoice
- `INVALID_FORMAT` (400): Unsupported output format requested
- `TIMEOUT_ERROR` (408): Query exceeded 15 second timeout

**Security Considerations**:
- Invoice access controls prevent unauthorized viewing
- Client information filtered based on permissions
- Audit logging for all invoice access attempts
- Sensitive data masking in certain output formats

**Performance Notes**:
- Single query optimization for complete invoice data
- Work items loaded efficiently with proper joins
- Output formatting optimized for each format type
- Caching implemented for frequently accessed invoices

**Related Tools**: `invoice_list`, `invoice_update`, `client_show`

---

#### Tool: `invoice_update`

**Description**: Update invoice details such as status, due date, or description with business rule validation and status transition controls

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "invoice_id": {
      "type": "string",
      "description": "Invoice ID to update",
      "minLength": 1
    },
    "invoice_number": {
      "type": "string",
      "description": "Invoice number to update", 
      "minLength": 1
    },
    "status": {
      "type": "string",
      "enum": ["draft", "sent", "paid", "overdue", "voided"],
      "description": "Update invoice status"
    },
    "due_date": {
      "type": "string",
      "format": "date",
      "description": "Update payment due date"
    },
    "description": {
      "type": "string",
      "maxLength": 500,
      "description": "Update invoice description"
    }
  },
  "allOf": [
    {
      "anyOf": [
        {"required": ["invoice_id"]},
        {"required": ["invoice_number"]}
      ]
    },
    {
      "anyOf": [
        {"required": ["status"]},
        {"required": ["due_date"]},
        {"required": ["description"]}
      ]
    }
  ]
}
```

**Output Format**:
```json
{
  "success": true,
  "invoice_id": "INV-001",
  "updates_applied": {
    "status": {
      "old_value": "draft",
      "new_value": "sent",
      "updated_at": "2025-08-03T14:30:00Z"
    }
  },
  "validation_warnings": [],
  "business_rules_applied": [
    "Status transition from 'draft' to 'sent' recorded",
    "Audit trail entry created"
  ]
}
```

**Usage Examples**:

*Mark Invoice as Sent*:
```json
{
  "invoice_number": "INV-001",
  "status": "sent"
}
```

*Extend Due Date*:
```json
{
  "invoice_id": "invoice_abc123",
  "due_date": "2025-09-30"
}
```

*Multiple Field Update*:
```json
{
  "invoice_number": "INV-042",
  "status": "sent",
  "due_date": "2025-09-15",
  "description": "August 2025 Consulting - Terms Updated"
}
```

**Error Conditions**:
- `INVOICE_NOT_FOUND` (404): Invoice with specified ID/number not found
- `INVALID_STATUS_TRANSITION` (409): Business rules prevent status change
- `PAID_INVOICE_IMMUTABLE` (409): Cannot modify paid invoices
- `VALIDATION_ERROR` (400): Invalid field values or formats
- `TIMEOUT_ERROR` (408): Update exceeded 20 second timeout

**Security Considerations**:
- Status transition validation prevents financial fraud
- Audit trail maintains complete change history
- Permission checks for invoice modification rights
- Business rule enforcement for compliance

**Performance Notes**:
- Atomic updates ensure data consistency
- Business rule validation optimized for speed
- Change tracking minimizes database overhead
- Concurrent update protection with optimistic locking

**Related Tools**: `invoice_show`, `invoice_list`, `invoice_delete`

---

#### Tool: `invoice_delete`

**Description**: Delete invoices with safety confirmations and business rule validation. Supports both soft delete (default) and permanent removal

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "invoice_id": {
      "type": "string",
      "description": "Invoice ID to delete",
      "minLength": 1
    },
    "invoice_number": {
      "type": "string",
      "description": "Invoice number to delete",
      "minLength": 1
    },
    "hard_delete": {
      "type": "boolean",
      "default": false,
      "description": "Permanently delete invoice (cannot be undone)"
    },
    "force": {
      "type": "boolean", 
      "default": false,
      "description": "Skip confirmation prompt (use with caution)"
    }
  },
  "anyOf": [
    {"required": ["invoice_id"]},
    {"required": ["invoice_number"]}
  ]
}
```

**Output Format**:
```json
{
  "success": true,
  "invoice_id": "INV-001",
  "deletion_type": "soft",
  "deleted_at": "2025-08-03T15:45:00Z",
  "recoverable": true,
  "backup_reference": "backup_inv_001_20250803",
  "business_rules_checked": [
    "Invoice status validated for deletion",
    "Payment history checked", 
    "Audit trail preserved"
  ]
}
```

**Usage Examples**:

*Safe Soft Delete*:
```json
{
  "invoice_number": "INV-001"
}
```

*Permanent Delete with Confirmation*:
```json
{
  "invoice_id": "invoice_test123",
  "hard_delete": true
}
```

*Force Delete (Dangerous)*:
```json
{
  "invoice_number": "DRAFT-999",
  "hard_delete": true,
  "force": true
}
```

**Error Conditions**:
- `INVOICE_NOT_FOUND` (404): Invoice with specified ID/number not found
- `PAID_INVOICE_PROTECTED` (409): Cannot delete paid invoices
- `SENT_INVOICE_RESTRICTED` (409): Sent invoices require special handling
- `USER_CANCELLED` (400): User declined confirmation prompt
- `BACKUP_FAILED` (500): Backup creation failed for hard delete

**Security Considerations**:
- Business rules prevent deletion of financial records
- Confirmation prompts protect against accidental deletion
- Audit trail preserved even for hard deletes
- Backup creation for permanent deletions
- Permission verification for delete operations

**Performance Notes**:
- Soft deletes are instantaneous operations
- Hard deletes include backup creation overhead
- Cascading deletes handled efficiently
- Index updates optimized for deletion operations

**Related Tools**: `invoice_show`, `invoice_list`, `invoice_update`

---

#### Tool: `invoice_add_item`

**Description**: Add work items to existing invoices with automatic total calculation. Supports single or batch work item entry for draft invoices

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "invoice_id": {
      "type": "string",
      "description": "Invoice ID to add work items to",
      "minLength": 1
    },
    "invoice_number": {
      "type": "string",
      "description": "Invoice number to add work items to",
      "minLength": 1
    },
    "work_items": {
      "type": "array",
      "minItems": 1,
      "description": "Work items to add to the invoice",
      "items": {
        "type": "object",
        "properties": {
          "date": {
            "type": "string",
            "format": "date",
            "description": "Date when work was performed"
          },
          "hours": {
            "type": "number",
            "minimum": 0.01,
            "maximum": 24.0,
            "description": "Number of hours worked"
          },
          "rate": {
            "type": "number",
            "minimum": 0.01,
            "description": "Hourly rate for this work item"
          },
          "description": {
            "type": "string",
            "minLength": 1,
            "maxLength": 500,
            "description": "Description of work performed"
          }
        },
        "required": ["date", "hours", "rate", "description"]
      }
    }
  },
  "anyOf": [
    {"required": ["invoice_id", "work_items"]},
    {"required": ["invoice_number", "work_items"]}
  ]
}
```

**Output Format**:
```json
{
  "success": true,
  "invoice_id": "INV-001",
  "work_items_added": 3,
  "work_item_ids": ["wi_004", "wi_005", "wi_006"],
  "totals_updated": {
    "previous_subtotal": 1000.00,
    "new_subtotal": 2400.00,
    "previous_total": 1085.00,
    "new_total": 2604.00,
    "items_total": 1400.00
  },
  "validation_results": {
    "all_valid": true,
    "duplicate_warnings": 0,
    "rate_variations": []
  }
}
```

**Usage Examples**:

*Add Single Work Item*:
```json
{
  "invoice_number": "INV-001",
  "work_items": [
    {
      "date": "2025-08-03",
      "hours": 6.0,
      "rate": 150.0,
      "description": "Bug fixes and performance optimization"
    }
  ]
}
```

*Batch Add Week's Work*:
```json
{
  "invoice_id": "invoice_abc123",
  "work_items": [
    {
      "date": "2025-08-01",
      "hours": 8.0,
      "rate": 125.0,
      "description": "Frontend component development"
    },
    {
      "date": "2025-08-02", 
      "hours": 7.5,
      "rate": 125.0,
      "description": "API integration and testing"
    },
    {
      "date": "2025-08-03",
      "hours": 4.0,
      "rate": 125.0,
      "description": "Code review and documentation"
    }
  ]
}
```

*Premium Rate Work*:
```json
{
  "invoice_number": "INV-CONSULTING-025",
  "work_items": [
    {
      "date": "2025-08-01",
      "hours": 2.5,
      "rate": 300.0,
      "description": "Executive strategy consulting session"
    }
  ]
}
```

**Error Conditions**:
- `INVOICE_NOT_FOUND` (404): Invoice with specified ID/number not found
- `INVOICE_NOT_DRAFT` (409): Can only add items to draft invoices
- `INVALID_WORK_ITEM` (400): Work item validation failed
- `DUPLICATE_WORK_ITEM` (409): Similar work item already exists
- `RATE_VALIDATION_FAILED` (400): Rate exceeds configured limits

**Security Considerations**:
- Invoice status validation prevents unauthorized modifications
- Rate limits prevent excessive billing rate entries
- Work item validation ensures data integrity
- Audit trail tracks all work item additions

**Performance Notes**:
- Batch processing optimized for multiple work items
- Total calculations performed efficiently
- Validation processed in parallel for speed
- Database transactions ensure data consistency

**Related Tools**: `invoice_remove_item`, `invoice_create`, `import_csv`

---

#### Tool: `invoice_remove_item`

**Description**: Remove work items from invoices using flexible identification methods. Supports removal by ID, description match, or date with automatic total recalculation

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "invoice_id": {
      "type": "string",
      "description": "Invoice ID to remove work items from",
      "minLength": 1
    },
    "invoice_number": {
      "type": "string",
      "description": "Invoice number to remove work items from",
      "minLength": 1
    },
    "work_item_id": {
      "type": "string",
      "description": "Specific work item ID to remove",
      "minLength": 1
    },
    "work_item_description": {
      "type": "string",
      "description": "Remove work items matching this description",
      "minLength": 1
    },
    "work_item_date": {
      "type": "string",
      "format": "date",
      "description": "Remove work items from this specific date"
    },
    "remove_all_matching": {
      "type": "boolean",
      "default": false,
      "description": "Remove all items matching criteria"
    },
    "confirm": {
      "type": "boolean",
      "default": false,
      "description": "Confirm removal without additional prompts"
    }
  },
  "allOf": [
    {
      "anyOf": [
        {"required": ["invoice_id"]},
        {"required": ["invoice_number"]}
      ]
    },
    {
      "anyOf": [
        {"required": ["work_item_id"]},
        {"required": ["work_item_description"]},
        {"required": ["work_item_date"]}
      ]
    }
  ]
}
```

**Output Format**:
```json
{
  "success": true,
  "invoice_id": "INV-001",
  "work_items_removed": 2,
  "removed_items": [
    {
      "id": "wi_003",
      "date": "2025-08-01",
      "hours": 4.0,
      "amount": 500.00,
      "description": "Bug fixes and testing"
    }
  ],
  "totals_updated": {
    "previous_subtotal": 2400.00,
    "new_subtotal": 1900.00,
    "adjustment_amount": -500.00,
    "new_total": 2061.50
  },
  "confirmation_required": false
}
```

**Usage Examples**:

*Remove Specific Work Item*:
```json
{
  "invoice_number": "INV-001",
  "work_item_id": "work_item_123",
  "confirm": true
}
```

*Remove by Description Match*:
```json
{
  "invoice_id": "invoice_abc123",
  "work_item_description": "Bug fixes",
  "remove_all_matching": false
}
```

*Remove All Work from Date*:
```json
{
  "invoice_number": "INV-025",
  "work_item_date": "2025-08-01",
  "remove_all_matching": true,
  "confirm": true
}
```

**Error Conditions**:
- `INVOICE_NOT_FOUND` (404): Invoice with specified ID/number not found
- `WORK_ITEM_NOT_FOUND` (404): No work items match the criteria
- `INVOICE_NOT_DRAFT` (409): Can only remove items from draft invoices
- `USER_CANCELLED` (400): User declined confirmation prompt
- `MULTIPLE_MATCHES` (409): Multiple items match, confirmation required

**Security Considerations**:
- Invoice status validation prevents unauthorized modifications
- Confirmation prompts protect against accidental removal
- Audit trail maintains record of all removals
- Business rule validation for data integrity

**Performance Notes**:
- Efficient matching algorithms for description/date searches
- Optimized total recalculation processes
- Batch removal operations for multiple items
- Transaction safety for data consistency

**Related Tools**: `invoice_add_item`, `invoice_show`, `invoice_update`

---

### Category: Client Management

**Purpose**: Complete client relationship and contact information management  
**Tools**: 5 tools for comprehensive client operations  
**Common Use Cases**: Client onboarding, contact management, relationship tracking

#### Tool: `client_create`

**Description**: Create new clients with contact information and business details. Validates email uniqueness and contact information completeness

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "Client name or company name",
      "minLength": 1,
      "maxLength": 200
    },
    "email": {
      "type": "string",
      "format": "email",
      "description": "Primary email address for the client"
    },
    "phone": {
      "type": "string",
      "pattern": "^[\\d\\s\\+\\-\\(\\)\\.\\/ext]+$",
      "description": "Phone number for the client"
    },
    "address": {
      "type": "string",
      "maxLength": 500,
      "description": "Full address for the client"
    },
    "tax_id": {
      "type": "string",
      "maxLength": 50,
      "description": "Tax ID or business registration number"
    },
    "payment_terms": {
      "type": "number",
      "minimum": 1,
      "maximum": 365,
      "default": 30,
      "description": "Default payment terms in days"
    },
    "notes": {
      "type": "string",
      "maxLength": 1000,
      "description": "Additional notes about the client"
    }
  },
  "required": ["name", "email"]
}
```

**Output Format**:
```json
{
  "success": true,
  "client_id": "client_abc123",
  "name": "Acme Corporation", 
  "email": "contact@acme.com",
  "created_at": "2025-08-03T10:00:00Z",
  "validation_results": {
    "email_unique": true,
    "contact_complete": true,
    "tax_id_valid": true
  }
}
```

**Usage Examples**:

*Simple Client Creation*:
```json
{
  "name": "Acme Corporation",
  "email": "contact@acme.com"
}
```

*Complete Business Client*:
```json
{
  "name": "Tech Solutions Inc",
  "email": "billing@techsolutions.com",
  "phone": "+1-555-123-4567",
  "address": "456 Business Ave, Suite 200, Tech City, TC 12345",
  "tax_id": "EIN-98-7654321",
  "payment_terms": 45,
  "notes": "Major enterprise client - handle with priority"
}
```

**Error Conditions**:
- `EMAIL_ALREADY_EXISTS` (409): Email address already in use
- `VALIDATION_ERROR` (400): Invalid email format or required fields missing
- `NAME_TOO_LONG` (400): Client name exceeds maximum length
- `INVALID_PHONE_FORMAT` (400): Phone number format validation failed

**Security Considerations**:
- Email uniqueness enforced to prevent duplicate clients
- Contact information validation for data integrity
- Permission checks for client creation rights
- Audit logging for all client creation activities

**Performance Notes**:
- Email uniqueness check optimized with database indexing
- Contact validation processed efficiently
- Client ID generation using secure random algorithms
- Average creation time under 2 seconds

**Related Tools**: `client_update`, `client_list`, `invoice_create`

---

#### Tool: `client_list`

**Description**: List and filter clients with flexible search criteria and contact information display

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "name_filter": {
      "type": "string",
      "description": "Filter by client name (partial matches supported)"
    },
    "email_filter": {
      "type": "string",
      "description": "Filter by email address (partial matches supported)"
    },
    "has_invoices": {
      "type": "boolean",
      "description": "Filter clients who have invoices"
    },
    "payment_overdue": {
      "type": "boolean",
      "description": "Filter clients with overdue payments"
    },
    "sort_by": {
      "type": "string",
      "enum": ["name", "email", "created_date", "last_invoice"],
      "default": "name"
    },
    "sort_order": {
      "type": "string",
      "enum": ["asc", "desc"],
      "default": "asc"
    },
    "limit": {
      "type": "number",
      "minimum": 1,
      "maximum": 500,
      "default": 50
    },
    "output_format": {
      "type": "string",
      "enum": ["table", "json", "csv"],
      "default": "table"
    },
    "include_contact_details": {
      "type": "boolean",
      "default": true,
      "description": "Include full contact information"
    },
    "include_invoice_summary": {
      "type": "boolean", 
      "default": false,
      "description": "Include invoice count and totals"
    }
  }
}
```

**Output Format**:

*Table Format*:
```
┌────────────────────┬─────────────────────────┬──────────────────┬─────────────┬──────────────┐
│ Name               │ Email                   │ Phone            │ Invoices    │ Total Billed │
├────────────────────┼─────────────────────────┼──────────────────┼─────────────┼──────────────┤
│ Acme Corporation   │ contact@acme.com        │ +1-555-123-4567  │ 5           │ $12,500.00   │
│ Tech Solutions Inc │ billing@techsolutions.c │ +1-555-987-6543  │ 3           │ $8,750.00    │
└────────────────────┴─────────────────────────┴──────────────────┴─────────────┴──────────────┘
```

*JSON Format*:
```json
{
  "clients": [
    {
      "id": "client_123",
      "name": "Acme Corporation",
      "email": "contact@acme.com",
      "phone": "+1-555-123-4567",
      "address": "123 Business St, City, State 12345",
      "payment_terms": 30,
      "created_at": "2025-01-15T09:00:00Z",
      "invoice_summary": {
        "total_invoices": 5,
        "total_billed": 12500.00,
        "unpaid_amount": 2500.00,
        "overdue_count": 1
      }
    }
  ],
  "summary": {
    "total_clients": 25,
    "active_clients": 23,
    "clients_with_overdue": 3
  }
}
```

**Usage Examples**:

*List All Clients*:
```json
{
  "output_format": "table",
  "include_invoice_summary": true
}
```

*Find Clients with Overdue Payments*:
```json
{
  "payment_overdue": true,
  "sort_by": "last_invoice",
  "sort_order": "desc"
}
```

*Search by Name*:
```json
{
  "name_filter": "Tech",
  "include_contact_details": true,
  "output_format": "json"
}
```

**Error Conditions**:
- `INVALID_SORT_FIELD` (400): Unknown sort field specified
- `LIMIT_EXCEEDED` (400): Limit parameter exceeds maximum
- `SEARCH_TIMEOUT` (408): Search operation exceeded timeout
- `INVALID_FILTER` (400): Filter parameters contain invalid values

**Security Considerations**:
- Client access controls based on user permissions
- Contact information filtering for privacy compliance
- Audit logging for client data access
- Rate limiting on search operations

**Performance Notes**:
- Search operations optimized with full-text indexing
- Invoice summaries calculated efficiently with aggregation queries
- Large result sets paginated automatically
- Contact details loaded on-demand for performance

**Related Tools**: `client_show`, `client_create`, `invoice_list`

---

#### Tool: `client_show`

**Description**: Display comprehensive details for a specific client including contact information, invoice history, and payment patterns

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "client_id": {
      "type": "string",
      "description": "Client ID to display details for",
      "minLength": 1
    },
    "client_name": {
      "type": "string",
      "description": "Client name to display details for",
      "minLength": 1
    },
    "client_email": {
      "type": "string",
      "format": "email",
      "description": "Client email to identify client"
    },
    "output_format": {
      "type": "string",
      "enum": ["text", "json", "yaml"],
      "default": "text"
    },
    "include_invoice_history": {
      "type": "boolean",
      "default": true,
      "description": "Include complete invoice history"
    },
    "include_payment_analysis": {
      "type": "boolean",
      "default": false,
      "description": "Include payment pattern analysis"
    },
    "invoice_limit": {
      "type": "number",
      "minimum": 1,
      "maximum": 100,
      "default": 10,
      "description": "Maximum number of recent invoices to show"
    }
  },
  "anyOf": [
    {"required": ["client_id"]},
    {"required": ["client_name"]},
    {"required": ["client_email"]}
  ]
}
```

**Output Format**:

*Text Format*:
```
Client Details: Acme Corporation
================================

Contact Information:
  ID: client_123
  Name: Acme Corporation
  Email: contact@acme.com
  Phone: +1-555-123-4567
  Address: 123 Business St, City, State 12345
  Tax ID: EIN-12-3456789
  Payment Terms: 30 days

Invoice History (Last 10):
  INV-001  │ 2025-08-01 │ Sent     │ $1,250.00 │ Due: 2025-08-31
  INV-002  │ 2025-07-01 │ Paid     │ $2,100.00 │ Paid: 2025-07-28
  INV-003  │ 2025-06-01 │ Paid     │ $1,875.00 │ Paid: 2025-06-25

Financial Summary:
  Total Invoiced: $15,750.00
  Total Paid: $13,250.00
  Outstanding: $2,500.00
  Average Payment Time: 22 days
  On-time Payment Rate: 85%
```

*JSON Format*:
```json
{
  "client": {
    "id": "client_123",
    "name": "Acme Corporation",
    "email": "contact@acme.com",
    "phone": "+1-555-123-4567",
    "address": "123 Business St, City, State 12345",
    "tax_id": "EIN-12-3456789",
    "payment_terms": 30,
    "created_at": "2025-01-15T09:00:00Z",
    "updated_at": "2025-08-01T14:30:00Z"
  },
  "invoice_history": [
    {
      "number": "INV-001",
      "date": "2025-08-01",
      "status": "sent",
      "amount": 1250.00,
      "due_date": "2025-08-31"
    }
  ],
  "financial_summary": {
    "total_invoiced": 15750.00,
    "total_paid": 13250.00,
    "outstanding_balance": 2500.00,
    "invoice_count": 8,
    "paid_invoice_count": 6
  },
  "payment_analysis": {
    "average_payment_days": 22,
    "on_time_percentage": 85,
    "payment_trend": "improving",
    "risk_score": "low"
  }
}
```

**Usage Examples**:

*Complete Client Profile*:
```json
{
  "client_name": "Acme Corporation",
  "include_invoice_history": true,
  "include_payment_analysis": true
}
```

*Quick Client Summary*:
```json
{
  "client_email": "contact@acme.com",
  "include_invoice_history": false,
  "output_format": "text"
}
```

*Financial Analysis*:
```json
{
  "client_id": "client_123",
  "output_format": "json",
  "include_payment_analysis": true,
  "invoice_limit": 50
}
```

**Error Conditions**:
- `CLIENT_NOT_FOUND` (404): Client with specified identifier not found
- `MULTIPLE_MATCHES` (409): Multiple clients match the search criteria
- `ACCESS_DENIED` (403): Insufficient permissions to view client details
- `TIMEOUT_ERROR` (408): Query exceeded 15 second timeout

**Security Considerations**:
- Client access controls prevent unauthorized viewing
- Financial data filtered based on user permissions
- Audit logging for all client data access
- Payment analysis data protected with additional security

**Performance Notes**:
- Client lookup optimized with proper indexing
- Invoice history loaded with efficient pagination
- Payment analysis calculated using cached aggregations
- Complex queries optimized for performance

**Related Tools**: `client_list`, `client_update`, `invoice_list`

---

#### Tool: `client_update`

**Description**: Update client information including contact details, payment terms, and business information with validation and change tracking

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "client_id": {
      "type": "string",
      "description": "Client ID to update",
      "minLength": 1
    },
    "client_name": {
      "type": "string",
      "description": "Client name to identify for update",
      "minLength": 1
    },
    "client_email": {
      "type": "string",
      "format": "email",
      "description": "Current client email to identify for update"
    },
    "name": {
      "type": "string",
      "maxLength": 200,
      "description": "Update client name"
    },
    "email": {
      "type": "string",
      "format": "email",
      "description": "Update email address"
    },
    "phone": {
      "type": "string",
      "pattern": "^[\\d\\s\\+\\-\\(\\)\\.\\/ext]+$",
      "description": "Update phone number"
    },
    "address": {
      "type": "string",
      "maxLength": 500,
      "description": "Update address"
    },
    "tax_id": {
      "type": "string",
      "maxLength": 50,
      "description": "Update tax ID"
    },
    "payment_terms": {
      "type": "number",
      "minimum": 1,
      "maximum": 365,
      "description": "Update payment terms in days"
    },
    "notes": {
      "type": "string",
      "maxLength": 1000,
      "description": "Update notes about the client"
    }
  },
  "allOf": [
    {
      "anyOf": [
        {"required": ["client_id"]},
        {"required": ["client_name"]},
        {"required": ["client_email"]}
      ]
    },
    {
      "anyOf": [
        {"required": ["name"]},
        {"required": ["email"]},
        {"required": ["phone"]},
        {"required": ["address"]},
        {"required": ["tax_id"]},
        {"required": ["payment_terms"]},
        {"required": ["notes"]}
      ]
    }
  ]
}
```

**Output Format**:
```json
{
  "success": true,
  "client_id": "client_123",
  "updates_applied": {
    "email": {
      "old_value": "old@acme.com",
      "new_value": "contact@acme.com",
      "updated_at": "2025-08-03T14:30:00Z"
    },
    "phone": {
      "old_value": "+1-555-000-0000",
      "new_value": "+1-555-123-4567",
      "updated_at": "2025-08-03T14:30:00Z"
    }
  },
  "validation_results": {
    "email_unique": true,
    "contact_valid": true,
    "business_rules_satisfied": true
  },
  "affected_invoices": 5
}
```

**Usage Examples**:

*Update Contact Information*:
```json
{
  "client_name": "Acme Corporation",
  "email": "newcontact@acme.com",
  "phone": "+1-555-999-8888"
}
```

*Update Payment Terms*:
```json
{
  "client_id": "client_123",
  "payment_terms": 45,
  "notes": "Extended payment terms approved by management"
}
```

*Update Address*:
```json
{
  "client_email": "contact@acme.com",
  "address": "456 New Business Plaza, Suite 300, New City, NC 54321"
}
```

**Error Conditions**:
- `CLIENT_NOT_FOUND` (404): Client with specified identifier not found
- `EMAIL_ALREADY_EXISTS` (409): New email address already in use
- `VALIDATION_ERROR` (400): Invalid field values or formats
- `MULTIPLE_MATCHES` (409): Multiple clients match the identifier
- `UPDATE_CONFLICT` (409): Concurrent update detected

**Security Considerations**:
- Email uniqueness validation prevents duplicate clients
- Change tracking maintains audit trail for compliance
- Permission checks for client modification rights
- Sensitive data updates require additional validation

**Performance Notes**:
- Optimistic locking prevents concurrent update conflicts
- Email uniqueness checked efficiently with database constraints
- Change tracking minimizes database update overhead
- Affected invoice updates processed asynchronously

**Related Tools**: `client_show`, `client_create`, `client_list`

---

#### Tool: `client_delete`

**Description**: Delete clients with dependency checking and safety confirmations. Prevents deletion of clients with active invoices

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "client_id": {
      "type": "string",
      "description": "Client ID to delete",
      "minLength": 1
    },
    "client_name": {
      "type": "string",
      "description": "Client name to identify for deletion",
      "minLength": 1
    },
    "client_email": {
      "type": "string",
      "format": "email",
      "description": "Client email to identify for deletion"
    },
    "force_delete": {
      "type": "boolean",
      "default": false,
      "description": "Force delete even with warnings (use with caution)"
    },
    "archive_invoices": {
      "type": "boolean",
      "default": true,
      "description": "Archive associated invoices instead of blocking deletion"
    }
  },
  "anyOf": [
    {"required": ["client_id"]},
    {"required": ["client_name"]},
    {"required": ["client_email"]}
  ]
}
```

**Output Format**:
```json
{
  "success": true,
  "client_id": "client_123",
  "client_name": "Acme Corporation",
  "deleted_at": "2025-08-03T15:45:00Z",
  "dependency_check": {
    "invoice_count": 0,
    "active_invoices": 0,
    "paid_invoices": 3,
    "archived_invoices": 3
  },
  "cleanup_actions": [
    "Invoice references updated to archived client",
    "Client data moved to archive table",
    "Audit trail preserved"
  ]
}
```

**Usage Examples**:

*Safe Client Deletion*:
```json
{
  "client_name": "Old Client Corp"
}
```

*Force Delete with Archive*:
```json
{
  "client_id": "client_inactive",
  "force_delete": true,
  "archive_invoices": true
}
```

*Delete by Email*:
```json
{
  "client_email": "defunct@company.com",
  "archive_invoices": true
}
```

**Error Conditions**:
- `CLIENT_NOT_FOUND` (404): Client with specified identifier not found
- `ACTIVE_INVOICES_EXIST` (409): Client has active invoices preventing deletion
- `PAID_INVOICES_PROTECTED` (409): Client has paid invoices requiring special handling
- `USER_CANCELLED` (400): User declined confirmation prompt
- `ARCHIVE_FAILED` (500): Invoice archiving process failed

**Security Considerations**:
- Dependency checking prevents data integrity issues
- Confirmation prompts protect against accidental deletion
- Invoice archiving preserves financial audit trail
- Permission verification for delete operations
- Backup creation before permanent deletion

**Performance Notes**:
- Dependency checking optimized with indexed queries
- Archive operations processed efficiently
- Cleanup actions performed in database transactions
- Large client deletions handled with batch processing

**Related Tools**: `client_show`, `client_list`, `invoice_list`

---

### Category: Data Import

**Purpose**: CSV import and validation workflows for timesheet and client data  
**Tools**: 3 tools for comprehensive data import operations  
**Common Use Cases**: Timesheet import, data validation, bulk client import

#### Tool: `import_csv`

**Description**: Import timesheet data from CSV files with mapping options and destination control. Supports flexible column mapping and data validation

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "file_path": {
      "type": "string",
      "description": "Path to the CSV file to import",
      "minLength": 1
    },
    "destination": {
      "type": "string",
      "enum": ["new_invoice", "existing_invoice", "preview_only"],
      "default": "new_invoice",
      "description": "Where to import the data"
    },
    "target_invoice_id": {
      "type": "string",
      "description": "Invoice ID for existing_invoice destination",
      "minLength": 1
    },
    "client_name": {
      "type": "string",
      "description": "Client name for new invoice creation",
      "minLength": 1
    },
    "client_id": {
      "type": "string",
      "description": "Client ID for new invoice creation",
      "minLength": 1
    },
    "column_mapping": {
      "type": "object",
      "description": "Custom column mapping for CSV fields",
      "properties": {
        "date": {"type": "string", "description": "Column name for date field"},
        "hours": {"type": "string", "description": "Column name for hours field"},
        "rate": {"type": "string", "description": "Column name for rate field"},
        "description": {"type": "string", "description": "Column name for description field"},
        "client": {"type": "string", "description": "Column name for client field"}
      }
    },
    "date_format": {
      "type": "string",
      "default": "auto",
      "description": "Date format in the CSV (auto-detect if not specified)",
      "examples": ["YYYY-MM-DD", "MM/DD/YYYY", "DD-MM-YYYY"]
    },
    "skip_header_rows": {
      "type": "number",
      "minimum": 0,
      "maximum": 10,
      "default": 1,
      "description": "Number of header rows to skip"
    },
    "validate_only": {
      "type": "boolean",
      "default": false,
      "description": "Only validate the CSV without importing"
    },
    "default_rate": {
      "type": "number",
      "minimum": 0.01,
      "description": "Default hourly rate if not specified in CSV"
    },
    "ignore_errors": {
      "type": "boolean",
      "default": false,
      "description": "Continue import despite validation errors"
    }
  },
  "required": ["file_path"],
  "allOf": [
    {
      "if": {
        "properties": {"destination": {"const": "existing_invoice"}}
      },
      "then": {
        "required": ["target_invoice_id"]
      }
    },
    {
      "if": {
        "properties": {"destination": {"const": "new_invoice"}}
      },
      "then": {
        "anyOf": [
          {"required": ["client_name"]},
          {"required": ["client_id"]}
        ]
      }
    }
  ]
}
```

**Output Format**:
```json
{
  "success": true,
  "import_summary": {
    "file_path": "/path/to/timesheet.csv",
    "rows_processed": 150,
    "rows_imported": 147,
    "rows_skipped": 3,
    "total_hours": 294.5,
    "total_amount": 36812.50
  },
  "destination_info": {
    "type": "new_invoice",
    "invoice_id": "INV-2025-042",
    "invoice_number": "INV-042",
    "client_name": "Acme Corporation"
  },
  "validation_results": {
    "errors": [
      {
        "row": 23,
        "field": "date",
        "message": "Invalid date format: '13/45/2025'",
        "suggestion": "Use YYYY-MM-DD format"
      }
    ],
    "warnings": [
      {
        "row": 45,
        "field": "rate",
        "message": "Rate seems high: $500/hour",
        "suggestion": "Verify premium rate is correct"
      }
    ]
  },
  "column_mapping_used": {
    "date": "Date",
    "hours": "Hours Worked",
    "rate": "Hourly Rate",
    "description": "Task Description"
  }
}
```

**Usage Examples**:

*Simple Timesheet Import*:
```json
{
  "file_path": "/Users/user/timesheets/august-2025.csv",
  "client_name": "Acme Corp",
  "destination": "new_invoice"
}
```

*Import to Existing Invoice*:
```json
{
  "file_path": "/Users/user/additional-hours.csv",
  "destination": "existing_invoice",
  "target_invoice_id": "INV-042"
}
```

*Custom Column Mapping*:
```json
{
  "file_path": "/Users/user/export.csv",
  "client_id": "client_123",
  "column_mapping": {
    "date": "Work Date",
    "hours": "Time Spent",
    "rate": "Bill Rate",
    "description": "Activity"
  },
  "date_format": "MM/DD/YYYY",
  "default_rate": 125.0
}
```

*Validation Only*:
```json
{
  "file_path": "/Users/user/timesheet-to-check.csv",
  "validate_only": true,
  "skip_header_rows": 2
}
```

**Error Conditions**:
- `FILE_NOT_FOUND` (404): CSV file path does not exist or is not accessible
- `INVALID_CSV_FORMAT` (400): File is not a valid CSV or is corrupted
- `COLUMN_MAPPING_FAILED` (400): Required columns not found in CSV
- `VALIDATION_ERRORS` (400): Data validation failed for multiple rows
- `CLIENT_NOT_FOUND` (404): Specified client for new invoice not found
- `INVOICE_NOT_FOUND` (404): Target invoice for import not found
- `IMPORT_TIMEOUT` (408): Import process exceeded timeout limit

**Security Considerations**:
- File path validation prevents directory traversal attacks
- CSV parsing with memory limits to prevent DoS attacks
- Client access validation prevents unauthorized data import
- Audit logging for all import operations with file metadata
- Rate validation prevents unrealistic billing rate imports

**Performance Notes**:
- Large CSV files processed in chunks to manage memory usage
- Parallel validation processing for improved speed
- Database batch inserts for efficient work item creation
- Progress tracking for long-running import operations
- Automatic cleanup of temporary processing files

**Related Tools**: `import_validate`, `import_preview`, `invoice_create`

---

#### Tool: `import_validate`

**Description**: Validate CSV structure and data before import execution. Provides comprehensive validation reports without making changes

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "file_path": {
      "type": "string",
      "description": "Path to the CSV file to validate",
      "minLength": 1
    },
    "expected_columns": {
      "type": "array",
      "description": "Expected column names in the CSV",
      "items": {"type": "string"},
      "minItems": 1
    },
    "column_mapping": {
      "type": "object",
      "description": "Column mapping to validate against",
      "properties": {
        "date": {"type": "string"},
        "hours": {"type": "string"},
        "rate": {"type": "string"},
        "description": {"type": "string"},
        "client": {"type": "string"}
      }
    },
    "date_format": {
      "type": "string",
      "default": "auto",
      "description": "Expected date format for validation"
    },
    "skip_header_rows": {
      "type": "number",
      "minimum": 0,
      "maximum": 10,
      "default": 1
    },
    "validation_rules": {
      "type": "object",
      "description": "Custom validation rules",
      "properties": {
        "max_hours_per_day": {"type": "number", "default": 24},
        "min_rate": {"type": "number", "default": 1},
        "max_rate": {"type": "number", "default": 1000},
        "require_description": {"type": "boolean", "default": true},
        "allow_future_dates": {"type": "boolean", "default": false}
      }
    },
    "sample_size": {
      "type": "number",
      "minimum": 10,
      "maximum": 1000,
      "default": 100,
      "description": "Number of rows to validate for performance"
    }
  },
  "required": ["file_path"]
}
```

**Output Format**:
```json
{
  "validation_summary": {
    "file_valid": true,
    "total_rows": 250,
    "rows_validated": 100,
    "error_count": 5,
    "warning_count": 12,
    "estimated_import_time": "45 seconds"
  },
  "file_analysis": {
    "file_size": "125.3 KB",
    "encoding": "UTF-8",
    "delimiter": ",",
    "quote_char": "\"",
    "columns_detected": ["Date", "Hours", "Rate", "Description", "Project"],
    "data_types": {
      "Date": "date",
      "Hours": "numeric",
      "Rate": "numeric", 
      "Description": "text",
      "Project": "text"
    }
  },
  "validation_errors": [
    {
      "row": 15,
      "column": "Date",
      "value": "2025-13-01",
      "error": "Invalid date: month 13 does not exist",
      "severity": "error",
      "suggestion": "Use valid month (01-12)"
    },
    {
      "row": 23,
      "column": "Hours",
      "value": "-2.5",
      "error": "Negative hours not allowed",
      "severity": "error",
      "suggestion": "Use positive values for hours worked"
    }
  ],
  "validation_warnings": [
    {
      "row": 45,
      "column": "Rate",
      "value": "350.00",
      "warning": "Rate above typical range",
      "severity": "warning",
      "suggestion": "Verify premium rate is intentional"
    }
  ],
  "column_mapping_analysis": {
    "recommended_mapping": {
      "date": "Date",
      "hours": "Hours",
      "rate": "Rate",
      "description": "Description"
    },
    "mapping_confidence": 0.95,
    "ambiguous_columns": []
  },
  "data_quality_metrics": {
    "completeness": 0.98,
    "consistency": 0.92,
    "accuracy": 0.96,
    "quality_score": "A-"
  }
}
```

**Usage Examples**:

*Basic CSV Validation*:
```json
{
  "file_path": "/Users/user/timesheet.csv"
}
```

*Validation with Custom Rules*:
```json
{
  "file_path": "/Users/user/contractor-hours.csv",
  "validation_rules": {
    "max_hours_per_day": 12,
    "min_rate": 50,
    "max_rate": 500,
    "allow_future_dates": true
  }
}
```

*Column Mapping Validation*:
```json
{
  "file_path": "/Users/user/export.csv",
  "column_mapping": {
    "date": "Work Date",
    "hours": "Time Spent",
    "rate": "Billing Rate",
    "description": "Task Details"
  },
  "date_format": "DD/MM/YYYY"
}
```

**Error Conditions**:
- `FILE_NOT_FOUND` (404): CSV file path does not exist
- `FILE_ACCESS_DENIED` (403): Insufficient permissions to read file
- `INVALID_CSV_FORMAT` (400): File is not a valid CSV format
- `ENCODING_ERROR` (400): File encoding not supported
- `FILE_TOO_LARGE` (413): File exceeds maximum size limit

**Security Considerations**:
- File access validation prevents unauthorized file system access
- Memory limits during validation prevent resource exhaustion
- Path traversal protection for file path parameters
- Audit logging for all validation operations

**Performance Notes**:
- Sample-based validation for large files improves performance
- Streaming CSV parsing minimizes memory usage
- Parallel validation of multiple data quality checks
- Caching of validation results for repeated operations

**Related Tools**: `import_csv`, `import_preview`, `export_data`

---

#### Tool: `import_preview`

**Description**: Preview import results without making any changes. Shows exactly what would be imported and how data would be processed

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "file_path": {
      "type": "string",
      "description": "Path to the CSV file to preview",
      "minLength": 1
    },
    "destination": {
      "type": "string",
      "enum": ["new_invoice", "existing_invoice"],
      "default": "new_invoice",
      "description": "Preview destination for import"
    },
    "target_invoice_id": {
      "type": "string",
      "description": "Invoice ID for existing_invoice preview",
      "minLength": 1
    },
    "client_name": {
      "type": "string",
      "description": "Client name for new invoice preview",
      "minLength": 1
    },
    "client_id": {
      "type": "string",
      "description": "Client ID for new invoice preview",
      "minLength": 1
    },
    "column_mapping": {
      "type": "object",
      "description": "Column mapping for preview",
      "properties": {
        "date": {"type": "string"},
        "hours": {"type": "string"},
        "rate": {"type": "string"},
        "description": {"type": "string"},
        "client": {"type": "string"}
      }
    },
    "date_format": {
      "type": "string",
      "default": "auto"
    },
    "skip_header_rows": {
      "type": "number",
      "minimum": 0,
      "maximum": 10,
      "default": 1
    },
    "preview_rows": {
      "type": "number",
      "minimum": 5,
      "maximum": 100,
      "default": 20,
      "description": "Number of rows to include in preview"
    },
    "default_rate": {
      "type": "number",
      "minimum": 0.01,
      "description": "Default rate for preview calculations"
    }
  },
  "required": ["file_path"]
}
```

**Output Format**:
```json
{
  "preview_summary": {
    "file_path": "/Users/user/timesheet.csv",
    "total_rows_in_file": 150,
    "rows_in_preview": 20,
    "estimated_total_hours": 294.5,
    "estimated_total_amount": 36812.50,
    "would_create_invoice": true
  },
  "destination_preview": {
    "type": "new_invoice",
    "client": {
      "id": "client_123",
      "name": "Acme Corporation",
      "email": "contact@acme.com"
    },
    "invoice_preview": {
      "estimated_number": "INV-043",
      "estimated_date": "2025-08-03",
      "estimated_due_date": "2025-09-02"
    }
  },
  "sample_work_items": [
    {
      "row_number": 2,
      "date": "2025-08-01",
      "hours": 8.0,
      "rate": 125.0,
      "description": "Frontend development and testing",
      "amount": 1000.0,
      "status": "valid"
    },
    {
      "row_number": 3,
      "date": "2025-08-02",
      "hours": 6.5,
      "rate": 125.0,
      "description": "Bug fixes and optimization",
      "amount": 812.5,
      "status": "valid"
    },
    {
      "row_number": 4,
      "date": "2025-08-03",
      "hours": 4.0,
      "rate": 150.0,
      "description": "Client consultation",
      "amount": 600.0,
      "status": "valid"
    }
  ],
  "validation_preview": {
    "estimated_errors": 3,
    "estimated_warnings": 8,
    "common_issues": [
      "2 rows with missing descriptions",
      "1 row with future date"
    ]
  },
  "import_plan": {
    "steps": [
      "Validate CSV structure and data",
      "Create new invoice for Acme Corporation",
      "Import 147 valid work items",
      "Calculate invoice totals",
      "Set invoice status to draft"
    ],
    "estimated_duration": "30-45 seconds",
    "reversible": true
  }
}
```

**Usage Examples**:

*Preview New Invoice Import*:
```json
{
  "file_path": "/Users/user/july-timesheet.csv",
  "client_name": "Tech Solutions Inc",
  "preview_rows": 25
}
```

*Preview Addition to Existing Invoice*:
```json
{
  "file_path": "/Users/user/additional-work.csv",
  "destination": "existing_invoice",
  "target_invoice_id": "INV-042",
  "preview_rows": 15
}
```

*Preview with Custom Mapping*:
```json
{
  "file_path": "/Users/user/custom-export.csv",
  "client_id": "client_456",
  "column_mapping": {
    "date": "Work Date",
    "hours": "Duration",
    "rate": "Bill Rate",
    "description": "Activity Description"
  },
  "default_rate": 135.0
}
```

**Error Conditions**:
- `FILE_NOT_FOUND` (404): CSV file path does not exist
- `CLIENT_NOT_FOUND` (404): Specified client not found for preview
- `INVOICE_NOT_FOUND` (404): Target invoice for preview not found
- `PREVIEW_GENERATION_FAILED` (500): Error generating preview data
- `INVALID_PREVIEW_SIZE` (400): Preview rows parameter out of range

**Security Considerations**:
- File access validation for security
- Client access verification for preview generation
- Preview data sanitization to prevent information leakage
- Audit logging for preview operations

**Performance Notes**:
- Preview generation optimized for speed over completeness
- Sample-based calculations for large files
- Cached preview results for repeated requests
- Memory-efficient preview data structures

**Related Tools**: `import_csv`, `import_validate`, `invoice_create`

---

### Category: Data Export

**Purpose**: Document generation and data export in various formats  
**Tools**: 3 tools for comprehensive export operations  
**Common Use Cases**: Invoice generation, report creation, data export

#### Tool: `generate_html`

**Description**: Generate HTML invoices with template options and customization. Supports multiple templates and styling options for professional invoice presentation

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "invoice_id": {
      "type": "string",
      "description": "Invoice ID to generate HTML for",
      "minLength": 1
    },
    "invoice_number": {
      "type": "string",
      "description": "Invoice number to generate HTML for",
      "minLength": 1
    },
    "template": {
      "type": "string",
      "enum": ["default", "modern", "minimal", "professional", "branded"],
      "default": "default",
      "description": "HTML template to use for generation"
    },
    "output_path": {
      "type": "string",
      "description": "Path where HTML file should be saved",
      "minLength": 1
    },
    "include_css": {
      "type": "boolean",
      "default": true,
      "description": "Include CSS styling in the HTML"
    },
    "css_file": {
      "type": "string",
      "description": "Path to custom CSS file for styling"
    },
    "company_logo": {
      "type": "string",
      "description": "Path to company logo image"
    },
    "watermark": {
      "type": "string",
      "description": "Watermark text for the invoice (e.g., 'DRAFT', 'PAID')"
    },
    "page_size": {
      "type": "string",
      "enum": ["A4", "Letter", "Legal"],
      "default": "A4",
      "description": "Page size for print formatting"
    },
    "currency_symbol": {
      "type": "string",
      "default": "$",
      "description": "Currency symbol for amounts"
    },
    "custom_fields": {
      "type": "object",
      "description": "Custom fields to include in the invoice",
      "additionalProperties": {"type": "string"}
    },
    "show_work_items": {
      "type": "boolean",
      "default": true,
      "description": "Include detailed work items in HTML"
    },
    "group_by_date": {
      "type": "boolean",
      "default": false,
      "description": "Group work items by date"
    }
  },
  "anyOf": [
    {"required": ["invoice_id"]},
    {"required": ["invoice_number"]}
  ]
}
```

**Output Format**:
```json
{
  "success": true,
  "html_file": "/Users/user/invoices/INV-001.html",
  "file_size": "45.2 KB",
  "template_used": "modern",
  "generation_time": "2.3 seconds",
  "invoice_details": {
    "invoice_number": "INV-001",
    "client_name": "Acme Corporation",
    "total_amount": 1695.31,
    "work_items_count": 5
  },
  "template_features": [
    "Responsive design",
    "Print-optimized layout",
    "Company branding",
    "Professional styling"
  ],
  "accessibility": {
    "screen_reader_compatible": true,
    "high_contrast_available": true,
    "keyboard_navigation": true
  }
}
```

**Usage Examples**:

*Basic HTML Generation*:
```json
{
  "invoice_number": "INV-001",
  "output_path": "/Users/user/invoices/INV-001.html"
}
```

*Professional Template with Branding*:
```json
{
  "invoice_id": "invoice_abc123",
  "template": "professional",
  "output_path": "/Users/user/invoices/",
  "company_logo": "/Users/user/assets/company-logo.png",
  "watermark": "CONFIDENTIAL",
  "custom_fields": {
    "purchase_order": "PO-2025-001",
    "project_code": "PROJ-WEB-001"
  }
}
```

*Minimal Template for Email*:
```json
{
  "invoice_number": "INV-042",
  "template": "minimal", 
  "output_path": "/tmp/invoice.html",
  "include_css": true,
  "show_work_items": false,
  "page_size": "Letter"
}
```

**Error Conditions**:
- `INVOICE_NOT_FOUND` (404): Invoice with specified ID/number not found
- `TEMPLATE_NOT_FOUND` (404): Specified template does not exist
- `OUTPUT_PATH_INVALID` (400): Output path is not writable or invalid
- `LOGO_FILE_NOT_FOUND` (404): Company logo file does not exist
- `CSS_FILE_INVALID` (400): Custom CSS file is invalid or not found
- `GENERATION_FAILED` (500): HTML generation process failed

**Security Considerations**:
- Output path validation prevents directory traversal
- Template validation prevents code injection
- Logo and CSS file validation for security
- HTML sanitization for safe output generation
- Audit logging for all document generation

**Performance Notes**:
- Template caching for improved generation speed
- Optimized HTML/CSS output for smaller file sizes
- Parallel processing for batch generation
- Memory-efficient template rendering
- Image optimization for embedded assets

**Related Tools**: `invoice_show`, `generate_summary`, `export_data`

---

#### Tool: `generate_summary`

**Description**: Create invoice summaries and reports with flexible aggregation options. Supports various time periods and grouping criteria

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "report_type": {
      "type": "string",
      "enum": ["client_summary", "monthly_summary", "status_summary", "payment_summary", "detailed_report"],
      "default": "monthly_summary",
      "description": "Type of summary report to generate"
    },
    "date_range": {
      "type": "object",
      "properties": {
        "from_date": {
          "type": "string",
          "format": "date",
          "description": "Start date for report period"
        },
        "to_date": {
          "type": "string", 
          "format": "date",
          "description": "End date for report period"
        },
        "period": {
          "type": "string",
          "enum": ["this_month", "last_month", "this_quarter", "last_quarter", "this_year", "last_year"],
          "description": "Predefined period for report"
        }
      },
      "anyOf": [
        {"required": ["from_date", "to_date"]},
        {"required": ["period"]}
      ]
    },
    "client_filter": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Filter by specific client IDs"
    },
    "status_filter": {
      "type": "array",
      "items": {"enum": ["draft", "sent", "paid", "overdue", "voided"]},
      "description": "Filter by invoice status"
    },
    "group_by": {
      "type": "string",
      "enum": ["client", "month", "status", "quarter"],
      "default": "month",
      "description": "How to group the summary data"
    },
    "output_format": {
      "type": "string",
      "enum": ["json", "html", "csv", "pdf"],
      "default": "json",
      "description": "Output format for the summary"
    },
    "output_path": {
      "type": "string",
      "description": "Path where summary file should be saved"
    },
    "include_charts": {
      "type": "boolean",
      "default": false,
      "description": "Include charts and graphs in HTML/PDF output"
    },
    "currency": {
      "type": "string",
      "default": "USD",
      "description": "Currency for financial calculations"
    },
    "include_projections": {
      "type": "boolean",
      "default": false,
      "description": "Include revenue projections based on trends"
    }
  }
}
```

**Output Format**:
```json
{
  "summary": {
    "report_type": "monthly_summary",
    "period": {
      "from_date": "2025-08-01",
      "to_date": "2025-08-31",
      "description": "August 2025"
    },
    "totals": {
      "total_invoices": 25,
      "total_amount": 125750.00,
      "paid_amount": 98250.00,
      "outstanding_amount": 27500.00,
      "overdue_amount": 5250.00
    },
    "by_status": {
      "draft": {"count": 3, "amount": 8750.00},
      "sent": {"count": 12, "amount": 18750.00},
      "paid": {"count": 8, "amount": 98250.00},
      "overdue": {"count": 2, "amount": 5250.00}
    },
    "by_client": [
      {
        "client_id": "client_123",
        "client_name": "Acme Corporation",
        "invoice_count": 5,
        "total_amount": 35250.00,
        "paid_amount": 28500.00,
        "outstanding": 6750.00
      }
    ],
    "trends": {
      "month_over_month_growth": 0.15,
      "average_invoice_value": 5030.00,
      "payment_cycle_days": 22,
      "collection_rate": 0.92
    }
  },
  "metadata": {
    "generated_at": "2025-08-03T16:30:00Z",
    "report_version": "1.2",
    "data_as_of": "2025-08-03T16:00:00Z",
    "filters_applied": ["status_filter", "date_range"],
    "currency": "USD"
  },
  "file_info": {
    "output_file": "/Users/user/reports/august-2025-summary.json",
    "file_size": "12.5 KB",
    "generation_time": "1.8 seconds"
  }
}
```

**Usage Examples**:

*Monthly Business Summary*:
```json
{
  "report_type": "monthly_summary",
  "date_range": {"period": "this_month"},
  "output_format": "html",
  "include_charts": true,
  "output_path": "/Users/user/reports/"
}
```

*Client Performance Report*:
```json
{
  "report_type": "client_summary",
  "date_range": {
    "from_date": "2025-01-01",
    "to_date": "2025-08-31"
  },
  "group_by": "client",
  "output_format": "pdf",
  "include_projections": true
}
```

*Outstanding Payments Report*:
```json
{
  "report_type": "payment_summary",
  "status_filter": ["sent", "overdue"],
  "date_range": {"period": "this_year"},
  "output_format": "csv",
  "group_by": "status"
}
```

**Error Conditions**:
- `INVALID_DATE_RANGE` (400): Date range is invalid or end date before start date
- `NO_DATA_FOUND` (404): No invoices match the specified criteria
- `OUTPUT_PATH_INVALID` (400): Output path is not writable
- `CHART_GENERATION_FAILED` (500): Chart generation failed for HTML/PDF output
- `REPORT_TIMEOUT` (408): Report generation exceeded timeout

**Security Considerations**:
- Client filtering based on user permissions
- Output path validation for security
- Financial data access controls
- Audit logging for all report generation
- Data anonymization options for sensitive reports

**Performance Notes**:
- Aggregation queries optimized with proper indexing
- Large dataset handling with pagination
- Chart generation optimized for performance
- Caching of common report calculations
- Asynchronous processing for complex reports

**Related Tools**: `invoice_list`, `export_data`, `generate_html`

---

#### Tool: `export_data`

**Description**: Export invoice data in various formats with filtering and customization options. Supports CSV, JSON, XML, and other formats for integration with external systems

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "export_type": {
      "type": "string",
      "enum": ["invoices", "clients", "work_items", "payments", "full_backup"],
      "default": "invoices",
      "description": "Type of data to export"
    },
    "output_format": {
      "type": "string",
      "enum": ["csv", "json", "xml", "xlsx", "yaml"],
      "default": "csv",
      "description": "Output format for exported data"
    },
    "output_path": {
      "type": "string",
      "description": "Path where export file should be saved",
      "minLength": 1
    },
    "date_range": {
      "type": "object",
      "properties": {
        "from_date": {
          "type": "string",
          "format": "date",
          "description": "Start date for export"
        },
        "to_date": {
          "type": "string",
          "format": "date", 
          "description": "End date for export"
        }
      }
    },
    "filters": {
      "type": "object",
      "properties": {
        "client_ids": {
          "type": "array",
          "items": {"type": "string"},
          "description": "Filter by specific client IDs"
        },
        "statuses": {
          "type": "array",
          "items": {"enum": ["draft", "sent", "paid", "overdue", "voided"]},
          "description": "Filter by invoice status"
        },
        "min_amount": {
          "type": "number",
          "minimum": 0,
          "description": "Minimum invoice amount"
        },
        "max_amount": {
          "type": "number",
          "minimum": 0,
          "description": "Maximum invoice amount"
        }
      }
    },
    "fields": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Specific fields to include in export"
    },
    "include_related": {
      "type": "boolean",
      "default": false,
      "description": "Include related data (work items, client details)"
    },
    "anonymize": {
      "type": "boolean",
      "default": false,
      "description": "Anonymize sensitive data for external sharing"
    },
    "compression": {
      "type": "string",
      "enum": ["none", "zip", "gzip"],
      "default": "none",
      "description": "Compression format for output file"
    },
    "split_large_files": {
      "type": "boolean",
      "default": false,
      "description": "Split large exports into multiple files"
    },
    "max_file_size": {
      "type": "string",
      "default": "50MB",
      "description": "Maximum size per file when splitting"
    }
  },
  "required": ["output_path"]
}
```

**Output Format**:
```json
{
  "success": true,
  "export_summary": {
    "export_type": "invoices",
    "output_format": "csv",
    "records_exported": 1250,
    "file_size": "2.8 MB",
    "generation_time": "15.2 seconds"
  },
  "files_created": [
    {
      "path": "/Users/user/exports/invoices-2025.csv",
      "size": "2.8 MB",
      "records": 1250,
      "checksum": "sha256:a1b2c3d4..."
    }
  ],
  "data_summary": {
    "date_range": {
      "from": "2025-01-01",
      "to": "2025-08-31"
    },
    "totals": {
      "total_amount": 458750.00,
      "invoice_count": 1250,
      "client_count": 85,
      "work_items": 8420
    },
    "filters_applied": {
      "status_filter": ["paid", "sent"],
      "date_range": true,
      "amount_range": false
    }
  },
  "format_details": {
    "csv_delimiter": ",",
    "encoding": "UTF-8",
    "header_row": true,
    "date_format": "YYYY-MM-DD",
    "number_format": "decimal"
  },
  "compatibility": {
    "excel_compatible": true,
    "quickbooks_ready": true,
    "accounting_standard": "compatible"
  }
}
```

**Usage Examples**:

*Export All Invoices to CSV*:
```json
{
  "export_type": "invoices",
  "output_format": "csv",
  "output_path": "/Users/user/exports/all-invoices.csv",
  "include_related": true
}
```

*Export Paid Invoices for Accounting*:
```json
{
  "export_type": "invoices",
  "output_format": "xlsx",
  "output_path": "/Users/user/accounting/",
  "filters": {
    "statuses": ["paid"],
    "min_amount": 100
  },
  "date_range": {
    "from_date": "2025-01-01",
    "to_date": "2025-12-31"
  }
}
```

*Full Backup Export*:
```json
{
  "export_type": "full_backup",
  "output_format": "json",
  "output_path": "/Users/user/backups/",
  "compression": "zip",
  "include_related": true,
  "split_large_files": true
}
```

*Client Data for CRM*:
```json
{
  "export_type": "clients",
  "output_format": "csv",
  "output_path": "/Users/user/crm-import.csv",
  "fields": ["name", "email", "phone", "total_billed"],
  "anonymize": false
}
```

**Error Conditions**:
- `OUTPUT_PATH_INVALID` (400): Output path is not writable or invalid
- `NO_DATA_TO_EXPORT` (404): No records match the specified criteria
- `FILE_SIZE_EXCEEDED` (413): Export would exceed maximum file size
- `FORMAT_NOT_SUPPORTED` (400): Requested output format not available
- `EXPORT_TIMEOUT` (408): Export process exceeded timeout limit
- `DISK_SPACE_INSUFFICIENT` (507): Not enough disk space for export

**Security Considerations**:
- Data access controls based on user permissions
- Anonymization options for sensitive data sharing
- Output path validation prevents unauthorized file access
- Audit logging for all data export operations
- Encryption options for sensitive exports

**Performance Notes**:
- Streaming export for large datasets
- Parallel processing for complex exports
- Compression algorithms optimized for speed
- Memory-efficient processing for large files
- Progress tracking for long-running exports

**Related Tools**: `invoice_list`, `client_list`, `generate_summary`

---

### Category: Configuration

**Purpose**: System configuration, settings management, and validation  
**Tools**: 3 tools for comprehensive configuration operations  
**Common Use Cases**: System setup, configuration validation, settings management

#### Tool: `config_show`

**Description**: Display current configuration with formatting options and comprehensive settings overview

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "output_format": {
      "type": "string",
      "enum": ["text", "json", "yaml", "table"],
      "default": "text",
      "description": "Output format for configuration display"
    },
    "section": {
      "type": "string",
      "enum": ["all", "database", "invoice", "email", "templates", "security", "performance"],
      "default": "all",
      "description": "Configuration section to display"
    },
    "show_sensitive": {
      "type": "boolean",
      "default": false,
      "description": "Include sensitive configuration values (masked)"
    },
    "show_defaults": {
      "type": "boolean",
      "default": false,
      "description": "Include default values that haven't been customized"
    },
    "include_validation": {
      "type": "boolean",
      "default": false,
      "description": "Include validation status for each setting"
    },
    "show_descriptions": {
      "type": "boolean",
      "default": true,
      "description": "Include descriptions for configuration options"
    }
  }
}
```

**Output Format**:

*Text Format*:
```
Configuration Overview
======================

Database Configuration:
  Host: localhost
  Port: 5432
  Database: go_invoice_db
  SSL Mode: require
  Connection Pool: 20 connections
  Status: ✓ Connected

Invoice Settings:
  Default Tax Rate: 8.50%
  Default Payment Terms: 30 days
  Invoice Number Format: INV-{YYYY}-{###}
  Currency: USD ($)
  Date Format: YYYY-MM-DD
  Status: ✓ Valid

Email Configuration:
  SMTP Host: smtp.company.com
  SMTP Port: 587
  Username: invoice@company.com
  Password: [MASKED]
  TLS Enabled: true
  Status: ⚠ Not tested

Template Settings:
  Default Template: modern
  Company Logo: /assets/logo.png
  Custom CSS: enabled
  Watermarks: enabled
  Status: ✓ Valid

Security Configuration:
  Rate Limiting: enabled
  Session Timeout: 30 minutes
  Password Policy: strong
  Audit Logging: enabled
  Status: ✓ Secure
```

*JSON Format*:
```json
{
  "configuration": {
    "database": {
      "host": "localhost",
      "port": 5432,
      "database": "go_invoice_db",
      "ssl_mode": "require",
      "connection_pool_size": 20,
      "status": "connected",
      "last_check": "2025-08-03T16:00:00Z"
    },
    "invoice": {
      "default_tax_rate": 0.085,
      "default_payment_terms": 30,
      "number_format": "INV-{YYYY}-{###}",
      "currency": "USD",
      "currency_symbol": "$",
      "date_format": "YYYY-MM-DD",
      "status": "valid"
    },
    "email": {
      "smtp_host": "smtp.company.com",
      "smtp_port": 587,
      "username": "invoice@company.com",
      "password": "[MASKED]",
      "tls_enabled": true,
      "status": "untested",
      "last_test": null
    },
    "templates": {
      "default_template": "modern",
      "logo_path": "/assets/logo.png",
      "custom_css_enabled": true,
      "watermarks_enabled": true,
      "available_templates": ["default", "modern", "minimal", "professional"],
      "status": "valid"
    },
    "security": {
      "rate_limiting_enabled": true,
      "session_timeout_minutes": 30,
      "password_policy": "strong",
      "audit_logging_enabled": true,
      "status": "secure"
    }
  },
  "metadata": {
    "config_version": "1.2.0",
    "last_updated": "2025-08-01T09:00:00Z",
    "environment": "production",
    "validation_status": "all_valid"
  }
}
```

**Usage Examples**:

*Complete Configuration Overview*:
```json
{
  "output_format": "text",
  "section": "all",
  "include_validation": true
}
```

*Database Settings Only*:
```json
{
  "output_format": "json",
  "section": "database",
  "show_sensitive": false
}
```

*Security Configuration Review*:
```json
{
  "output_format": "table",
  "section": "security",
  "show_descriptions": true,
  "include_validation": true
}
```

**Error Conditions**:
- `CONFIG_FILE_NOT_FOUND` (404): Configuration file does not exist
- `CONFIG_ACCESS_DENIED` (403): Insufficient permissions to view configuration
- `INVALID_SECTION` (400): Requested section does not exist
- `CONFIG_PARSE_ERROR` (500): Configuration file format is invalid

**Security Considerations**:
- Sensitive values masked by default
- Permission checks for configuration access
- Audit logging for configuration viewing
- Role-based access to different configuration sections

**Performance Notes**:
- Configuration cached for fast access
- Validation status computed efficiently
- Large configurations paginated appropriately
- Network settings tested on-demand

**Related Tools**: `config_validate`, `config_init`

---

#### Tool: `config_validate`

**Description**: Validate configuration integrity and report issues with comprehensive validation checks and recommendations

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "validation_level": {
      "type": "string",
      "enum": ["basic", "standard", "comprehensive", "security_audit"],
      "default": "standard",
      "description": "Level of validation to perform"
    },
    "sections": {
      "type": "array",
      "items": {"enum": ["database", "invoice", "email", "templates", "security", "performance"]},
      "description": "Specific sections to validate (all if not specified)"
    },
    "fix_issues": {
      "type": "boolean",
      "default": false,
      "description": "Automatically fix issues that can be safely resolved"
    },
    "test_connections": {
      "type": "boolean",
      "default": true,
      "description": "Test external connections (database, email, etc.)"
    },
    "output_format": {
      "type": "string",
      "enum": ["text", "json", "html"],
      "default": "text",
      "description": "Output format for validation report"
    },
    "include_recommendations": {
      "type": "boolean",
      "default": true,
      "description": "Include optimization recommendations"
    },
    "save_report": {
      "type": "boolean",
      "default": false,
      "description": "Save validation report to file"
    },
    "report_path": {
      "type": "string",
      "description": "Path to save validation report"
    }
  }
}
```

**Output Format**:
```json
{
  "validation_summary": {
    "overall_status": "warning",
    "validation_level": "standard",
    "sections_checked": 6,
    "issues_found": 3,
    "critical_issues": 0,
    "warnings": 2,
    "recommendations": 5,
    "auto_fixes_applied": 1
  },
  "section_results": {
    "database": {
      "status": "valid",
      "checks_performed": [
        "Connection test",
        "Schema validation",
        "Index optimization",
        "Performance metrics"
      ],
      "issues": [],
      "recommendations": [
        "Consider increasing connection pool size for better performance"
      ]
    },
    "email": {
      "status": "warning",
      "checks_performed": [
        "SMTP connection test",
        "Authentication validation",
        "TLS configuration"
      ],
      "issues": [
        {
          "severity": "warning",
          "code": "EMAIL_CONNECTION_FAILED",
          "message": "SMTP connection test failed: timeout after 30 seconds",
          "recommendation": "Check SMTP server settings and firewall configuration",
          "auto_fixable": false
        }
      ],
      "recommendations": [
        "Configure backup SMTP server for redundancy"
      ]
    },
    "security": {
      "status": "valid",
      "checks_performed": [
        "Password policy validation",
        "Session security",
        "Rate limiting",
        "Audit logging"
      ],
      "issues": [],
      "recommendations": [
        "Consider enabling two-factor authentication",
        "Update password policy to require special characters"
      ]
    }
  },
  "connection_tests": {
    "database": {
      "status": "success",
      "response_time": "12ms",
      "last_tested": "2025-08-03T16:30:00Z"
    },
    "smtp": {
      "status": "failed",
      "error": "Connection timeout",
      "last_tested": "2025-08-03T16:30:15Z"
    }
  },
  "auto_fixes": [
    {
      "issue": "Missing database index on invoices.client_id",
      "action": "Created index for improved query performance",
      "status": "applied"
    }
  ],
  "next_steps": [
    "Review SMTP configuration and test email sending",
    "Consider implementing backup email provider",
    "Schedule regular configuration validation"
  ]
}
```

**Usage Examples**:

*Comprehensive System Validation*:
```json
{
  "validation_level": "comprehensive",
  "test_connections": true,
  "include_recommendations": true,
  "save_report": true,
  "report_path": "/Users/user/reports/config-validation.html"
}
```

*Quick Database Check*:
```json
{
  "validation_level": "basic",
  "sections": ["database"],
  "test_connections": true,
  "output_format": "json"
}
```

*Security Audit*:
```json
{
  "validation_level": "security_audit",
  "sections": ["security"],
  "fix_issues": false,
  "output_format": "html",
  "save_report": true
}
```

**Error Conditions**:
- `CONFIG_FILE_CORRUPT` (500): Configuration file is corrupted or unreadable
- `VALIDATION_TIMEOUT` (408): Validation process exceeded timeout
- `CONNECTION_TEST_FAILED` (503): Critical connection tests failed
- `REPORT_SAVE_FAILED` (500): Unable to save validation report
- `INSUFFICIENT_PERMISSIONS` (403): Cannot perform certain validation checks

**Security Considerations**:
- Sensitive data protection during validation
- Security audit mode for enhanced checks
- Permission validation for configuration modifications
- Secure handling of connection credentials during testing

**Performance Notes**:
- Parallel validation of independent sections
- Caching of validation results for repeated checks
- Optimized connection testing with proper timeouts
- Resource usage monitoring during validation

**Related Tools**: `config_show`, `config_init`

---

#### Tool: `config_init`

**Description**: Initialize new configuration with guided setup and best practice recommendations

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "setup_mode": {
      "type": "string",
      "enum": ["interactive", "automated", "template"],
      "default": "interactive",
      "description": "Configuration setup mode"
    },
    "template": {
      "type": "string",
      "enum": ["development", "production", "testing", "minimal"],
      "description": "Configuration template to use (required for template mode)"
    },
    "config_path": {
      "type": "string",
      "description": "Path where configuration file should be created"
    },
    "overwrite_existing": {
      "type": "boolean",
      "default": false,
      "description": "Overwrite existing configuration file"
    },
    "database_config": {
      "type": "object",
      "properties": {
        "host": {"type": "string", "default": "localhost"},
        "port": {"type": "number", "default": 5432},
        "database": {"type": "string"},
        "username": {"type": "string"},
        "password": {"type": "string"},
        "ssl_mode": {"type": "string", "enum": ["disable", "require", "verify-full"], "default": "require"}
      }
    },
    "email_config": {
      "type": "object",
      "properties": {
        "smtp_host": {"type": "string"},
        "smtp_port": {"type": "number", "default": 587},
        "username": {"type": "string"},
        "password": {"type": "string"},
        "from_address": {"type": "string", "format": "email"},
        "tls_enabled": {"type": "boolean", "default": true}
      }
    },
    "company_info": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "address": {"type": "string"},
        "phone": {"type": "string"},
        "email": {"type": "string", "format": "email"},
        "website": {"type": "string", "format": "uri"},
        "tax_id": {"type": "string"},
        "logo_path": {"type": "string"}
      }
    },
    "invoice_defaults": {
      "type": "object",
      "properties": {
        "tax_rate": {"type": "number", "minimum": 0, "maximum": 1},
        "payment_terms": {"type": "number", "minimum": 1, "maximum": 365, "default": 30},
        "currency": {"type": "string", "default": "USD"},
        "number_format": {"type": "string", "default": "INV-{YYYY}-{###}"}
      }
    },
    "skip_validation": {
      "type": "boolean",
      "default": false,
      "description": "Skip configuration validation after creation"
    },
    "create_sample_data": {
      "type": "boolean",
      "default": false,
      "description": "Create sample clients and invoices for testing"
    }
  }
}
```

**Output Format**:
```json
{
  "success": true,
  "initialization_summary": {
    "setup_mode": "interactive",
    "template_used": "production",
    "config_file": "/Users/user/.go-invoice/config.yaml",
    "backup_created": "/Users/user/.go-invoice/config.yaml.backup",
    "initialization_time": "45 seconds"
  },
  "configuration_created": {
    "database": {
      "configured": true,
      "connection_tested": true,
      "schema_created": true
    },
    "email": {
      "configured": true,
      "connection_tested": false,
      "test_email_sent": false
    },
    "company": {
      "configured": true,
      "logo_validated": true
    },
    "invoicing": {
      "configured": true,
      "defaults_set": true,
      "templates_ready": true
    }
  },
  "validation_results": {
    "overall_status": "valid",
    "issues_found": 1,
    "warnings": [
      {
        "section": "email",
        "message": "Email configuration not tested - SMTP connection failed",
        "recommendation": "Verify SMTP settings and test email functionality"
      }
    ]
  },
  "next_steps": [
    "Test email configuration by sending a test invoice",
    "Create your first client using client_create tool",
    "Generate your first invoice using invoice_create tool",
    "Review and customize invoice templates if needed"
  ],
  "sample_data": {
    "created": false,
    "reason": "create_sample_data was false"
  },
  "security_recommendations": [
    "Change default database password",
    "Enable audit logging for production use",
    "Configure rate limiting for API access",
    "Set up regular configuration backups"
  ]
}
```

**Usage Examples**:

*Interactive Setup*:
```json
{
  "setup_mode": "interactive",
  "config_path": "/Users/user/.go-invoice/config.yaml"
}
```

*Production Template Setup*:
```json
{
  "setup_mode": "template",
  "template": "production",
  "config_path": "/etc/go-invoice/config.yaml",
  "database_config": {
    "host": "db.company.com",
    "database": "invoice_prod",
    "username": "invoice_user",
    "ssl_mode": "verify-full"
  },
  "company_info": {
    "name": "Acme Corporation",
    "address": "123 Business St, City, State 12345",
    "email": "invoices@acme.com",
    "phone": "+1-555-123-4567"
  }
}
```

*Development Setup with Sample Data*:
```json
{
  "setup_mode": "template",
  "template": "development",
  "create_sample_data": true,
  "skip_validation": false,
  "invoice_defaults": {
    "tax_rate": 0.08,
    "payment_terms": 15,
    "currency": "USD"
  }
}
```

**Error Conditions**:
- `CONFIG_PATH_INVALID` (400): Configuration path is invalid or not writable
- `FILE_EXISTS_ERROR` (409): Configuration file exists and overwrite_existing is false
- `DATABASE_CONNECTION_FAILED` (503): Cannot connect to specified database
- `EMAIL_CONFIG_INVALID` (400): Email configuration is invalid or incomplete
- `TEMPLATE_NOT_FOUND` (404): Specified template does not exist
- `INITIALIZATION_TIMEOUT` (408): Setup process exceeded timeout

**Security Considerations**:
- Secure handling of database and email credentials
- Configuration file permissions set appropriately
- Backup creation before overwriting existing configurations
- Validation of external connection security settings
- Audit logging for configuration initialization

**Performance Notes**:
- Efficient template loading and processing
- Parallel validation of configuration sections
- Optimized database schema creation for large installations
- Progress tracking for long initialization processes

**Related Tools**: `config_show`, `config_validate`

---

## Error Handling and Troubleshooting

### Common Error Categories

#### 1. Validation Errors (400 series)
- **CLIENT_NOT_FOUND**: Client specified does not exist
- **INVALID_DATE_FORMAT**: Date parameter not in YYYY-MM-DD format
- **VALIDATION_ERROR**: Input parameters fail schema validation
- **REQUIRED_FIELD_MISSING**: Required parameter not provided

#### 2. Permission Errors (403 series)
- **ACCESS_DENIED**: Insufficient permissions for operation
- **INVOICE_NOT_EDITABLE**: Cannot modify paid/sent invoices
- **CLIENT_ACCESS_RESTRICTED**: Cannot access client data

#### 3. Not Found Errors (404 series)
- **INVOICE_NOT_FOUND**: Invoice with specified ID/number not found
- **WORK_ITEM_NOT_FOUND**: Work item does not exist
- **FILE_NOT_FOUND**: Specified file path does not exist

#### 4. Conflict Errors (409 series)
- **EMAIL_ALREADY_EXISTS**: Client email address already in use
- **DUPLICATE_INVOICE_NUMBER**: Invoice number already exists
- **INVALID_STATUS_TRANSITION**: Business rules prevent status change

#### 5. Server Errors (500 series)
- **DATABASE_CONNECTION_FAILED**: Cannot connect to database
- **FILE_OPERATION_FAILED**: File read/write operation failed
- **EXPORT_GENERATION_FAILED**: Document generation failed

### Error Response Format

All errors follow a consistent JSON structure:

```json
{
  "success": false,
  "error": {
    "code": "INVOICE_NOT_FOUND",
    "message": "Invoice with number 'INV-999' was not found",
    "details": {
      "requested_invoice": "INV-999",
      "available_suggestions": ["INV-001", "INV-002", "INV-003"]
    },
    "timestamp": "2025-08-03T16:45:00Z",
    "request_id": "req_abc123"
  },
  "suggestions": [
    "Check the invoice number spelling",
    "Use invoice_list to see available invoices",
    "Try searching by client name instead"
  ]
}
```

### Troubleshooting Guide

#### Performance Issues
- Check database connection and indexing
- Monitor memory usage during large operations
- Use pagination for large result sets
- Consider splitting large imports into smaller batches

#### Connection Problems
- Verify database credentials and connectivity
- Check email server settings and firewall rules
- Validate file paths and permissions
- Test network connectivity to external services

#### Validation Failures
- Review input parameter formats and types
- Check required field completeness
- Validate date formats and ranges
- Ensure business rule compliance

## Security and Compliance

### Data Protection
- All financial data encrypted at rest and in transit
- Client information protected with access controls
- Audit trails maintained for all operations
- Regular security updates and vulnerability assessments

### Access Control
- Role-based permissions for different operations
- Client data isolation and access restrictions
- API rate limiting to prevent abuse
- Session management and timeout controls

### Compliance Features
- Audit logging for financial operations
- Data retention policies for legal requirements
- Export capabilities for tax reporting
- Backup and recovery procedures

### Best Practices
- Regular configuration validation
- Secure credential management
- Network security configuration
- Regular backup verification

## Performance Optimization

### Database Optimization
- Proper indexing on frequently queried fields
- Connection pooling for improved performance
- Query optimization for large datasets
- Regular database maintenance and cleanup

### Memory Management
- Streaming processing for large files
- Efficient pagination for large result sets
- Memory-efficient template rendering
- Garbage collection optimization

### Network Optimization
- Response compression for large datasets
- Caching of frequently accessed data
- Parallel processing where appropriate
- Timeout configuration for responsiveness

### File Operations
- Efficient CSV parsing and validation
- Optimized document generation
- Batch processing for multiple operations
- Temporary file cleanup procedures

---

## Appendices

### A. JSON Schema References
All tool schemas follow JSON Schema Draft 7 specification and include:
- Comprehensive validation rules
- Clear error messages
- Type definitions and constraints
- Example values and use cases

### B. CLI Command Mapping
Each MCP tool maps to specific go-invoice CLI commands:
- Argument translation and validation
- Output format standardization
- Error code mapping
- Timeout and resource management

### C. Integration Examples
Sample integration patterns for:
- Claude Desktop configuration
- Automation workflows
- Batch processing scripts
- External system integration

### D. Migration and Upgrade
Guidelines for:
- Tool version compatibility
- Configuration migration
- Data format updates
- Breaking change handling

---

**Document Version**: 1.0.0  
**Last Updated**: August 3, 2025  
**Go-Invoice MCP Server**: v1.0.0  
**Schema Standard**: JSON Schema Draft 7  

For additional support and examples, see the complete documentation suite in the `docs/mcp/` directory.