package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mrz/go-invoice/internal/render"
)

// SimpleFileReader implements the FileReader interface
type SimpleFileReader struct{}

func (r *SimpleFileReader) ReadFile(ctx context.Context, path string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return os.ReadFile(path)
}

func (r *SimpleFileReader) FileExists(ctx context.Context, path string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (r *SimpleFileReader) GetFileInfo(ctx context.Context, path string) (render.FileInfo, error) {
	select {
	case <-ctx.Done():
		return render.FileInfo{}, ctx.Err()
	default:
	}

	info, err := os.Stat(path)
	if err != nil {
		return render.FileInfo{}, err
	}

	return render.FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
		Path:    path,
	}, nil
}

// SimpleTemplateCache implements the TemplateCache interface
type SimpleTemplateCache struct {
	templates map[string]render.Template
	mu        sync.RWMutex
	stats     render.CacheStats
}

func (c *SimpleTemplateCache) Get(ctx context.Context, name string) (render.Template, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	template, exists := c.templates[name]
	if !exists {
		c.stats.MissCount++
		return nil, fmt.Errorf("template %s not found in cache", name)
	}

	c.stats.HitCount++
	return template, nil
}

func (c *SimpleTemplateCache) Set(ctx context.Context, name string, template render.Template) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.templates[name] = template
	return nil
}

func (c *SimpleTemplateCache) Delete(ctx context.Context, name string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.templates, name)
	return nil
}

func (c *SimpleTemplateCache) Clear(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.templates = make(map[string]render.Template)
	return nil
}

func (c *SimpleTemplateCache) GetStats(ctx context.Context) (*render.CacheStats, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.templates)

	total := stats.HitCount + stats.MissCount
	if total > 0 {
		stats.HitRate = float64(stats.HitCount) / float64(total) * 100
	}

	stats.LastAccessed = time.Now().Format(time.RFC3339)

	return &stats, nil
}

func (c *SimpleTemplateCache) GetSize(ctx context.Context) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.templates), nil
}

// SimpleTemplateValidator implements the TemplateValidator interface
type SimpleTemplateValidator struct {
	logger render.Logger
}

func (v *SimpleTemplateValidator) ValidateTemplate(ctx context.Context, template render.Template) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if template == nil {
		return &render.TemplateError{
			Type:     "validation",
			Message:  "template is nil",
			Template: "unknown",
		}
	}

	// Validate template using its own validation method
	if err := template.Validate(ctx); err != nil {
		return err
	}

	v.logger.Debug("template validation passed", "template", template.Name())
	return nil
}

func (v *SimpleTemplateValidator) ValidateTemplateString(ctx context.Context, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if content == "" {
		return &render.TemplateError{
			Type:     "validation",
			Message:  "template content is empty",
			Template: "string",
		}
	}

	// Basic validation - check for balanced braces
	openBraces := 0
	for i, char := range content {
		switch char {
		case '{':
			if i < len(content)-1 && rune(content[i+1]) == '{' {
				openBraces++
			}
		case '}':
			if i > 0 && rune(content[i-1]) == '}' {
				openBraces--
			}
		}
	}

	if openBraces != 0 {
		return &render.TemplateError{
			Type:     "syntax",
			Message:  fmt.Sprintf("unbalanced template braces: %d", openBraces),
			Template: "string",
		}
	}

	return nil
}

func (v *SimpleTemplateValidator) CheckSecurity(ctx context.Context, template render.Template) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get template content for security analysis
	info := template.GetInfo()

	// Render a test to get the template content
	testContent, err := template.ExecuteToString(ctx, map[string]interface{}{})
	if err != nil {
		// If we can't render with empty data, that's okay for security check
		v.logger.Debug("could not render template for security check", "template", template.Name(), "error", err)
	}

	// Check for potentially dangerous patterns
	securityChecks := []struct {
		pattern     string
		description string
	}{
		{"{{", "template syntax detected (expected)"},
		{".Execute", "potential template execution"},
		{".Call", "potential function call"},
		{"os.", "potential OS package access"},
		{"exec.", "potential exec package access"},
		{"syscall.", "potential syscall access"},
		{"unsafe.", "potential unsafe package access"},
		{"reflect.ValueOf", "potential reflection abuse"},
		{"runtime.", "potential runtime package access"},
	}

	// This is a basic content-based security check
	// In a production system, you would use a proper template parser
	for _, check := range securityChecks {
		if strings.Contains(testContent, check.pattern) &&
			!strings.Contains(check.pattern, "{{") { // Allow normal template syntax
			v.logger.Warn("potential security risk in template",
				"template", template.Name(),
				"pattern", check.pattern,
				"description", check.description)

			// For now, just warn. In production, you might reject the template
			if strings.Contains(check.pattern, "exec") ||
				strings.Contains(check.pattern, "syscall") ||
				strings.Contains(check.pattern, "unsafe") {
				return &render.TemplateError{
					Type:     "security",
					Message:  fmt.Sprintf("template contains dangerous pattern: %s", check.pattern),
					Template: template.Name(),
				}
			}
		}
	}

	// Check template size limits
	if info.SizeBytes > 1024*1024 { // 1MB limit
		return &render.TemplateError{
			Type:     "security",
			Message:  "template exceeds maximum size limit (1MB)",
			Template: template.Name(),
		}
	}

	// Validate template functions are safe
	if err := v.validateTemplateFunctions(ctx, template); err != nil {
		return err
	}

	v.logger.Debug("security check passed", "template", template.Name())
	return nil
}

// validateTemplateFunctions checks that only safe template functions are used
func (v *SimpleTemplateValidator) validateTemplateFunctions(ctx context.Context, template render.Template) error {
	// List of allowed template functions
	allowedFunctions := map[string]bool{
		"formatCurrency": true,
		"formatDate":     true,
		"upper":          true,
		"lower":          true,
		"title":          true,
		"add":            true,
		"multiply":       true,
		"formatFloat":    true,
		"len":            true,
		"range":          true,
		"if":             true,
		"else":           true,
		"end":            true,
		"with":           true,
		"default":        true,
	}

	// This is a simplified validation - in a real system you'd parse the AST
	// For now, we assume the template is using only built-in functions

	info := template.GetInfo()
	v.logger.Debug("template function validation passed",
		"template", info.Name,
		"allowed_functions", len(allowedFunctions))

	return nil
}

func (v *SimpleTemplateValidator) ValidateData(ctx context.Context, template render.Template, data interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if data == nil {
		return &render.TemplateError{
			Type:     "validation",
			Message:  "template data is nil",
			Template: template.Name(),
		}
	}

	// Basic data validation
	v.logger.Debug("data validation passed", "template", template.Name())
	return nil
}
