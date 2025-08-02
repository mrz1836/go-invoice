package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/mrz/go-invoice/internal/models"
)

// TimesheetParser defines the interface for CSV parsing operations
// Consumer-driven interface defined at point of use
type TimesheetParser interface {
	ParseTimesheet(ctx context.Context, reader io.Reader, options ParseOptions) (*ParseResult, error)
	DetectFormat(ctx context.Context, reader io.Reader) (*FormatInfo, error)
	ValidateFormat(ctx context.Context, reader io.Reader) error
}

// Logger interface for CSV operations
type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
	Debug(msg string, fields ...any)
}

// IDGenerator defines the interface for generating unique IDs
type IDGenerator interface {
	GenerateID() string
}

// CSVParser implements TimesheetParser with context-first design
type CSVParser struct {
	validator   CSVValidator
	logger      Logger
	idGenerator IDGenerator
}

// NewCSVParser creates a new CSV parser with dependency injection
func NewCSVParser(validator CSVValidator, logger Logger, idGenerator IDGenerator) *CSVParser {
	return &CSVParser{
		validator:   validator,
		logger:      logger,
		idGenerator: idGenerator,
	}
}

// ParseTimesheet parses CSV timesheet data into work items with context support
func (p *CSVParser) ParseTimesheet(ctx context.Context, reader io.Reader, options ParseOptions) (*ParseResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	p.logger.Info("starting timesheet parsing", "format", options.Format)

	// Create CSV reader with format-specific configuration
	csvReader := csv.NewReader(reader)
	if err := p.configureReader(csvReader, options.Format); err != nil {
		return nil, fmt.Errorf("failed to configure CSV reader: %w", err)
	}

	// Read all rows
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV data: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Detect or validate header
	headerMap, dataStartRow, err := p.processHeader(ctx, rows, options)
	if err != nil {
		return nil, fmt.Errorf("header processing failed: %w", err)
	}

	// Parse data rows
	var workItems []models.WorkItem
	var parseErrors []ParseError

	for i := dataStartRow; i < len(rows); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		lineNum := i + 1 // 1-based line numbers for user display
		row := rows[i]

		workItem, err := p.parseRow(ctx, row, headerMap, lineNum)
		if err != nil {
			parseError := ParseError{
				Line:    lineNum,
				Message: err.Error(),
				Row:     row,
			}
			parseErrors = append(parseErrors, parseError)

			if !options.ContinueOnError {
				return nil, fmt.Errorf("parsing failed at line %d: %w", lineNum, err)
			}
			continue
		}

		// Validate work item
		if err := p.validator.ValidateWorkItem(ctx, workItem); err != nil {
			parseError := ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("validation failed: %v", err),
				Row:     row,
			}
			parseErrors = append(parseErrors, parseError)

			if !options.ContinueOnError {
				return nil, fmt.Errorf("validation failed at line %d: %w", lineNum, err)
			}
			continue
		}

		workItems = append(workItems, *workItem)
	}

	result := &ParseResult{
		WorkItems:   workItems,
		TotalRows:   len(rows) - dataStartRow,
		SuccessRows: len(workItems),
		ErrorRows:   len(parseErrors),
		Errors:      parseErrors,
		HeaderMap:   headerMap,
		Format:      options.Format,
	}

	p.logger.Info("timesheet parsing completed",
		"total_rows", result.TotalRows,
		"success_rows", result.SuccessRows,
		"error_rows", result.ErrorRows)

	return result, nil
}

// DetectFormat attempts to detect the CSV format from the data
func (p *CSVParser) DetectFormat(ctx context.Context, reader io.Reader) (*FormatInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Read first few lines to analyze format
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content for format detection: %w", err)
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("cannot detect format of empty file")
	}

	// Analyze content to detect format
	format := p.analyzeFormat(string(content))

	p.logger.Debug("format detection completed", "detected_format", format.Name, "delimiter", format.Delimiter)

	return format, nil
}

// ValidateFormat validates that the CSV format is supported
func (p *CSVParser) ValidateFormat(ctx context.Context, reader io.Reader) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	format, err := p.DetectFormat(ctx, reader)
	if err != nil {
		return fmt.Errorf("format detection failed: %w", err)
	}

	if !p.isSupportedFormat(format) {
		return fmt.Errorf("unsupported CSV format: %s", format.Name)
	}

	return nil
}

// parseRow parses a single CSV row into a WorkItem
func (p *CSVParser) parseRow(ctx context.Context, row []string, headerMap map[string]int, lineNum int) (*models.WorkItem, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(row) == 0 {
		return nil, fmt.Errorf("empty row")
	}

	// Extract required fields using header mapping
	dateStr, err := p.getFieldValue(row, headerMap, "date", lineNum)
	if err != nil {
		return nil, err
	}

	hoursStr, err := p.getFieldValue(row, headerMap, "hours", lineNum)
	if err != nil {
		return nil, err
	}

	rateStr, err := p.getFieldValue(row, headerMap, "rate", lineNum)
	if err != nil {
		return nil, err
	}

	description, err := p.getFieldValue(row, headerMap, "description", lineNum)
	if err != nil {
		return nil, err
	}

	// Parse date
	date, err := p.parseDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date '%s': %w", dateStr, err)
	}

	// Parse hours
	hours, err := strconv.ParseFloat(strings.TrimSpace(hoursStr), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid hours '%s': %w", hoursStr, err)
	}

	// Parse rate
	rate, err := strconv.ParseFloat(strings.TrimSpace(rateStr), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid rate '%s': %w", rateStr, err)
	}

	// Generate ID for work item
	id := p.idGenerator.GenerateID()

	// Create work item
	workItem, err := models.NewWorkItem(ctx, id, date, hours, rate, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create work item: %w", err)
	}

	return workItem, nil
}

// processHeader processes the header row and returns field mapping
func (p *CSVParser) processHeader(ctx context.Context, rows [][]string, options ParseOptions) (map[string]int, int, error) {
	if len(rows) == 0 {
		return nil, 0, fmt.Errorf("no rows to process")
	}

	headerRow := rows[0]
	headerMap := make(map[string]int)

	// Map header fields to column indices
	for i, header := range headerRow {
		normalizedHeader := p.normalizeHeaderName(header)
		headerMap[normalizedHeader] = i
	}

	// Validate required fields are present
	requiredFields := []string{"date", "hours", "rate", "description"}
	for _, field := range requiredFields {
		if _, exists := headerMap[field]; !exists {
			return nil, 0, fmt.Errorf("required field '%s' not found in header", field)
		}
	}

	p.logger.Debug("header processed", "fields", len(headerMap))
	return headerMap, 1, nil // Data starts at row 1 (0-based)
}

// getFieldValue retrieves a field value from a row using header mapping
func (p *CSVParser) getFieldValue(row []string, headerMap map[string]int, fieldName string, lineNum int) (string, error) {
	colIndex, exists := headerMap[fieldName]
	if !exists {
		return "", fmt.Errorf("field '%s' not found in header", fieldName)
	}

	if colIndex >= len(row) {
		return "", fmt.Errorf("field '%s' missing in row", fieldName)
	}

	value := strings.TrimSpace(row[colIndex])
	if value == "" {
		return "", fmt.Errorf("field '%s' is empty", fieldName)
	}

	return value, nil
}

// normalizeHeaderName normalizes header names for consistent mapping
func (p *CSVParser) normalizeHeaderName(header string) string {
	normalized := strings.ToLower(strings.TrimSpace(header))

	// Handle common variations
	switch normalized {
	case "date", "work_date", "day":
		return "date"
	case "hours", "time", "duration", "hours_worked":
		return "hours"
	case "rate", "hourly_rate", "hour_rate", "billing_rate":
		return "rate"
	case "description", "desc", "task", "work_description", "notes":
		return "description"
	default:
		return normalized
	}
}

// parseDate parses date string in various formats
func (p *CSVParser) parseDate(dateStr string) (time.Time, error) {
	dateFormats := []string{
		"2006-01-02",          // ISO format
		"01/02/2006",          // US format
		"02/01/2006",          // EU format
		"2006/01/02",          // Alternative ISO
		"Jan 2, 2006",         // Month name format
		"January 2, 2006",     // Full month name
		"2006-01-02 15:04:05", // ISO with time
	}

	dateStr = strings.TrimSpace(dateStr)

	for _, format := range dateFormats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date format")
}

// configureReader configures CSV reader based on format
func (p *CSVParser) configureReader(reader *csv.Reader, format string) error {
	switch format {
	case "standard", "rfc4180":
		reader.Comma = ','
	case "tab", "tsv":
		reader.Comma = '\t'
	case "semicolon":
		reader.Comma = ';'
	case "excel":
		reader.Comma = ','
		reader.LazyQuotes = true
	default:
		reader.Comma = ','
	}

	reader.TrimLeadingSpace = true
	return nil
}

// analyzeFormat analyzes content to detect CSV format
func (p *CSVParser) analyzeFormat(content string) *FormatInfo {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return &FormatInfo{Name: "standard", Delimiter: ','}
	}

	firstLine := lines[0]

	// Count different delimiters
	commaCount := strings.Count(firstLine, ",")
	tabCount := strings.Count(firstLine, "\t")
	semicolonCount := strings.Count(firstLine, ";")

	// Determine most likely delimiter
	if tabCount > commaCount && tabCount > semicolonCount {
		return &FormatInfo{Name: "tab", Delimiter: '\t'}
	}
	if semicolonCount > commaCount && semicolonCount > tabCount {
		return &FormatInfo{Name: "semicolon", Delimiter: ';'}
	}

	// Default to comma (standard CSV)
	return &FormatInfo{Name: "standard", Delimiter: ','}
}

// isSupportedFormat checks if the detected format is supported
func (p *CSVParser) isSupportedFormat(format *FormatInfo) bool {
	supportedFormats := []string{"standard", "rfc4180", "tab", "tsv", "semicolon", "excel"}

	for _, supported := range supportedFormats {
		if format.Name == supported {
			return true
		}
	}

	return false
}
