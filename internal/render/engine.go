// Package render provides template rendering and PDF generation services.
package render

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/mrz/go-invoice/internal/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Logger defines the interface for logging operations
// Consumer-driven interface for logging needs in template rendering
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// FileReader defines the interface for reading template files
// Consumer-driven interface for file system operations
type FileReader interface {
	ReadFile(ctx context.Context, path string) ([]byte, error)
	FileExists(ctx context.Context, path string) (bool, error)
	GetFileInfo(ctx context.Context, path string) (FileInfo, error)
}

// FileInfo represents file metadata
type FileInfo struct {
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
	Path    string
}

// TemplateRenderer implements the InvoiceRenderer interface
type TemplateRenderer struct {
	engine    TemplateEngine
	cache     TemplateCache
	validator TemplateValidator
	logger    Logger
	options   *RendererOptions
}

// RendererOptions represents configuration options for the template renderer
type RendererOptions struct {
	TemplateDir       string        `json:"template_dir"`
	CacheSize         int           `json:"cache_size"`
	CacheExpiry       time.Duration `json:"cache_expiry"`
	EnableSecurity    bool          `json:"enable_security"`
	EnableCompression bool          `json:"enable_compression"`
	DefaultTemplate   string        `json:"default_template"`
	MaxRenderTime     time.Duration `json:"max_render_time"`
}

// NewTemplateRenderer creates a new template renderer with dependency injection
func NewTemplateRenderer(engine TemplateEngine, cache TemplateCache, validator TemplateValidator, logger Logger, options *RendererOptions) *TemplateRenderer {
	if options == nil {
		options = &RendererOptions{
			TemplateDir:       "templates",
			CacheSize:         100,
			CacheExpiry:       30 * time.Minute,
			EnableSecurity:    true,
			EnableCompression: false,
			DefaultTemplate:   "default",
			MaxRenderTime:     30 * time.Second,
		}
	}

	return &TemplateRenderer{
		engine:    engine,
		cache:     cache,
		validator: validator,
		logger:    logger,
		options:   options,
	}
}

// RenderInvoice renders an invoice to HTML using the specified template
func (r *TemplateRenderer) RenderInvoice(ctx context.Context, invoice *models.Invoice, templateName string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if templateName == "" {
		templateName = r.options.DefaultTemplate
	}

	start := time.Now()
	r.logger.Debug("starting invoice rendering", "invoice_id", invoice.ID, "template", templateName)

	// Create render context with timeout
	renderCtx, cancel := context.WithTimeout(ctx, r.options.MaxRenderTime)
	defer cancel()

	// Get or load template
	tmpl, err := r.getTemplate(renderCtx, templateName)
	if err != nil {
		return "", fmt.Errorf("failed to get template %s: %w", templateName, err)
	}

	// Validate data if security is enabled
	if r.options.EnableSecurity {
		if validationErr := r.validator.ValidateData(renderCtx, tmpl, invoice); validationErr != nil {
			return "", fmt.Errorf("data validation failed for template %s: %w", templateName, validationErr)
		}
	}

	// Render template
	content, err := tmpl.ExecuteToString(renderCtx, invoice)
	if err != nil {
		return "", fmt.Errorf("template execution failed for template %s with invoice %s: %w", templateName, invoice.ID, err)
	}

	renderTime := time.Since(start)
	r.logger.Info("invoice rendered successfully",
		"invoice_id", invoice.ID,
		"template", templateName,
		"size", len(content),
		"render_time_ms", renderTime.Milliseconds())

	return content, nil
}

// RenderData renders any data to HTML using the specified template (supports data structures)
func (r *TemplateRenderer) RenderData(ctx context.Context, data interface{}, templateName string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if templateName == "" {
		templateName = r.options.DefaultTemplate
	}

	start := time.Now()
	r.logger.Debug("starting data rendering", "template", templateName)

	// Create render context with timeout
	renderCtx, cancel := context.WithTimeout(ctx, r.options.MaxRenderTime)
	defer cancel()

	// Get or load template
	tmpl, err := r.getTemplate(renderCtx, templateName)
	if err != nil {
		return "", fmt.Errorf("failed to get template %s: %w", templateName, err)
	}

	// Validate data if security is enabled
	if r.options.EnableSecurity {
		if validationErr := r.validator.ValidateData(renderCtx, tmpl, data); validationErr != nil {
			return "", fmt.Errorf("data validation failed for template %s: %w", templateName, validationErr)
		}
	}

	// Render template
	content, err := tmpl.ExecuteToString(renderCtx, data)
	if err != nil {
		return "", fmt.Errorf("template execution failed for template %s: %w", templateName, err)
	}

	renderTime := time.Since(start)
	r.logger.Info("data rendered successfully",
		"template", templateName,
		"size", len(content),
		"render_time_ms", renderTime.Milliseconds())

	return content, nil
}

// RenderInvoiceToWriter renders an invoice to HTML and writes it to the provided writer
func (r *TemplateRenderer) RenderInvoiceToWriter(ctx context.Context, invoice *models.Invoice, templateName string, writer io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if templateName == "" {
		templateName = r.options.DefaultTemplate
	}

	start := time.Now()
	r.logger.Debug("starting invoice rendering to writer", "invoice_id", invoice.ID, "template", templateName)

	// Create render context with timeout
	renderCtx, cancel := context.WithTimeout(ctx, r.options.MaxRenderTime)
	defer cancel()

	// Get or load template
	tmpl, err := r.getTemplate(renderCtx, templateName)
	if err != nil {
		return fmt.Errorf("failed to get template %s: %w", templateName, err)
	}

	// Validate data if security is enabled
	if r.options.EnableSecurity {
		if err := r.validator.ValidateData(renderCtx, tmpl, invoice); err != nil {
			return fmt.Errorf("data validation failed for template %s: %w", templateName, err)
		}
	}

	// Render template directly to writer
	if err := tmpl.Execute(renderCtx, invoice, writer); err != nil {
		return fmt.Errorf("template execution failed for template %s with invoice %s: %w", templateName, invoice.ID, err)
	}

	renderTime := time.Since(start)
	r.logger.Info("invoice rendered to writer successfully",
		"invoice_id", invoice.ID,
		"template", templateName,
		"render_time_ms", renderTime.Milliseconds())

	return nil
}

// ValidateTemplate checks if a template is valid and can be rendered
func (r *TemplateRenderer) ValidateTemplate(ctx context.Context, templateName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tmpl, err := r.getTemplate(ctx, templateName)
	if err != nil {
		return fmt.Errorf("failed to get template %s: %w", templateName, err)
	}

	return tmpl.Validate(ctx)
}

// ListAvailableTemplates returns the names of all available templates
func (r *TemplateRenderer) ListAvailableTemplates(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// This would typically scan the template directory
	// For now, return built-in templates
	return []string{"default", "professional", "minimal"}, nil
}

// GetTemplateInfo returns metadata about a specific template
func (r *TemplateRenderer) GetTemplateInfo(ctx context.Context, templateName string) (*TemplateInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	tmpl, err := r.getTemplate(ctx, templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to get template %s: %w", templateName, err)
	}

	return tmpl.GetInfo(), nil
}

// getTemplate retrieves a template from cache or loads it
func (r *TemplateRenderer) getTemplate(ctx context.Context, templateName string) (Template, error) {
	// Try cache first
	if tmpl, err := r.cache.Get(ctx, templateName); err == nil {
		return tmpl, nil
	}

	// Load template from engine
	tmpl, err := r.engine.GetTemplate(ctx, templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %s: %w", templateName, err)
	}

	// Cache the template
	if err := r.cache.Set(ctx, templateName, tmpl); err != nil {
		r.logger.Warn("failed to cache template", "template", templateName, "error", err)
	}

	return tmpl, nil
}

// HTMLTemplateEngine implements the TemplateEngine interface using Go's html/template
type HTMLTemplateEngine struct {
	templates  map[string]*GoTemplate
	fileReader FileReader
	logger     Logger
	mu         sync.RWMutex
}

// NewHTMLTemplateEngine creates a new HTML template engine with dependency injection
func NewHTMLTemplateEngine(fileReader FileReader, logger Logger) *HTMLTemplateEngine {
	return &HTMLTemplateEngine{
		templates:  make(map[string]*GoTemplate),
		fileReader: fileReader,
		logger:     logger,
	}
}

// LoadTemplate loads a template from the specified path
func (e *HTMLTemplateEngine) LoadTemplate(ctx context.Context, name, path string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	content, err := e.fileReader.ReadFile(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", path, err)
	}

	return e.ParseTemplateString(ctx, name, string(content))
}

// ParseTemplateString parses a template from a string
func (e *HTMLTemplateEngine) ParseTemplateString(ctx context.Context, name, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create new template with useful functions
	tmpl, err := template.New(name).Funcs(e.getTemplateFunctions()).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	// Create metadata
	info := &TemplateInfo{
		Name:       name,
		SizeBytes:  int64(len(content)),
		CreatedAt:  time.Now().Format(time.RFC3339),
		ModifiedAt: time.Now().Format(time.RFC3339),
		IsBuiltIn:  false,
		IsValid:    true,
	}

	goTemplate := &GoTemplate{
		template: tmpl,
		info:     info,
		content:  content,
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.templates[name] = goTemplate

	e.logger.Info("template loaded successfully", "name", name, "size", len(content))
	return nil
}

// UnloadTemplate removes a template from memory
func (e *HTMLTemplateEngine) UnloadTemplate(ctx context.Context, name string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.templates[name]; !exists {
		return fmt.Errorf("%w: %s", models.ErrTemplateNotFound, name)
	}

	delete(e.templates, name)
	e.logger.Info("template unloaded", "name", name)
	return nil
}

// ReloadTemplate reloads a template from disk
func (e *HTMLTemplateEngine) ReloadTemplate(ctx context.Context, name string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	e.mu.RLock()
	tmpl, exists := e.templates[name]
	e.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: %s", models.ErrTemplateNotFound, name)
	}

	// If template has a path, reload from file
	if tmpl.info.Path != "" {
		return e.LoadTemplate(ctx, name, tmpl.info.Path)
	}

	return fmt.Errorf("%w: %s", models.ErrTemplateCannotReload, name)
}

// GetTemplate returns a loaded template by name
func (e *HTMLTemplateEngine) GetTemplate(ctx context.Context, name string) (Template, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	tmpl, exists := e.templates[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", models.ErrTemplateNotFound, name)
	}

	return tmpl, nil
}

// ClearCache clears all cached templates
func (e *HTMLTemplateEngine) ClearCache(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	count := len(e.templates)
	e.templates = make(map[string]*GoTemplate)

	e.logger.Info("template cache cleared", "count", count)
	return nil
}

// getCurrencySymbol converts currency codes to symbols
func getCurrencySymbol(currency string) string {
	switch currency {
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	case "CAD":
		return "C$"
	case "AUD":
		return "A$"
	default:
		return currency // fallback to currency code
	}
}

// getTemplateFunctions returns useful template functions
func (e *HTMLTemplateEngine) getTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"formatCurrency": func(amount float64, currency string) string {
			symbol := getCurrencySymbol(currency)
			return fmt.Sprintf("%s%.2f", symbol, amount)
		},
		"formatDate": func(t time.Time, format string) string {
			if format == "" {
				format = "2006-01-02"
			}
			return t.Format(format)
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": cases.Title(language.English).String,
		"add": func(a, b float64) float64 {
			return a + b
		},
		"multiply": func(a, b float64) float64 {
			return a * b
		},
		"formatFloat": func(f float64, precision interface{}) string {
			var p int
			switch v := precision.(type) {
			case int:
				p = v
			case float64:
				p = int(v)
			default:
				p = 2 // default precision
			}
			return fmt.Sprintf("%.*f", p, f)
		},
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil || value == "" {
				return defaultValue
			}
			return value
		},
		"minDate": func(workItems interface{}) time.Time {
			return getMinDateFromWorkItems(workItems)
		},
		"maxDate": func(workItems interface{}) time.Time {
			return getMaxDateFromWorkItems(workItems)
		},
	}
}

// getMinDateFromWorkItems finds the earliest date from work items
func getMinDateFromWorkItems(workItems interface{}) time.Time {
	switch items := workItems.(type) {
	case []models.WorkItem:
		if len(items) == 0 {
			return time.Time{}
		}
		minDate := items[0].Date
		for _, item := range items[1:] {
			if item.Date.Before(minDate) {
				minDate = item.Date
			}
		}
		return minDate
	default:
		return time.Time{}
	}
}

// getMaxDateFromWorkItems finds the latest date from work items
func getMaxDateFromWorkItems(workItems interface{}) time.Time {
	switch items := workItems.(type) {
	case []models.WorkItem:
		if len(items) == 0 {
			return time.Time{}
		}
		maxDate := items[0].Date
		for _, item := range items[1:] {
			if item.Date.After(maxDate) {
				maxDate = item.Date
			}
		}
		return maxDate
	default:
		return time.Time{}
	}
}

// GoTemplate implements the Template interface using Go's html/template
type GoTemplate struct {
	template *template.Template
	info     *TemplateInfo
	content  string
	mu       sync.RWMutex
}

// Execute renders the template with the provided data
func (t *GoTemplate) Execute(ctx context.Context, data interface{}, writer io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.template.Execute(writer, data)
}

// ExecuteToString renders the template to a string
func (t *GoTemplate) ExecuteToString(ctx context.Context, data interface{}) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	var buf bytes.Buffer
	if err := t.Execute(ctx, data, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Name returns the template name
func (t *GoTemplate) Name() string {
	return t.info.Name
}

// GetInfo returns template metadata
func (t *GoTemplate) GetInfo() *TemplateInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to prevent mutation
	info := *t.info
	return &info
}

// Validate checks if the template is valid
func (t *GoTemplate) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Basic validation - try to parse the template
	_, err := template.New("validation").Parse(t.content)
	if err != nil {
		return &TemplateError{
			Type:     "syntax",
			Message:  err.Error(),
			Template: t.info.Name,
		}
	}

	return nil
}
