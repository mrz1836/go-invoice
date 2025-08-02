package json

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mrz/go-invoice/internal/models"
	"github.com/mrz/go-invoice/internal/storage"
)

// Client storage errors
var (
	ErrClientCannotBeNil     = fmt.Errorf("client cannot be nil")
	ErrClientIDCannotBeEmpty = fmt.Errorf("client ID cannot be empty")
	ErrEmailCannotBeEmpty    = fmt.Errorf("email cannot be empty")
)

// Client storage implementation methods for JSONStorage

// CreateClient stores a new client
func (s *JSONStorage) CreateClient(ctx context.Context, client *models.Client) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if client == nil {
		return ErrClientCannotBeNil
	}

	// Validate client
	if err := client.Validate(ctx); err != nil {
		return fmt.Errorf("invalid client: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if client already exists
	clientPath := s.getClientPath(client.ID)
	if _, err := os.Stat(clientPath); err == nil {
		return storage.NewConflictError("client", string(client.ID), "")
	}

	// Write client file atomically
	if err := s.writeJSONFile(ctx, clientPath, client); err != nil {
		return fmt.Errorf("failed to write client file: %w", err)
	}

	s.logger.Info("client created", "id", client.ID, "name", client.Name)
	return nil
}

// GetClient retrieves a client by ID
func (s *JSONStorage) GetClient(ctx context.Context, id models.ClientID) (*models.Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if strings.TrimSpace(string(id)) == "" {
		return nil, ErrClientIDCannotBeEmpty
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	clientPath := s.getClientPath(id)
	var client models.Client

	if err := s.readJSONFile(ctx, clientPath, &client); err != nil {
		if os.IsNotExist(err) {
			return nil, storage.NewNotFoundError("client", string(id))
		}
		return nil, fmt.Errorf("failed to read client file: %w", err)
	}

	return &client, nil
}

// UpdateClient updates an existing client
func (s *JSONStorage) UpdateClient(ctx context.Context, client *models.Client) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if client == nil {
		return ErrClientCannotBeNil
	}

	// Validate client
	if err := client.Validate(ctx); err != nil {
		return fmt.Errorf("invalid client: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if client exists
	clientPath := s.getClientPath(client.ID)
	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		return storage.NewNotFoundError("client", string(client.ID))
	}

	// Update timestamp
	client.UpdatedAt = time.Now()

	// Write updated client atomically
	if err := s.writeJSONFile(ctx, clientPath, client); err != nil {
		return fmt.Errorf("failed to write updated client: %w", err)
	}

	s.logger.Info("client updated", "id", client.ID, "name", client.Name)
	return nil
}

// DeleteClient removes a client by ID (soft delete - marks as inactive)
func (s *JSONStorage) DeleteClient(ctx context.Context, id models.ClientID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if strings.TrimSpace(string(id)) == "" {
		return ErrClientIDCannotBeEmpty
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Read existing client
	client, err := s.getClientUnsafe(ctx, id)
	if err != nil {
		if storage.IsNotFound(err) {
			return storage.NewNotFoundError("client", string(id))
		}
		return fmt.Errorf("failed to read client: %w", err)
	}

	// Soft delete - mark as inactive
	client.Active = false
	client.UpdatedAt = time.Now()

	// Write updated client
	clientPath := s.getClientPath(id)
	if err := s.writeJSONFile(ctx, clientPath, client); err != nil {
		return fmt.Errorf("failed to update client for deletion: %w", err)
	}

	s.logger.Info("client deleted (soft)", "id", id)
	return nil
}

// ListClients retrieves all clients with pagination
func (s *JSONStorage) ListClients(ctx context.Context, activeOnly bool, limit, offset int) (*storage.ClientListResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read all client files
	clientFiles, err := filepath.Glob(filepath.Join(s.clientsDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list client files: %w", err)
	}

	var allClients []*models.Client
	for _, filePath := range clientFiles {
		var client models.Client
		if err := s.readJSONFile(ctx, filePath, &client); err != nil {
			s.logger.Error("failed to read client file", "file", filePath, "error", err)
			continue // Skip corrupted files
		}

		// Filter by active status if requested
		if activeOnly && !client.Active {
			continue
		}

		allClients = append(allClients, &client)
	}

	// Sort clients by name
	sort.Slice(allClients, func(i, j int) bool {
		return strings.ToLower(allClients[i].Name) < strings.ToLower(allClients[j].Name)
	})

	// Apply pagination
	totalCount := int64(len(allClients))
	start := offset
	if start > len(allClients) {
		start = len(allClients)
	}

	end := start + limit
	if limit <= 0 {
		end = len(allClients)
	} else if end > len(allClients) {
		end = len(allClients)
	}

	result := &storage.ClientListResult{
		Clients:    allClients[start:end],
		TotalCount: totalCount,
		HasMore:    end < len(allClients),
	}

	if result.HasMore {
		result.NextOffset = end
	}

	return result, nil
}

// FindClientByEmail finds a client by email address
func (s *JSONStorage) FindClientByEmail(ctx context.Context, email string) (*models.Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if strings.TrimSpace(email) == "" {
		return nil, ErrEmailCannotBeEmpty
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read all client files and search for matching email
	clientFiles, err := filepath.Glob(filepath.Join(s.clientsDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list client files: %w", err)
	}

	email = strings.ToLower(strings.TrimSpace(email))

	for _, filePath := range clientFiles {
		var client models.Client
		if err := s.readJSONFile(ctx, filePath, &client); err != nil {
			s.logger.Error("failed to read client file", "file", filePath, "error", err)
			continue // Skip corrupted files
		}

		if strings.ToLower(client.Email) == email {
			return &client, nil
		}
	}

	return nil, storage.NewNotFoundError("client", fmt.Sprintf("email:%s", email))
}

// ExistsClient checks if a client exists
func (s *JSONStorage) ExistsClient(ctx context.Context, id models.ClientID) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	clientPath := s.getClientPath(id)
	_, err := os.Stat(clientPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Helper method for client operations (internal use)
func (s *JSONStorage) getClientUnsafe(ctx context.Context, id models.ClientID) (*models.Client, error) {
	clientPath := s.getClientPath(id)
	var client models.Client

	if err := s.readJSONFile(ctx, clientPath, &client); err != nil {
		if os.IsNotExist(err) {
			return nil, storage.NewNotFoundError("client", string(id))
		}
		return nil, fmt.Errorf("failed to read client file: %w", err)
	}

	return &client, nil
}

// HardDeleteClient completely removes a client from storage (for cleanup/admin use)
func (s *JSONStorage) HardDeleteClient(ctx context.Context, id models.ClientID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if strings.TrimSpace(string(id)) == "" {
		return ErrClientIDCannotBeEmpty
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	clientPath := s.getClientPath(id)

	// Check if client exists
	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		return storage.NewNotFoundError("client", string(id))
	}

	// Remove client file completely
	if err := os.Remove(clientPath); err != nil {
		return fmt.Errorf("failed to delete client file: %w", err)
	}

	s.logger.Info("client hard deleted", "id", id)
	return nil
}

// RestoreClient reactivates a soft-deleted client
func (s *JSONStorage) RestoreClient(ctx context.Context, id models.ClientID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Read existing client
	client, err := s.getClientUnsafe(ctx, id)
	if err != nil {
		return err
	}

	// Reactivate client
	client.Active = true
	client.UpdatedAt = time.Now()

	// Write updated client
	clientPath := s.getClientPath(id)
	if err := s.writeJSONFile(ctx, clientPath, client); err != nil {
		return fmt.Errorf("failed to restore client: %w", err)
	}

	s.logger.Info("client restored", "id", id)
	return nil
}

// GetClientInvoiceCount returns the number of invoices for a specific client
func (s *JSONStorage) GetClientInvoiceCount(ctx context.Context, clientID models.ClientID) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	// Use the invoice filter to count invoices for this client
	count, err := s.CountInvoices(ctx, models.InvoiceFilter{
		ClientID: clientID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count client invoices: %w", err)
	}

	return count, nil
}

// GetClientActiveInvoices returns active (unpaid) invoices for a specific client
func (s *JSONStorage) GetClientActiveInvoices(ctx context.Context, clientID models.ClientID) ([]*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get invoices that are not paid or voided
	filters := []models.InvoiceFilter{
		{ClientID: clientID, Status: models.StatusDraft},
		{ClientID: clientID, Status: models.StatusSent},
		{ClientID: clientID, Status: models.StatusOverdue},
	}

	var activeInvoices []*models.Invoice
	for _, filter := range filters {
		result, err := s.ListInvoices(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to get client invoices: %w", err)
		}
		activeInvoices = append(activeInvoices, result.Invoices...)
	}

	// Sort by due date
	sort.Slice(activeInvoices, func(i, j int) bool {
		return activeInvoices[i].DueDate.Before(activeInvoices[j].DueDate)
	})

	return activeInvoices, nil
}
