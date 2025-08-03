// Package main provides an example of how to use the complete MCP tool registry system.
//
// This example demonstrates the unified tool registry and validation system that integrates
// all 21 tools from phases 2.2-2.4 into a comprehensive MCP integration. It shows how to
// initialize the system, discover tools, and validate inputs.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mrz/go-invoice/internal/mcp/tools"
)

// ExampleUsage demonstrates the complete MCP tool system functionality.
//
// This function shows how to:
// 1. Initialize the complete tool system with all 21 tools
// 2. Use the discovery service to find tools
// 3. Validate tool inputs
// 4. Get system metrics and status
//
// This serves as both documentation and a working example for MCP server integration.
func ExampleUsage() {
	ctx := context.Background()

	fmt.Println("=== Go-Invoice MCP Tool System Example ===")
	fmt.Println()

	// Step 1: Initialize the complete tool system
	fmt.Println("1. Initializing tool system...")
	components, err := tools.InitializeToolSystem(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to initialize tool system: %v", err)
	}
	fmt.Printf("✓ Tool system initialized with %d tools\n", 21)
	fmt.Println()

	// Step 2: Explore the tool registry
	fmt.Println("2. Exploring tool registry...")

	// Get all categories
	categories, err := components.Registry.GetCategories(ctx)
	if err != nil {
		log.Fatalf("Failed to get categories: %v", err)
	}

	fmt.Printf("✓ Found %d categories:\n", len(categories))
	for _, category := range categories {
		toolsInCategory, err := components.Registry.ListTools(ctx, category)
		if err != nil {
			continue
		}
		fmt.Printf("  - %s: %d tools\n", category, len(toolsInCategory))
	}
	fmt.Println()

	// Step 3: Demonstrate tool discovery
	fmt.Println("3. Demonstrating tool discovery...")

	// Search for invoice-related tools
	searchCriteria := &tools.ToolSearchCriteria{
		Query:             "invoice",
		MaxResults:        5,
		MinRelevanceScore: 0.5,
		SortBy:            "relevance",
		SortOrder:         "desc",
	}

	searchResults, err := components.DiscoveryService.SearchTools(ctx, searchCriteria)
	if err != nil {
		log.Fatalf("Failed to search tools: %v", err)
	}

	fmt.Printf("✓ Found %d tools matching 'invoice':\n", len(searchResults))
	for _, result := range searchResults {
		fmt.Printf("  - %s (relevance: %.2f): %s\n",
			result.Tool.Name,
			result.RelevanceScore,
			result.Tool.Description[:min(80, len(result.Tool.Description))])
	}
	fmt.Println()

	// Step 4: Demonstrate category-based discovery
	fmt.Println("4. Demonstrating category-based discovery...")

	categoryResult, err := components.DiscoveryService.DiscoverToolsByCategory(ctx, tools.CategoryInvoiceManagement)
	if err != nil {
		log.Fatalf("Failed to discover category tools: %v", err)
	}

	fmt.Printf("✓ Invoice Management category has %d tools:\n", categoryResult.ToolCount)
	for _, tool := range categoryResult.Tools {
		fmt.Printf("  - %s: %s\n", tool.Name, tool.Description[:min(60, len(tool.Description))])
	}
	fmt.Println()

	// Step 5: Demonstrate tool validation
	fmt.Println("5. Demonstrating tool validation...")

	// Get a specific tool and test validation
	invoiceCreateTool, err := components.Registry.GetTool(ctx, "invoice_create")
	if err != nil {
		log.Fatalf("Failed to get invoice_create tool: %v", err)
	}

	// Test with invalid input (empty)
	emptyInput := map[string]interface{}{}
	err = components.Registry.ValidateToolInput(ctx, "invoice_create", emptyInput)
	if err != nil {
		fmt.Printf("✓ Validation correctly rejected empty input: %s\n", err.Error()[:min(80, len(err.Error()))])
	} else {
		fmt.Printf("✓ Validation passed for empty input (no required fields)\n")
	}

	// Test with valid input
	validInput := map[string]interface{}{
		"client_name": "Example Client",
		"description": "Test invoice for demonstration",
	}
	err = components.Registry.ValidateToolInput(ctx, "invoice_create", validInput)
	if err != nil {
		fmt.Printf("✗ Validation failed for valid input: %s\n", err.Error())
	} else {
		fmt.Printf("✓ Validation passed for valid input\n")
	}
	fmt.Println()

	// Step 6: Show system metrics
	fmt.Println("6. System metrics...")

	metrics, err := components.Registry.GetRegistrationMetrics(ctx)
	if err != nil {
		log.Fatalf("Failed to get metrics: %v", err)
	}

	fmt.Printf("✓ System metrics:\n")
	fmt.Printf("  - Total tools: %d\n", metrics.TotalTools)
	fmt.Printf("  - Total categories: %d\n", metrics.TotalCategories)
	fmt.Printf("  - System uptime: %s\n", metrics.Uptime.Round(time.Millisecond))
	fmt.Printf("  - Initialization time: %s\n", metrics.InitializationTime.Format(time.RFC3339))

	fmt.Printf("  - Tools by category:\n")
	for category, count := range metrics.ToolsByCategory {
		fmt.Printf("    - %s: %d tools\n", category, count)
	}
	fmt.Println()

	// Step 7: Show tool schema example
	fmt.Println("7. Example tool schema...")

	fmt.Printf("✓ Schema for %s:\n", invoiceCreateTool.Name)
	schemaJSON, err := json.MarshalIndent(invoiceCreateTool.InputSchema, "  ", "  ")
	if err != nil {
		fmt.Printf("  Error marshaling schema: %v\n", err)
	} else {
		// Show first few lines of schema
		schemaStr := string(schemaJSON)
		lines := splitLines(schemaStr, 10)
		for _, line := range lines {
			fmt.Printf("  %s\n", line)
		}
		if len(schemaStr) > 500 {
			fmt.Printf("  ... (schema truncated for display)\n")
		}
	}
	fmt.Println()

	// Step 8: Tool recommendations
	fmt.Println("8. Tool recommendations...")

	recommendations, err := components.DiscoveryService.GetToolRecommendations(ctx, "I need to create and send invoices", 3)
	if err != nil {
		log.Fatalf("Failed to get recommendations: %v", err)
	}

	fmt.Printf("✓ Recommendations for 'creating and sending invoices' (%d):\n", len(recommendations))
	for _, rec := range recommendations {
		if rec.Tool != nil {
			fmt.Printf("  - %s (confidence: %.2f): %s\n",
				rec.Tool.Name, rec.Confidence, rec.Rationale)
		}
	}
	fmt.Println()

	fmt.Println("=== Example completed successfully! ===")
	fmt.Printf("The MCP tool system is ready for integration with Claude Desktop.\n")
	fmt.Printf("All 21 tools across 5 categories are available for use.\n")
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func splitLines(text string, maxLines int) []string {
	var lines []string
	current := ""
	lineCount := 0

	for _, char := range text {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
			lineCount++
			if lineCount >= maxLines {
				break
			}
		} else {
			current += string(char)
		}
	}

	if current != "" && lineCount < maxLines {
		lines = append(lines, current)
	}

	return lines
}

// main runs the example if this file is executed directly.
func main() {
	if len(os.Args) > 1 && os.Args[1] == "example" {
		ExampleUsage()
	} else {
		fmt.Println("Use 'go run example_usage.go example' to run the demonstration")
	}
}
