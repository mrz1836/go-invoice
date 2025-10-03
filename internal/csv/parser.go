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

// CSV parsing errors
var (
	ErrCSVFileEmpty          = fmt.Errorf("CSV file is empty")
	ErrCannotDetectFormat    = fmt.Errorf("cannot detect format of empty file")
	ErrUnsupportedCSVFormat  = fmt.Errorf("unsupported CSV format")
	ErrEmptyRow              = fmt.Errorf("empty row")
	ErrNoRowsToProcess       = fmt.Errorf("no rows to process")
	ErrRequiredFieldMissing  = fmt.Errorf("required field not found in header")
	ErrFieldNotInHeader      = fmt.Errorf("field not found in header")
	ErrFieldMissingInRow     = fmt.Errorf("field missing in row")
	ErrFieldEmpty            = fmt.Errorf("field is empty")
	ErrUnsupportedDateFormat = fmt.Errorf("unsupported date format")
	ErrNoContentToAnalyze    = fmt.Errorf("no content to analyze")
	ErrFirstLineEmpty        = fmt.Errorf("first line is empty")
	ErrAmbiguousFormat       = fmt.Errorf("ambiguous format: multiple delimiter types detected")
	ErrNoDelimitersFound     = fmt.Errorf("no clear delimiters found")
	ErrTooFewColumns         = fmt.Errorf("too few columns detected")
	ErrTooManyColumns        = fmt.Errorf("too many columns detected")
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
	p.configureReader(csvReader, options.Format)

	// Read all rows
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV data: %w", err)
	}

	if len(rows) == 0 {
		return nil, ErrCSVFileEmpty
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
		return nil, ErrCannotDetectFormat
	}

	// Analyze content to detect format
	format, err := p.analyzeFormat(string(content))
	if err != nil {
		return nil, fmt.Errorf("format detection failed: %w", err)
	}

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
		return fmt.Errorf("%w: %s", ErrUnsupportedCSVFormat, format.Name)
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
		return nil, ErrEmptyRow
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
func (p *CSVParser) processHeader(_ context.Context, rows [][]string, _ ParseOptions) (map[string]int, int, error) {
	if len(rows) == 0 {
		return nil, 0, ErrNoRowsToProcess
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
			return nil, 0, fmt.Errorf("%w: %s", ErrRequiredFieldMissing, field)
		}
	}

	p.logger.Debug("header processed", "fields", len(headerMap))

	return headerMap, 1, nil // Data starts at row 1 (0-based)
}

// getFieldValue retrieves a field value from a row using header mapping
func (p *CSVParser) getFieldValue(row []string, headerMap map[string]int, fieldName string, _ int) (string, error) {
	colIndex, exists := headerMap[fieldName]
	if !exists {
		return "", fmt.Errorf("%w: %s", ErrFieldNotInHeader, fieldName)
	}

	if colIndex >= len(row) {
		return "", fmt.Errorf("%w: %s", ErrFieldMissingInRow, fieldName)
	}

	value := strings.TrimSpace(row[colIndex])
	if value == "" {
		return "", fmt.Errorf("%w: %s", ErrFieldEmpty, fieldName)
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

// parseDate parses date string in various formats with smart year inference.
//
// Supports full year formats (YYYY), 2-digit years (YY), and dates without years.
// For dates without years, assumes current year unless that would put the date
// more than 6 months in the future, in which case it uses the previous year.
// For 2-digit years: 00-50 → 2000-2050, 51-99 → 1951-1999.
func (p *CSVParser) parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	now := time.Now()

	// Try formats with full 4-digit years first (most reliable)
	fullYearFormats := []string{
		"2006-01-02",          // ISO format
		"01/02/2006",          // US format
		"02/01/2006",          // EU format
		"2006/01/02",          // Alternative ISO
		"Jan 2, 2006",         // Month name format
		"January 2, 2006",     // Full month name
		"2006-01-02 15:04:05", // ISO with time
	}

	for _, format := range fullYearFormats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	// Try formats with 2-digit years
	twoDigitYearFormats := []string{
		"01/02/06", // US format with 2-digit year
		"02/01/06", // EU format with 2-digit year
		"1/2/06",   // US format with 2-digit year (no leading zeros)
		"2/1/06",   // EU format with 2-digit year (no leading zeros)
		"06-01-02", // ISO-like with 2-digit year
	}

	for _, format := range twoDigitYearFormats {
		if parsedDate, err := time.Parse(format, dateStr); err == nil {
			// Smart 2-digit year handling: 00-50 → 2000-2050, 51-99 → 1951-1999
			year := parsedDate.Year()
			if year >= 0 && year <= 50 {
				parsedDate = parsedDate.AddDate(2000, 0, 0)
			} else if year >= 51 && year <= 99 {
				parsedDate = parsedDate.AddDate(1900, 0, 0)
			}
			return parsedDate, nil
		}
	}

	// Try formats without years (day/month only)
	noYearFormats := []string{
		"01/02", // US format M/D
		"1/2",   // US format M/D (no leading zeros)
		"Jan 2", // Month name format
	}

	for _, format := range noYearFormats {
		if parsedDate, err := time.Parse(format, dateStr); err == nil {
			// Smart year inference: assume current year unless that puts us >6 months in future
			month := parsedDate.Month()
			day := parsedDate.Day()

			// Try current year
			candidateDate := time.Date(now.Year(), month, day, 0, 0, 0, 0, time.UTC)

			// If more than 6 months in the future, use previous year
			sixMonthsFromNow := now.AddDate(0, 6, 0)
			if candidateDate.After(sixMonthsFromNow) {
				candidateDate = candidateDate.AddDate(-1, 0, 0)
			}

			return candidateDate, nil
		}
	}

	return time.Time{}, ErrUnsupportedDateFormat
}

// configureReader configures CSV reader based on format
func (p *CSVParser) configureReader(reader *csv.Reader, format string) {
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
}

// analyzeFormat analyzes content to detect CSV format
func (p *CSVParser) analyzeFormat(content string) (*FormatInfo, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil, ErrNoContentToAnalyze
	}

	firstLine := strings.TrimSpace(lines[0])
	if firstLine == "" {
		return nil, ErrFirstLineEmpty
	}

	// Count different delimiters
	commaCount := strings.Count(firstLine, ",")
	tabCount := strings.Count(firstLine, "\t")
	semicolonCount := strings.Count(firstLine, ";")

	// Check for mixed delimiters (ambiguous format)
	totalDelimiters := commaCount + tabCount + semicolonCount
	if totalDelimiters > 0 {
		// Count how many different delimiter types are present
		delimiterTypes := 0
		if commaCount > 0 {
			delimiterTypes++
		}
		if tabCount > 0 {
			delimiterTypes++
		}
		if semicolonCount > 0 {
			delimiterTypes++
		}

		// If multiple delimiter types, it's ambiguous
		if delimiterTypes > 1 {
			return nil, ErrAmbiguousFormat
		}
	}

	// Check if no clear delimiters found
	if totalDelimiters == 0 {
		return nil, ErrNoDelimitersFound
	}

	// Check for too few or too many columns
	maxDelimiterCount := commaCount
	if tabCount > maxDelimiterCount {
		maxDelimiterCount = tabCount
	}
	if semicolonCount > maxDelimiterCount {
		maxDelimiterCount = semicolonCount
	}

	columnCount := maxDelimiterCount + 1
	if columnCount < 3 {
		return nil, fmt.Errorf("%w (%d), need at least 3 for work items", ErrTooFewColumns, columnCount)
	}
	if columnCount > 50 {
		return nil, fmt.Errorf("%w (%d), maximum supported is 50", ErrTooManyColumns, columnCount)
	}

	// Determine most likely delimiter
	if tabCount > commaCount && tabCount > semicolonCount {
		return &FormatInfo{Name: "tab", Delimiter: '\t'}, nil
	}
	if semicolonCount > commaCount && semicolonCount > tabCount {
		return &FormatInfo{Name: "semicolon", Delimiter: ';'}, nil
	}

	// Default to comma (standard CSV)
	return &FormatInfo{Name: "standard", Delimiter: ','}, nil
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
