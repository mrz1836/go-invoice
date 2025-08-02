package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/mrz/go-invoice/internal/models"
	"github.com/mrz/go-invoice/internal/storage"
)

var (
	ErrInvalidCreateClientRequest = fmt.Errorf("invalid create client request")
	ErrFailedToGenerateClientID   = fmt.Errorf("failed to generate client ID")
	ErrFailedToCreateClientModel  = fmt.Errorf("failed to create client model")
	ErrFailedToStoreClient        = fmt.Errorf("failed to store client")
	ErrFailedToRetrieveClient     = fmt.Errorf("failed to retrieve client")
)

// ClientService provides high-level client management operations
// Follows dependency injection pattern with consumer-driven interfaces
type ClientService struct {
	clientStorage  storage.ClientStorage
	invoiceStorage storage.InvoiceStorage
	logger         Logger
	idGenerator    IDGenerator
}

// NewClientService creates a new client service with injected dependencies
func NewClientService(
	clientStorage storage.ClientStorage,
	invoiceStorage storage.InvoiceStorage,
	logger Logger,
	idGenerator IDGenerator,
) *ClientService {
	return &ClientService{
		clientStorage:  clientStorage,
		invoiceStorage: invoiceStorage,
		logger:         logger,
		idGenerator:    idGenerator,
	}
}

// CreateClient creates a new client with business logic validation
func (s *ClientService) CreateClient(ctx context.Context, req models.CreateClientRequest) (*models.Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("creating client", "name", req.Name, "email", req.Email)

	// Validate request
	if err := req.Validate(ctx); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCreateClientRequest, err)
	}

	// Check if client with this email already exists
	if err := s.validateUniqueClientEmail(ctx, req.Email); err != nil {
		return nil, err
	}

	// Generate client ID
	clientID, err := s.idGenerator.GenerateClientID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToGenerateClientID, err)
	}

	// Create client
	client, err := models.NewClient(ctx, clientID, req.Name, req.Email)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToCreateClientModel, err)
	}

	// Set optional fields
	if req.Phone != "" {
		if err := client.UpdatePhone(ctx, req.Phone); err != nil {
			return nil, fmt.Errorf("failed to set client phone: %w", err)
		}
	}

	if req.Address != "" {
		if err := client.UpdateAddress(ctx, req.Address); err != nil {
			return nil, fmt.Errorf("failed to set client address: %w", err)
		}
	}

	if req.TaxID != "" {
		if err := client.UpdateTaxID(ctx, req.TaxID); err != nil {
			return nil, fmt.Errorf("failed to set client tax ID: %w", err)
		}
	}

	// Store client
	if err := s.clientStorage.CreateClient(ctx, client); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToStoreClient, err)
	}

	s.logger.Info("client created successfully", "id", client.ID, "name", client.Name)
	return client, nil
}

// GetClient retrieves a client by ID
func (s *ClientService) GetClient(ctx context.Context, id models.ClientID) (*models.Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if strings.TrimSpace(string(id)) == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}

	client, err := s.clientStorage.GetClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToRetrieveClient, err)
	}

	return client, nil
}

// UpdateClient updates an existing client
func (s *ClientService) UpdateClient(ctx context.Context, client *models.Client) (*models.Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if client == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}

	s.logger.Info("updating client", "id", client.ID, "name", client.Name)

	// Validate client
	if err := client.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid client: %w", err)
	}

	// Check if email is being changed and is unique
	existingClient, err := s.clientStorage.GetClient(ctx, client.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve existing client: %w", err)
	}

	if existingClient.Email != client.Email {
		if err := s.validateUniqueClientEmail(ctx, client.Email); err != nil {
			return nil, err
		}
	}

	// Update client in storage
	if err := s.clientStorage.UpdateClient(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to update client in storage: %w", err)
	}

	s.logger.Info("client updated successfully", "id", client.ID, "name", client.Name)
	return client, nil
}

// DeleteClient deactivates a client (soft delete)
func (s *ClientService) DeleteClient(ctx context.Context, id models.ClientID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.logger.Info("deleting client", "id", id)

	// Check if client has active invoices
	hasActiveInvoices, err := s.clientHasActiveInvoices(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check client invoices: %w", err)
	}

	if hasActiveInvoices {
		return fmt.Errorf("cannot delete client with active invoices - please mark all invoices as paid or voided first")
	}

	// Soft delete client
	if err := s.clientStorage.DeleteClient(ctx, id); err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	s.logger.Info("client deleted successfully", "id", id)
	return nil
}

// ListClients retrieves clients with pagination
func (s *ClientService) ListClients(ctx context.Context, activeOnly bool, limit, offset int) (*storage.ClientListResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result, err := s.clientStorage.ListClients(ctx, activeOnly, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}

	s.logger.Debug("listed clients", "count", len(result.Clients), "total", result.TotalCount)
	return result, nil
}

// FindClientByEmail finds a client by email address
func (s *ClientService) FindClientByEmail(ctx context.Context, email string) (*models.Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if strings.TrimSpace(email) == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	client, err := s.clientStorage.FindClientByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find client by email: %w", err)
	}

	return client, nil
}

// GetClientWithInvoices returns a client along with their invoice information
func (s *ClientService) GetClientWithInvoices(ctx context.Context, id models.ClientID) (*ClientWithInvoices, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get client
	client, err := s.clientStorage.GetClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToRetrieveClient, err)
	}

	// Get client's invoices
	filter := models.InvoiceFilter{ClientID: id}
	invoiceResult, err := s.invoiceStorage.ListInvoices(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get client invoices: %w", err)
	}

	// Calculate statistics
	var totalAmount, paidAmount, outstandingAmount float64
	var activeInvoiceCount int

	for _, invoice := range invoiceResult.Invoices {
		totalAmount += invoice.Total

		switch invoice.Status {
		case models.StatusPaid:
			paidAmount += invoice.Total
		case models.StatusSent, models.StatusOverdue:
			outstandingAmount += invoice.Total
			activeInvoiceCount++
		case models.StatusDraft:
			activeInvoiceCount++
		}
	}

	result := &ClientWithInvoices{
		Client:            client,
		Invoices:          invoiceResult.Invoices,
		TotalInvoices:     len(invoiceResult.Invoices),
		ActiveInvoices:    activeInvoiceCount,
		TotalAmount:       totalAmount,
		PaidAmount:        paidAmount,
		OutstandingAmount: outstandingAmount,
	}

	return result, nil
}

// ActivateClient reactivates a deactivated client
func (s *ClientService) ActivateClient(ctx context.Context, id models.ClientID) (*models.Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("activating client", "id", id)

	// Get client
	client, err := s.clientStorage.GetClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToRetrieveClient, err)
	}

	if client.Active {
		return client, nil // Already active
	}

	// Activate client
	if err := client.Activate(ctx); err != nil {
		return nil, fmt.Errorf("failed to activate client: %w", err)
	}

	// Update in storage
	if err := s.clientStorage.UpdateClient(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to update client activation status: %w", err)
	}

	s.logger.Info("client activated successfully", "id", id, "name", client.Name)
	return client, nil
}

// DeactivateClient deactivates a client
func (s *ClientService) DeactivateClient(ctx context.Context, id models.ClientID) (*models.Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Info("deactivating client", "id", id)

	// Check if client has active invoices
	hasActiveInvoices, err := s.clientHasActiveInvoices(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check client invoices: %w", err)
	}

	if hasActiveInvoices {
		return nil, fmt.Errorf("cannot deactivate client with active invoices")
	}

	// Get client
	client, err := s.clientStorage.GetClient(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToRetrieveClient, err)
	}

	if !client.Active {
		return client, nil // Already inactive
	}

	// Deactivate client
	if err := client.Deactivate(ctx); err != nil {
		return nil, fmt.Errorf("failed to deactivate client: %w", err)
	}

	// Update in storage
	if err := s.clientStorage.UpdateClient(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to update client deactivation status: %w", err)
	}

	s.logger.Info("client deactivated successfully", "id", id, "name", client.Name)
	return client, nil
}

// GetClientStatistics returns summary statistics for all clients
func (s *ClientService) GetClientStatistics(ctx context.Context) (*ClientStatistics, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get all clients
	result, err := s.clientStorage.ListClients(ctx, false, -1, 0) // Get all clients
	if err != nil {
		return nil, fmt.Errorf("failed to get clients for statistics: %w", err)
	}

	stats := &ClientStatistics{
		TotalClients: int64(len(result.Clients)),
	}

	for _, client := range result.Clients {
		if client.Active {
			stats.ActiveClients++
		} else {
			stats.InactiveClients++
		}
	}

	return stats, nil
}

// Helper methods

func (s *ClientService) validateUniqueClientEmail(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	_, err := s.clientStorage.FindClientByEmail(ctx, email)
	if err == nil {
		return fmt.Errorf("client with email '%s' already exists", email)
	}

	if !storage.IsNotFound(err) {
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}

	return nil
}

func (s *ClientService) clientHasActiveInvoices(ctx context.Context, clientID models.ClientID) (bool, error) {
	// Check for invoices in active statuses
	activeStatuses := []string{models.StatusDraft, models.StatusSent, models.StatusOverdue}

	for _, status := range activeStatuses {
		filter := models.InvoiceFilter{
			ClientID: clientID,
			Status:   status,
			Limit:    1, // We only need to know if any exist
		}

		result, err := s.invoiceStorage.ListInvoices(ctx, filter)
		if err != nil {
			return false, err
		}

		if len(result.Invoices) > 0 {
			return true, nil
		}
	}

	return false, nil
}

// ClientWithInvoices represents a client with their invoice information
type ClientWithInvoices struct {
	Client            *models.Client    `json:"client"`
	Invoices          []*models.Invoice `json:"invoices"`
	TotalInvoices     int               `json:"total_invoices"`
	ActiveInvoices    int               `json:"active_invoices"`
	TotalAmount       float64           `json:"total_amount"`
	PaidAmount        float64           `json:"paid_amount"`
	OutstandingAmount float64           `json:"outstanding_amount"`
}

// ClientStatistics represents summary statistics for clients
type ClientStatistics struct {
	TotalClients    int64 `json:"total_clients"`
	ActiveClients   int64 `json:"active_clients"`
	InactiveClients int64 `json:"inactive_clients"`
}
