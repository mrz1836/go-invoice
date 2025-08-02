package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/mrz/go-invoice/internal/csv"
	"github.com/mrz/go-invoice/internal/models"
)

var (
	// CSVParsingFailedError indicates that CSV parsing failed.
	CSVParsingFailedError = fmt.Errorf("CSV parsing failed")
	// BatchValidationFailedError indicates that batch validation failed.
	BatchValidationFailedError = fmt.Errorf("batch validation failed")
	// ClientVerificationFailedError indicates that client verification failed.
	ClientVerificationFailedError = fmt.Errorf("client verification failed")
	// DuplicateDetectionFailedError indicates that duplicate detection failed.
	DuplicateDetectionFailedError = fmt.Errorf("duplicate detection failed")
	// BatchImportFailedError indicates that batch import failed.
	BatchImportFailedError = fmt.Errorf("batch import failed")
)

// ImportService provides high-level import orchestration operations
// Follows dependency injection pattern with consumer-driven interfaces
type ImportService struct {
	parser         csv.TimesheetParser
	invoiceService *InvoiceService
	clientService  *ClientService
	validator      csv.CSVValidator
	logger         Logger
	idGenerator    IDGenerator
}

// NewImportService creates a new import service with injected dependencies
func NewImportService(
	parser csv.TimesheetParser,
	invoiceService *InvoiceService,
	clientService *ClientService,
	validator csv.CSVValidator,
	logger Logger,
	idGenerator IDGenerator,
) *ImportService {
	return &ImportService{
		parser:         parser,
		invoiceService: invoiceService,
		clientService:  clientService,
		validator:      validator,
		logger:         logger,
		idGenerator:    idGenerator,
	}
}

// ImportToNewInvoice imports CSV data and creates a new invoice
func (s *ImportService) ImportToNewInvoice(ctx context.Context, reader io.Reader, req ImportToNewInvoiceRequest) (*csv.ImportResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("starting import to new invoice", "client_id", req.ClientID, "dry_run", req.DryRun)

	// Parse CSV data
	parseResult, err := s.parser.ParseTimesheet(ctx, reader, req.ParseOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", CSVParsingFailedError, err)
	}

	if len(parseResult.WorkItems) == 0 {
		return &csv.ImportResult{
			ParseResult:    parseResult,
			WorkItemsAdded: 0,
			DryRun:         req.DryRun,
		}, nil
	}

	// Validate batch of work items
	if validationErr := s.validator.ValidateBatch(ctx, parseResult.WorkItems); validationErr != nil {
		return nil, fmt.Errorf("%w: %w", BatchValidationFailedError, validationErr)
	}

	if req.DryRun {
		s.logger.Info("dry run completed", "work_items", len(parseResult.WorkItems))
		return s.createDryRunResult(parseResult), nil
	}

	// Verify client exists
	_, err = s.clientService.GetClient(ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ClientVerificationFailedError, err)
	}

	// Generate invoice number if not provided
	invoiceNumber := req.InvoiceNumber
	if invoiceNumber == "" {
		invoiceNumber, err = s.generateInvoiceNumber(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate invoice number: %w", err)
		}
	}

	// Create invoice
	invoiceReq := models.CreateInvoiceRequest{
		Number:      invoiceNumber,
		ClientID:    req.ClientID,
		Date:        req.InvoiceDate,
		DueDate:     req.DueDate,
		Description: req.Description,
		WorkItems:   s.convertToWorkItemRequests(parseResult.WorkItems),
	}

	invoice, err := s.invoiceService.CreateInvoice(ctx, invoiceReq)
	if err != nil {
		return nil, fmt.Errorf("invoice creation failed: %w", err)
	}

	// Calculate total amount
	totalAmount := s.calculateTotalAmount(parseResult.WorkItems)

	result := &csv.ImportResult{
		ParseResult:    parseResult,
		InvoiceID:      string(invoice.ID),
		WorkItemsAdded: len(parseResult.WorkItems),
		TotalAmount:    totalAmount,
		DryRun:         false,
	}

	s.logger.Info("import to new invoice completed",
		"invoice_id", invoice.ID,
		"work_items", len(parseResult.WorkItems),
		"total_amount", totalAmount)

	return result, nil
}

// AppendToInvoice imports CSV data and appends to existing invoice
func (s *ImportService) AppendToInvoice(ctx context.Context, reader io.Reader, req AppendToInvoiceRequest) (*csv.ImportResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("starting import append to invoice", "invoice_id", req.InvoiceID, "dry_run", req.DryRun)

	// Parse CSV data
	parseResult, err := s.parser.ParseTimesheet(ctx, reader, req.ParseOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", CSVParsingFailedError, err)
	}

	if len(parseResult.WorkItems) == 0 {
		return &csv.ImportResult{
			ParseResult:    parseResult,
			InvoiceID:      req.InvoiceID,
			WorkItemsAdded: 0,
			DryRun:         req.DryRun,
		}, nil
	}

	// Validate batch
	if validationErr := s.validator.ValidateBatch(ctx, parseResult.WorkItems); validationErr != nil {
		return nil, fmt.Errorf("%w: %w", BatchValidationFailedError, validationErr)
	}

	// Check for duplicates with existing invoice work items
	warnings, err := s.detectDuplicates(ctx, models.InvoiceID(req.InvoiceID), parseResult.WorkItems)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", DuplicateDetectionFailedError, err)
	}

	if req.DryRun {
		result := s.createDryRunResult(parseResult)
		result.InvoiceID = req.InvoiceID
		result.Warnings = warnings
		s.logger.Info("dry run append completed", "work_items", len(parseResult.WorkItems))
		return result, nil
	}

	// Add work items to invoice
	successCount := 0
	for _, workItem := range parseResult.WorkItems {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		_, err := s.invoiceService.AddWorkItemToInvoice(ctx, models.InvoiceID(req.InvoiceID), workItem)
		if err != nil {
			s.logger.Error("failed to add work item to invoice",
				"invoice_id", req.InvoiceID,
				"work_item_date", workItem.Date,
				"error", err)

			continue
		}

		successCount++
	}

	totalAmount := s.calculateTotalAmount(parseResult.WorkItems[:successCount])

	result := &csv.ImportResult{
		ParseResult:    parseResult,
		InvoiceID:      req.InvoiceID,
		WorkItemsAdded: successCount,
		TotalAmount:    totalAmount,
		Warnings:       warnings,
		DryRun:         false,
	}

	s.logger.Info("import append completed",
		"invoice_id", req.InvoiceID,
		"work_items_added", successCount,
		"total_amount", totalAmount)

	return result, nil
}

// ValidateImport validates CSV data without importing
func (s *ImportService) ValidateImport(ctx context.Context, reader io.Reader, req csv.ValidateImportRequest) (*csv.ValidationResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("starting import validation")

	// Parse CSV data
	parseResult, err := s.parser.ParseTimesheet(ctx, reader, req.Options)
	if err != nil {
		return &csv.ValidationResult{
			Valid:       false,
			ParseResult: parseResult,
			Suggestions: []string{"Check CSV format and field mappings"},
		}, nil
	}

	// Validate batch
	batchErr := s.validator.ValidateBatch(ctx, parseResult.WorkItems)

	valid := parseResult.ErrorRows == 0 && batchErr == nil
	estimatedTotal := s.calculateTotalAmount(parseResult.WorkItems)

	suggestions := s.generateValidationSuggestions(parseResult, batchErr)
	warnings := s.generateValidationWarnings(parseResult.WorkItems)

	result := &csv.ValidationResult{
		Valid:          valid,
		ParseResult:    parseResult,
		Warnings:       warnings,
		Suggestions:    suggestions,
		EstimatedTotal: estimatedTotal,
	}

	s.logger.Info("import validation completed", "valid", valid, "work_items", len(parseResult.WorkItems))
	return result, nil
}

// BatchImport processes multiple import requests
func (s *ImportService) BatchImport(ctx context.Context, req csv.BatchImportRequest) (*csv.BatchResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("starting batch import", "requests", len(req.Requests))

	result := &csv.BatchResult{
		TotalRequests: len(req.Requests),
		Results:       make([]csv.ImportResult, 0, len(req.Requests)),
	}

	for i, importReq := range req.Requests {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		s.logger.Debug("processing batch import request", "index", i+1, "total", len(req.Requests))

		var importResult *csv.ImportResult
		var err error

		// Determine import type and process accordingly
		if importReq.InvoiceID != "" {
			// Append to existing invoice
			appendReq := AppendToInvoiceRequest{
				InvoiceID:    importReq.InvoiceID,
				ParseOptions: importReq.Options,
				DryRun:       importReq.DryRun,
			}
			importResult, err = s.AppendToInvoice(ctx, importReq.Reader.(io.Reader), appendReq)
		} else {
			// Create new invoice
			newInvoiceReq := ImportToNewInvoiceRequest{
				ClientID:     models.ClientID(importReq.ClientID),
				ParseOptions: importReq.Options,
				InvoiceDate:  time.Now(),
				DueDate:      time.Now().AddDate(0, 0, 30),
				DryRun:       importReq.DryRun,
			}
			importResult, err = s.ImportToNewInvoice(ctx, importReq.Reader.(io.Reader), newInvoiceReq)
		}

		if err != nil {
			s.logger.Error("batch import request failed", "index", i+1, "error", err)
			result.FailedRequests++

			if !req.Options.ContinueOnError {
				return nil, fmt.Errorf("batch import failed at request %d: %w", i+1, err)
			}
			continue
		}

		result.SuccessRequests++
		result.TotalWorkItems += importResult.WorkItemsAdded
		result.TotalAmount += importResult.TotalAmount
		result.Results = append(result.Results, *importResult)
	}

	s.logger.Info("batch import completed",
		"total", result.TotalRequests,
		"success", result.SuccessRequests,
		"failed", result.FailedRequests,
		"work_items", result.TotalWorkItems)

	return result, nil
}

// Helper methods

func (s *ImportService) createDryRunResult(parseResult *csv.ParseResult) *csv.ImportResult {
	totalAmount := s.calculateTotalAmount(parseResult.WorkItems)

	return &csv.ImportResult{
		ParseResult:    parseResult,
		WorkItemsAdded: len(parseResult.WorkItems),
		TotalAmount:    totalAmount,
		DryRun:         true,
	}
}

func (s *ImportService) calculateTotalAmount(workItems []models.WorkItem) float64 {
	total := 0.0
	for _, item := range workItems {
		total += item.Total
	}
	return total
}

func (s *ImportService) convertToWorkItemRequests(workItems []models.WorkItem) []models.WorkItem {
	// Since models.WorkItem is already the correct type, return as-is
	return workItems
}

func (s *ImportService) generateInvoiceNumber(ctx context.Context) (string, error) {
	// Generate a simple invoice number based on current date and time
	now := time.Now()
	return fmt.Sprintf("INV-%s", now.Format("20060102-150405")), nil
}

func (s *ImportService) detectDuplicates(ctx context.Context, invoiceID models.InvoiceID, newWorkItems []models.WorkItem) ([]csv.ImportWarning, error) {
	// Get existing invoice
	invoice, err := s.invoiceService.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice for duplicate detection: %w", err)
	}

	var warnings []csv.ImportWarning

	// Simple duplicate detection based on date and hours
	for _, newItem := range newWorkItems {
		for _, existingItem := range invoice.WorkItems {
			if s.workItemsAreSimilar(newItem, existingItem) {
				warning := csv.ImportWarning{
					Type: "duplicate",
					Message: fmt.Sprintf("Potential duplicate work item on %s: %v hours",
						newItem.Date.Format("2006-01-02"), newItem.Hours),
				}
				warnings = append(warnings, warning)
				break
			}
		}
	}

	return warnings, nil
}

func (s *ImportService) workItemsAreSimilar(item1, item2 models.WorkItem) bool {
	// Consider items similar if same date and similar hours (within 0.1)
	return item1.Date.Equal(item2.Date) &&
		abs(item1.Hours-item2.Hours) < 0.1
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func (s *ImportService) generateValidationSuggestions(parseResult *csv.ParseResult, batchErr error) []string {
	var suggestions []string

	if parseResult.ErrorRows > 0 {
		suggestions = append(suggestions,
			"Check data format in rows with errors",
			"Ensure dates are in YYYY-MM-DD format",
			"Verify numeric fields (hours, rates) contain valid numbers")
	}

	if batchErr != nil {
		suggestions = append(suggestions,
			"Review work item validation rules",
			"Check for unusual values (very high hours, extreme rates)")
	}

	if len(parseResult.WorkItems) == 0 {
		suggestions = append(suggestions,
			"File appears to be empty or header-only",
			"Ensure CSV contains data rows after header")
	}

	return suggestions
}

func (s *ImportService) generateValidationWarnings(workItems []models.WorkItem) []csv.ImportWarning {
	var warnings []csv.ImportWarning

	// Check for weekend work
	for _, item := range workItems {
		if item.Date.Weekday() == time.Saturday || item.Date.Weekday() == time.Sunday {
			warnings = append(warnings, csv.ImportWarning{
				Type:    "weekend_work",
				Message: fmt.Sprintf("Work item on weekend: %s", item.Date.Format("2006-01-02")),
			})
		}
	}

	// Check for high hours
	for _, item := range workItems {
		if item.Hours > 10 {
			warnings = append(warnings, csv.ImportWarning{
				Type:    "high_hours",
				Message: fmt.Sprintf("High hours on %s: %v hours", item.Date.Format("2006-01-02"), item.Hours),
			})
		}
	}

	return warnings
}

// Request types for import operations

// ImportToNewInvoiceRequest represents a request to import CSV data into a new invoice
type ImportToNewInvoiceRequest struct {
	ClientID      models.ClientID  `json:"client_id"`      // Client for the new invoice
	ParseOptions  csv.ParseOptions `json:"parse_options"`  // CSV parsing options
	InvoiceNumber string           `json:"invoice_number"` // Optional invoice number (generated if empty)
	InvoiceDate   time.Time        `json:"invoice_date"`   // Invoice date
	DueDate       time.Time        `json:"due_date"`       // Due date
	Description   string           `json:"description"`    // Invoice description
	DryRun        bool             `json:"dry_run"`        // Validate only, don't create
}

// AppendToInvoiceRequest represents a request to append CSV data to existing invoice
type AppendToInvoiceRequest struct {
	InvoiceID    string           `json:"invoice_id"`    // Existing invoice ID
	ParseOptions csv.ParseOptions `json:"parse_options"` // CSV parsing options
	DryRun       bool             `json:"dry_run"`       // Validate only, don't append
}
