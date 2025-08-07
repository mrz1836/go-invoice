package json

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/mrz/go-invoice/internal/csv"
	"github.com/mrz/go-invoice/internal/models"
)

// JSON parsing errors
var (
	ErrJSONFileEmpty        = fmt.Errorf("JSON file is empty")
	ErrInvalidJSONFormat    = fmt.Errorf("invalid JSON format")
	ErrNoWorkItems          = fmt.Errorf("no work items found in JSON")
	ErrInvalidDateFormat    = fmt.Errorf("invalid date format in JSON")
	ErrMissingRequiredField = fmt.Errorf("missing required field in JSON")
	ErrNotStructuredFormat  = fmt.Errorf("not structured format")
)

// JSONParser implements TimesheetParser for JSON format
type JSONParser struct {
	validator   csv.CSVValidator
	logger      csv.Logger
	idGenerator csv.IDGenerator
}

// NewJSONParser creates a new JSON parser with dependency injection
func NewJSONParser(validator csv.CSVValidator, logger csv.Logger, idGenerator csv.IDGenerator) *JSONParser {
	return &JSONParser{
		validator:   validator,
		logger:      logger,
		idGenerator: idGenerator,
	}
}

// ParseTimesheet parses JSON timesheet data into work items
func (p *JSONParser) ParseTimesheet(ctx context.Context, reader io.Reader, options csv.ParseOptions) (*csv.ParseResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	p.logger.Info("starting JSON timesheet parsing")

	// Read all data from reader
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON data: %w", err)
	}

	if len(data) == 0 {
		return nil, ErrJSONFileEmpty
	}

	// Try to parse as structured format first
	workItems, metadata, err := p.parseStructuredFormat(data)
	if err != nil {
		// Fallback to simple array format
		workItems, err = p.parseSimpleFormat(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidJSONFormat, err)
		}
	}

	if len(workItems) == 0 {
		return nil, ErrNoWorkItems
	}

	// Convert JSON work items to models.WorkItem
	modelWorkItems, parseErrors := p.convertToModelWorkItems(workItems, options)

	result := &csv.ParseResult{
		WorkItems:   modelWorkItems,
		TotalRows:   len(workItems),
		SuccessRows: len(modelWorkItems),
		ErrorRows:   len(parseErrors),
		Errors:      parseErrors,
		Format:      "JSON",
	}

	// Log metadata if available
	if metadata != nil {
		p.logger.Info("JSON metadata found",
			"client", metadata.Client,
			"period", metadata.Period,
			"description", metadata.Description,
			"currency", metadata.Currency)
	}

	p.logger.Info("JSON parsing completed",
		"total_rows", result.TotalRows,
		"success_rows", result.SuccessRows,
		"error_rows", result.ErrorRows)

	return result, nil
}

// parseStructuredFormat attempts to parse JSON as structured format with metadata
func (p *JSONParser) parseStructuredFormat(data []byte) ([]WorkItemJSON, *ImportMetadata, error) {
	var importData JSONImport
	if err := json.Unmarshal(data, &importData); err != nil {
		return nil, nil, err
	}

	// Check if we have work_items field (structured format)
	if importData.WorkItems == nil {
		return nil, nil, ErrNotStructuredFormat
	}

	return importData.WorkItems, importData.Metadata, nil
}

// parseSimpleFormat attempts to parse JSON as simple array format
func (p *JSONParser) parseSimpleFormat(data []byte) ([]WorkItemJSON, error) {
	var workItems []WorkItemJSON
	if err := json.Unmarshal(data, &workItems); err != nil {
		return nil, err
	}
	return workItems, nil
}

// convertToModelWorkItems converts JSON work items to model work items
func (p *JSONParser) convertToModelWorkItems(jsonItems []WorkItemJSON, _ csv.ParseOptions) ([]models.WorkItem, []csv.ParseError) {
	workItems := make([]models.WorkItem, 0, len(jsonItems))
	var errors []csv.ParseError

	for i, item := range jsonItems {
		rowNum := i + 1

		// Validate required fields
		if item.Date == "" {
			errors = append(errors, csv.ParseError{
				Line:    rowNum,
				Column:  "date",
				Value:   "",
				Message: "date is required",
			})
			continue
		}

		if item.Description == "" {
			errors = append(errors, csv.ParseError{
				Line:    rowNum,
				Column:  "description",
				Value:   "",
				Message: "description is required",
			})
			continue
		}

		// Parse date
		date, err := p.parseDate(item.Date, nil)
		if err != nil {
			errors = append(errors, csv.ParseError{
				Line:    rowNum,
				Column:  "date",
				Value:   item.Date,
				Message: fmt.Sprintf("invalid date format: %v", err),
			})
			continue
		}

		// Use the rate from the JSON
		rate := item.Rate

		// Create work item
		workItem := models.WorkItem{
			ID:          p.idGenerator.GenerateID(),
			Date:        date,
			Hours:       item.Hours,
			Rate:        rate,
			Description: item.Description,
			Total:       item.Hours * rate,
			CreatedAt:   time.Now(),
		}

		workItems = append(workItems, workItem)
	}

	return workItems, errors
}

// parseDate attempts to parse a date string using multiple formats
func (p *JSONParser) parseDate(dateStr string, formats []string) (time.Time, error) {
	// Default formats if none provided
	if len(formats) == 0 {
		formats = []string{
			"2006-01-02",          // ISO date
			"2006-01-02T15:04:05", // ISO datetime
			"01/02/2006",          // US format
			"02/01/2006",          // EU format
			"1/2/2006",            // Short US
			"2/1/2006",            // Short EU
		}
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("%w: %s", ErrInvalidDateFormat, dateStr)
}

// DetectFormat checks if the reader contains valid JSON
func (p *JSONParser) DetectFormat(ctx context.Context, reader io.Reader) (*csv.FormatInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON
	var test interface{}
	if err := json.Unmarshal(data, &test); err != nil {
		return nil, fmt.Errorf("not valid JSON: %w", err)
	}

	return &csv.FormatInfo{
		Name:      "JSON",
		Delimiter: 0, // No delimiter for JSON
		HasHeader: false,
	}, nil
}

// ValidateFormat validates that the JSON format is correct
func (p *JSONParser) ValidateFormat(ctx context.Context, reader io.Reader) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := p.DetectFormat(ctx, reader)
	return err
}
