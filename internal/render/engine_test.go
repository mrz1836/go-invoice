package render

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mrz/go-invoice/internal/models"
	"github.com/stretchr/testify/suite"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	logs []LogEntry
}

type LogEntry struct {
	Level   string
	Message string
	Fields  []interface{}
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "debug", Message: msg, Fields: fields})
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "info", Message: msg, Fields: fields})
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "warn", Message: msg, Fields: fields})
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "error", Message: msg, Fields: fields})
}

// MockFileReader implements the FileReader interface for testing
type MockFileReader struct {
	files map[string][]byte
	err   error
}

func NewMockFileReader() *MockFileReader {
	return &MockFileReader{
		files: make(map[string][]byte),
	}
}

func (m *MockFileReader) AddFile(path string, content []byte) {
	m.files[path] = content
}

func (m *MockFileReader) SetError(err error) {
	m.err = err
}

func (m *MockFileReader) ReadFile(_ context.Context, path string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}

	content, exists := m.files[path]
	if !exists {
		return nil, &TemplateError{
			Type:     "file",
			Message:  "file not found",
			Template: path,
		}
	}

	return content, nil
}

func (m *MockFileReader) FileExists(_ context.Context, path string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}

	_, exists := m.files[path]
	return exists, nil
}

func (m *MockFileReader) GetFileInfo(_ context.Context, path string) (FileInfo, error) {
	if m.err != nil {
		return FileInfo{}, m.err
	}

	content, exists := m.files[path]
	if !exists {
		return FileInfo{}, &TemplateError{
			Type:     "file",
			Message:  "file not found",
			Template: path,
		}
	}

	return FileInfo{
		Name:    path,
		Size:    int64(len(content)),
		ModTime: time.Now(),
		IsDir:   false,
		Path:    path,
	}, nil
}

// RenderTestSuite defines a test suite for the template rendering system
type RenderTestSuite struct {
	suite.Suite

	engine       *HTMLTemplateEngine
	fileReader   *MockFileReader
	logger       *MockLogger
	testTemplate string
}

// SetupTest runs before each test method
func (suite *RenderTestSuite) SetupTest() {
	suite.logger = &MockLogger{}
	suite.fileReader = NewMockFileReader()
	suite.engine = NewHTMLTemplateEngine(suite.fileReader, suite.logger)

	suite.testTemplate = `<!DOCTYPE html>
<html>
<head><title>Invoice {{.Number}}</title></head>
<body>
	<h1>Invoice #{{.Number}}</h1>
	<p>Client: {{.Client.Name}}</p>
	<p>Date: {{formatDate .Date "2006-01-02"}}</p>
	<p>Total: {{formatCurrency .Total "USD"}}</p>
	<table>
		{{range .WorkItems}}
		<tr>
			<td>{{.Description}}</td>
			<td>{{.Hours}}</td>
			<td>{{formatCurrency .Total "USD"}}</td>
		</tr>
		{{end}}
	</table>
</body>
</html>`
}

// TestHTMLTemplateEngine_ParseTemplateString tests template parsing from string
func (suite *RenderTestSuite) TestHTMLTemplateEngine_ParseTemplateString() {
	ctx := context.Background()

	testCases := []struct {
		name         string
		templateName string
		content      string
		expectError  bool
	}{
		{
			name:         "ValidTemplate",
			templateName: "test",
			content:      suite.testTemplate,
		},
		{
			name:         "EmptyTemplate",
			templateName: "empty",
			content:      "",
		},
		{
			name:         "InvalidTemplate",
			templateName: "invalid",
			content:      "{{.Invalid.Template.Syntax",
			expectError:  true,
		},
		{
			name:         "TemplateWithFunctions",
			templateName: "functions",
			content:      "{{formatCurrency .Total \"USD\"}} {{upper .Client.Name}}",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.engine.ParseTemplateString(ctx, tc.templateName, tc.content)

			if tc.expectError {
				suite.Error(err)
			} else {
				suite.Require().NoError(err)

				// Verify template was stored
				template, err := suite.engine.GetTemplate(ctx, tc.templateName)
				suite.Require().NoError(err)
				suite.Equal(tc.templateName, template.Name())
			}
		})
	}
}

// TestHTMLTemplateEngine_LoadTemplate tests template loading from file
func (suite *RenderTestSuite) TestHTMLTemplateEngine_LoadTemplate() {
	ctx := context.Background()

	// Add template file to mock file reader
	suite.fileReader.AddFile("test.html", []byte(suite.testTemplate))

	err := suite.engine.LoadTemplate(ctx, "test", "test.html")
	suite.Require().NoError(err)

	// Verify template was loaded
	template, err := suite.engine.GetTemplate(ctx, "test")
	suite.Require().NoError(err)
	suite.Equal("test", template.Name())
}

// TestHTMLTemplateEngine_LoadTemplate_FileNotFound tests handling of missing files
func (suite *RenderTestSuite) TestHTMLTemplateEngine_LoadTemplate_FileNotFound() {
	ctx := context.Background()

	err := suite.engine.LoadTemplate(ctx, "missing", "missing.html")
	suite.Error(err)
}

// TestGoTemplate_Execute tests template execution
func (suite *RenderTestSuite) TestGoTemplate_Execute() {
	ctx := context.Background()

	// Parse test template
	err := suite.engine.ParseTemplateString(ctx, "test", suite.testTemplate)
	suite.Require().NoError(err)

	// Get template
	template, err := suite.engine.GetTemplate(ctx, "test")
	suite.Require().NoError(err)

	// Create test data
	invoice := suite.createTestInvoice()

	// Execute template
	var buf bytes.Buffer
	err = template.Execute(ctx, invoice, &buf)
	suite.Require().NoError(err)

	output := buf.String()
	suite.Contains(output, "Invoice #TEST-001")
	suite.Contains(output, "Test Client")
	suite.Contains(output, "1993.75 USD")
}

// TestGoTemplate_ExecuteToString tests template execution to string
func (suite *RenderTestSuite) TestGoTemplate_ExecuteToString() {
	ctx := context.Background()

	// Parse test template
	err := suite.engine.ParseTemplateString(ctx, "test", suite.testTemplate)
	suite.Require().NoError(err)

	// Get template
	template, err := suite.engine.GetTemplate(ctx, "test")
	suite.Require().NoError(err)

	// Create test data
	invoice := suite.createTestInvoice()

	// Execute template to string
	output, err := template.ExecuteToString(ctx, invoice)
	suite.Require().NoError(err)

	suite.Contains(output, "Invoice #TEST-001")
	suite.Contains(output, "Test Client")
	suite.Contains(output, "1993.75 USD")
}

// TestGoTemplate_Validate tests template validation
func (suite *RenderTestSuite) TestGoTemplate_Validate() {
	ctx := context.Background()

	testCases := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:    "ValidTemplate",
			content: "<h1>{{.Title}}</h1><p>{{.Description}}</p>",
		},
		{
			name:    "EmptyTemplate",
			content: "",
		},
		{
			name:        "InvalidSyntax",
			content:     "{{.Invalid.Template.Syntax",
			expectError: true,
		},
		{
			name:    "TemplateWithComments",
			content: "{{/* This is a comment */}}Hello {{.Name}}",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.engine.ParseTemplateString(ctx, tc.name, tc.content)

			if tc.expectError {
				suite.Error(err)
				return
			}

			suite.Require().NoError(err)

			template, err := suite.engine.GetTemplate(ctx, tc.name)
			suite.Require().NoError(err)

			err = template.Validate(ctx)
			suite.NoError(err)
		})
	}
}

// TestHTMLTemplateEngine_UnloadTemplate tests template unloading
func (suite *RenderTestSuite) TestHTMLTemplateEngine_UnloadTemplate() {
	ctx := context.Background()

	// Parse template
	err := suite.engine.ParseTemplateString(ctx, "test", suite.testTemplate)
	suite.Require().NoError(err)

	// Verify template exists
	_, err = suite.engine.GetTemplate(ctx, "test")
	suite.Require().NoError(err)

	// Unload template
	err = suite.engine.UnloadTemplate(ctx, "test")
	suite.Require().NoError(err)

	// Verify template no longer exists
	_, err = suite.engine.GetTemplate(ctx, "test")
	suite.Error(err)
}

// TestHTMLTemplateEngine_ClearCache tests cache clearing
func (suite *RenderTestSuite) TestHTMLTemplateEngine_ClearCache() {
	ctx := context.Background()

	// Parse multiple templates
	templates := []string{"test1", "test2", "test3"}
	for _, name := range templates {
		err := suite.engine.ParseTemplateString(ctx, name, suite.testTemplate)
		suite.Require().NoError(err)
	}

	// Verify templates exist
	for _, name := range templates {
		_, err := suite.engine.GetTemplate(ctx, name)
		suite.Require().NoError(err)
	}

	// Clear cache
	err := suite.engine.ClearCache(ctx)
	suite.Require().NoError(err)

	// Verify templates no longer exist
	for _, name := range templates {
		_, err := suite.engine.GetTemplate(ctx, name)
		suite.Error(err)
	}
}

// TestTemplateFunctions tests template helper functions
func (suite *RenderTestSuite) TestTemplateFunctions() {
	ctx := context.Background()

	testCases := []struct {
		name     string
		template string
		data     interface{}
		expected []string
	}{
		{
			name:     "FormatCurrency",
			template: "{{formatCurrency 1234.56 \"USD\"}}",
			data:     nil,
			expected: []string{"1234.56 USD"},
		},
		{
			name:     "FormatDate",
			template: "{{formatDate .Date \"2006-01-02\"}}",
			data:     map[string]interface{}{"Date": time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
			expected: []string{"2024-01-15"},
		},
		{
			name:     "StringFunctions",
			template: "{{upper \"hello\"}} {{lower \"WORLD\"}} {{title \"test\"}}",
			data:     nil,
			expected: []string{"HELLO", "world", "Test"},
		},
		{
			name:     "MathFunctions",
			template: "{{add 10 5}} {{multiply 3 4}}",
			data:     nil,
			expected: []string{"15", "12"},
		},
		{
			name:     "FormatFloat",
			template: "{{formatFloat 3.14159 2}}",
			data:     nil,
			expected: []string{"3.14"},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.engine.ParseTemplateString(ctx, tc.name, tc.template)
			suite.Require().NoError(err)

			template, err := suite.engine.GetTemplate(ctx, tc.name)
			suite.Require().NoError(err)

			output, err := template.ExecuteToString(ctx, tc.data)
			suite.Require().NoError(err)

			for _, expected := range tc.expected {
				suite.Contains(output, expected)
			}
		})
	}
}

// TestContextCancellation tests context cancellation handling
func (suite *RenderTestSuite) TestContextCancellation() {
	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Test various operations with canceled context
	err := suite.engine.ParseTemplateString(ctx, "test", suite.testTemplate)
	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)

	_, err = suite.engine.GetTemplate(ctx, "test")
	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)

	err = suite.engine.ClearCache(ctx)
	suite.Require().Error(err)
	suite.Equal(context.Canceled, err)
}

// TestConcurrentAccess tests concurrent template operations
func (suite *RenderTestSuite) TestConcurrentAccess() {
	ctx := context.Background()

	// Parse initial template
	err := suite.engine.ParseTemplateString(ctx, "test", suite.testTemplate)
	suite.Require().NoError(err)

	// Run concurrent operations
	const numGoroutines = 10
	errChan := make(chan error, numGoroutines*2)

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := suite.engine.GetTemplate(ctx, "test")
			errChan <- err
		}()
	}

	// Concurrent template parsing
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			templateName := fmt.Sprintf("test_%d", idx)
			err := suite.engine.ParseTemplateString(ctx, templateName, suite.testTemplate)
			errChan <- err
		}(i)
	}

	// Check results
	for i := 0; i < numGoroutines*2; i++ {
		err := <-errChan
		suite.NoError(err)
	}
}

// Helper methods

func (suite *RenderTestSuite) createTestInvoice() *models.Invoice {
	workItems := []models.WorkItem{
		{
			ID:          "work_001",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Hours:       8.0,
			Rate:        125.00,
			Description: "Web development",
			Total:       1000.00,
		},
		{
			ID:          "work_002",
			Date:        time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			Hours:       6.5,
			Rate:        125.00,
			Description: "Database optimization",
			Total:       812.50,
		},
	}

	client := models.Client{
		ID:      models.ClientID("test_client"),
		Name:    "Test Client",
		Email:   "test@example.com",
		Address: "123 Test St, Test City, TC 12345",
		Active:  true,
	}

	return &models.Invoice{
		ID:        models.InvoiceID("test_invoice"),
		Number:    "TEST-001",
		Date:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		DueDate:   time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
		Client:    client,
		WorkItems: workItems,
		Status:    models.StatusDraft,
		Subtotal:  1812.50,
		TaxRate:   0.10,
		TaxAmount: 181.25,
		Total:     1993.75,
		CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Version:   1,
	}
}

// Run the test suite
func TestRenderTestSuite(t *testing.T) {
	suite.Run(t, new(RenderTestSuite))
}

// Benchmark tests for performance verification

func BenchmarkTemplateExecution(b *testing.B) {
	logger := &MockLogger{}
	fileReader := NewMockFileReader()
	engine := NewHTMLTemplateEngine(fileReader, logger)
	ctx := context.Background()

	// Setup template
	testTemplate := `<h1>{{.Number}}</h1><p>{{.Client.Name}}</p><p>{{formatCurrency .Total "USD"}}</p>`
	err := engine.ParseTemplateString(ctx, "bench", testTemplate)
	if err != nil {
		b.Fatal(err)
	}

	template, err := engine.GetTemplate(ctx, "bench")
	if err != nil {
		b.Fatal(err)
	}

	// Create test data
	suite := &RenderTestSuite{}
	invoice := suite.createTestInvoice()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = template.ExecuteToString(ctx, invoice)
	}
}

func BenchmarkTemplateParsingAndExecution(b *testing.B) {
	logger := &MockLogger{}
	fileReader := NewMockFileReader()
	engine := NewHTMLTemplateEngine(fileReader, logger)
	ctx := context.Background()

	testTemplate := `<h1>{{.Number}}</h1><p>{{.Client.Name}}</p>`
	suite := &RenderTestSuite{}
	invoice := suite.createTestInvoice()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		templateName := fmt.Sprintf("bench_%d", i)
		err := engine.ParseTemplateString(ctx, templateName, testTemplate)
		if err != nil {
			b.Fatal(err)
		}

		template, err := engine.GetTemplate(ctx, templateName)
		if err != nil {
			b.Fatal(err)
		}

		_, _ = template.ExecuteToString(ctx, invoice)
	}
}
