#!/bin/bash

# Client Setup Script for go-invoice
# This script helps set up a new client with directory structure and sample files
# Usage: ./setup-client.sh "Client Name" [email] [rate]

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
DATA_DIR="${DATA_DIR:-$PROJECT_ROOT/data}"
TEMPLATES_DIR="${TEMPLATES_DIR:-$PROJECT_ROOT/examples/templates}"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Function to display usage
usage() {
    echo "Usage: $0 \"Client Name\" [email] [hourly_rate]"
    echo ""
    echo "Set up a new client with directory structure and sample files"
    echo ""
    echo "Arguments:"
    echo "  client_name   Client company name (required, use quotes for spaces)"
    echo "  email         Client email address (optional)"
    echo "  hourly_rate   Default hourly rate (optional, default: 100.00)"
    echo ""
    echo "Examples:"
    echo "  $0 \"TechCorp Solutions\""
    echo "  $0 \"Startup Inc\" \"billing@startup.com\" \"125.00\""
    echo ""
    echo "This script will:"
    echo "  • Add the client to go-invoice"
    echo "  • Create directory structure for timesheets and invoices"
    echo "  • Generate sample timesheet template"
    echo "  • Add client to clients.txt file"
    echo "  • Create quick-start commands script"
}

# Function to create directory structure
create_directories() {
    local client_slug="$1"
    local current_year=$(date +%Y)
    local current_month=$(date +%m)
    local month_name=$(date +%B)
    
    echo -e "${BLUE}📁 Creating directory structure...${NC}"
    
    # Create base directories
    mkdir -p "$DATA_DIR/timesheets/$current_year/$current_month-$month_name"
    mkdir -p "$DATA_DIR/invoices/$current_year/$current_month-$month_name"
    mkdir -p "$DATA_DIR/templates"
    
    echo -e "${GREEN}✅ Created directories for $current_year/$current_month-$month_name${NC}"
}

# Function to add client to go-invoice
add_client_to_go_invoice() {
    local client_name="$1"
    local client_email="$2"
    
    echo -e "${BLUE}👤 Adding client to go-invoice...${NC}"
    
    local cmd="go-invoice client add --name \"$client_name\""
    if [[ -n "$client_email" ]]; then
        cmd="$cmd --email \"$client_email\""
    fi
    
    if eval "$cmd"; then
        echo -e "${GREEN}✅ Client added successfully${NC}"
    else
        echo -e "${YELLOW}⚠️  Client may already exist or there was an error${NC}"
        echo -e "${BLUE}ℹ️  You can check with: go-invoice client list${NC}"
    fi
}

# Function to create sample timesheet
create_sample_timesheet() {
    local client_name="$1"
    local client_slug="$2"
    local hourly_rate="$3"
    local current_year=$(date +%Y)
    local current_month=$(date +%m)
    local month_name=$(date +%B)
    
    local timesheet_file="$DATA_DIR/timesheets/$current_year/$current_month-$month_name/${client_slug}.csv"
    
    echo -e "${BLUE}📄 Creating sample timesheet...${NC}"
    
    # Generate some sample dates (last 5 working days)
    local dates=()
    local current_date=$(date +%Y-%m-%d)
    local temp_date
    local day_of_week
    local days_back=0
    local working_days=0
    
    while [[ $working_days -lt 5 ]]; do
        temp_date=$(date -d "$current_date -$days_back days" +%Y-%m-%d)
        day_of_week=$(date -d "$temp_date" +%u)
        
        # Only include weekdays (1-5, Monday-Friday)
        if [[ $day_of_week -le 5 ]]; then
            dates+=("$temp_date")
            working_days=$((working_days + 1))
        fi
        
        days_back=$((days_back + 1))
        
        # Safety check to avoid infinite loop
        if [[ $days_back -gt 30 ]]; then
            break
        fi
    done
    
    # Create the timesheet with sample data
    cat > "$timesheet_file" << EOF
date,description,hours,rate
${dates[4]},Requirements analysis and project planning,6.5,$hourly_rate
${dates[3]},Database schema design and implementation,8.0,$hourly_rate
${dates[2]},API development and endpoint testing,7.5,$hourly_rate
${dates[1]},Frontend integration and UI improvements,8.0,$hourly_rate
${dates[0]},Code review and bug fixes,4.0,$hourly_rate
EOF
    
    echo -e "${GREEN}✅ Sample timesheet created: $timesheet_file${NC}"
    echo -e "${BLUE}ℹ️  Edit this file with your actual work hours${NC}"
}

# Function to add client to clients.txt
add_to_clients_file() {
    local client_name="$1"
    local clients_file="$DATA_DIR/clients.txt"
    
    echo -e "${BLUE}📝 Adding client to automation list...${NC}"
    
    # Create clients file if it doesn't exist
    if [[ ! -f "$clients_file" ]]; then
        cat > "$clients_file" << 'EOF'
# go-invoice Clients List
# One client name per line (used by automation scripts)
# Lines starting with # are comments and will be ignored
#
# Example:
# TechCorp Solutions
# Startup Inc
# Marketing Agency LLC

EOF
    fi
    
    # Check if client already exists
    if grep -Fxq "$client_name" "$clients_file"; then
        echo -e "${YELLOW}⚠️  Client already exists in $clients_file${NC}"
    else
        echo "$client_name" >> "$clients_file"
        echo -e "${GREEN}✅ Added client to $clients_file${NC}"
    fi
}

# Function to create quick commands script
create_quick_commands() {
    local client_name="$1"
    local client_slug="$2"
    local current_year=$(date +%Y)
    local current_month=$(date +%m)
    local month_name=$(date +%B)
    
    local commands_file="$DATA_DIR/quick-commands-${client_slug}.sh"
    
    echo -e "${BLUE}⚡ Creating quick commands script...${NC}"
    
    cat > "$commands_file" << EOF
#!/bin/bash

# Quick Commands for $client_name
# Generated on $(date)

CLIENT_NAME="$client_name"
CLIENT_SLUG="$client_slug"
CURRENT_YEAR="$current_year"
CURRENT_MONTH="$current_month"
MONTH_NAME="$month_name"

echo "Quick Commands for \$CLIENT_NAME"
echo "================================="
echo ""

# Set data directory
DATA_DIR="\$(dirname "\$0")"

# Common paths
TIMESHEET_FILE="\$DATA_DIR/timesheets/\$CURRENT_YEAR/\$CURRENT_MONTH-\$MONTH_NAME/\$CLIENT_SLUG.csv"
INVOICE_DIR="\$DATA_DIR/invoices/\$CURRENT_YEAR/\$CURRENT_MONTH-\$MONTH_NAME"

case "\$1" in
    "validate")
        echo "🔍 Validating timesheet..."
        go-invoice import csv "\$TIMESHEET_FILE" --client "\$CLIENT_NAME" --validate --dry-run
        ;;
    "import")
        echo "📥 Importing timesheet..."
        go-invoice import csv "\$TIMESHEET_FILE" --client "\$CLIENT_NAME" --validate
        ;;
    "invoice")
        echo "🧾 Creating invoice..."
        mkdir -p "\$INVOICE_DIR"
        go-invoice invoice create \\
            --client "\$CLIENT_NAME" \\
            --description "\$MONTH_NAME \$CURRENT_YEAR Services" \\
            --output "\$INVOICE_DIR/\$CLIENT_SLUG.html"
        ;;
    "full")
        echo "🚀 Running full workflow (import + invoice)..."
        echo ""
        echo "1. Validating timesheet..."
        go-invoice import csv "\$TIMESHEET_FILE" --client "\$CLIENT_NAME" --validate --dry-run
        echo ""
        echo "2. Importing timesheet..."
        go-invoice import csv "\$TIMESHEET_FILE" --client "\$CLIENT_NAME" --validate
        echo ""
        echo "3. Creating invoice..."
        mkdir -p "\$INVOICE_DIR"
        go-invoice invoice create \\
            --client "\$CLIENT_NAME" \\
            --description "\$MONTH_NAME \$CURRENT_YEAR Services" \\
            --output "\$INVOICE_DIR/\$CLIENT_SLUG.html"
        echo ""
        echo "✅ Workflow completed!"
        echo "📁 Invoice saved to: \$INVOICE_DIR/\$CLIENT_SLUG.html"
        ;;
    "list")
        echo "📋 Listing invoices for \$CLIENT_NAME..."
        go-invoice invoice list --client "\$CLIENT_NAME"
        ;;
    "status")
        echo "📊 Client information:"
        go-invoice client show --name "\$CLIENT_NAME"
        echo ""
        echo "📁 Files:"
        echo "  Timesheet: \$TIMESHEET_FILE"
        if [[ -f "\$TIMESHEET_FILE" ]]; then
            echo "    Status: ✅ Exists (\$(wc -l < "\$TIMESHEET_FILE") lines)"
        else
            echo "    Status: ❌ Not found"
        fi
        echo "  Invoice Dir: \$INVOICE_DIR"
        if [[ -d "\$INVOICE_DIR" ]]; then
            echo "    Status: ✅ Exists (\$(ls "\$INVOICE_DIR"/*.html 2>/dev/null | wc -l) invoices)"
        else
            echo "    Status: ❌ Not found"
        fi
        ;;
    *)
        echo "Usage: \$0 {validate|import|invoice|full|list|status}"
        echo ""
        echo "Commands:"
        echo "  validate  - Validate timesheet format without importing"
        echo "  import    - Import timesheet data"
        echo "  invoice   - Create invoice from imported data"
        echo "  full      - Run complete workflow (validate + import + invoice)"
        echo "  list      - List all invoices for this client"
        echo "  status    - Show client and file status"
        echo ""
        echo "Files:"
        echo "  Timesheet: \$TIMESHEET_FILE"
        echo "  Invoices:  \$INVOICE_DIR/"
        echo ""
        echo "Examples:"
        echo "  \$0 validate    # Check timesheet format"
        echo "  \$0 full        # Complete workflow"
        echo "  \$0 status      # Show current status"
        ;;
esac
EOF
    
    chmod +x "$commands_file"
    echo -e "${GREEN}✅ Quick commands script created: $commands_file${NC}"
}

# Function to display next steps
show_next_steps() {
    local client_name="$1"
    local client_slug="$2"
    local current_year=$(date +%Y)
    local current_month=$(date +%m)
    local month_name=$(date +%B)
    
    echo ""
    echo -e "${GREEN}🎉 Client setup completed for: $client_name${NC}"
    echo ""
    echo -e "${BLUE}📋 Next Steps:${NC}"
    echo ""
    echo "1. Edit your timesheet:"
    echo "   vim $DATA_DIR/timesheets/$current_year/$current_month-$month_name/${client_slug}.csv"
    echo ""
    echo "2. Validate the timesheet:"
    echo "   go-invoice import csv \"$DATA_DIR/timesheets/$current_year/$current_month-$month_name/${client_slug}.csv\" --client \"$client_name\" --validate --dry-run"
    echo ""
    echo "3. Import and create invoice:"
    echo "   $DATA_DIR/quick-commands-${client_slug}.sh full"
    echo ""
    echo "4. Or use individual commands:"
    echo "   $DATA_DIR/quick-commands-${client_slug}.sh validate"
    echo "   $DATA_DIR/quick-commands-${client_slug}.sh import"
    echo "   $DATA_DIR/quick-commands-${client_slug}.sh invoice"
    echo ""
    echo -e "${BLUE}📁 Files created:${NC}"
    echo "   • Timesheet template: $DATA_DIR/timesheets/$current_year/$current_month-$month_name/${client_slug}.csv"
    echo "   • Quick commands: $DATA_DIR/quick-commands-${client_slug}.sh"
    echo "   • Client list: $DATA_DIR/clients.txt"
    echo ""
    echo -e "${BLUE}🔍 Check client status:${NC}"
    echo "   go-invoice client show --name \"$client_name\""
    echo ""
    echo -e "${BLUE}🤖 For automation:${NC}"
    echo "   examples/scripts/monthly-billing.sh"
}

# Main function
main() {
    local client_name="$1"
    local client_email="$2"
    local hourly_rate="${3:-100.00}"
    
    # Validate arguments
    if [[ -z "$client_name" ]]; then
        echo -e "${RED}❌ Error: Client name is required${NC}"
        echo ""
        usage
        exit 1
    fi
    
    # Create client slug (safe filename)
    local client_slug="${client_name// /-}"
    client_slug="${client_slug//[^a-zA-Z0-9-]/}"
    
    echo -e "${GREEN}🚀 Setting up client: $client_name${NC}"
    echo -e "${BLUE}📧 Email: ${client_email:-'Not provided'}${NC}"
    echo -e "${BLUE}💰 Rate: \$${hourly_rate}/hour${NC}"
    echo -e "${BLUE}🏷️  Slug: ${client_slug}${NC}"
    echo ""
    
    # Check if go-invoice is available
    if ! command -v go-invoice &> /dev/null; then
        echo -e "${RED}❌ Error: go-invoice is not installed or not in PATH${NC}"
        exit 1
    fi
    
    # Create directory structure
    create_directories "$client_slug"
    
    # Add client to go-invoice
    add_client_to_go_invoice "$client_name" "$client_email"
    
    # Create sample timesheet
    create_sample_timesheet "$client_name" "$client_slug" "$hourly_rate"
    
    # Add to clients.txt
    add_to_clients_file "$client_name"
    
    # Create quick commands
    create_quick_commands "$client_name" "$client_slug"
    
    # Show next steps
    show_next_steps "$client_name" "$client_slug"
}

# Handle help flag
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    usage
    exit 0
fi

# Run main function
main "$@"