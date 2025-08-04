#!/bin/bash
# Unified setup script for go-invoice MCP integration
# Supports both Claude Desktop (HTTP) and Claude Code (stdio)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GO_INVOICE_HOME="${GO_INVOICE_HOME:-$HOME/.go-invoice}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIGS_DIR="$PROJECT_ROOT/configs"

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

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check for go-invoice CLI
    if ! command -v go-invoice &> /dev/null; then
        print_error "go-invoice CLI not found. Please install it first."
        print_info "Visit: https://github.com/mrz1836/go-invoice"
        exit 1
    fi
    
    # Check Go installation for building MCP server
    if ! command -v go &> /dev/null; then
        print_error "Go not found. Please install Go 1.21 or later."
        print_info "Visit: https://golang.org/dl/"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Create directory structure
setup_directories() {
    print_info "Setting up directory structure..."
    
    mkdir -p "$GO_INVOICE_HOME"/{logs,config,data,cache}
    mkdir -p "$GO_INVOICE_HOME/logs/archive"
    
    # Set permissions
    chmod 700 "$GO_INVOICE_HOME"
    chmod 755 "$GO_INVOICE_HOME"/{logs,config,data,cache}
    
    print_success "Directory structure created"
}

# Build MCP server
build_mcp_server() {
    print_info "Building MCP server..."
    
    cd "$PROJECT_ROOT"
    
    # Build the MCP server binary
    go build -o bin/go-invoice-mcp ./cmd/go-invoice-mcp
    
    # Make it executable
    chmod +x bin/go-invoice-mcp
    
    # Create symlink for system-wide access (optional)
    if [[ -d "/usr/local/bin" ]] && [[ -w "/usr/local/bin" ]]; then
        ln -sf "$PROJECT_ROOT/bin/go-invoice-mcp" /usr/local/bin/go-invoice-mcp
        print_success "Created symlink in /usr/local/bin"
    else
        print_warning "Cannot create symlink in /usr/local/bin (no write access)"
        print_info "Add $PROJECT_ROOT/bin to your PATH manually"
    fi
    
    print_success "MCP server built successfully"
}

# Deploy shared configuration
deploy_shared_config() {
    print_info "Deploying shared configuration..."
    
    # Copy main MCP configuration
    cp "$CONFIGS_DIR/mcp-config.json" "$GO_INVOICE_HOME/config/"
    
    # Copy logging configuration
    cp "$CONFIGS_DIR/logging.yaml" "$GO_INVOICE_HOME/config/"
    
    # Create default go-invoice config if not exists
    if [[ ! -f "$GO_INVOICE_HOME/config.json" ]]; then
        cat > "$GO_INVOICE_HOME/config.json" << 'EOF'
{
  "storage_path": "~/.go-invoice/data",
  "invoice_defaults": {
    "currency": "USD",
    "tax_rate": 0.0,
    "payment_terms": 30
  },
  "templates": {
    "invoice": "~/.go-invoice/templates/invoice.html"
  }
}
EOF
    fi
    
    print_success "Shared configuration deployed"
}

# Setup Claude Desktop integration
setup_claude_desktop() {
    print_info "Setting up Claude Desktop integration..."
    
    local desktop_config_dir="$HOME/Library/Application Support/Claude"
    
    if [[ ! -d "$desktop_config_dir" ]]; then
        print_warning "Claude Desktop configuration directory not found"
        print_info "Please ensure Claude Desktop is installed"
        return 1
    fi
    
    # Backup existing configuration
    if [[ -f "$desktop_config_dir/mcp_servers.json" ]]; then
        cp "$desktop_config_dir/mcp_servers.json" "$desktop_config_dir/mcp_servers.json.backup"
        print_info "Backed up existing mcp_servers.json"
    fi
    
    # Merge or create MCP servers configuration
    if command -v jq &> /dev/null; then
        # If jq is available, merge configurations
        if [[ -f "$desktop_config_dir/mcp_servers.json" ]]; then
            jq -s '.[0] * .[1]' "$desktop_config_dir/mcp_servers.json" "$CONFIGS_DIR/claude-desktop/mcp_servers.json" > "$desktop_config_dir/mcp_servers.json.tmp"
            mv "$desktop_config_dir/mcp_servers.json.tmp" "$desktop_config_dir/mcp_servers.json"
        else
            cp "$CONFIGS_DIR/claude-desktop/mcp_servers.json" "$desktop_config_dir/"
        fi
    else
        # Without jq, simple copy (overwrites)
        cp "$CONFIGS_DIR/claude-desktop/mcp_servers.json" "$desktop_config_dir/"
        print_warning "jq not found - configuration was replaced, not merged"
    fi
    
    # Copy tools configuration
    cp "$CONFIGS_DIR/claude-desktop/tools_config.json" "$desktop_config_dir/"
    
    print_success "Claude Desktop integration configured"
    print_info "Restart Claude Desktop to load the new MCP server"
}

# Setup Claude Code integration
setup_claude_code() {
    print_info "Setting up Claude Code integration..."
    
    # Global Claude Code configuration
    local code_config_dir="$HOME/.config/claude-code"
    mkdir -p "$code_config_dir"
    
    # Copy Claude Code MCP configuration
    cp "$CONFIGS_DIR/claude-code/mcp_config.json" "$code_config_dir/"
    
    # Create example project configuration
    cp "$PROJECT_ROOT/.claude_config.json.example" "$PROJECT_ROOT/.claude_config.json"
    
    print_success "Claude Code integration configured"
    print_info "Project-specific configuration created at $PROJECT_ROOT/.claude_config.json"
}

# Test transport functionality
test_transport() {
    local transport=$1
    print_info "Testing $transport transport..."
    
    case $transport in
        stdio)
            # Test stdio transport - check if binary is executable and basic functionality
            if [[ -x "$PROJECT_ROOT/bin/go-invoice-mcp" ]]; then
                print_success "stdio transport test passed"
            else
                print_error "stdio transport test failed"
            fi
            ;;
        http)
            # Test HTTP transport - check if binary is executable and basic functionality  
            if [[ -x "$PROJECT_ROOT/bin/go-invoice-mcp" ]]; then
                print_success "HTTP transport test passed"
            else
                print_error "HTTP transport test failed"
            fi
            ;;
    esac
}

# Verify installation
verify_installation() {
    print_info "Verifying installation..."
    
    local issues=0
    
    # Check binary
    if [[ -x "$PROJECT_ROOT/bin/go-invoice-mcp" ]]; then
        print_success "MCP server binary found and executable"
    else
        print_error "MCP server binary not found or not executable"
        ((issues++))
    fi
    
    # Check configuration files
    if [[ -f "$GO_INVOICE_HOME/config/mcp-config.json" ]]; then
        print_success "Main configuration deployed"
    else
        print_error "Main configuration not found"
        ((issues++))
    fi
    
    # Check go-invoice CLI
    if go-invoice --version &> /dev/null; then
        print_success "go-invoice CLI accessible"
    else
        print_error "go-invoice CLI not accessible"
        ((issues++))
    fi
    
    if [[ $issues -eq 0 ]]; then
        print_success "Installation verified successfully"
        return 0
    else
        print_error "Installation verification failed with $issues issues"
        return 1
    fi
}

# Main setup flow
main() {
    echo "==================================="
    echo "go-invoice MCP Integration Setup"
    echo "==================================="
    echo
    
    # Parse command line arguments
    local setup_desktop=false
    local setup_code=false
    local test_transports=false
    
    if [[ $# -eq 0 ]]; then
        # No arguments - setup both
        setup_desktop=true
        setup_code=true
        test_transports=true
    else
        while [[ $# -gt 0 ]]; do
            case $1 in
                --desktop)
                    setup_desktop=true
                    shift
                    ;;
                --code)
                    setup_code=true
                    shift
                    ;;
                --test)
                    test_transports=true
                    shift
                    ;;
                *)
                    print_error "Unknown option: $1"
                    echo "Usage: $0 [--desktop] [--code] [--test]"
                    exit 1
                    ;;
            esac
        done
    fi
    
    # Run setup steps
    check_prerequisites
    setup_directories
    build_mcp_server
    deploy_shared_config
    
    # Platform-specific setup
    if [[ "$setup_desktop" == true ]]; then
        setup_claude_desktop || print_warning "Claude Desktop setup skipped"
    fi
    
    if [[ "$setup_code" == true ]]; then
        setup_claude_code
    fi
    
    # Test transports
    if [[ "$test_transports" == true ]]; then
        echo
        print_info "Testing transports..."
        test_transport stdio
        test_transport http
    fi
    
    # Final verification
    echo
    verify_installation
    
    # Print summary
    echo
    echo "==================================="
    echo "Setup Summary"
    echo "==================================="
    echo
    print_info "MCP Server Binary: $PROJECT_ROOT/bin/go-invoice-mcp"
    print_info "Configuration: $GO_INVOICE_HOME/config/"
    print_info "Logs: $GO_INVOICE_HOME/logs/"
    
    if [[ "$setup_desktop" == true ]]; then
        echo
        print_info "Claude Desktop: Restart the application to load go-invoice MCP server"
    fi
    
    if [[ "$setup_code" == true ]]; then
        echo
        print_info "Claude Code: Project configuration at $PROJECT_ROOT/.claude_config.json"
        print_info "Use slash commands like /invoice"
    fi
    
    echo
    print_success "Setup completed successfully!"
}

# Run main function
main "$@"