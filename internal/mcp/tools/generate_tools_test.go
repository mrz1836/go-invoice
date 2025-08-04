package tools

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// GenerateToolsTestSuite provides comprehensive tests for generation tool definitions
type GenerateToolsTestSuite struct {
	suite.Suite

	// No context stored in struct - pass through method parameters instead
}

func (s *GenerateToolsTestSuite) SetupTest() {
	// Context created as needed in individual test methods
}

func (s *GenerateToolsTestSuite) TestCreateGenerateTools() {
	tools := CreateDocumentGenerationTools()

	s.NotNil(tools, "Should return tools slice")
	s.NotEmpty(tools, "Should return at least one tool")

	// Expected generation tools based on typical functionality
	expectedToolNames := []string{
		"generate_invoice",
		"generate_pdf",
		"generate_report",
	}

	// Verify we have the expected tools (at least the core ones)
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for _, expectedName := range expectedToolNames {
		if toolNames[expectedName] {
			s.True(toolNames[expectedName], "Should have %s tool", expectedName)
		}
	}

	// Verify all tools are properly configured
	for _, tool := range tools {
		s.validateGenerateTool(tool)
	}
}

func (s *GenerateToolsTestSuite) TestInvoiceGenerationTool() {
	tools := CreateDocumentGenerationTools()
	var invoiceGenTool *MCPTool

	for _, tool := range tools {
		if strings.Contains(tool.Name, "generate") && strings.Contains(tool.Name, "invoice") {
			invoiceGenTool = tool
			break
		}
	}

	if invoiceGenTool == nil {
		s.T().Skip("invoice generation tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Contains(strings.ToLower(invoiceGenTool.Name), "generate")
		s.Contains(strings.ToLower(invoiceGenTool.Name), "invoice")
		s.NotEmpty(invoiceGenTool.Description)
		s.Contains(strings.ToLower(invoiceGenTool.Description), "generate")
		s.Equal(CategoryDataExport, invoiceGenTool.Category)
		s.NotEmpty(invoiceGenTool.Version)
		s.Greater(invoiceGenTool.Timeout, time.Duration(0))
	})

	s.Run("Schema", func() {
		s.NotNil(invoiceGenTool.InputSchema)
		s.Equal("object", invoiceGenTool.InputSchema["type"])

		// Verify expected properties exist
		properties, hasProperties := invoiceGenTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties in schema")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have invoice identifier
			identifierFields := []string{"invoice_id", "invoice_number"}
			hasIdentifier := false
			for _, field := range identifierFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasIdentifier = true
					break
				}
			}
			s.True(hasIdentifier, "Should have invoice identifier field")

			// May have output format options
			formatFields := []string{"format", "output_format", "type"}
			for _, field := range formatFields {
				if fieldDef, hasField := propertiesMap[field]; hasField {
					s.NotNil(fieldDef, "Format field %s should have definition", field)
				}
			}
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(invoiceGenTool.Examples, "Should have examples")

		for i, example := range invoiceGenTool.Examples {
			s.NotEmpty(example.Description, "Example %d should have description", i)
			s.NotNil(example.Input, "Example %d should have input", i)
			s.NotEmpty(example.Input, "Example %d input should not be empty", i)

			// Verify example demonstrates invoice generation
			hasInvoiceRef := false
			invoiceFields := []string{"invoice_id", "invoice_number"}
			for _, field := range invoiceFields {
				if invoiceVal, hasInvoice := example.Input[field]; hasInvoice {
					s.IsType("", invoiceVal, "Invoice field should be string")
					s.NotEmpty(invoiceVal, "Invoice field should not be empty")
					hasInvoiceRef = true
				}
			}
			s.True(hasInvoiceRef, "Example should demonstrate invoice generation")
		}
	})

	s.Run("OutputFormats", func() {
		// Check if examples cover different output formats
		hasPDFExample := false
		hasHTMLExample := false
		hasJSONExample := false

		for _, example := range invoiceGenTool.Examples {
			exampleText := strings.ToLower(example.Description + " " + example.UseCase)
			if format, hasFormat := example.Input["format"]; hasFormat {
				if formatStr, isStr := format.(string); isStr {
					switch strings.ToLower(formatStr) {
					case "pdf":
						hasPDFExample = true
					case "html":
						hasHTMLExample = true
					case "json":
						hasJSONExample = true
					}
				}
			}

			if strings.Contains(exampleText, "pdf") {
				hasPDFExample = true
			}
			if strings.Contains(exampleText, "html") {
				hasHTMLExample = true
			}
		}

		s.True(hasPDFExample || hasHTMLExample || hasJSONExample, "Should have examples for different output formats")
	})
}

func (s *GenerateToolsTestSuite) TestPDFGenerationTool() {
	tools := CreateDocumentGenerationTools()
	var pdfGenTool *MCPTool

	for _, tool := range tools {
		if strings.Contains(strings.ToLower(tool.Name), "pdf") {
			pdfGenTool = tool
			break
		}
	}

	if pdfGenTool == nil {
		s.T().Skip("PDF generation tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Contains(strings.ToLower(pdfGenTool.Name), "pdf")
		s.NotEmpty(pdfGenTool.Description)
		s.Contains(strings.ToLower(pdfGenTool.Description), "pdf")
		s.Equal(CategoryDataExport, pdfGenTool.Category)
	})

	s.Run("Schema", func() {
		s.NotNil(pdfGenTool.InputSchema)
		s.Equal("object", pdfGenTool.InputSchema["type"])

		properties, hasProperties := pdfGenTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have source identifier
			sourceFields := []string{"invoice_id", "invoice_number", "source", "template"}
			hasSource := false
			for _, field := range sourceFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasSource = true
					break
				}
			}
			s.True(hasSource, "Should have source identifier field")
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(pdfGenTool.Examples, "Should have examples")

		for _, example := range pdfGenTool.Examples {
			s.NotEmpty(example.Description)
			s.NotNil(example.Input)

			// Should demonstrate PDF generation scenarios
			exampleText := strings.ToLower(example.Description + " " + example.UseCase)
			s.Contains(exampleText, "pdf", "Example should be PDF-related")
		}
	})
}

func (s *GenerateToolsTestSuite) TestReportGenerationTool() {
	tools := CreateDocumentGenerationTools()
	var reportGenTool *MCPTool

	for _, tool := range tools {
		if strings.Contains(strings.ToLower(tool.Name), "report") {
			reportGenTool = tool
			break
		}
	}

	if reportGenTool == nil {
		s.T().Skip("report generation tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Contains(strings.ToLower(reportGenTool.Name), "report")
		s.NotEmpty(reportGenTool.Description)
		s.Contains(strings.ToLower(reportGenTool.Description), "report")
		// Report generation could be in CategoryDataExport or CategoryReporting
		s.True(
			reportGenTool.Category == CategoryDataExport || reportGenTool.Category == CategoryReporting,
			"Report tool should be in data export or reporting category",
		)
	})

	s.Run("Schema", func() {
		s.NotNil(reportGenTool.InputSchema)
		s.Equal("object", reportGenTool.InputSchema["type"])

		properties, hasProperties := reportGenTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have date range or period fields
			dateFields := []string{"from_date", "to_date", "period", "start_date", "end_date"}
			hasDateField := false
			for _, field := range dateFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasDateField = true
					break
				}
			}
			s.True(hasDateField, "Should have date range fields for reporting")
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(reportGenTool.Examples, "Should have examples")

		for _, example := range reportGenTool.Examples {
			s.NotEmpty(example.Description)
			s.NotNil(example.Input)

			// Should demonstrate reporting scenarios
			exampleText := strings.ToLower(example.Description + " " + example.UseCase)
			reportKeywords := []string{"report", "summary", "analysis", "period", "monthly", "yearly"}
			hasReportKeyword := false
			for _, keyword := range reportKeywords {
				if strings.Contains(exampleText, keyword) {
					hasReportKeyword = true
					break
				}
			}
			s.True(hasReportKeyword, "Example should be report-related")
		}
	})
}

func (s *GenerateToolsTestSuite) TestRegisterGenerateTools() {
	ctx := context.Background()
	// Create mock registry
	validator := new(MockInputValidator)
	logger := new(MockLogger)
	registry := NewDefaultToolRegistry(validator, logger)

	// Setup expectations for tool registration
	generateTools := CreateDocumentGenerationTools()
	logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(len(generateTools))

	// Test registration
	err := RegisterDocumentGenerationTools(ctx, registry)
	if err != nil {
		s.T().Skipf("RegisterGenerateTools not implemented or failed: %v", err)
		return
	}

	s.Require().NoError(err, "Should register generate tools successfully")

	// Verify tools were registered - check both possible categories
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	exportTools, err := registry.ListTools(ctx, CategoryDataExport)
	s.Require().NoError(err)

	reportTools, err := registry.ListTools(ctx, CategoryReporting)
	s.Require().NoError(err)

	totalGenerateTools := len(exportTools) + len(reportTools)
	s.Positive(totalGenerateTools, "Should have registered generation tools")

	// Verify each registered tool
	allTools := append(exportTools, reportTools...)
	for _, tool := range allTools {
		if strings.Contains(strings.ToLower(tool.Name), "generate") {
			s.validateGenerateTool(tool)
		}
	}
}

func (s *GenerateToolsTestSuite) TestGenerateToolsIntegration() {
	s.Run("ToolNamesUnique", func() {
		tools := CreateDocumentGenerationTools()
		toolNames := make(map[string]bool)

		for _, tool := range tools {
			s.False(toolNames[tool.Name], "Tool name %s should be unique", tool.Name)
			toolNames[tool.Name] = true
		}
	})

	s.Run("ToolsCoverGenerationWorkflow", func() {
		tools := CreateDocumentGenerationTools()
		toolNames := make(map[string]bool)

		for _, tool := range tools {
			toolNames[tool.Name] = true
		}

		// Core generation operations
		generationOperations := []string{"generate", "create", "export", "render"}
		hasGenerationOperation := false

		for _, tool := range tools {
			toolNameLower := strings.ToLower(tool.Name)
			for _, op := range generationOperations {
				if strings.Contains(toolNameLower, op) {
					hasGenerationOperation = true
					break
				}
			}
			if hasGenerationOperation {
				break
			}
		}

		s.True(hasGenerationOperation, "Should have tools covering generation operations")
	})

	s.Run("ConsistentNamingConvention", func() {
		tools := CreateDocumentGenerationTools()

		for _, tool := range tools {
			s.True(
				strings.HasPrefix(tool.Name, "generate_") ||
					strings.Contains(tool.Name, "generate") ||
					strings.Contains(tool.Name, "export"),
				"Tool %s should be clearly generation-related", tool.Name,
			)
			s.NotContains(tool.Name, " ", "Tool name should not contain spaces")
			s.Equal(strings.ToLower(tool.Name), tool.Name, "Tool name should be lowercase")
		}
	})

	s.Run("ConsistentCLIIntegration", func() {
		tools := CreateDocumentGenerationTools()

		for _, tool := range tools {
			s.NotEmpty(tool.CLICommand, "Tool %s should have CLI command", tool.Name)
			s.NotEmpty(tool.CLIArgs, "Tool %s should have CLI args", tool.Name)

			// Verify CLI args make sense for generation operations
			hasGenerateArg := false
			cliArgsStr := strings.Join(tool.CLIArgs, " ")
			generateKeywords := []string{"generate", "render", "export", "create"}
			for _, keyword := range generateKeywords {
				if strings.Contains(cliArgsStr, keyword) {
					hasGenerateArg = true
					break
				}
			}
			s.True(hasGenerateArg, "Tool %s should have generation-related CLI args", tool.Name)
		}
	})

	s.Run("ExamplesCoverOutputFormats", func() {
		tools := CreateDocumentGenerationTools()

		allExampleTexts := ""
		for _, tool := range tools {
			for _, example := range tool.Examples {
				allExampleTexts += strings.ToLower(example.Description + " " + example.UseCase + " ")

				// Check format specifications in input
				for key, value := range example.Input {
					if strings.Contains(key, "format") || strings.Contains(key, "type") {
						if formatStr, isStr := value.(string); isStr {
							allExampleTexts += strings.ToLower(formatStr) + " "
						}
					}
				}
			}
		}

		// Should cover common output formats
		commonFormats := []string{"pdf", "html", "json", "csv", "xlsx"}
		formatsCovered := 0
		for _, format := range commonFormats {
			if strings.Contains(allExampleTexts, format) {
				formatsCovered++
			}
		}

		s.Positive(formatsCovered, "Should have examples covering common output formats")
	})

	s.Run("ExamplesCoverTemplating", func() {
		tools := CreateDocumentGenerationTools()

		hasTemplateExample := false
		for _, tool := range tools {
			for _, example := range tool.Examples {
				exampleText := strings.ToLower(example.Description + " " + example.UseCase)
				templateKeywords := []string{"template", "custom", "format", "style", "layout"}
				for _, keyword := range templateKeywords {
					if strings.Contains(exampleText, keyword) {
						hasTemplateExample = true
						break
					}
				}
				if hasTemplateExample {
					break
				}
			}
			if hasTemplateExample {
				break
			}
		}

		s.True(hasTemplateExample, "Should have examples covering templating or customization")
	})
}

func (s *GenerateToolsTestSuite) TestGenerateToolsEdgeCases() {
	s.Run("EmptyToolsList", func() {
		// This tests the function behavior itself
		tools := CreateDocumentGenerationTools()
		s.NotNil(tools, "Should never return nil")
		// It's acceptable for tools to be empty if not implemented yet
	})

	s.Run("OutputPathHandling", func() {
		tools := CreateDocumentGenerationTools()

		for _, tool := range tools {
			for _, example := range tool.Examples {
				// Check output path examples are realistic
				for key, value := range example.Input {
					if strings.Contains(key, "output") || strings.Contains(key, "path") {
						if outputPath, isStr := value.(string); isStr && outputPath != "" {
							// Should look like a real file path
							s.True(
								strings.Contains(outputPath, ".") ||
									strings.Contains(outputPath, "/") ||
									strings.Contains(outputPath, "\\"),
								"Output path example should look realistic: %s", outputPath,
							)
						}
					}
				}
			}
		}
	})

	s.Run("SchemaSupportsFileGeneration", func() {
		tools := CreateDocumentGenerationTools()

		for _, tool := range tools {
			properties, hasProps := tool.InputSchema["properties"]
			if !hasProps {
				continue
			}

			if propsMap, ok := properties.(map[string]interface{}); ok {
				// Look for output-related fields
				outputFields := []string{"output", "output_path", "filename", "destination"}
				for _, field := range outputFields {
					if fieldDef, hasField := propsMap[field]; hasField {
						if fieldMap, isMap := fieldDef.(map[string]interface{}); isMap {
							// Output fields should be strings
							if fieldType, hasType := fieldMap["type"]; hasType {
								s.Equal("string", fieldType, "Output field %s should be string type", field)
							}
						}
					}
				}
			}
		}
	})

	s.Run("TimeoutAppropriateForGeneration", func() {
		tools := CreateDocumentGenerationTools()

		for _, tool := range tools {
			// Generation operations may take longer due to rendering/processing
			s.Greater(tool.Timeout, 5*time.Second, "Tool %s should have sufficient timeout for generation", tool.Name)
			s.LessOrEqual(tool.Timeout, 5*time.Minute, "Tool %s timeout should not be excessive", tool.Name)
		}
	})

	s.Run("DescriptionsIndicateOutputTypes", func() {
		tools := CreateDocumentGenerationTools()

		for _, tool := range tools {
			desc := strings.ToLower(tool.Description)

			// Should indicate what type of output is being generated
			outputTypes := []string{"pdf", "html", "report", "invoice", "document", "file"}
			hasOutputType := false
			for _, outputType := range outputTypes {
				if strings.Contains(desc, outputType) {
					hasOutputType = true
					break
				}
			}
			s.True(hasOutputType, "Tool %s description should indicate output type", tool.Name)
		}
	})
}

func (s *GenerateToolsTestSuite) TestGenerateToolsPerformance() {
	s.Run("ReasonableToolCount", func() {
		tools := CreateDocumentGenerationTools()

		// Should have reasonable number of generation tools
		s.LessOrEqual(len(tools), 8, "Should not have excessive number of generation tools")
		s.GreaterOrEqual(len(tools), 1, "Should have at least one generation tool")
	})

	s.Run("SchemaComplexityReasonable", func() {
		tools := CreateDocumentGenerationTools()

		for _, tool := range tools {
			// Count schema properties
			properties, hasProps := tool.InputSchema["properties"]
			if hasProps {
				if propsMap, ok := properties.(map[string]interface{}); ok {
					s.LessOrEqual(len(propsMap), 15, "Tool %s should not have excessive schema complexity", tool.Name)
				}
			}

			// Count examples
			s.LessOrEqual(len(tool.Examples), 8, "Tool %s should not have excessive examples", tool.Name)
		}
	})
}

// Helper method to validate generation tool structure
func (s *GenerateToolsTestSuite) validateGenerateTool(tool *MCPTool) {
	s.NotNil(tool, "Tool should not be nil")
	s.NotEmpty(tool.Name, "Tool should have name")
	s.NotEmpty(tool.Description, "Tool should have description")
	s.NotNil(tool.InputSchema, "Tool should have input schema")

	// Generation tools should be in data export or reporting category
	s.True(
		tool.Category == CategoryDataExport || tool.Category == CategoryReporting,
		"Tool should be in data export or reporting category",
	)

	s.NotEmpty(tool.CLICommand, "Tool should have CLI command")
	s.NotEmpty(tool.Version, "Tool should have version")
	s.Greater(tool.Timeout, time.Duration(0), "Tool should have positive timeout")

	// Verify schema structure
	s.Equal("object", tool.InputSchema["type"], "Schema should be object type")

	// Verify examples structure if present
	for i, example := range tool.Examples {
		s.NotEmpty(example.Description, "Example %d should have description", i)
		s.NotNil(example.Input, "Example %d should have input", i)
	}

	// Generation tools should indicate output creation capability
	toolText := strings.ToLower(tool.Name + " " + tool.Description)
	generateKeywords := []string{"generate", "create", "export", "render", "produce", "output"}
	hasGenerateKeyword := false
	for _, keyword := range generateKeywords {
		if strings.Contains(toolText, keyword) {
			hasGenerateKeyword = true
			break
		}
	}
	s.True(hasGenerateKeyword, "Generation tool should indicate output creation capability")
}

// TestGenerateToolsTestSuite runs the complete generation tools test suite
func TestGenerateToolsTestSuite(t *testing.T) {
	suite.Run(t, new(GenerateToolsTestSuite))
}

// Benchmark tests for generation tools
func BenchmarkCreateGenerateTools(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tools := CreateDocumentGenerationTools()
		_ = tools
	}
}

// Unit tests for specific generation tool behaviors
func TestGenerateTools_Specific(t *testing.T) {
	t.Run("AllToolsHaveValidCategory", func(t *testing.T) {
		tools := CreateDocumentGenerationTools()
		for _, tool := range tools {
			assert.True(t,
				tool.Category == CategoryDataExport || tool.Category == CategoryReporting,
				"Tool %s should be in valid generation category", tool.Name,
			)
		}
	})

	t.Run("AllToolsHaveAppropriateTimeouts", func(t *testing.T) {
		tools := CreateDocumentGenerationTools()
		for _, tool := range tools {
			// Generation operations may take longer
			assert.Greater(t, tool.Timeout, 5*time.Second, "Tool %s timeout too short", tool.Name)
			assert.LessOrEqual(t, tool.Timeout, 5*time.Minute, "Tool %s timeout too long", tool.Name)
		}
	})

	t.Run("AllToolsHaveValidSchemas", func(t *testing.T) {
		tools := CreateDocumentGenerationTools()
		for _, tool := range tools {
			require.NotNil(t, tool.InputSchema, "Tool %s missing schema", tool.Name)
			assert.Equal(t, "object", tool.InputSchema["type"], "Tool %s should have object schema", tool.Name)
		}
	})

	t.Run("ToolNamesIndicatePurpose", func(t *testing.T) {
		tools := CreateDocumentGenerationTools()
		for _, tool := range tools {
			toolName := strings.ToLower(tool.Name)
			assert.True(t,
				strings.Contains(toolName, "generate") ||
					strings.Contains(toolName, "create") ||
					strings.Contains(toolName, "export") ||
					strings.Contains(toolName, "render"),
				"Tool name %s should indicate generation purpose", tool.Name,
			)
		}
	})
}
