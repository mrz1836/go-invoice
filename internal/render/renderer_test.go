package render

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz/go-invoice/internal/models"
)

// MockTemplateCache implements TemplateCache for testing
type MockTemplateCache struct {
	templates map[string]Template
	stats     *CacheStats
}

func NewMockTemplateCache() *MockTemplateCache {
	return &MockTemplateCache{
		templates: make(map[string]Template),
		stats: &CacheStats{
			Size:      0,
			HitCount:  0,
			MissCount: 0,
			HitRate:   0.0,
		},
	}
}

func (c *MockTemplateCache) Get(ctx context.Context, name string) (Template, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	tmpl, exists := c.templates[name]
	if !exists {
		c.stats.MissCount++
		return nil, fmt.Errorf("template not found in cache")
	}

	c.stats.HitCount++
	c.updateHitRate()
	return tmpl, nil
}

func (c *MockTemplateCache) Set(ctx context.Context, name string, template Template) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.templates[name] = template
	c.stats.Size = len(c.templates)
	c.updateHitRate()
	return nil
}

func (c *MockTemplateCache) Delete(ctx context.Context, name string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	delete(c.templates, name)
	c.stats.Size = len(c.templates)
	return nil
}

func (c *MockTemplateCache) Clear(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.templates = make(map[string]Template)
	c.stats.Size = 0
	return nil
}

func (c *MockTemplateCache) GetStats(ctx context.Context) (*CacheStats, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return c.stats, nil
}

func (c *MockTemplateCache) GetSize(ctx context.Context) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	return len(c.templates), nil
}

func (c *MockTemplateCache) updateHitRate() {
	total := c.stats.HitCount + c.stats.MissCount
	if total > 0 {
		c.stats.HitRate = float64(c.stats.HitCount) / float64(total) * 100
	}
}

// MockTemplateValidator implements TemplateValidator for testing
type MockTemplateValidator struct {
	shouldError        bool
	shouldSecError     bool
	shouldDataError    bool
	validationDelay    time.Duration
	validatedTemplates []string
}

func NewMockTemplateValidator() *MockTemplateValidator {
	return &MockTemplateValidator{
		validatedTemplates: make([]string, 0),
	}
}

func (v *MockTemplateValidator) SetError(shouldError bool) {
	v.shouldError = shouldError
}

func (v *MockTemplateValidator) SetSecurityError(shouldSecError bool) {
	v.shouldSecError = shouldSecError
}

func (v *MockTemplateValidator) SetDataError(shouldDataError bool) {
	v.shouldDataError = shouldDataError
}

func (v *MockTemplateValidator) SetValidationDelay(delay time.Duration) {
	v.validationDelay = delay
}

func (v *MockTemplateValidator) ValidateTemplate(ctx context.Context, template Template) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if v.validationDelay > 0 {
		select {
		case <-time.After(v.validationDelay):
			// Delay completed
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	v.validatedTemplates = append(v.validatedTemplates, template.Name())

	if v.shouldError {
		return &TemplateError{
			Type:     "validation",
			Message:  "template validation failed",
			Template: template.Name(),
		}
	}

	return nil
}

func (v *MockTemplateValidator) ValidateTemplateString(ctx context.Context, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if v.shouldError {
		return &TemplateError{
			Type:    "validation",
			Message: "template string validation failed",
		}
	}

	return nil
}

func (v *MockTemplateValidator) CheckSecurity(ctx context.Context, template Template) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if v.shouldSecError {
		return &TemplateError{
			Type:     "security",
			Message:  "security validation failed",
			Template: template.Name(),
		}
	}

	return nil
}

func (v *MockTemplateValidator) ValidateData(ctx context.Context, template Template, data interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if v.shouldDataError {
		return &TemplateError{
			Type:     "data",
			Message:  "data validation failed",
			Template: template.Name(),
		}
	}

	return nil
}

// TemplateRendererTestSuite tests the TemplateRenderer
type TemplateRendererTestSuite struct {
	suite.Suite
	renderer  *TemplateRenderer
	engine    *HTMLTemplateEngine
	cache     *MockTemplateCache
	validator *MockTemplateValidator
	logger    *MockLogger
	invoice   *models.Invoice
}

func (suite *TemplateRendererTestSuite) SetupTest() {
	suite.logger = &MockLogger{}
	fileReader := NewMockFileReader()
	suite.engine = NewHTMLTemplateEngine(fileReader, suite.logger)
	suite.cache = NewMockTemplateCache()
	suite.validator = NewMockTemplateValidator()

	options := &RendererOptions{
		TemplateDir:       "templates",
		CacheSize:         10,
		CacheExpiry:       5 * time.Minute,
		EnableSecurity:    true,
		EnableCompression: false,
		DefaultTemplate:   "default",
		MaxRenderTime:     5 * time.Second,
	}

	suite.renderer = NewTemplateRenderer(suite.engine, suite.cache, suite.validator, suite.logger, options)
	suite.invoice = suite.createTestInvoice()

	// Setup default template (without functions for validation compatibility)
	testTemplate := `<h1>Invoice {{.Number}}</h1><p>{{.Client.Name}}</p><p>Total: {{.Total}}</p>`
	err := suite.engine.ParseTemplateString(context.Background(), "default", testTemplate)
	suite.Require().NoError(err)

	// Setup template with functions for rendering tests
	funcTemplate := `<h1>Invoice {{.Number}}</h1><p>{{.Client.Name}}</p><p>{{formatCurrency .Total "USD"}}</p>`
	err = suite.engine.ParseTemplateString(context.Background(), "with_functions", funcTemplate)
	suite.Require().NoError(err)
}

func (suite *TemplateRendererTestSuite) TestRenderInvoice_Success() {
	ctx := context.Background()

	// Use template with functions for this test
	result, err := suite.renderer.RenderInvoice(ctx, suite.invoice, "with_functions")
	suite.Require().NoError(err)

	suite.Contains(result, "Invoice TEST-001")
	suite.Contains(result, "Test Client")
	suite.Contains(result, "1993.75 USD")
}

func (suite *TemplateRendererTestSuite) TestRenderInvoice_DefaultTemplate() {
	ctx := context.Background()

	// Test with empty template name - should use default
	result, err := suite.renderer.RenderInvoice(ctx, suite.invoice, "")
	suite.Require().NoError(err)

	suite.Contains(result, "Invoice TEST-001")
}

func (suite *TemplateRendererTestSuite) TestRenderInvoice_TemplateNotFound() {
	ctx := context.Background()

	_, err := suite.renderer.RenderInvoice(ctx, suite.invoice, "nonexistent")
	suite.Require().Error(err)
	suite.Contains(err.Error(), "failed to get template")
}

func (suite *TemplateRendererTestSuite) TestRenderInvoice_SecurityValidationError() {
	ctx := context.Background()

	// Enable data validation error
	suite.validator.SetDataError(true)

	_, err := suite.renderer.RenderInvoice(ctx, suite.invoice, "default")
	suite.Error(err)
	assert.Contains(suite.T(), err.Error(), "data validation failed")
}

func (suite *TemplateRendererTestSuite) TestRenderInvoice_SecurityDisabled() {
	ctx := context.Background()

	// Disable security
	suite.renderer.options.EnableSecurity = false
	suite.validator.SetDataError(true) // This should be ignored

	result, err := suite.renderer.RenderInvoice(ctx, suite.invoice, "default")
	suite.Require().NoError(err)
	assert.Contains(suite.T(), result, "Invoice TEST-001")
}

func (suite *TemplateRendererTestSuite) TestRenderInvoice_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := suite.renderer.RenderInvoice(ctx, suite.invoice, "default")
	suite.Error(err)
	suite.Equal(context.Canceled, err)
}

func (suite *TemplateRendererTestSuite) TestRenderInvoice_Timeout() {
	ctx := context.Background()

	// Set a very short timeout
	suite.renderer.options.MaxRenderTime = 1 * time.Nanosecond

	// Add delay to validation to trigger timeout
	suite.validator.SetValidationDelay(10 * time.Millisecond)

	_, err := suite.renderer.RenderInvoice(ctx, suite.invoice, "default")
	suite.Error(err)
	// Error should be related to context timeout, not cancellation
}

func (suite *TemplateRendererTestSuite) TestRenderInvoiceToWriter_Success() {
	ctx := context.Background()
	var buf bytes.Buffer

	// Use simple template without functions
	err := suite.renderer.RenderInvoiceToWriter(ctx, suite.invoice, "default", &buf)
	suite.Require().NoError(err)

	result := buf.String()
	assert.Contains(suite.T(), result, "Invoice TEST-001")
	assert.Contains(suite.T(), result, "Test Client")
}

func (suite *TemplateRendererTestSuite) TestRenderInvoiceToWriter_TemplateNotFound() {
	ctx := context.Background()
	var buf bytes.Buffer

	err := suite.renderer.RenderInvoiceToWriter(ctx, suite.invoice, "nonexistent", &buf)
	suite.Error(err)
	assert.Contains(suite.T(), err.Error(), "failed to get template")
}

func (suite *TemplateRendererTestSuite) TestValidateTemplate_Success() {
	ctx := context.Background()

	err := suite.renderer.ValidateTemplate(ctx, "default")
	suite.NoError(err)
}

func (suite *TemplateRendererTestSuite) TestValidateTemplate_NotFound() {
	ctx := context.Background()

	err := suite.renderer.ValidateTemplate(ctx, "nonexistent")
	suite.Error(err)
	assert.Contains(suite.T(), err.Error(), "failed to get template")
}

func (suite *TemplateRendererTestSuite) TestListAvailableTemplates() {
	ctx := context.Background()

	templates, err := suite.renderer.ListAvailableTemplates(ctx)
	suite.Require().NoError(err)

	// Should return built-in templates
	expected := []string{"default", "professional", "minimal"}
	suite.Equal(expected, templates)
}

func (suite *TemplateRendererTestSuite) TestGetTemplateInfo() {
	ctx := context.Background()

	info, err := suite.renderer.GetTemplateInfo(ctx, "default")
	suite.Require().NoError(err)

	suite.Equal("default", info.Name)
	suite.Positive(info.SizeBytes)
}

func (suite *TemplateRendererTestSuite) TestGetTemplateInfo_NotFound() {
	ctx := context.Background()

	_, err := suite.renderer.GetTemplateInfo(ctx, "nonexistent")
	suite.Require().Error(err)
	assert.Contains(suite.T(), err.Error(), "failed to get template")
}

func (suite *TemplateRendererTestSuite) TestCacheIntegration() {
	ctx := context.Background()

	// First call should load template and cache it
	_, err := suite.renderer.RenderInvoice(ctx, suite.invoice, "default")
	suite.Require().NoError(err)

	// Check cache stats
	stats, err := suite.cache.GetStats(ctx)
	suite.Require().NoError(err)

	// Should have at least one cache operation
	suite.True(stats.HitCount > 0 || stats.MissCount > 0)
}

func (suite *TemplateRendererTestSuite) TestNewTemplateRenderer_DefaultOptions() {
	renderer := NewTemplateRenderer(suite.engine, suite.cache, suite.validator, suite.logger, nil)

	suite.NotNil(renderer.options)
	assert.Equal(suite.T(), "templates", renderer.options.TemplateDir)
	assert.Equal(suite.T(), 100, renderer.options.CacheSize)
	assert.Equal(suite.T(), 30*time.Minute, renderer.options.CacheExpiry)
	assert.True(suite.T(), renderer.options.EnableSecurity)
	suite.False(renderer.options.EnableCompression)
	assert.Equal(suite.T(), "default", renderer.options.DefaultTemplate)
	assert.Equal(suite.T(), 30*time.Second, renderer.options.MaxRenderTime)
}

func (suite *TemplateRendererTestSuite) createTestInvoice() *models.Invoice {
	workItems := []models.WorkItem{
		{
			ID:          "work_001",
			Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Hours:       8.0,
			Rate:        125.00,
			Description: "Web development",
			Total:       1000.00,
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
		Subtotal:  1000.00,
		TaxRate:   0.10,
		TaxAmount: 100.00,
		Total:     1993.75, // Note: intentionally different from subtotal+tax for testing
		CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		Version:   1,
	}
}

func TestTemplateRendererTestSuite(t *testing.T) {
	suite.Run(t, new(TemplateRendererTestSuite))
}

// TestTemplateError tests the TemplateError type
func TestTemplateError(t *testing.T) {
	t.Run("ErrorWithLineNumber", func(t *testing.T) {
		err := &TemplateError{
			Type:     "syntax",
			Message:  "unexpected token",
			Template: "test.html",
			Line:     10,
			Column:   5,
		}

		expected := "template error in test.html at line 10: unexpected token"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("ErrorWithoutLineNumber", func(t *testing.T) {
		err := &TemplateError{
			Type:     "validation",
			Message:  "template not found",
			Template: "missing.html",
		}

		expected := "template error in missing.html: template not found"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("IsSecurityError", func(t *testing.T) {
		secErr := &TemplateError{Type: "security", Message: "unsafe operation"}
		syntaxErr := &TemplateError{Type: "syntax", Message: "parse error"}

		assert.True(t, secErr.IsSecurityError())
		assert.False(t, syntaxErr.IsSecurityError())
	})

	t.Run("IsSyntaxError", func(t *testing.T) {
		secErr := &TemplateError{Type: "security", Message: "unsafe operation"}
		syntaxErr := &TemplateError{Type: "syntax", Message: "parse error"}

		assert.False(t, secErr.IsSyntaxError())
		assert.True(t, syntaxErr.IsSyntaxError())
	})
}

// TestSecurityValidation tests security-related functionality
func TestSecurityValidation(t *testing.T) {
	t.Run("XSSPrevention", func(t *testing.T) {
		logger := &MockLogger{}
		fileReader := NewMockFileReader()
		engine := NewHTMLTemplateEngine(fileReader, logger)
		ctx := context.Background()

		// Template that outputs user data (should be HTML escaped)
		template := `<div>{{.Client.Name}}</div>`
		err := engine.ParseTemplateString(ctx, "xss_test", template)
		require.NoError(t, err)

		tmpl, err := engine.GetTemplate(ctx, "xss_test")
		require.NoError(t, err)

		// Create invoice with potentially dangerous content
		suite := &TemplateRendererTestSuite{}
		invoice := suite.createTestInvoice()
		invoice.Client.Name = "<script>alert('xss')</script>"

		result, err := tmpl.ExecuteToString(ctx, invoice)
		require.NoError(t, err)

		// Should be HTML escaped
		assert.Contains(t, result, "&lt;script&gt;")
		assert.NotContains(t, result, "<script>")
	})

	t.Run("SafeTemplateFunctions", func(t *testing.T) {
		logger := &MockLogger{}
		fileReader := NewMockFileReader()
		engine := NewHTMLTemplateEngine(fileReader, logger)
		ctx := context.Background()

		// Test that template functions work safely
		template := `{{formatCurrency 1234.56 "USD"}} {{upper "hello"}} {{add 5 3}}`
		err := engine.ParseTemplateString(ctx, "func_test", template)
		require.NoError(t, err)

		tmpl, err := engine.GetTemplate(ctx, "func_test")
		require.NoError(t, err)

		result, err := tmpl.ExecuteToString(ctx, nil)
		require.NoError(t, err)

		assert.Contains(t, result, "1234.56 USD")
		assert.Contains(t, result, "HELLO")
		assert.Contains(t, result, "8")
	})
}

// Benchmark tests for TemplateRenderer
func BenchmarkTemplateRenderer_RenderInvoice(b *testing.B) {
	logger := &MockLogger{}
	fileReader := NewMockFileReader()
	engine := NewHTMLTemplateEngine(fileReader, logger)
	cache := NewMockTemplateCache()
	validator := NewMockTemplateValidator()

	renderer := NewTemplateRenderer(engine, cache, validator, logger, nil)
	ctx := context.Background()

	// Setup template
	testTemplate := `<h1>{{.Number}}</h1><p>{{.Client.Name}}</p><p>{{formatCurrency .Total "USD"}}</p>`
	err := engine.ParseTemplateString(ctx, "default", testTemplate)
	if err != nil {
		b.Fatal(err)
	}

	// Create test invoice
	suite := &TemplateRendererTestSuite{}
	invoice := suite.createTestInvoice()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = renderer.RenderInvoice(ctx, invoice, "default")
	}
}

func BenchmarkTemplateRenderer_RenderToWriter(b *testing.B) {
	logger := &MockLogger{}
	fileReader := NewMockFileReader()
	engine := NewHTMLTemplateEngine(fileReader, logger)
	cache := NewMockTemplateCache()
	validator := NewMockTemplateValidator()

	renderer := NewTemplateRenderer(engine, cache, validator, logger, nil)
	ctx := context.Background()

	// Setup template
	testTemplate := `<h1>{{.Number}}</h1><p>{{.Client.Name}}</p>`
	err := engine.ParseTemplateString(ctx, "default", testTemplate)
	if err != nil {
		b.Fatal(err)
	}

	// Create test invoice
	suite := &TemplateRendererTestSuite{}
	invoice := suite.createTestInvoice()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = renderer.RenderInvoiceToWriter(ctx, invoice, "default", &buf)
	}
}
