package executor

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Parser errors
var (
	ErrInvalidJSON      = errors.New("invalid JSON output")
	ErrInvalidTable     = errors.New("invalid table format")
	ErrInvalidKeyValue  = errors.New("invalid key-value format")
	ErrParsingFailed    = errors.New("output parsing failed")
	ErrNoDataFound      = errors.New("no data found in output")
	ErrUnexpectedFormat = errors.New("unexpected output format")
	ErrCommandExitCode  = errors.New("command failed with exit code")
)

// DefaultOutputParser implements OutputParser with support for multiple formats.
type DefaultOutputParser struct {
	logger Logger
}

// NewDefaultOutputParser creates a new output parser.
func NewDefaultOutputParser(logger Logger) *DefaultOutputParser {
	if logger == nil {
		panic("logger is required")
	}

	return &DefaultOutputParser{
		logger: logger,
	}
}

// ParseJSON parses JSON output from a command.
func (p *DefaultOutputParser) ParseJSON(ctx context.Context, output string) (map[string]interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Trim whitespace and check for empty output
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, ErrNoDataFound
	}

	// Try to parse as JSON object
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		// Try to extract JSON from mixed output
		jsonData := p.extractJSON(output)
		if jsonData == "" {
			p.logger.Debug("JSON parsing failed",
				"error", err,
				"outputLen", len(output),
			)
			return nil, fmt.Errorf("%w: %w", ErrInvalidJSON, err)
		}

		// Retry with extracted JSON
		if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidJSON, err)
		}
	}

	p.logger.Debug("JSON parsed successfully",
		"keys", len(result),
	)

	return result, nil
}

// ParseTable parses table-formatted output.
func (p *DefaultOutputParser) ParseTable(ctx context.Context, output string) ([]map[string]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Trim and check for empty output
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, ErrNoDataFound
	}

	// Split into lines
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("%w: insufficient lines", ErrInvalidTable)
	}

	// Try different table formats
	// Format 1: Header with separator line (e.g., "Name | Email | Status")
	if strings.Contains(lines[0], "|") || strings.Contains(lines[0], "\t") {
		return p.parsePipeTable(lines)
	}

	// Format 2: Space-separated columns with fixed width
	if len(lines) > 2 && p.looksLikeFixedWidthTable(lines) {
		return p.parseFixedWidthTable(lines)
	}

	// Format 3: Simple space-separated values
	return p.parseSpaceSeparatedTable(lines)
}

// ParseKeyValue parses key-value formatted output.
func (p *DefaultOutputParser) ParseKeyValue(ctx context.Context, output string) (map[string]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Trim and check for empty output
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, ErrNoDataFound
	}

	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Common key-value patterns
	colonPattern := regexp.MustCompile(`^([^:]+):\s*(.*)$`)
	equalPattern := regexp.MustCompile(`^([^=]+)=\s*(.*)$`)
	spacePattern := regexp.MustCompile(`^(\S+)\s+(.*)$`)

	lineCount := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lineCount++

		// Try colon separator (most common)
		if matches := colonPattern.FindStringSubmatch(line); len(matches) == 3 {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			result[key] = value
			continue
		}

		// Try equals separator
		if matches := equalPattern.FindStringSubmatch(line); len(matches) == 3 {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			result[key] = value
			continue
		}

		// Try space separator (less reliable)
		if matches := spacePattern.FindStringSubmatch(line); len(matches) == 3 {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			result[key] = value
			continue
		}

		// If no pattern matches, log but continue
		p.logger.Debug("skipping unparseable line",
			"line", line,
			"lineNumber", lineCount,
		)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("%w: no key-value pairs found", ErrInvalidKeyValue)
	}

	p.logger.Debug("key-value pairs parsed",
		"count", len(result),
	)

	return result, nil
}

// ExtractError extracts error information from output.
func (p *DefaultOutputParser) ExtractError(ctx context.Context, stdout, stderr string, exitCode int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// If exit code is 0, no error
	if exitCode == 0 {
		return nil
	}

	// Common error patterns
	errorPatterns := []struct {
		pattern *regexp.Regexp
		format  string
	}{
		{regexp.MustCompile(`(?i)error:\s*(.+)`), "error: %s"},
		{regexp.MustCompile(`(?i)fatal:\s*(.+)`), "fatal: %s"},
		{regexp.MustCompile(`(?i)failed:\s*(.+)`), "failed: %s"},
		{regexp.MustCompile(`(?i)exception:\s*(.+)`), "exception: %s"},
		{regexp.MustCompile(`(?i)panic:\s*(.+)`), "panic: %s"},
	}

	// Check stderr first
	for _, ep := range errorPatterns {
		if matches := ep.pattern.FindStringSubmatch(stderr); len(matches) > 1 {
			return fmt.Errorf("%w: %s", ErrCommandFailed, strings.TrimSpace(matches[1]))
		}
	}

	// Then check stdout
	for _, ep := range errorPatterns {
		if matches := ep.pattern.FindStringSubmatch(stdout); len(matches) > 1 {
			return fmt.Errorf("%w: %s", ErrCommandFailed, strings.TrimSpace(matches[1]))
		}
	}

	// If no specific error found, use generic message
	if stderr != "" {
		// Take first non-empty line from stderr
		lines := strings.Split(strings.TrimSpace(stderr), "\n")
		for _, line := range lines {
			if line = strings.TrimSpace(line); line != "" {
				return fmt.Errorf("%w: %s", ErrCommandFailed, line)
			}
		}
	}

	// Default error message
	return fmt.Errorf("%w: %d", ErrCommandExitCode, exitCode)
}

// extractJSON attempts to extract JSON from mixed output.
func (p *DefaultOutputParser) extractJSON(output string) string {
	// Look for JSON object boundaries
	start := strings.Index(output, "{")
	if start == -1 {
		// Try array
		start = strings.Index(output, "[")
		if start == -1 {
			return ""
		}
		end := strings.LastIndex(output, "]")
		if end > start {
			return output[start : end+1]
		}
	}

	// Find matching closing brace
	end := strings.LastIndex(output, "}")
	if end > start {
		return output[start : end+1]
	}

	return ""
}

// parsePipeTable parses table with pipe or tab separators.
func (p *DefaultOutputParser) parsePipeTable(lines []string) ([]map[string]string, error) {
	if len(lines) < 2 {
		return nil, ErrInvalidTable
	}

	// Determine separator
	separator := "|"
	if !strings.Contains(lines[0], "|") && strings.Contains(lines[0], "\t") {
		separator = "\t"
	}

	// Parse header
	headers := p.splitTableRow(lines[0], separator)
	if len(headers) == 0 {
		return nil, fmt.Errorf("%w: no headers found", ErrInvalidTable)
	}

	// Clean headers
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	// Skip separator line if present
	startLine := 1
	if len(lines) > 2 && p.isSeparatorLine(lines[1]) {
		startLine = 2
	}

	// Parse data rows
	var result []map[string]string
	for i := startLine; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		values := p.splitTableRow(line, separator)
		if len(values) != len(headers) {
			p.logger.Debug("skipping malformed row",
				"row", i,
				"expected", len(headers),
				"got", len(values),
			)
			continue
		}

		row := make(map[string]string)
		for j, header := range headers {
			row[header] = strings.TrimSpace(values[j])
		}
		result = append(result, row)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("%w: no data rows found", ErrInvalidTable)
	}

	return result, nil
}

// parseFixedWidthTable parses fixed-width column tables.
func (p *DefaultOutputParser) parseFixedWidthTable(lines []string) ([]map[string]string, error) {
	if len(lines) < 2 {
		return nil, ErrInvalidTable
	}

	// Detect column boundaries based on header and separator line
	headers, columnBounds := p.detectColumnBoundaries(lines[0], lines[1])
	if len(headers) == 0 {
		return nil, fmt.Errorf("%w: could not detect columns", ErrInvalidTable)
	}

	// Parse data rows
	var result []map[string]string
	for i := 2; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}

		row := make(map[string]string)
		for j, header := range headers {
			start, end := columnBounds[j], columnBounds[j+1]
			if start < len(line) {
				if end > len(line) {
					end = len(line)
				}
				row[header] = strings.TrimSpace(line[start:end])
			} else {
				row[header] = ""
			}
		}
		result = append(result, row)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("%w: no data rows found", ErrInvalidTable)
	}

	return result, nil
}

// parseSpaceSeparatedTable parses simple space-separated tables.
func (p *DefaultOutputParser) parseSpaceSeparatedTable(lines []string) ([]map[string]string, error) {
	if len(lines) < 2 {
		return nil, ErrInvalidTable
	}

	// Parse header
	headers := strings.Fields(lines[0])
	if len(headers) == 0 {
		return nil, fmt.Errorf("%w: no headers found", ErrInvalidTable)
	}

	// Parse data rows
	var result []map[string]string
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || p.isSeparatorLine(line) {
			continue
		}

		values := strings.Fields(line)
		if len(values) != len(headers) {
			// Try to handle quoted values
			values = p.parseQuotedFields(line, len(headers))
			if len(values) != len(headers) {
				p.logger.Debug("skipping malformed row",
					"row", i,
					"expected", len(headers),
					"got", len(values),
				)
				continue
			}
		}

		row := make(map[string]string)
		for j, header := range headers {
			row[header] = values[j]
		}
		result = append(result, row)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("%w: no data rows found", ErrInvalidTable)
	}

	return result, nil
}

// Helper methods

func (p *DefaultOutputParser) splitTableRow(row, separator string) []string {
	if separator == "|" {
		// Handle pipe-separated with potential spaces
		parts := strings.Split(row, separator)
		// Trim first and last if empty (table borders)
		if len(parts) > 0 && strings.TrimSpace(parts[0]) == "" {
			parts = parts[1:]
		}
		if len(parts) > 0 && strings.TrimSpace(parts[len(parts)-1]) == "" {
			parts = parts[:len(parts)-1]
		}
		return parts
	}
	return strings.Split(row, separator)
}

func (p *DefaultOutputParser) isSeparatorLine(line string) bool {
	// Check if line contains only separator characters
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	// Common separator patterns
	separatorChars := []rune{'-', '=', '+', '|', ' ', '\t'}
	for _, r := range trimmed {
		found := false
		for _, sep := range separatorChars {
			if r == sep {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (p *DefaultOutputParser) looksLikeFixedWidthTable(lines []string) bool {
	// Check if second line is a separator line with dashes
	if len(lines) < 3 {
		return false
	}
	return p.isSeparatorLine(lines[1]) && strings.Contains(lines[1], "-")
}

func (p *DefaultOutputParser) detectColumnBoundaries(header, separator string) ([]string, []int) {
	// Find column boundaries based on separator line
	bounds := []int{0}
	inColumn := false

	for i, r := range separator {
		if r == '-' || r == '=' {
			if !inColumn {
				inColumn = true
			}
		} else if inColumn {
			bounds = append(bounds, i)
			inColumn = false
		}
	}
	if inColumn {
		bounds = append(bounds, len(separator))
	}

	// Extract headers based on boundaries
	var headers []string
	for i := 0; i < len(bounds)-1; i++ {
		start, end := bounds[i], bounds[i+1]
		if start < len(header) {
			if end > len(header) {
				end = len(header)
			}
			h := strings.TrimSpace(header[start:end])
			if h != "" {
				headers = append(headers, h)
			}
		}
	}

	return headers, bounds
}

func (p *DefaultOutputParser) parseQuotedFields(line string, _ int) []string {
	// Simple quoted field parser for space-separated values with quotes
	var fields []string
	var current strings.Builder
	inQuote := false
	quote := rune(0)

	for _, r := range line {
		if !inQuote && (r == '"' || r == '\'') {
			inQuote = true
			quote = r
		} else if inQuote && r == quote {
			inQuote = false
			quote = 0
		} else if !inQuote && r == ' ' {
			if current.Len() > 0 {
				fields = append(fields, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		fields = append(fields, current.String())
	}

	return fields
}
