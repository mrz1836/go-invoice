#!/bin/bash

# Monthly Billing Automation Script for go-invoice
# This script automates the monthly billing process for multiple clients
# Usage: ./monthly-billing.sh [month] [year]
# Example: ./monthly-billing.sh 01 2024

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
TIMESHEETS_DIR="${TIMESHEETS_DIR:-$PROJECT_ROOT/data/timesheets}"
INVOICES_DIR="${INVOICES_DIR:-$PROJECT_ROOT/data/invoices}"
CLIENTS_FILE="${CLIENTS_FILE:-$PROJECT_ROOT/data/clients.txt}"
LOG_FILE="${LOG_FILE:-$PROJECT_ROOT/logs/monthly-billing.log}"

# Default to current month/year if not provided
MONTH="${1:-$(date +%m)}"
YEAR="${2:-$(date +%Y)}"
MONTH_NAME=$(date -d "${YEAR}-${MONTH}-01" +%B)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    # Create log directory if it doesn't exist
    mkdir -p "$(dirname "$LOG_FILE")"

    # Log to file
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"

    # Also output to console with colors
    case "$level" in
        "ERROR")
            echo -e "${RED}❌ $message${NC}" >&2
            ;;
        "SUCCESS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  $message${NC}"
            ;;
        *)
            echo "$message"
            ;;
    esac
}

# Function to check prerequisites
check_prerequisites() {
    log "INFO" "Checking prerequisites..."

    # Check if go-invoice is installed
    if ! command -v go-invoice &> /dev/null; then
        log "ERROR" "go-invoice is not installed or not in PATH"
        exit 1
    fi

    # Check if jq is installed (for JSON processing)
    if ! command -v jq &> /dev/null; then
        log "WARNING" "jq is not installed. Some features may not work properly"
        log "INFO" "Install jq with: sudo apt-get install jq (Ubuntu) or brew install jq (macOS)"
    fi

    # Create necessary directories
    mkdir -p "$TIMESHEETS_DIR/$YEAR/$MONTH-$MONTH_NAME"
    mkdir -p "$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME"
    mkdir -p "$(dirname "$LOG_FILE")"

    log "SUCCESS" "Prerequisites check completed"
}

# Function to validate client file
validate_clients_file() {
    if [[ ! -f "$CLIENTS_FILE" ]]; then
        log "ERROR" "Clients file not found: $CLIENTS_FILE"
        log "INFO" "Create a clients file with one client name per line:"
        log "INFO" "  echo 'Client Company Inc' >> $CLIENTS_FILE"
        exit 1
    fi

    local client_count=$(grep -v '^#' "$CLIENTS_FILE" | grep -v '^$' | wc -l)
    log "INFO" "Found $client_count active clients in $CLIENTS_FILE"

    if [[ $client_count -eq 0 ]]; then
        log "ERROR" "No active clients found in $CLIENTS_FILE"
        exit 1
    fi
}

# Function to process a single client
process_client() {
    local client="$1"
    local client_slug="${client// /-}"
    local timesheet_file="$TIMESHEETS_DIR/$YEAR/$MONTH-$MONTH_NAME/${client_slug}.csv"
    local invoice_file="$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/${client_slug}.html"

    log "INFO" "Processing client: $client"

    # Check if timesheet exists
    if [[ ! -f "$timesheet_file" ]]; then
        log "WARNING" "Timesheet not found for $client: $timesheet_file"
        log "INFO" "Skipping $client - create timesheet file first"
        echo "$client: SKIPPED (no timesheet)" >> "$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/summary.txt"
        return 0
    fi

    # Validate timesheet format
    log "INFO" "Validating timesheet for $client..."
    if ! go-invoice import csv "$timesheet_file" --client "$client" --validate --dry-run > /dev/null 2>&1; then
        log "ERROR" "Invalid timesheet format for $client"
        log "INFO" "Run: go-invoice import csv '$timesheet_file' --client '$client' --validate --dry-run"
        echo "$client: FAILED (invalid timesheet)" >> "$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/summary.txt"
        return 1
    fi

    # Import timesheet
    log "INFO" "Importing timesheet for $client..."
    if ! go-invoice import csv "$timesheet_file" --client "$client" --validate; then
        log "ERROR" "Failed to import timesheet for $client"
        echo "$client: FAILED (import error)" >> "$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/summary.txt"
        return 1
    fi

    # Create invoice
    log "INFO" "Creating invoice for $client..."
    local invoice_description="$MONTH_NAME $YEAR Services"
    if ! go-invoice invoice create \
        --client "$client" \
        --description "$invoice_description" \
        --output "$invoice_file"; then
        log "ERROR" "Failed to create invoice for $client"
        echo "$client: FAILED (invoice creation)" >> "$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/summary.txt"
        return 1
    fi

    # Get the latest invoice ID for this client
    local invoice_id
    if command -v jq &> /dev/null; then
        invoice_id=$(go-invoice invoice list --client "$client" --status draft --format json 2>/dev/null | \
                    jq -r '.[0].id // empty' 2>/dev/null || echo "")
    else
        # Fallback without jq - get the first draft invoice ID from text output
        invoice_id=$(go-invoice invoice list --client "$client" --status draft 2>/dev/null | \
                    grep -E '^[A-Z0-9-]+' | head -n1 | awk '{print $1}' || echo "")
    fi

    if [[ -n "$invoice_id" ]]; then
        # Send invoice
        log "INFO" "Sending invoice $invoice_id for $client..."
        if go-invoice invoice send --invoice "$invoice_id"; then
            log "SUCCESS" "Invoice $invoice_id sent for $client"
            echo "$client: SUCCESS ($invoice_id)" >> "$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/summary.txt"
        else
            log "WARNING" "Failed to send invoice $invoice_id for $client"
            echo "$client: PARTIAL SUCCESS ($invoice_id created but not sent)" >> "$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/summary.txt"
        fi
    else
        log "WARNING" "Could not find invoice ID for $client, but invoice was created"
        echo "$client: PARTIAL SUCCESS (invoice created)" >> "$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/summary.txt"
    fi

    log "SUCCESS" "Completed processing for $client"
    return 0
}

# Function to generate summary report
generate_summary() {
    local summary_file="$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/summary.txt"
    local report_file="$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/monthly-report.html"

    log "INFO" "Generating monthly summary report..."

    # Create summary header
    {
        echo "# Monthly Billing Summary - $MONTH_NAME $YEAR"
        echo "# Generated on: $(date)"
        echo "# Log file: $LOG_FILE"
        echo ""
        echo "## Processing Results:"
    } > "$summary_file.header"

    # Combine header with results
    if [[ -f "$summary_file" ]]; then
        cat "$summary_file.header" "$summary_file" > "$summary_file.tmp"
        mv "$summary_file.tmp" "$summary_file"
    else
        mv "$summary_file.header" "$summary_file"
    fi

    # Generate detailed HTML report if go-invoice supports it
    if go-invoice report summary --month "$YEAR-$MONTH" --output "$report_file" 2>/dev/null; then
        log "SUCCESS" "Detailed HTML report generated: $report_file"
    else
        log "INFO" "HTML report generation not supported or failed"
    fi

    log "SUCCESS" "Summary report created: $summary_file"
}

# Function to backup data
backup_data() {
    local backup_dir="${BACKUP_DIR:-$PROJECT_ROOT/backups}"
    local backup_name="monthly-billing-$YEAR-$MONTH-$(date +%Y%m%d_%H%M%S)"
    local backup_path="$backup_dir/$backup_name"

    if [[ "${ENABLE_BACKUP:-true}" == "true" ]]; then
        log "INFO" "Creating backup..."
        mkdir -p "$backup_path"

        # Backup invoice data
        if [[ -d "$PROJECT_ROOT/data" ]]; then
            cp -r "$PROJECT_ROOT/data" "$backup_path/"
        fi

        # Backup logs
        if [[ -f "$LOG_FILE" ]]; then
            cp "$LOG_FILE" "$backup_path/"
        fi

        # Create backup info
        {
            echo "Backup created: $(date)"
            echo "Month: $MONTH_NAME $YEAR"
            echo "Script version: $(grep '^# Version:' "$0" || echo 'Unknown')"
        } > "$backup_path/backup-info.txt"

        log "SUCCESS" "Backup created: $backup_path"
    else
        log "INFO" "Backup disabled via ENABLE_BACKUP=false"
    fi
}

# Function to display usage
usage() {
    echo "Usage: $0 [month] [year]"
    echo ""
    echo "Monthly billing automation for go-invoice"
    echo ""
    echo "Arguments:"
    echo "  month    Month number (01-12), defaults to current month"
    echo "  year     Year (YYYY), defaults to current year"
    echo ""
    echo "Environment Variables:"
    echo "  TIMESHEETS_DIR    Directory containing timesheet CSV files (default: data/timesheets)"
    echo "  INVOICES_DIR      Directory for generated invoices (default: data/invoices)"
    echo "  CLIENTS_FILE      File containing client names (default: data/clients.txt)"
    echo "  LOG_FILE          Log file path (default: logs/monthly-billing.log)"
    echo "  BACKUP_DIR        Backup directory (default: backups)"
    echo "  ENABLE_BACKUP     Enable/disable backup (default: true)"
    echo ""
    echo "Example:"
    echo "  $0 01 2024    # Process January 2024"
    echo "  $0            # Process current month"
    echo ""
    echo "Setup:"
    echo "  1. Create clients file: echo 'Client Name' >> data/clients.txt"
    echo "  2. Add timesheet files: data/timesheets/YYYY/MM-MonthName/Client-Name.csv"
    echo "  3. Run this script: $0"
}

# Main execution
main() {
    local start_time=$(date +%s)

    # Handle help flag
    if [[ "$1" == "-h" || "$1" == "--help" ]]; then
        usage
        exit 0
    fi

    log "INFO" "Starting monthly billing automation for $MONTH_NAME $YEAR"
    log "INFO" "Script: $0"
    log "INFO" "Arguments: $*"

    # Check prerequisites
    check_prerequisites

    # Validate clients file
    validate_clients_file

    # Initialize summary file
    rm -f "$INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/summary.txt"

    # Process each client
    local total_clients=0
    local successful_clients=0
    local failed_clients=0

    while IFS= read -r client || [[ -n "$client" ]]; do
        # Skip empty lines and comments
        [[ -z "$client" || "$client" =~ ^[[:space:]]*# ]] && continue

        total_clients=$((total_clients + 1))

        if process_client "$client"; then
            successful_clients=$((successful_clients + 1))
        else
            failed_clients=$((failed_clients + 1))
        fi

        echo ""  # Add spacing between clients

    done < "$CLIENTS_FILE"

    # Generate summary
    generate_summary

    # Backup data
    backup_data

    # Final statistics
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    log "INFO" "Monthly billing automation completed"
    log "INFO" "Processed: $total_clients clients"
    log "SUCCESS" "Successful: $successful_clients clients"
    if [[ $failed_clients -gt 0 ]]; then
        log "ERROR" "Failed: $failed_clients clients"
    fi
    log "INFO" "Duration: ${duration}s"
    log "INFO" "Check detailed results in: $INVOICES_DIR/$YEAR/$MONTH-$MONTH_NAME/"

    # Exit with error code if any clients failed
    if [[ $failed_clients -gt 0 ]]; then
        exit 1
    fi
}

# Run main function with all arguments
main "$@"
