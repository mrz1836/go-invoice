package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mrz1836/go-invoice/internal/models"
	"github.com/mrz1836/go-invoice/internal/storage"
)

// Static errors to avoid dynamic error creation
var (
	ErrInvoiceNumberEmpty    = errors.New("invoice number cannot be empty")
	ErrInvoiceNumberNotFound = errors.New("invoice not found")
)

// Logger interface for service operations
type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
	Debug(msg string, fields ...any)
}

// IDGenerator interface for generating unique IDs
type IDGenerator interface {
	GenerateInvoiceID(ctx context.Context) (models.InvoiceID, error)
	GenerateClientID(ctx context.Context) (models.ClientID, error)
	GenerateWorkItemID(ctx context.Context) (string, error)
}

// InvoiceService provides high-level invoice management operations
// Follows dependency injection pattern with consumer-driven interfaces
type InvoiceService struct {
	invoiceStorage storage.InvoiceStorage
	clientStorage  storage.ClientStorage
	logger         Logger
	idGenerator    IDGenerator
}

// NewInvoiceService creates a new invoice service with injected dependencies
func NewInvoiceService(
	invoiceStorage storage.InvoiceStorage,
	clientStorage storage.ClientStorage,
	logger Logger,
	idGenerator IDGenerator,
) *InvoiceService {
	return &InvoiceService{
		invoiceStorage: invoiceStorage,
		clientStorage:  clientStorage,
		logger:         logger,
		idGenerator:    idGenerator,
	}
}

// CreateInvoice creates a new invoice with business logic validation
func (s *InvoiceService) CreateInvoice(ctx context.Context, req models.CreateInvoiceRequest) (*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("creating invoice", "number", req.Number, "client_id", req.ClientID)

	// Validate request
	if err := req.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid create invoice request: %w", err)
	}

	// Verify client exists and is active
	client, err := s.clientStorage.GetClient(ctx, req.ClientID)
	if err != nil {
		if storage.IsNotFound(err) {
			return nil, fmt.Errorf("%w: %s", models.ErrClientNotFound, req.ClientID)
		}
		return nil, fmt.Errorf("failed to retrieve client: %w", err)
	}

	if !client.Active {
		return nil, fmt.Errorf("%w: %s", models.ErrClientInactive, client.Name)
	}

	// Generate invoice ID
	invoiceID, err := s.idGenerator.GenerateInvoiceID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice ID: %w", err)
	}

	// Check if invoice number already exists
	if validateErr := s.validateUniqueInvoiceNumber(ctx, req.Number); validateErr != nil {
		return nil, validateErr
	}

	// Create invoice with work items
	invoice, err := models.NewInvoice(ctx, invoiceID, req.Number, req.Date, req.DueDate, *client, 0.0)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice model: %w", err)
	}

	// Set description if provided
	if req.Description != "" {
		invoice.Description = req.Description
	}

	// Set crypto address overrides if provided
	if req.USDCAddress != nil {
		invoice.USDCAddressOverride = req.USDCAddress
	}
	if req.BSVAddress != nil {
		invoice.BSVAddressOverride = req.BSVAddress
	}

	// Add work items if provided
	for _, workItemReq := range req.WorkItems {
		workItemID, err := s.idGenerator.GenerateWorkItemID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate work item ID: %w", err)
		}

		workItem, err := models.NewWorkItem(ctx, workItemID, workItemReq.Date,
			workItemReq.Hours, workItemReq.Rate, workItemReq.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to create work item: %w", err)
		}

		if err := invoice.AddWorkItem(ctx, *workItem); err != nil {
			return nil, fmt.Errorf("failed to add work item to invoice: %w", err)
		}
	}

	// Store invoice
	if err := s.invoiceStorage.CreateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to store invoice: %w", err)
	}

	s.logger.Info("invoice created successfully", "id", invoice.ID, "number", invoice.Number, "total", invoice.Total)
	return invoice, nil
}

// GetInvoice retrieves an invoice by ID
func (s *InvoiceService) GetInvoice(ctx context.Context, id models.InvoiceID) (*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if strings.TrimSpace(string(id)) == "" {
		return nil, models.ErrInvoiceIDEmpty
	}

	invoice, err := s.invoiceStorage.GetInvoice(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoice: %w", err)
	}

	return invoice, nil
}

// GetInvoiceByNumber retrieves an invoice by its number
func (s *InvoiceService) GetInvoiceByNumber(ctx context.Context, number string) (*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if strings.TrimSpace(number) == "" {
		return nil, ErrInvoiceNumberEmpty
	}

	// Use list functionality to find invoice by number
	filter := models.InvoiceFilter{}
	result, err := s.invoiceStorage.ListInvoices(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search for invoice: %w", err)
	}

	for _, invoice := range result.Invoices {
		if invoice.Number == number {
			return invoice, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrInvoiceNumberNotFound, number)
}

// UpdateInvoice updates an existing invoice
func (s *InvoiceService) UpdateInvoice(ctx context.Context, req models.UpdateInvoiceRequest) (*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("updating invoice", "id", req.ID)

	// Validate request
	if err := req.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid update invoice request: %w", err)
	}

	// Get existing invoice
	invoice, err := s.invoiceStorage.GetInvoice(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoice for update: %w", err)
	}

	// Apply updates
	if req.Number != nil {
		// Check if new number is unique (excluding this invoice)
		if *req.Number != invoice.Number {
			if err := s.validateUniqueInvoiceNumber(ctx, *req.Number); err != nil {
				return nil, err
			}
		}
		invoice.Number = *req.Number
	}

	if req.Date != nil {
		invoice.Date = *req.Date
	}

	if req.DueDate != nil {
		invoice.DueDate = *req.DueDate
	}

	if req.Status != nil {
		if err := invoice.UpdateStatus(ctx, *req.Status); err != nil {
			return nil, fmt.Errorf("failed to update invoice status: %w", err)
		}
	}

	if req.Description != nil {
		invoice.Description = *req.Description
	}

	// Update crypto address overrides if provided
	if req.USDCAddress != nil {
		invoice.USDCAddressOverride = req.USDCAddress
	}
	if req.BSVAddress != nil {
		invoice.BSVAddressOverride = req.BSVAddress
	}

	// Update invoice in storage
	if err := s.invoiceStorage.UpdateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to update invoice in storage: %w", err)
	}

	s.logger.Info("invoice updated successfully", "id", invoice.ID, "version", invoice.Version)
	return invoice, nil
}

// DeleteInvoice deletes an invoice
func (s *InvoiceService) DeleteInvoice(ctx context.Context, id models.InvoiceID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.logger.Info("deleting invoice", "id", id)

	// Check if invoice exists and can be deleted - try by ID first, then by number
	invoice, err := s.invoiceStorage.GetInvoice(ctx, id)
	if err != nil {
		// If not found by ID, try by number
		invoice, err = s.GetInvoiceByNumber(ctx, string(id))
		if err != nil {
			return fmt.Errorf("failed to retrieve invoice for deletion: %w: invoice with ID '%s' not found", models.ErrInvoiceNotFound, id)
		}
	}

	// Business rule: don't delete paid invoices
	if invoice.Status == models.StatusPaid {
		return fmt.Errorf("%w: %s", models.ErrCannotDeletePaidInvoice, invoice.Number)
	}

	// Delete invoice using the actual invoice ID (in case we found it by number)
	if err := s.invoiceStorage.DeleteInvoice(ctx, invoice.ID); err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	s.logger.Info("invoice deleted successfully", "id", id, "number", invoice.Number)
	return nil
}

// ListInvoices retrieves invoices with filtering and pagination
func (s *InvoiceService) ListInvoices(ctx context.Context, filter models.InvoiceFilter) (*storage.InvoiceListResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result, err := s.invoiceStorage.ListInvoices(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}

	s.logger.Debug("listed invoices", "count", len(result.Invoices), "total", result.TotalCount)
	return result, nil
}

// AddWorkItemToInvoice adds a work item to an existing invoice
func (s *InvoiceService) AddWorkItemToInvoice(ctx context.Context, invoiceID models.InvoiceID, workItemData models.WorkItem) (*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("adding work item to invoice", "invoice_id", invoiceID)

	// Get existing invoice
	invoice, err := s.invoiceStorage.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoice: %w", err)
	}

	// Business rule: can only add work items to draft invoices
	if invoice.Status != models.StatusDraft {
		return nil, fmt.Errorf("%w, current status: %s", models.ErrCannotAddWorkItemToNonDraft, invoice.Status)
	}

	// Generate work item ID if not provided
	if workItemData.ID == "" {
		workItemID, err := s.idGenerator.GenerateWorkItemID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate work item ID: %w", err)
		}
		workItemData.ID = workItemID
	}

	// Set creation time
	workItemData.CreatedAt = time.Now()

	// Add work item to invoice
	if err := invoice.AddWorkItemWithoutVersionIncrement(ctx, workItemData); err != nil {
		return nil, fmt.Errorf("failed to add work item: %w", err)
	}

	// Update invoice in storage
	if err := s.invoiceStorage.UpdateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to update invoice with new work item: %w", err)
	}

	s.logger.Info("work item added successfully", "invoice_id", invoiceID, "work_item_id", workItemData.ID)
	return invoice, nil
}

// AddLineItemToInvoice adds a line item to an existing invoice
func (s *InvoiceService) AddLineItemToInvoice(ctx context.Context, invoiceID models.InvoiceID, lineItemData models.LineItem) (*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("adding line item to invoice", "invoice_id", invoiceID, "type", lineItemData.Type)

	// Get existing invoice
	invoice, err := s.invoiceStorage.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoice: %w", err)
	}

	// Business rule: can only add line items to draft invoices
	if invoice.Status != models.StatusDraft {
		return nil, fmt.Errorf("%w, current status: %s", models.ErrCannotAddWorkItemToNonDraft, invoice.Status)
	}

	// Generate line item ID if not provided
	if lineItemData.ID == "" {
		lineItemID, err := s.idGenerator.GenerateWorkItemID(ctx) // Reuse work item ID generator
		if err != nil {
			return nil, fmt.Errorf("failed to generate line item ID: %w", err)
		}
		lineItemData.ID = lineItemID
	}

	// Set creation time
	lineItemData.CreatedAt = time.Now()

	// Add line item to invoice
	if err := invoice.AddLineItemWithoutVersionIncrement(ctx, lineItemData); err != nil {
		return nil, fmt.Errorf("failed to add line item: %w", err)
	}

	// Update invoice in storage
	if err := s.invoiceStorage.UpdateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to update invoice with new line item: %w", err)
	}

	s.logger.Info("line item added successfully", "invoice_id", invoiceID, "line_item_id", lineItemData.ID, "type", lineItemData.Type)
	return invoice, nil
}

// UpdateInvoiceDirectly updates an invoice directly in storage without additional validation
// This method is used internally by import service to avoid optimistic locking conflicts
func (s *InvoiceService) UpdateInvoiceDirectly(ctx context.Context, invoice *models.Invoice) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.logger.Info("updating invoice directly", "invoice_id", invoice.ID)

	// Update timestamp
	invoice.UpdatedAt = time.Now()

	// Update invoice in storage
	if err := s.invoiceStorage.UpdateInvoice(ctx, invoice); err != nil {
		return fmt.Errorf("failed to update invoice in storage: %w", err)
	}

	s.logger.Info("invoice updated directly", "invoice_id", invoice.ID, "version", invoice.Version)
	return nil
}

// RemoveWorkItemFromInvoice removes a work item from an invoice
func (s *InvoiceService) RemoveWorkItemFromInvoice(ctx context.Context, invoiceID models.InvoiceID, workItemID string) (*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("removing work item from invoice", "invoice_id", invoiceID, "work_item_id", workItemID)

	// Get existing invoice
	invoice, err := s.invoiceStorage.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoice: %w", err)
	}

	// Business rule: can only remove work items from draft invoices
	if invoice.Status != models.StatusDraft {
		return nil, fmt.Errorf("%w, current status: %s", models.ErrCannotRemoveWorkItemFromNonDraft, invoice.Status)
	}

	// Remove work item from invoice
	if err := invoice.RemoveWorkItem(ctx, workItemID); err != nil {
		return nil, fmt.Errorf("failed to remove work item: %w", err)
	}

	// Update invoice in storage
	if err := s.invoiceStorage.UpdateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to update invoice after removing work item: %w", err)
	}

	s.logger.Info("work item removed successfully", "invoice_id", invoiceID, "work_item_id", workItemID)
	return invoice, nil
}

// SendInvoice marks an invoice as sent and updates the status
func (s *InvoiceService) SendInvoice(ctx context.Context, id models.InvoiceID) (*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("sending invoice", "id", id)

	// Get existing invoice
	invoice, err := s.invoiceStorage.GetInvoice(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoice: %w", err)
	}

	// Business rule: can only send draft invoices
	if invoice.Status != models.StatusDraft {
		return nil, fmt.Errorf("%w, current status: %s", models.ErrCannotSendNonDraftInvoice, invoice.Status)
	}

	// Business rule: invoice must have work items
	if len(invoice.WorkItems) == 0 {
		return nil, models.ErrCannotSendEmptyInvoice
	}

	// Update status to sent
	if err := invoice.UpdateStatus(ctx, models.StatusSent); err != nil {
		return nil, fmt.Errorf("failed to update invoice status: %w", err)
	}

	// Update invoice in storage
	if err := s.invoiceStorage.UpdateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to update invoice status in storage: %w", err)
	}

	s.logger.Info("invoice sent successfully", "id", id, "number", invoice.Number)
	return invoice, nil
}

// MarkInvoicePaid marks an invoice as paid
func (s *InvoiceService) MarkInvoicePaid(ctx context.Context, id models.InvoiceID) (*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("marking invoice as paid", "id", id)

	// Get existing invoice
	invoice, err := s.invoiceStorage.GetInvoice(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invoice: %w", err)
	}

	// Business rule: can only mark sent or overdue invoices as paid
	if invoice.Status != models.StatusSent && invoice.Status != models.StatusOverdue {
		return nil, fmt.Errorf("%w, current status: %s", models.ErrCannotMarkNonSentAsPaid, invoice.Status)
	}

	// Update status to paid
	if err := invoice.UpdateStatus(ctx, models.StatusPaid); err != nil {
		return nil, fmt.Errorf("failed to update invoice status: %w", err)
	}

	// Update invoice in storage
	if err := s.invoiceStorage.UpdateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to update invoice status in storage: %w", err)
	}

	s.logger.Info("invoice marked as paid", "id", id, "number", invoice.Number, "amount", invoice.Total)
	return invoice, nil
}

// GetOverdueInvoices returns all overdue invoices
func (s *InvoiceService) GetOverdueInvoices(ctx context.Context) ([]*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get all sent invoices that are past due date
	filter := models.InvoiceFilter{
		Status:    models.StatusSent,
		DueDateTo: time.Now(),
	}

	result, err := s.invoiceStorage.ListInvoices(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}

	// Filter for truly overdue invoices and update status
	var overdueInvoices []*models.Invoice
	for _, invoice := range result.Invoices {
		if invoice.IsOverdue() {
			// Update status to overdue
			if err := invoice.UpdateStatus(ctx, models.StatusOverdue); err != nil {
				s.logger.Error("failed to update overdue invoice status", "id", invoice.ID, "error", err)
				continue
			}

			// Update in storage
			if err := s.invoiceStorage.UpdateInvoice(ctx, invoice); err != nil {
				s.logger.Error("failed to update overdue invoice in storage", "id", invoice.ID, "error", err)
				continue
			}

			overdueInvoices = append(overdueInvoices, invoice)
		}
	}

	s.logger.Info("found overdue invoices", "count", len(overdueInvoices))
	return overdueInvoices, nil
}

// GetInvoiceStatistics returns summary statistics for all invoices
func (s *InvoiceService) GetInvoiceStatistics(ctx context.Context) (*InvoiceStatistics, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get all invoices
	filter := models.InvoiceFilter{}
	result, err := s.invoiceStorage.ListInvoices(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices for statistics: %w", err)
	}

	stats := &InvoiceStatistics{}
	stats.TotalInvoices = int(result.TotalCount)

	var totalAmount, paidAmount, outstandingAmount float64

	for _, invoice := range result.Invoices {
		totalAmount += invoice.Total

		switch invoice.Status {
		case models.StatusDraft:
			stats.DraftCount++
		case models.StatusSent:
			stats.SentCount++
			outstandingAmount += invoice.Total
		case models.StatusPaid:
			stats.PaidCount++
			paidAmount += invoice.Total
		case models.StatusOverdue:
			stats.OverdueCount++
			outstandingAmount += invoice.Total
		case models.StatusVoided:
			stats.VoidedCount++
		}
	}

	stats.TotalAmount = totalAmount
	stats.PaidAmount = paidAmount
	stats.OutstandingAmount = outstandingAmount

	return stats, nil
}

// Helper methods

func (s *InvoiceService) validateUniqueInvoiceNumber(ctx context.Context, number string) error {
	// This is a simplified implementation - in a real system with many invoices,
	// you'd want a more efficient approach using an index
	filter := models.InvoiceFilter{}
	result, err := s.invoiceStorage.ListInvoices(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to check invoice number uniqueness: %w", err)
	}

	for _, invoice := range result.Invoices {
		if invoice.Number == number {
			return fmt.Errorf("%w: %s", models.ErrInvoiceNumberExists, number)
		}
	}

	return nil
}

// InvoiceStatistics represents summary statistics for invoices
type InvoiceStatistics struct {
	TotalInvoices     int     `json:"total_invoices"`
	DraftCount        int     `json:"draft_count"`
	SentCount         int     `json:"sent_count"`
	PaidCount         int     `json:"paid_count"`
	OverdueCount      int     `json:"overdue_count"`
	VoidedCount       int     `json:"voided_count"`
	TotalAmount       float64 `json:"total_amount"`
	PaidAmount        float64 `json:"paid_amount"`
	OutstandingAmount float64 `json:"outstanding_amount"`
}
