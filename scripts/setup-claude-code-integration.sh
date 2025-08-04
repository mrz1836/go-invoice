#!/bin/bash
# Claude Code specific setup script for go-invoice MCP integration
# Focuses on project-level stdio transport configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIGS_DIR="$PROJECT_ROOT/configs/claude-code"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running in a project directory
check_project_context() {
    print_info "Checking project context..."
    
    if [[ ! -f "package.json" ]] && [[ ! -f "go.mod" ]] && [[ ! -f "Cargo.toml" ]] && [[ ! -f "pom.xml" ]]; then
        print_warning "No project file detected. This might not be a project root."
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    print_success "Project context verified"
}

# Create project structure for invoices
setup_project_structure() {
    print_info "Setting up project invoice structure..."
    
    # Create directories
    mkdir -p invoices/{drafts,sent,paid}
    mkdir -p timesheets/{pending,processed}
    mkdir -p templates
    mkdir -p .go-invoice/{logs,cache}
    
    # Create .gitignore entries
    if [[ -f .gitignore ]]; then
        # Check if entries already exist
        if ! grep -q "^\.go-invoice/" .gitignore; then
            echo "" >> .gitignore
            echo "# go-invoice MCP integration" >> .gitignore
            echo ".go-invoice/" >> .gitignore
            echo "*.log" >> .gitignore
            print_info "Updated .gitignore"
        fi
    else
        cat > .gitignore << 'EOF'
# go-invoice MCP integration
.go-invoice/
*.log

# Invoice files (uncomment to ignore)
# invoices/
# timesheets/
EOF
        print_info "Created .gitignore"
    fi
    
    print_success "Project structure created"
}

# Deploy Claude Code configuration
deploy_claude_code_config() {
    print_info "Deploying Claude Code configuration..."
    
    # Check for existing configuration
    if [[ -f .claude_config.json ]]; then
        cp .claude_config.json .claude_config.json.backup
        print_warning "Backed up existing .claude_config.json"
    fi
    
    # Get project information
    local project_name=$(basename "$PWD")
    local invoice_prefix="INV"
    
    # Try to detect project type and suggest prefix
    if [[ -f package.json ]]; then
        project_name=$(jq -r '.name // empty' package.json 2>/dev/null || echo "$project_name")
    elif [[ -f go.mod ]]; then
        project_name=$(head -n1 go.mod | awk '{print $2}' | xargs basename)
    fi
    
    # Interactive configuration
    echo
    print_info "Project Configuration"
    echo "---------------------"
    read -p "Project name [$project_name]: " input_name
    project_name=${input_name:-$project_name}
    
    read -p "Invoice prefix [$invoice_prefix]: " input_prefix
    invoice_prefix=${input_prefix:-$invoice_prefix}
    
    read -p "Default client name (leave empty for none): " default_client
    
    # Create configuration from template
    cat > .claude_config.json << EOF
{
  "mcp": {
    "servers": {
      "go-invoice": {
        "command": "${PROJECT_ROOT}/bin/go-invoice-mcp",
        "args": [
          "--stdio",
          "--config",
          "./.go-invoice/config.json"
        ],
        "workingDirectory": ".",
        "env": {
          "GO_INVOICE_PROJECT": "./",
          "MCP_TRANSPORT": "stdio",
          "MCP_LOG_FILE": "./.go-invoice/mcp.log"
        },
        "projectSettings": {
          "projectName": "$project_name",
          "defaultClient": "$default_client",
          "invoicePrefix": "$invoice_prefix",
          "autoImportPath": "./timesheets",
          "outputPath": "./invoices",
          "templatePath": "./templates/invoice.html"
        }
      }
    },
    "shortcuts": {
      "invoice": {
        "description": "Quick invoice creation for $project_name",
        "server": "go-invoice",
        "tool": "invoice_create"
      },
      "import": {
        "description": "Import timesheet",
        "server": "go-invoice",
        "tool": "import_csv"
      },
      "generate": {
        "description": "Generate invoice HTML",
        "server": "go-invoice",
        "tool": "generate_html"
      },
      "summary": {
        "description": "Show invoice summary",
        "server": "go-invoice",
        "tool": "invoice_summary"
      }
    }
  },
  "workspace": {
    "invoices": {
      "path": "./invoices",
      "watch": true,
      "autoOpen": true
    },
    "timesheets": {
      "path": "./timesheets",
      "watch": true,
      "patterns": ["*.csv", "*.xlsx"]
    },
    "templates": {
      "path": "./templates",
      "watch": false
    }
  }
}
EOF
    
    print_success "Claude Code configuration created"
}

# Create project-specific go-invoice configuration
create_project_invoice_config() {
    print_info "Creating project invoice configuration..."
    
    cat > .go-invoice/config.json << EOF
{
  "storage_path": "./.go-invoice/data",
  "invoice_defaults": {
    "currency": "USD",
    "tax_rate": 0.0,
    "payment_terms": 30,
    "prefix": "${invoice_prefix:-INV}"
  },
  "templates": {
    "invoice": "./templates/invoice.html",
    "quote": "./templates/quote.html"
  },
  "import": {
    "csv": {
      "delimiter": ",",
      "date_format": "2006-01-02",
      "time_format": "15:04"
    }
  },
  "export": {
    "html": {
      "output_dir": "./invoices",
      "open_after_generate": true
    },
    "pdf": {
      "output_dir": "./invoices",
      "engine": "wkhtmltopdf"
    }
  }
}
EOF
    
    print_success "Project invoice configuration created"
}

# Create sample invoice template
create_sample_template() {
    print_info "Creating sample invoice template..."
    
    cat > templates/invoice.html << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Invoice {{.Invoice.Number}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 40px 20px;
        }
        .header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 40px;
            padding-bottom: 20px;
            border-bottom: 2px solid #f0f0f0;
        }
        .invoice-title {
            font-size: 32px;
            font-weight: 300;
            color: #2c3e50;
            margin: 0;
        }
        .invoice-number {
            text-align: right;
            color: #7f8c8d;
        }
        .invoice-number strong {
            color: #2c3e50;
        }
        .parties {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 40px;
            margin-bottom: 40px;
        }
        .party h3 {
            color: #2c3e50;
            margin-bottom: 10px;
        }
        .party p {
            margin: 5px 0;
            color: #555;
        }
        .items-table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 30px;
        }
        .items-table th {
            background-color: #f8f9fa;
            padding: 12px;
            text-align: left;
            font-weight: 600;
            color: #2c3e50;
            border-bottom: 2px solid #dee2e6;
        }
        .items-table td {
            padding: 12px;
            border-bottom: 1px solid #f0f0f0;
        }
        .items-table .amount {
            text-align: right;
        }
        .totals {
            margin-left: auto;
            width: 300px;
        }
        .totals-row {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
        }
        .totals-row.total {
            font-size: 20px;
            font-weight: 600;
            color: #2c3e50;
            border-top: 2px solid #dee2e6;
            margin-top: 10px;
            padding-top: 15px;
        }
        .footer {
            margin-top: 60px;
            padding-top: 30px;
            border-top: 1px solid #f0f0f0;
            text-align: center;
            color: #7f8c8d;
            font-size: 14px;
        }
        @media print {
            body {
                padding: 0;
            }
            .header {
                page-break-after: avoid;
            }
            .items-table {
                page-break-inside: avoid;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1 class="invoice-title">Invoice</h1>
        <div class="invoice-number">
            <strong>{{.Invoice.Number}}</strong><br>
            {{.Invoice.Date.Format "January 2, 2006"}}
        </div>
    </div>

    <div class="parties">
        <div class="party">
            <h3>From</h3>
            <p><strong>{{.From.Name}}</strong></p>
            {{if .From.Address}}<p>{{.From.Address}}</p>{{end}}
            {{if .From.Email}}<p>{{.From.Email}}</p>{{end}}
            {{if .From.Phone}}<p>{{.From.Phone}}</p>{{end}}
        </div>
        <div class="party">
            <h3>Bill To</h3>
            <p><strong>{{.Client.Name}}</strong></p>
            {{if .Client.Address}}<p>{{.Client.Address}}</p>{{end}}
            {{if .Client.Email}}<p>{{.Client.Email}}</p>{{end}}
            {{if .Client.Phone}}<p>{{.Client.Phone}}</p>{{end}}
        </div>
    </div>

    <table class="items-table">
        <thead>
            <tr>
                <th>Description</th>
                <th style="width: 100px">Hours</th>
                <th style="width: 120px">Rate</th>
                <th style="width: 120px" class="amount">Amount</th>
            </tr>
        </thead>
        <tbody>
            {{range .Invoice.Items}}
            <tr>
                <td>{{.Description}}</td>
                <td>{{.Quantity}}</td>
                <td>${{.Rate}}/hr</td>
                <td class="amount">${{.Total}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>

    <div class="totals">
        <div class="totals-row">
            <span>Subtotal</span>
            <span>${{.Invoice.Subtotal}}</span>
        </div>
        {{if gt .Invoice.Tax 0}}
        <div class="totals-row">
            <span>Tax ({{.Invoice.TaxRate}}%)</span>
            <span>${{.Invoice.Tax}}</span>
        </div>
        {{end}}
        <div class="totals-row total">
            <span>Total</span>
            <span>${{.Invoice.Total}}</span>
        </div>
    </div>

    <div class="footer">
        <p>Payment is due within {{.Invoice.PaymentTerms}} days.</p>
        <p>Thank you for your business!</p>
    </div>
</body>
</html>
EOF
    
    print_success "Sample invoice template created"
}

# Create sample timesheet
create_sample_timesheet() {
    print_info "Creating sample timesheet..."
    
    cat > timesheets/sample-timesheet.csv << 'EOF'
Date,Hours,Description,Project,Tags
2025-01-15,2.5,"Initial project setup and configuration",Setup,setup;config
2025-01-16,4.0,"Implement authentication system",Development,auth;backend
2025-01-17,3.5,"Create user dashboard UI",Development,frontend;ui
2025-01-18,2.0,"Write unit tests for auth module",Testing,testing;backend
2025-01-19,1.5,"Code review and documentation",Review,review;docs
EOF
    
    print_success "Sample timesheet created at timesheets/sample-timesheet.csv"
}

# Test Claude Code integration
test_claude_code_integration() {
    print_info "Testing Claude Code integration..."
    
    # Check if MCP server exists
    if [[ ! -x "${PROJECT_ROOT}/bin/go-invoice-mcp" ]]; then
        print_error "MCP server not found. Run the main setup script first."
        return 1
    fi
    
    # Test stdio communication
    local test_result=$(echo '{"jsonrpc":"2.0","method":"initialize","params":{"capabilities":{}},"id":1}' | \
        "${PROJECT_ROOT}/bin/go-invoice-mcp" --stdio --config "./.go-invoice/config.json" 2>/dev/null | \
        head -n1)
    
    if echo "$test_result" | grep -q '"result"'; then
        print_success "stdio transport test passed"
    else
        print_error "stdio transport test failed"
        return 1
    fi
    
    print_success "Claude Code integration test completed"
}

# Print usage instructions
print_usage_instructions() {
    echo
    echo "==================================="
    echo "Claude Code Integration Usage"
    echo "==================================="
    echo
    print_info "Available slash commands:"
    echo "  /mcp__go_invoice__create_invoice    - Create a new invoice"
    echo "  /mcp__go_invoice__list_invoices     - List all invoices"
    echo "  /mcp__go_invoice__import_csv        - Import timesheet from CSV"
    echo "  /mcp__go_invoice__generate_html     - Generate HTML invoice"
    echo "  /mcp__go_invoice__show_config       - Display configuration"
    echo
    print_info "Resource mentions:"
    echo "  @invoice:INV-2025-001  - Reference a specific invoice"
    echo "  @client:\"Acme Corp\"    - Reference a client"
    echo "  @timesheet:./hours.csv - Reference a timesheet file"
    echo "  @config:invoice_defaults - Reference configuration"
    echo
    print_info "Example workflow:"
    echo "  1. Add hours to timesheets/january.csv"
    echo "  2. Use: /mcp__go_invoice__import_csv @timesheet:./timesheets/january.csv"
    echo "  3. Use: /mcp__go_invoice__create_invoice for \"Acme Corp\""
    echo "  4. Use: /mcp__go_invoice__generate_html @invoice:latest"
    echo
}

# Main setup flow
main() {
    echo "============================================"
    echo "Claude Code Integration Setup for go-invoice"
    echo "============================================"
    echo
    
    # Check if running from project root
    check_project_context
    
    # Create project structure
    setup_project_structure
    
    # Deploy configuration
    deploy_claude_code_config
    create_project_invoice_config
    
    # Create templates and samples
    create_sample_template
    create_sample_timesheet
    
    # Test integration
    echo
    test_claude_code_integration
    
    # Print usage instructions
    print_usage_instructions
    
    echo
    print_success "Claude Code integration setup completed!"
    print_info "Open this project in Claude Code to start using go-invoice MCP integration"
}

# Run main function
main "$@"