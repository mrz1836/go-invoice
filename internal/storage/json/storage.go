package json

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mrz/go-invoice/internal/models"
	"github.com/mrz/go-invoice/internal/storage"
)

// Invoice storage errors
var (
	ErrInvoiceCannotBeNil     = fmt.Errorf("invoice cannot be nil")
	ErrInvoiceIDCannotBeEmpty = fmt.Errorf("invoice ID cannot be empty")
)

// JSONStorage provides file-based JSON storage with concurrent safety
type JSONStorage struct {
	basePath    string
	invoicesDir string
	clientsDir  string
	indexDir    string
	backupDir   string
	mu          sync.RWMutex
	initialized bool
	stats       *storage.StorageStats
	logger      Logger
}

// Logger interface for storage operations
type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
	Debug(msg string, fields ...any)
}

// NewJSONStorage creates a new JSON storage instance
func NewJSONStorage(basePath string, logger Logger) *JSONStorage {
	return &JSONStorage{
		basePath:    basePath,
		invoicesDir: filepath.Join(basePath, "invoices"),
		clientsDir:  filepath.Join(basePath, "clients"),
		indexDir:    filepath.Join(basePath, "index"),
		backupDir:   filepath.Join(basePath, "backups"),
		logger:      logger,
		stats: &storage.StorageStats{
			HealthStatus: storage.HealthStatusHealthy,
		},
	}
}

// Initialize sets up the storage directory structure
func (s *JSONStorage) Initialize(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("initializing JSON storage", "path", s.basePath)

	// Create all required directories
	dirs := []string{
		s.basePath,
		s.invoicesDir,
		s.clientsDir,
		s.indexDir,
		s.backupDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return storage.NewStorageUnavailableError(
				fmt.Sprintf("failed to create directory %s", dir), err)
		}
	}

	// Create metadata file
	metadataPath := filepath.Join(s.basePath, "metadata.json")
	metadata := map[string]interface{}{
		"version":      "1.0",
		"created_at":   time.Now().Format(time.RFC3339),
		"storage_type": "json",
	}

	if err := s.writeJSONFile(ctx, metadataPath, metadata); err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}

	// Initialize index files
	if err := s.initializeIndexes(ctx); err != nil {
		return fmt.Errorf("failed to initialize indexes: %w", err)
	}

	s.initialized = true
	s.logger.Info("JSON storage initialized successfully")
	return nil
}

// IsInitialized checks if the storage is properly initialized
func (s *JSONStorage) IsInitialized(ctx context.Context) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.initialized {
		return true, nil
	}

	// Check if directories exist
	dirs := []string{s.basePath, s.invoicesDir, s.clientsDir, s.indexDir}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return false, nil
		}
	}

	// Check metadata file
	metadataPath := filepath.Join(s.basePath, "metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return false, nil
	}

	s.initialized = true
	return true, nil
}

// GetStorageInfo returns information about the storage system
func (s *JSONStorage) GetStorageInfo(ctx context.Context) (*storage.StorageInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	initialized, err := s.IsInitialized(ctx)
	if err != nil {
		return nil, err
	}

	return &storage.StorageInfo{
		Type:             "json",
		Version:          "1.0",
		Path:             s.basePath,
		Initialized:      initialized,
		ReadOnly:         false,
		SupportsBackups:  true,
		SupportsIndexing: true,
	}, nil
}

// Validate performs integrity checks on the storage system
func (s *JSONStorage) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.logger.Info("validating storage integrity")

	// Check directory structure
	dirs := []string{s.basePath, s.invoicesDir, s.clientsDir, s.indexDir}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			return storage.NewStorageUnavailableError(
				fmt.Sprintf("directory %s is not accessible", dir), err)
		}
	}

	// Validate invoice files
	if err := s.validateInvoiceFiles(ctx); err != nil {
		return fmt.Errorf("invoice validation failed: %w", err)
	}

	// Validate client files
	if err := s.validateClientFiles(ctx); err != nil {
		return fmt.Errorf("client validation failed: %w", err)
	}

	s.logger.Info("storage validation completed successfully")
	return nil
}

// CreateInvoice stores a new invoice
func (s *JSONStorage) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if invoice == nil {
		return ErrInvoiceCannotBeNil
	}

	// Validate invoice
	if err := invoice.Validate(ctx); err != nil {
		return fmt.Errorf("invalid invoice: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if invoice already exists
	invoicePath := s.getInvoicePath(invoice.ID)
	if _, err := os.Stat(invoicePath); err == nil {
		return storage.NewConflictError("invoice", string(invoice.ID), "")
	}

	// Write invoice file atomically
	if err := s.writeJSONFile(ctx, invoicePath, invoice); err != nil {
		return fmt.Errorf("failed to write invoice file: %w", err)
	}

	// Update index
	if err := s.updateInvoiceIndex(ctx, invoice, "create"); err != nil {
		s.logger.Error("failed to update invoice index", "error", err, "invoice_id", invoice.ID)
		// Don't fail the operation for index errors
	}

	s.logger.Info("invoice created", "id", invoice.ID, "number", invoice.Number)
	return nil
}

// GetInvoice retrieves an invoice by ID
func (s *JSONStorage) GetInvoice(ctx context.Context, id models.InvoiceID) (*models.Invoice, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if strings.TrimSpace(string(id)) == "" {
		return nil, ErrInvoiceIDCannotBeEmpty
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	invoicePath := s.getInvoicePath(id)
	var invoice models.Invoice

	if err := s.readJSONFile(ctx, invoicePath, &invoice); err != nil {
		if os.IsNotExist(err) {
			return nil, storage.NewNotFoundError("invoice", string(id))
		}
		return nil, fmt.Errorf("failed to read invoice file: %w", err)
	}

	return &invoice, nil
}

// UpdateInvoice updates an existing invoice with optimistic locking
func (s *JSONStorage) UpdateInvoice(ctx context.Context, invoice *models.Invoice) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if invoice == nil {
		return ErrInvoiceCannotBeNil
	}

	// Validate invoice
	if err := invoice.Validate(ctx); err != nil {
		return fmt.Errorf("invalid invoice: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Read existing invoice for optimistic locking
	existing, err := s.getInvoiceUnsafe(ctx, invoice.ID)
	if err != nil {
		if storage.IsNotFound(err) {
			return storage.NewNotFoundError("invoice", string(invoice.ID))
		}
		return fmt.Errorf("failed to read existing invoice: %w", err)
	}

	// Check version for optimistic locking
	if existing.Version != invoice.Version {
		return storage.NewVersionMismatchError("invoice", string(invoice.ID),
			invoice.Version, existing.Version)
	}

	// Increment version
	invoice.Version++
	invoice.UpdatedAt = time.Now()

	// Write updated invoice atomically
	invoicePath := s.getInvoicePath(invoice.ID)
	if err := s.writeJSONFile(ctx, invoicePath, invoice); err != nil {
		return fmt.Errorf("failed to write updated invoice: %w", err)
	}

	// Update index
	if err := s.updateInvoiceIndex(ctx, invoice, "update"); err != nil {
		s.logger.Error("failed to update invoice index", "error", err, "invoice_id", invoice.ID)
	}

	s.logger.Info("invoice updated", "id", invoice.ID, "version", invoice.Version)
	return nil
}

// DeleteInvoice removes an invoice by ID
func (s *JSONStorage) DeleteInvoice(ctx context.Context, id models.InvoiceID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if strings.TrimSpace(string(id)) == "" {
		return ErrInvoiceIDCannotBeEmpty
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	invoicePath := s.getInvoicePath(id)

	// Check if invoice exists
	if _, err := os.Stat(invoicePath); os.IsNotExist(err) {
		return storage.NewNotFoundError("invoice", string(id))
	}

	// Remove invoice file
	if err := os.Remove(invoicePath); err != nil {
		return fmt.Errorf("failed to delete invoice file: %w", err)
	}

	// Update index
	if err := s.updateInvoiceIndex(ctx, &models.Invoice{ID: id}, "delete"); err != nil {
		s.logger.Error("failed to update invoice index", "error", err, "invoice_id", id)
	}

	s.logger.Info("invoice deleted", "id", id)
	return nil
}

// ExistsInvoice checks if an invoice exists
func (s *JSONStorage) ExistsInvoice(ctx context.Context, id models.InvoiceID) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	invoicePath := s.getInvoicePath(id)
	_, err := os.Stat(invoicePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ListInvoices retrieves invoices based on filter criteria
func (s *JSONStorage) ListInvoices(ctx context.Context, filter models.InvoiceFilter) (*storage.InvoiceListResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate filter
	if err := filter.Validate(ctx); err != nil {
		return nil, storage.NewInvalidFilterError("filter", filter, err.Error())
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read all invoice files
	invoiceFiles, err := filepath.Glob(filepath.Join(s.invoicesDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list invoice files: %w", err)
	}

	var allInvoices []*models.Invoice
	for _, filePath := range invoiceFiles {
		var invoice models.Invoice
		if err := s.readJSONFile(ctx, filePath, &invoice); err != nil {
			s.logger.Error("failed to read invoice file", "file", filePath, "error", err)
			continue // Skip corrupted files
		}

		// Apply filter
		if s.matchesFilter(&invoice, filter) {
			allInvoices = append(allInvoices, &invoice)
		}
	}

	// Sort invoices (by date descending by default)
	sort.Slice(allInvoices, func(i, j int) bool {
		return allInvoices[i].Date.After(allInvoices[j].Date)
	})

	// Apply pagination
	totalCount := int64(len(allInvoices))
	start := filter.Offset
	if start > len(allInvoices) {
		start = len(allInvoices)
	}

	end := start + filter.Limit
	if filter.Limit <= 0 {
		end = len(allInvoices)
	} else if end > len(allInvoices) {
		end = len(allInvoices)
	}

	result := &storage.InvoiceListResult{
		Invoices:   allInvoices[start:end],
		TotalCount: totalCount,
		HasMore:    end < len(allInvoices),
	}

	if result.HasMore {
		result.NextOffset = end
	}

	return result, nil
}

// CountInvoices returns the total count of invoices matching the filter
func (s *JSONStorage) CountInvoices(ctx context.Context, filter models.InvoiceFilter) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	// Use ListInvoices but only count
	result, err := s.ListInvoices(ctx, models.InvoiceFilter{
		Status:      filter.Status,
		ClientID:    filter.ClientID,
		DateFrom:    filter.DateFrom,
		DateTo:      filter.DateTo,
		DueDateFrom: filter.DueDateFrom,
		DueDateTo:   filter.DueDateTo,
		AmountMin:   filter.AmountMin,
		AmountMax:   filter.AmountMax,
		Limit:       0, // No limit for counting
	})
	if err != nil {
		return 0, err
	}

	return result.TotalCount, nil
}

// Helper methods

func (s *JSONStorage) getInvoicePath(id models.InvoiceID) string {
	return filepath.Join(s.invoicesDir, fmt.Sprintf("%s.json", string(id)))
}

func (s *JSONStorage) getClientPath(id models.ClientID) string {
	return filepath.Join(s.clientsDir, fmt.Sprintf("%s.json", string(id)))
}

func (s *JSONStorage) getInvoiceUnsafe(ctx context.Context, id models.InvoiceID) (*models.Invoice, error) {
	invoicePath := s.getInvoicePath(id)
	var invoice models.Invoice

	if err := s.readJSONFile(ctx, invoicePath, &invoice); err != nil {
		if os.IsNotExist(err) {
			return nil, storage.NewNotFoundError("invoice", string(id))
		}
		return nil, fmt.Errorf("failed to read invoice file: %w", err)
	}

	return &invoice, nil
}

func (s *JSONStorage) writeJSONFile(ctx context.Context, path string, data interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create temporary file for atomic write
	tempPath := path + ".tmp"
	file, err := os.Create(tempPath) // #nosec G304 -- Path is derived from validated storage directory
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			s.logger.Error("failed to close temp file", "error", err, "path", tempPath)
		}
	}()

	// Check context after file creation
	select {
	case <-ctx.Done():
		if removeErr := os.Remove(tempPath); removeErr != nil {
			s.logger.Error("failed to remove temp file", "path", tempPath, "error", removeErr)
		}
		return ctx.Err()
	default:
	}

	// Encode JSON with indentation for readability
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		if removeErr := os.Remove(tempPath); removeErr != nil {
			s.logger.Error("failed to remove temp file", "path", tempPath, "error", removeErr)
		}
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	// Check context after JSON encoding
	select {
	case <-ctx.Done():
		if removeErr := os.Remove(tempPath); removeErr != nil {
			s.logger.Error("failed to remove temp file", "path", tempPath, "error", removeErr)
		}
		return ctx.Err()
	default:
	}

	// Sync to disk
	if err := file.Sync(); err != nil {
		if removeErr := os.Remove(tempPath); removeErr != nil {
			s.logger.Error("failed to remove temp file", "path", tempPath, "error", removeErr)
		}
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// Check context after sync
	select {
	case <-ctx.Done():
		if removeErr := os.Remove(tempPath); removeErr != nil {
			s.logger.Error("failed to remove temp file", "path", tempPath, "error", removeErr)
		}
		return ctx.Err()
	default:
	}

	if err := file.Close(); err != nil {
		s.logger.Error("failed to close temp file before rename", "error", err, "path", tempPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Check context before final rename
	select {
	case <-ctx.Done():
		if removeErr := os.Remove(tempPath); removeErr != nil {
			s.logger.Error("failed to remove temp file", "path", tempPath, "error", removeErr)
		}
		return ctx.Err()
	default:
	}

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		if removeErr := os.Remove(tempPath); removeErr != nil {
			s.logger.Error("failed to remove temp file", "path", tempPath, "error", removeErr)
		}
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func (s *JSONStorage) readJSONFile(ctx context.Context, path string, data interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	file, err := os.Open(path) // #nosec G304 -- Path is derived from validated storage directory
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			s.logger.Error("failed to close file", "error", err, "path", path)
		}
	}()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

func (s *JSONStorage) matchesFilter(invoice *models.Invoice, filter models.InvoiceFilter) bool {
	// Status filter
	if filter.Status != "" && invoice.Status != filter.Status {
		return false
	}

	// Client ID filter
	if filter.ClientID != "" && invoice.Client.ID != filter.ClientID {
		return false
	}

	// Date range filters
	if !filter.DateFrom.IsZero() && invoice.Date.Before(filter.DateFrom) {
		return false
	}
	if !filter.DateTo.IsZero() && invoice.Date.After(filter.DateTo) {
		return false
	}

	// Due date range filters
	if !filter.DueDateFrom.IsZero() && invoice.DueDate.Before(filter.DueDateFrom) {
		return false
	}
	if !filter.DueDateTo.IsZero() && invoice.DueDate.After(filter.DueDateTo) {
		return false
	}

	// Amount range filters
	if filter.AmountMin > 0 && invoice.Total < filter.AmountMin {
		return false
	}
	if filter.AmountMax > 0 && invoice.Total > filter.AmountMax {
		return false
	}

	return true
}

func (s *JSONStorage) initializeIndexes(ctx context.Context) error {
	// Create invoice index file
	invoiceIndexPath := filepath.Join(s.indexDir, "invoices.json")
	invoiceIndex := make(map[string]interface{})
	if err := s.writeJSONFile(ctx, invoiceIndexPath, invoiceIndex); err != nil {
		return fmt.Errorf("failed to create invoice index: %w", err)
	}

	// Create client index file
	clientIndexPath := filepath.Join(s.indexDir, "clients.json")
	clientIndex := make(map[string]interface{})
	if err := s.writeJSONFile(ctx, clientIndexPath, clientIndex); err != nil {
		return fmt.Errorf("failed to create client index: %w", err)
	}

	return nil
}

func (s *JSONStorage) updateInvoiceIndex(_ context.Context, invoice *models.Invoice, operation string) error { //nolint:unparam // Placeholder for future index implementation
	// For now, this is a placeholder - a full implementation would maintain
	// search indexes for faster querying
	s.logger.Debug("updating invoice index", "invoice_id", invoice.ID, "operation", operation)
	return nil
}

func (s *JSONStorage) validateInvoiceFiles(ctx context.Context) error {
	invoiceFiles, err := filepath.Glob(filepath.Join(s.invoicesDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list invoice files: %w", err)
	}

	for _, filePath := range invoiceFiles {
		var invoice models.Invoice
		if err := s.readJSONFile(ctx, filePath, &invoice); err != nil {
			return storage.NewCorruptedError("invoice", filepath.Base(filePath), err.Error())
		}

		if err := invoice.Validate(ctx); err != nil {
			return storage.NewCorruptedError("invoice", string(invoice.ID), err.Error())
		}
	}

	return nil
}

func (s *JSONStorage) validateClientFiles(ctx context.Context) error {
	clientFiles, err := filepath.Glob(filepath.Join(s.clientsDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list client files: %w", err)
	}

	for _, filePath := range clientFiles {
		var client models.Client
		if err := s.readJSONFile(ctx, filePath, &client); err != nil {
			return storage.NewCorruptedError("client", filepath.Base(filePath), err.Error())
		}

		if err := client.Validate(ctx); err != nil {
			return storage.NewCorruptedError("client", string(client.ID), err.Error())
		}
	}

	return nil
}
