package render

import (
	"context"
	"fmt"
	"io"

	"github.com/mrz/go-invoice/internal/models"
)

// InvoiceRenderer defines the interface for rendering invoices to HTML
// Consumer-driven interface focusing on invoice generation needs
type InvoiceRenderer interface {
	// RenderInvoice renders an invoice to HTML using the specified template
	// Returns the rendered HTML content as a string
	RenderInvoice(ctx context.Context, invoice *models.Invoice, templateName string) (string, error)

	// RenderInvoiceToWriter renders an invoice to HTML and writes it to the provided writer
	// Useful for streaming large invoices or writing directly to files
	RenderInvoiceToWriter(ctx context.Context, invoice *models.Invoice, templateName string, writer io.Writer) error

	// ValidateTemplate checks if a template is valid and can be rendered
	// Returns detailed validation errors if template is invalid
	ValidateTemplate(ctx context.Context, templateName string) error

	// ListAvailableTemplates returns the names of all available templates
	ListAvailableTemplates(ctx context.Context) ([]string, error)

	// GetTemplateInfo returns metadata about a specific template
	GetTemplateInfo(ctx context.Context, templateName string) (*TemplateInfo, error)
}

// TemplateEngine defines the interface for managing template operations
// Consumer-driven interface for template system management
type TemplateEngine interface {
	// LoadTemplate loads a template from the specified path
	// Returns error if template cannot be loaded or parsed
	LoadTemplate(ctx context.Context, name, path string) error

	// ParseTemplateString parses a template from a string
	// Useful for dynamic template creation or testing
	ParseTemplateString(ctx context.Context, name, content string) error

	// UnloadTemplate removes a template from memory
	UnloadTemplate(ctx context.Context, name string) error

	// ReloadTemplate reloads a template from disk
	// Useful for development and template updates
	ReloadTemplate(ctx context.Context, name string) error

	// GetTemplate returns a loaded template by name
	GetTemplate(ctx context.Context, name string) (Template, error)

	// ClearCache clears all cached templates
	ClearCache(ctx context.Context) error
}

// Template defines the interface for individual templates
// Provides rendering capabilities for loaded templates
type Template interface {
	// Execute renders the template with the provided data
	Execute(ctx context.Context, data interface{}, writer io.Writer) error

	// ExecuteToString renders the template to a string
	ExecuteToString(ctx context.Context, data interface{}) (string, error)

	// Name returns the template name
	Name() string

	// GetInfo returns template metadata
	GetInfo() *TemplateInfo

	// Validate checks if the template is valid
	Validate(ctx context.Context) error
}

// TemplateCache defines the interface for template caching
// Consumer-driven interface for performance optimization
type TemplateCache interface {
	// Get retrieves a template from cache
	// Returns ErrCacheMiss if template is not cached
	Get(ctx context.Context, name string) (Template, error)

	// Set stores a template in cache
	Set(ctx context.Context, name string, template Template) error

	// Delete removes a template from cache
	Delete(ctx context.Context, name string) error

	// Clear removes all templates from cache
	Clear(ctx context.Context) error

	// GetStats returns cache statistics
	GetStats(ctx context.Context) (*CacheStats, error)

	// GetSize returns the number of cached templates
	GetSize(ctx context.Context) (int, error)
}

// TemplateValidator defines the interface for template validation
// Consumer-driven interface for security and correctness validation
type TemplateValidator interface {
	// ValidateTemplate performs comprehensive template validation
	// Checks for security issues, syntax errors, and best practices
	ValidateTemplate(ctx context.Context, template Template) error

	// ValidateTemplateString validates a template string
	ValidateTemplateString(ctx context.Context, content string) error

	// CheckSecurity performs security-specific validation
	// Checks for potential code injection and unsafe operations
	CheckSecurity(ctx context.Context, template Template) error

	// ValidateData validates that data can be safely used with template
	ValidateData(ctx context.Context, template Template, data interface{}) error
}

// TemplateInfo represents metadata about a template
type TemplateInfo struct {
	Name        string            `json:"name"`
	Path        string            `json:"path,omitempty"`
	Description string            `json:"description,omitempty"`
	Version     string            `json:"version,omitempty"`
	Author      string            `json:"author,omitempty"`
	CreatedAt   string            `json:"created_at,omitempty"`
	ModifiedAt  string            `json:"modified_at,omitempty"`
	SizeBytes   int64             `json:"size_bytes"`
	Variables   []string          `json:"variables,omitempty"`  // Required template variables
	Includes    []string          `json:"includes,omitempty"`   // Other templates this depends on
	Tags        []string          `json:"tags,omitempty"`       // Template tags/categories
	Metadata    map[string]string `json:"metadata,omitempty"`   // Additional metadata
	IsBuiltIn   bool              `json:"is_built_in"`          // Whether this is a built-in template
	IsValid     bool              `json:"is_valid"`             // Whether template passed validation
	LastError   string            `json:"last_error,omitempty"` // Last validation error
}

// CacheStats represents template cache statistics
type CacheStats struct {
	Size         int     `json:"size"`          // Number of cached templates
	HitCount     int64   `json:"hit_count"`     // Cache hits
	MissCount    int64   `json:"miss_count"`    // Cache misses
	HitRate      float64 `json:"hit_rate"`      // Hit rate percentage
	MemoryUsage  int64   `json:"memory_usage"`  // Estimated memory usage in bytes
	LastAccessed string  `json:"last_accessed"` // Last cache access time
}

// RenderOptions represents options for rendering operations
type RenderOptions struct {
	TemplateName    string            `json:"template_name"`            // Template to use
	OutputFormat    string            `json:"output_format"`            // "html", "pdf", etc.
	Minify          bool              `json:"minify"`                   // Whether to minify output
	PrettyPrint     bool              `json:"pretty_print"`             // Whether to format output
	Variables       map[string]string `json:"variables,omitempty"`      // Template variables
	IncludeStyles   bool              `json:"include_styles"`           // Whether to include CSS
	Theme           string            `json:"theme,omitempty"`          // Theme name
	Language        string            `json:"language,omitempty"`       // Output language
	CustomCSS       string            `json:"custom_css,omitempty"`     // Custom CSS to include
	WatermarkText   string            `json:"watermark_text,omitempty"` // Watermark text
	ShowLineNumbers bool              `json:"show_line_numbers"`        // Show line numbers in output
}

// RenderResult represents the result of a rendering operation
type RenderResult struct {
	Content          string            `json:"content"`                     // Rendered content
	Size             int64             `json:"size"`                        // Content size in bytes
	RenderTimeMs     float64           `json:"render_time_ms"`              // Rendering time in milliseconds
	TemplateName     string            `json:"template_name"`               // Template used
	Variables        map[string]string `json:"variables,omitempty"`         // Variables used
	Warnings         []string          `json:"warnings,omitempty"`          // Non-fatal warnings
	CacheHit         bool              `json:"cache_hit"`                   // Whether template was cached
	CompressionRatio float64           `json:"compression_ratio,omitempty"` // If content was compressed
}

// TemplateError represents template-specific errors
type TemplateError struct {
	Type       string `json:"type"`       // Error type: "syntax", "security", "validation", etc.
	Message    string `json:"message"`    // Human-readable error message
	Template   string `json:"template"`   // Template name where error occurred
	Line       int    `json:"line"`       // Line number (if applicable)
	Column     int    `json:"column"`     // Column number (if applicable)
	Context    string `json:"context"`    // Context around the error
	Suggestion string `json:"suggestion"` // Suggested fix
}

// Error implements the error interface
func (e *TemplateError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("template error in %s at line %d: %s", e.Template, e.Line, e.Message)
	}
	return fmt.Sprintf("template error in %s: %s", e.Template, e.Message)
}

// IsSecurityError returns true if this is a security-related error
func (e *TemplateError) IsSecurityError() bool {
	return e.Type == "security"
}

// IsSyntaxError returns true if this is a syntax error
func (e *TemplateError) IsSyntaxError() bool {
	return e.Type == "syntax"
}
