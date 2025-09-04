# Error Recovery - Handling Problems and Edge Cases

This example demonstrates comprehensive error handling, data recovery, and problem-solving workflows. This scenario showcases how **Robert**, a service provider, encounters various problems and uses the MCP tools to diagnose, fix, and prevent future issues.

## Scenario Overview

**Robert** runs a professional services firm and encounters various challenging situations that require problem-solving and recovery. He needs to:
1. Handle corrupted or invalid data imports
2. Recover from accidentally deleted or modified records
3. Resolve client billing disputes and invoice corrections
4. Fix system configuration issues
5. Handle edge cases in payment processing
6. Implement preventive measures to avoid future problems

This workflow demonstrates **15 different MCP tools** with focus on error handling, validation, and recovery procedures.

---

## The Complete Conversation

### Part 1: Data Import Problems and Recovery

**Robert**: "Hi Claude! I'm having a serious problem. I just tried to import a large timesheet CSV file and something went wrong. The import seemed to complete, but now I have duplicate entries and some of the data looks corrupted. Can you help me figure out what happened and fix it?"

**Claude**: Oh no! Data import issues can be stressful, but don't worry - we can diagnose and fix this. Let me start by checking what actually got imported and then we'll develop a recovery plan. First, let me examine your recent import activity.

*Using invoice_list tool:*
```json
{
  "tool": "invoice_list",
  "parameters": {
    "from_date": "2025-08-03",
    "to_date": "2025-08-03",
    "include_import_metadata": true,
    "show_data_quality_flags": true,
    "output_format": "detailed"
  }
}
```

I can see there was an import today that created invoice INV-015 with some data quality flags. Let me get the detailed information:

*Using invoice_show tool:*
```json
{
  "tool": "invoice_show",
  "parameters": {
    "invoice_number": "INV-015",
    "show_work_items": true,
    "show_import_details": true,
    "flag_potential_issues": true,
    "output_format": "diagnostic"
  }
}
```

**Diagnosis Results:**
- **Invoice**: INV-015 for ProjectCorp
- **Import Issues Found**:
  - 12 duplicate time entries (same date/description)
  - 3 entries with invalid date formats
  - 5 entries with missing descriptions
  - Total discrepancy: $2,400 in duplicated charges

Let me validate the original CSV file to understand what went wrong:

*Using import_validate tool:*
```json
{
  "tool": "import_validate",
  "parameters": {
    "file_path": "/data/timesheets/projectcorp-july-final.csv",
    "check_for_duplicates": true,
    "validate_data_integrity": true,
    "check_encoding_issues": true,
    "detailed_error_report": true
  }
}
```

**Root Cause Identified:**
- **CSV File Issues**: The file has UTF-8 encoding problems causing duplicate header detection
- **Data Problems**: 12 actual duplicate rows in the source file
- **Import Logic**: System imported all rows without deduplication

**Recovery Plan:**
1. Remove the corrupted invoice
2. Clean the CSV file
3. Re-import with proper validation
4. Verify results

Let me start the recovery process:

*Using invoice_delete tool:*
```json
{
  "tool": "invoice_delete",
  "parameters": {
    "invoice_number": "INV-015",
    "reason": "Data corruption during import - duplicate entries detected",
    "confirm": true
  }
}
```

Now let me preview the corrected import to make sure we handle the duplicates properly:

*Using import_preview tool:*
```json
{
  "tool": "import_preview",
  "parameters": {
    "file_path": "/data/timesheets/projectcorp-july-final.csv",
    "deduplicate_rows": true,
    "preview_rows": 15,
    "show_deduplication_summary": true
  }
}
```

Perfect! The preview shows:
- **Original rows**: 87
- **After deduplication**: 75 unique entries
- **Duplicates removed**: 12 rows
- **Data integrity**: All entries now valid

Now let's re-import correctly:

*Using import_csv tool:*
```json
{
  "tool": "import_csv",
  "parameters": {
    "file_path": "/data/timesheets/projectcorp-july-final.csv",
    "client_name": "ProjectCorp",
    "description": "July 2025 Development Services - CORRECTED IMPORT",
    "import_mode": "new_invoice",
    "deduplicate_entries": true,
    "validate_before_import": true,
    "due_days": 30
  }
}
```

Excellent! Recovery complete:
- **New Invoice**: INV-016 (replacing corrupted INV-015)
- **Total Amount**: $11,250.00 (correct amount, $2,400 less than corrupted version)
- **Work Entries**: 75 valid entries (12 duplicates removed)
- **Data Quality**: 100% clean

The data is now correctly imported and you've avoided overbilling your client by $2,400!

### Part 2: Accidental Deletion Recovery

**Robert**: "Thank you for fixing that! Now I have another problem - I think I accidentally deleted an important client record yesterday. The client is GlobalManufacturing Inc, and I can't find them anywhere. I'm supposed to invoice them tomorrow. Is there any way to recover their information?"

**Claude**: Let me help you search for GlobalManufacturing Inc and see what happened. Accidental deletions can often be recovered depending on how the system is configured.

*Using client_list tool:*
```json
{
  "tool": "client_list",
  "parameters": {
    "search_term": "GlobalManufacturing",
    "include_deleted": true,
    "show_audit_trail": true,
    "output_format": "detailed"
  }
}
```

Good news! I found GlobalManufacturing Inc in the deleted records:
- **Status**: Soft deleted on August 2, 2025
- **Original Data**: Complete client record with contact information preserved
- **Invoice History**: 6 previous invoices totaling $45,600

Since it was a soft delete, all the data is recoverable. Let me restore the client record:

*Using client_update tool:*
```json
{
  "tool": "client_update",
  "parameters": {
    "client_id": "client_gmi_001",
    "restore_from_deleted": true,
    "update_status": "active",
    "add_note": "Restored on 2025-08-03 - accidental deletion recovery"
  }
}
```

Perfect! Let me verify the restoration worked correctly:

*Using client_show tool:*
```json
{
  "tool": "client_show",
  "parameters": {
    "client_name": "GlobalManufacturing Inc",
    "show_invoice_history": true,
    "show_contact_details": true,
    "verify_data_integrity": true
  }
}
```

**Recovery Successful!**
- **Client**: GlobalManufacturing Inc - RESTORED ✅
- **Contact**: purchasing@globalmanufacturing.com
- **Address**: 1500 Industrial Parkway, Detroit, MI 48201
- **History**: All 6 previous invoices intact
- **Status**: Active and ready for new invoicing

Your client is fully restored and you can proceed with tomorrow's invoice as planned!

### Part 3: Invoice Correction and Billing Disputes

**Robert**: "Fantastic! Now I have a billing dispute to resolve. TechStartup LLC claims that invoice INV-012 has incorrect hours - they say I billed them for work that was done for a different client. They have documentation showing the discrepancy. How can I investigate and fix this?"

**Claude**: Billing disputes require careful investigation and transparent resolution. Let me examine invoice INV-012 in detail and help you resolve this professionally.

*Using invoice_show tool:*
```json
{
  "tool": "invoice_show",
  "parameters": {
    "invoice_number": "INV-012",
    "show_work_items": true,
    "show_detailed_breakdown": true,
    "include_audit_trail": true,
    "output_format": "audit"
  }
}
```

**Invoice INV-012 Analysis:**
- **Client**: TechStartup LLC
- **Total**: $8,750.00
- **Work Period**: July 22-28, 2025
- **Total Hours**: 35 hours

Let me examine the specific work items that might be disputed:

Looking at the breakdown, I see some entries on July 25-26 that have generic descriptions like "API development" and "database optimization." These could potentially be confused with work for other clients.

Let me check what other invoices you generated around the same time period to see if there might have been a mix-up:

*Using invoice_list tool:*
```json
{
  "tool": "invoice_list",
  "parameters": {
    "from_date": "2025-07-20",
    "to_date": "2025-07-30",
    "exclude_invoice": "INV-012",
    "show_work_details": true,
    "search_description": "API development OR database optimization"
  }
}
```

**Investigation Results:**
I found the issue! Invoice INV-011 for "DataCorp Solutions" also has "API development" and "database optimization" entries for the exact same dates (July 25-26). It appears there was a data entry error where some work items got duplicated across both invoices.

**Specific Problem:**
- **July 25**: 4 hours "API development" appears on both INV-011 and INV-012
- **July 26**: 3 hours "database optimization" appears on both invoices
- **Total Overcharge**: 7 hours × $250/hour = $1,750

Let me create a corrected invoice for TechStartup LLC:

*Using invoice_remove_item tool:*
```json
{
  "tool": "invoice_remove_item",
  "parameters": {
    "invoice_number": "INV-012",
    "work_item_date": "2025-07-25",
    "work_item_description": "API development",
    "confirm": true,
    "reason": "Billing dispute resolution - work was for DataCorp Solutions (INV-011)"
  }
}
```

*Using invoice_remove_item tool:*
```json
{
  "tool": "invoice_remove_item",
  "parameters": {
    "invoice_number": "INV-012",
    "work_item_date": "2025-07-26",
    "work_item_description": "database optimization",
    "confirm": true,
    "reason": "Billing dispute resolution - work was for DataCorp Solutions (INV-011)"
  }
}
```

Now let me update the invoice with a note about the correction:

*Using invoice_update tool:*
```json
{
  "tool": "invoice_update",
  "parameters": {
    "invoice_number": "INV-012",
    "description": "TechStartup LLC - July 2025 Services - CORRECTED (removed duplicate entries)",
    "add_note": "Invoice corrected on 2025-08-03: Removed 7 hours incorrectly billed (work was for DataCorp Solutions). Original total: $8,750, Corrected total: $7,000. Credit memo issued for $1,750."
  }
}
```

**Dispute Resolution Complete:**
- **Original Invoice**: INV-012 - $8,750.00
- **Corrected Invoice**: INV-012 - $7,000.00
- **Adjustment**: -$1,750.00 credit
- **Resolution**: Professional acknowledgment of error with transparent correction

Let me generate a formal credit memo:

*Using invoice_create tool:*
```json
{
  "tool": "invoice_create",
  "parameters": {
    "client_name": "TechStartup LLC",
    "description": "Credit Memo - Billing Error Correction for INV-012",
    "invoice_date": "2025-08-03",
    "work_items": [
      {
        "date": "2025-08-03",
        "hours": -7.0,
        "rate": 250.0,
        "description": "Credit for duplicate billing on INV-012 (July 25-26 API/database work belonged to DataCorp Solutions)"
      }
    ]
  }
}
```

Perfect! Credit memo INV-017 created for -$1,750.00. Your professional handling of this dispute will actually strengthen the client relationship by demonstrating integrity and transparency.

### Part 4: System Configuration Issues

**Robert**: "This professional dispute resolution is exactly what I needed! Now I'm noticing some configuration issues. My invoices are generating with the wrong tax settings, and I think my default rates might be outdated. Can you help me review and fix my system configuration?"

**Claude**: Absolutely! Configuration issues can cause ongoing problems if not addressed properly. Let me examine your current system settings and identify any issues.

*Using config_show tool:*
```json
{
  "tool": "config_show",
  "parameters": {
    "show_all_sections": true,
    "include_validation_status": true,
    "flag_potential_issues": true,
    "output_format": "diagnostic"
  }
}
```

**Configuration Issues Identified:**

**Tax Settings:**
- **Sales Tax Rate**: 8.5% (should be 7.25% for Texas)
- **Tax-Exempt Clients**: Configuration missing for non-profits
- **Multi-State Handling**: Not configured for out-of-state clients

**Default Rates:**
- **Standard Rate**: $200/hour (last updated 6 months ago)
- **Premium Rate**: $275/hour (may be below market)
- **Client-Specific Overrides**: Only 3 clients configured

**Other Issues:**
- **Payment Terms**: Default NET-30 for all clients (some need NET-15 or NET-45)
- **Invoice Numbering**: Running low on current sequence
- **Backup Settings**: Last backup 15 days ago (recommended: daily)

Let me validate these settings against best practices:

*Using config_validate tool:*
```json
{
  "tool": "config_validate",
  "parameters": {
    "comprehensive_check": true,
    "include_compliance_check": true,
    "check_data_integrity": true,
    "validate_business_rules": true,
    "generate_recommendations": true
  }
}
```

**Validation Results:**
- **Critical Issues**: 2 (tax rate, backup frequency)
- **Warnings**: 5 (rates, payment terms, numbering)
- **Compliance**: Tax settings need immediate correction
- **Data Integrity**: All business data is sound

Let me walk you through fixing these issues. First, let's correct the tax rate:

*Using config_init tool:*
```json
{
  "tool": "config_init",
  "parameters": {
    "update_existing": true,
    "tax_settings": {
      "default_tax_rate": 7.25,
      "tax_jurisdiction": "Texas",
      "tax_exempt_handling": true,
      "multi_state_rules": "destination_based"
    },
    "backup_current_config": true
  }
}
```

Now let's update the default rates based on current market conditions:

*Using config_init tool:*
```json
{
  "tool": "config_init",
  "parameters": {
    "update_existing": true,
    "billing_rates": {
      "standard_rate": 225.0,
      "premium_rate": 300.0,
      "rush_rate": 350.0,
      "non_profit_rate": 175.0
    },
    "payment_terms": {
      "default": 30,
      "corporate": 30,
      "startup": 15,
      "non_profit": 45
    }
  }
}
```

Let me verify the configuration changes:

*Using config_validate tool:*
```json
{
  "tool": "config_validate",
  "parameters": {
    "quick_check": true,
    "focus_areas": ["tax_settings", "billing_rates", "payment_terms"],
    "compare_to_previous": true
  }
}
```

**Configuration Update Complete:**
- **Tax Rate**: Corrected to 7.25% ✅
- **Default Rates**: Updated to market rates (12.5% increase) ✅
- **Payment Terms**: Client-type specific terms configured ✅
- **Compliance**: All critical issues resolved ✅

**Impact on Future Invoices:**
- Correct tax calculations going forward
- Market-appropriate billing rates
- Appropriate payment terms by client type
- Improved compliance and accuracy

### Part 5: Preventive Measures and System Health

**Robert**: "Excellent! All these fixes give me much more confidence in the system. Can you help me set up some preventive measures so I can avoid these kinds of problems in the future?"

**Claude**: Absolutely! Prevention is always better than recovery. Let me help you establish monitoring, validation, and backup procedures to prevent future issues.

First, let me create a comprehensive system health check that you can run regularly:

*Using config_validate tool:*
```json
{
  "tool": "config_validate",
  "parameters": {
    "create_health_check_template": true,
    "template_name": "monthly_system_health",
    "check_categories": [
      "data_integrity",
      "configuration_compliance",
      "backup_status",
      "performance_metrics",
      "security_validation"
    ],
    "automated_scheduling": "monthly",
    "alert_thresholds": "conservative"
  }
}
```

Now let me set up data validation rules for imports:

*Using import_validate tool:*
```json
{
  "tool": "import_validate",
  "parameters": {
    "create_validation_rules": true,
    "rule_set_name": "standard_import_validation",
    "rules": {
      "require_deduplication": true,
      "max_duplicate_threshold": 2,
      "require_date_validation": true,
      "require_description_minimum": 10,
      "flag_unusual_hours": true,
      "validate_client_existence": true
    },
    "auto_apply_to_imports": true
  }
}
```

Let me also create a backup and recovery protocol:

*Using export_data tool:*
```json
{
  "tool": "export_data",
  "parameters": {
    "create_backup_schedule": true,
    "backup_frequency": "daily",
    "backup_types": ["full_system", "incremental_data"],
    "retention_policy": "30_days_full, 90_days_incremental",
    "backup_location": "./backups/automated/",
    "include_configuration": true,
    "verify_backup_integrity": true
  }
}
```

Finally, let me establish monitoring for critical business metrics:

*Using generate_summary tool:*
```json
{
  "tool": "generate_summary",
  "parameters": {
    "create_monitoring_dashboard": true,
    "dashboard_name": "business_health_monitor",
    "key_indicators": [
      "data_quality_score",
      "invoice_accuracy_rate",
      "client_payment_patterns",
      "system_performance_metrics",
      "backup_success_rate"
    ],
    "alert_conditions": {
      "data_quality_below": 95,
      "invoice_disputes_above": 2,
      "backup_failures": 1,
      "unusual_payment_delays": "15_days_over_normal"
    },
    "reporting_frequency": "weekly"
  }
}
```

**Preventive System Established:**

**1. Automated Health Checks:**
- Monthly comprehensive system validation
- Data integrity monitoring
- Configuration compliance verification
- Performance tracking

**2. Import Protection:**
- Automatic deduplication for all CSV imports
- Pre-import validation rules
- Unusual data pattern detection
- Client existence verification

**3. Backup & Recovery:**
- Daily automated backups
- 30-day retention for full backups
- Automated integrity verification
- Quick recovery procedures documented

**4. Business Monitoring:**
- Weekly health dashboard reports
- Automated alerts for critical issues
- Payment pattern anomaly detection
- Client relationship health tracking

**5. Error Prevention Checklist:**
- ✅ Validate all CSV files before import
- ✅ Review invoices before sending (especially after imports)
- ✅ Confirm client information changes
- ✅ Regular system health checks
- ✅ Backup verification routine

Now let me generate a comprehensive incident response playbook:

*Using export_data tool:*
```json
{
  "tool": "export_data",
  "parameters": {
    "create_document": "incident_response_playbook",
    "include_scenarios": [
      "data_corruption_recovery",
      "accidental_deletion_procedures",
      "billing_dispute_resolution",
      "configuration_issue_diagnosis",
      "system_performance_problems"
    ],
    "include_step_by_step_procedures": true,
    "include_contact_information": true,
    "export_path": "./documentation/Incident-Response-Playbook.pdf"
  }
}
```

**You're now protected against future issues with:**
- Comprehensive monitoring and alerting
- Automated backup and recovery systems
- Preventive validation for all data operations
- Clear procedures for common problems
- Regular health check routines

This systematic approach will help you catch problems early and resolve them quickly when they do occur!

---

## Tools Demonstrated

This workflow showcased **15 different MCP tools**:

1. **invoice_list** - Diagnosing recent import issues and data problems
2. **invoice_show** - Detailed investigation of corrupted or disputed invoices
3. **import_validate** - Identifying root causes of CSV import problems
4. **invoice_delete** - Removing corrupted data safely
5. **import_preview** - Validating corrected data before re-import
6. **import_csv** - Re-importing data with proper validation and deduplication
7. **client_list** - Finding deleted client records and audit trails
8. **client_update** - Restoring accidentally deleted client information
9. **client_show** - Verifying successful data recovery
10. **invoice_remove_item** - Correcting billing disputes by removing incorrect entries
11. **invoice_update** - Adding correction notes and updating invoice status
12. **invoice_create** - Generating credit memos for billing corrections
13. **config_show** - Diagnosing system configuration issues
14. **config_validate** - Comprehensive system health checking
15. **config_init** - Correcting configuration problems and establishing preventive measures

## Key Error Recovery Benefits

- **Rapid Problem Diagnosis**: Quickly identify root causes of data and system issues
- **Safe Data Recovery**: Restore deleted or corrupted information without data loss
- **Professional Dispute Resolution**: Handle billing disputes transparently and maintain client trust
- **System Health Monitoring**: Prevent problems before they impact business operations
- **Automated Protection**: Built-in safeguards for common error scenarios
- **Documentation and Procedures**: Clear playbooks for handling various problem types

## Problem Prevention Strategies

- **Data Validation**: Multi-layer checking prevents bad data from entering the system
- **Backup Systems**: Multiple backup strategies ensure data recovery is always possible
- **Configuration Management**: Regular validation prevents system drift and compliance issues
- **Monitoring and Alerting**: Early warning systems catch problems before they become critical
- **Process Standardization**: Consistent procedures reduce human error opportunities
- **Audit Trails**: Complete tracking enables rapid problem investigation and resolution

## Business Continuity Advantages

This workflow demonstrates how comprehensive error handling protects business operations:

- **Client Relationship Protection**: Professional handling of disputes strengthens rather than damages relationships
- **Financial Accuracy**: Systematic correction procedures ensure billing integrity
- **Operational Resilience**: Quick recovery from problems minimizes business disruption
- **Compliance Assurance**: Regular validation maintains regulatory and tax compliance
- **Risk Mitigation**: Preventive measures reduce the likelihood and impact of future problems
- **Professional Image**: Systematic problem resolution demonstrates business maturity and reliability

The go-invoice MCP integration transforms error handling from reactive firefighting into proactive system management that protects and enhances business operations.
