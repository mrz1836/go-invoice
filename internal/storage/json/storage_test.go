package json

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz/go-invoice/internal/models"
	storageTypes "github.com/mrz/go-invoice/internal/storage"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	mu       sync.Mutex
	messages []LogMessage
}

type LogMessage struct {
	Level  string
	Msg    string
	Fields []any
}

func (m *MockLogger) Info(msg string, fields ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, LogMessage{Level: "info", Msg: msg, Fields: fields})
}

func (m *MockLogger) Error(msg string, fields ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, LogMessage{Level: "error", Msg: msg, Fields: fields})
}

func (m *MockLogger) Debug(msg string, fields ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, LogMessage{Level: "debug", Msg: msg, Fields: fields})
}

func (m *MockLogger) GetMessages() []LogMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]LogMessage, len(m.messages))
	copy(result, m.messages)
	return result
}

func (m *MockLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = nil
}

// JSONStorageTestSuite tests JSON storage operations
type JSONStorageTestSuite struct {
	suite.Suite
	ctx        context.Context
	cancelFunc context.CancelFunc
	storage    *JSONStorage
	tempDir    string
	logger     *MockLogger
}

func (suite *JSONStorageTestSuite) SetupTest() {
	suite.ctx, suite.cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)

	// Create temporary directory for tests
	tempDir, err := os.MkdirTemp("", "json-storage-test-*")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir

	// Create mock logger
	suite.logger = &MockLogger{}

	// Create storage instance
	suite.storage = NewJSONStorage(suite.tempDir, suite.logger)
}

func (suite *JSONStorageTestSuite) TearDownTest() {
	suite.cancelFunc()

	// Clean up temporary directory
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

func TestJSONStorageTestSuite(t *testing.T) {
	suite.Run(t, new(JSONStorageTestSuite))
}

func (suite *JSONStorageTestSuite) TestNewJSONStorage() {
	t := suite.T()

	storage := NewJSONStorage("/test/path", suite.logger)
	require.NotNil(t, storage)

	assert.Equal(t, "/test/path", storage.basePath)
	assert.Equal(t, filepath.Join("/test/path", "invoices"), storage.invoicesDir)
	assert.Equal(t, filepath.Join("/test/path", "clients"), storage.clientsDir)
	assert.Equal(t, filepath.Join("/test/path", "index"), storage.indexDir)
	assert.Equal(t, filepath.Join("/test/path", "backups"), storage.backupDir)
	assert.NotNil(t, storage.logger)
	assert.NotNil(t, storage.stats)
	assert.Equal(t, storageTypes.HealthStatusHealthy, storage.stats.HealthStatus)
	assert.False(t, storage.initialized)
}

func (suite *JSONStorageTestSuite) TestInitialize() {
	t := suite.T()

	// Test successful initialization
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Verify directories were created
	dirs := []string{
		suite.tempDir,
		filepath.Join(suite.tempDir, "invoices"),
		filepath.Join(suite.tempDir, "clients"),
		filepath.Join(suite.tempDir, "index"),
		filepath.Join(suite.tempDir, "backups"),
	}

	for _, dir := range dirs {
		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, info.IsDir(), "Expected %s to be a directory", dir)
	}

	// Verify metadata file was created
	metadataPath := filepath.Join(suite.tempDir, "metadata.json")
	info, err := os.Stat(metadataPath)
	require.NoError(t, err)
	assert.False(t, info.IsDir())

	// Read and verify metadata content
	var metadata map[string]interface{}
	data, err := os.ReadFile(metadataPath)
	require.NoError(t, err)
	err = json.Unmarshal(data, &metadata)
	require.NoError(t, err)

	assert.Equal(t, "1.0", metadata["version"])
	assert.Equal(t, "json", metadata["storage_type"])
	assert.NotEmpty(t, metadata["created_at"])

	// Verify index files were created
	invoiceIndexPath := filepath.Join(suite.tempDir, "index", "invoices.json")
	_, err = os.Stat(invoiceIndexPath)
	require.NoError(t, err)

	clientIndexPath := filepath.Join(suite.tempDir, "index", "clients.json")
	_, err = os.Stat(clientIndexPath)
	require.NoError(t, err)

	// Verify storage is marked as initialized
	assert.True(t, suite.storage.initialized)

	// Verify log messages
	messages := suite.logger.GetMessages()
	assert.Greater(t, len(messages), 0)
	assert.Equal(t, "info", messages[0].Level)
	assert.Contains(t, messages[0].Msg, "initializing JSON storage")
}

func (suite *JSONStorageTestSuite) TestInitializeWithContextCancellation() {
	t := suite.T()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := suite.storage.Initialize(ctx)
	assert.Equal(t, context.Canceled, err)
}

func (suite *JSONStorageTestSuite) TestInitializeWithDirectoryCreationError() {
	t := suite.T()

	// Create a file where directory should be
	badPath := filepath.Join(suite.tempDir, "bad-storage")
	err := os.WriteFile(badPath, []byte("content"), 0o644)
	require.NoError(t, err)

	storage := NewJSONStorage(filepath.Join(badPath, "subdir"), suite.logger)
	err = storage.Initialize(suite.ctx)
	require.Error(t, err)

	var storageErr storageTypes.ErrStorageUnavailable
	assert.ErrorAs(t, err, &storageErr)
	assert.Contains(t, err.Error(), "failed to create directory")
}

func (suite *JSONStorageTestSuite) TestIsInitialized() {
	t := suite.T()

	// Test uninitialized storage
	initialized, err := suite.storage.IsInitialized(suite.ctx)
	require.NoError(t, err)
	assert.False(t, initialized)

	// Initialize storage
	err = suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Test initialized storage
	initialized, err = suite.storage.IsInitialized(suite.ctx)
	require.NoError(t, err)
	assert.True(t, initialized)

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	initialized, err = suite.storage.IsInitialized(ctx)
	assert.Equal(t, context.Canceled, err)
	assert.False(t, initialized)
}

func (suite *JSONStorageTestSuite) TestIsInitializedWithMissingDirectory() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Remove a required directory
	err = os.RemoveAll(filepath.Join(suite.tempDir, "invoices"))
	require.NoError(t, err)

	// Reset initialized flag
	suite.storage.initialized = false

	// Check initialization status
	initialized, err := suite.storage.IsInitialized(suite.ctx)
	require.NoError(t, err)
	assert.False(t, initialized)
}

func (suite *JSONStorageTestSuite) TestIsInitializedWithMissingMetadata() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Remove metadata file
	err = os.Remove(filepath.Join(suite.tempDir, "metadata.json"))
	require.NoError(t, err)

	// Reset initialized flag
	suite.storage.initialized = false

	// Check initialization status
	initialized, err := suite.storage.IsInitialized(suite.ctx)
	require.NoError(t, err)
	assert.False(t, initialized)
}

func (suite *JSONStorageTestSuite) TestGetStorageInfo() {
	t := suite.T()

	// Test before initialization
	info, err := suite.storage.GetStorageInfo(suite.ctx)
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, "json", info.Type)
	assert.Equal(t, "1.0", info.Version)
	assert.Equal(t, suite.tempDir, info.Path)
	assert.False(t, info.Initialized)
	assert.False(t, info.ReadOnly)
	assert.True(t, info.SupportsBackups)
	assert.True(t, info.SupportsIndexing)

	// Initialize storage
	err = suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Test after initialization
	info, err = suite.storage.GetStorageInfo(suite.ctx)
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.True(t, info.Initialized)

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	info, err = suite.storage.GetStorageInfo(ctx)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, info)
}

func (suite *JSONStorageTestSuite) TestValidate() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Test successful validation
	err = suite.storage.Validate(suite.ctx)
	require.NoError(t, err)

	// Create a valid invoice file
	invoice := &models.Invoice{
		ID:     "INV-001",
		Number: "INV-2024-001",
		Client: models.Client{
			ID:        "CLIENT-001",
			Name:      "Test Client",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Version:   1,
		Date:      time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
		Status:    models.StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	invoiceData, err := json.MarshalIndent(invoice, "", "  ")
	require.NoError(t, err)

	invoicePath := filepath.Join(suite.tempDir, "invoices", "INV-001.json")
	err = os.WriteFile(invoicePath, invoiceData, 0o644)
	require.NoError(t, err)

	// Validate with valid files
	err = suite.storage.Validate(suite.ctx)
	require.NoError(t, err)

	// Test with corrupted invoice file
	err = os.WriteFile(invoicePath, []byte("invalid json"), 0o644)
	require.NoError(t, err)

	err = suite.storage.Validate(suite.ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invoice validation failed")

	// Test with missing directory
	err = os.RemoveAll(filepath.Join(suite.tempDir, "invoices"))
	require.NoError(t, err)

	err = suite.storage.Validate(suite.ctx)
	require.Error(t, err)
	var storageErr storageTypes.ErrStorageUnavailable
	assert.ErrorAs(t, err, &storageErr)
}

func (suite *JSONStorageTestSuite) TestCreateInvoice() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Create test invoice
	invoice := &models.Invoice{
		ID:     "INV-001",
		Number: "INV-2024-001",
		Client: models.Client{
			ID:        "CLIENT-001",
			Name:      "Test Client",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Version:   1,
		Date:      time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
		Status:    models.StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test successful creation
	err = suite.storage.CreateInvoice(suite.ctx, invoice)
	require.NoError(t, err)

	// Verify file was created
	invoicePath := filepath.Join(suite.tempDir, "invoices", "INV-001.json")
	_, err = os.Stat(invoicePath)
	require.NoError(t, err)

	// Verify file content
	var savedInvoice models.Invoice
	data, err := os.ReadFile(invoicePath)
	require.NoError(t, err)
	err = json.Unmarshal(data, &savedInvoice)
	require.NoError(t, err)

	assert.Equal(t, invoice.ID, savedInvoice.ID)
	assert.Equal(t, invoice.Number, savedInvoice.Number)
	assert.Equal(t, invoice.Client.ID, savedInvoice.Client.ID)

	// Test duplicate creation
	err = suite.storage.CreateInvoice(suite.ctx, invoice)
	require.Error(t, err)
	var conflictErr storageTypes.ErrConflict
	assert.ErrorAs(t, err, &conflictErr)
	assert.Equal(t, "invoice", conflictErr.Resource)
	assert.Equal(t, "INV-001", conflictErr.ID)

	// Test with nil invoice
	err = suite.storage.CreateInvoice(suite.ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invoice cannot be nil")

	// Test with invalid invoice
	invalidInvoice := &models.Invoice{
		ID: "", // Invalid - empty ID
	}
	err = suite.storage.CreateInvoice(suite.ctx, invalidInvoice)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid invoice")

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = suite.storage.CreateInvoice(ctx, invoice)
	assert.Equal(t, context.Canceled, err)
}

func (suite *JSONStorageTestSuite) TestGetInvoice() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Create test invoice
	invoice := &models.Invoice{
		ID:     "INV-001",
		Number: "INV-2024-001",
		Client: models.Client{
			ID:        "CLIENT-001",
			Name:      "Test Client",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Version:   1,
		Date:      time.Now().Truncate(time.Second),
		DueDate:   time.Now().AddDate(0, 0, 30).Truncate(time.Second),
		Status:    models.StatusDraft,
		CreatedAt: time.Now().Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}

	// Save invoice
	err = suite.storage.CreateInvoice(suite.ctx, invoice)
	require.NoError(t, err)

	// Test successful retrieval
	retrieved, err := suite.storage.GetInvoice(suite.ctx, "INV-001")
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, invoice.ID, retrieved.ID)
	assert.Equal(t, invoice.Number, retrieved.Number)
	assert.Equal(t, invoice.Client.ID, retrieved.Client.ID)
	assert.WithinDuration(t, invoice.Date, retrieved.Date, time.Second)
	assert.WithinDuration(t, invoice.DueDate, retrieved.DueDate, time.Second)

	// Test non-existent invoice
	retrieved, err = suite.storage.GetInvoice(suite.ctx, "INV-999")
	require.Error(t, err)
	assert.Nil(t, retrieved)
	var notFoundErr storageTypes.ErrNotFound
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, "invoice", notFoundErr.Resource)
	assert.Equal(t, "INV-999", notFoundErr.ID)

	// Test with empty ID
	retrieved, err = suite.storage.GetInvoice(suite.ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invoice ID cannot be empty")

	// Test with whitespace ID
	retrieved, err = suite.storage.GetInvoice(suite.ctx, "   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invoice ID cannot be empty")

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	retrieved, err = suite.storage.GetInvoice(ctx, "INV-001")
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, retrieved)
}

func (suite *JSONStorageTestSuite) TestUpdateInvoice() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Create test invoice
	invoice := &models.Invoice{
		ID:      "INV-001",
		Number:  "INV-2024-001",
		Version: 1,
		Client: models.Client{
			ID:        "CLIENT-001",
			Name:      "Test Client",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Date:      time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
		Status:    models.StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save invoice
	err = suite.storage.CreateInvoice(suite.ctx, invoice)
	require.NoError(t, err)

	// Update invoice
	invoice.Status = models.StatusSent
	invoice.Description = "Updated description"

	err = suite.storage.UpdateInvoice(suite.ctx, invoice)
	require.NoError(t, err)

	// Verify version was incremented
	assert.Equal(t, 2, invoice.Version)

	// Verify changes were saved
	retrieved, err := suite.storage.GetInvoice(suite.ctx, "INV-001")
	require.NoError(t, err)
	assert.Equal(t, models.StatusSent, retrieved.Status)
	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, 2, retrieved.Version)

	// Test optimistic locking
	outdatedInvoice := &models.Invoice{
		ID:        "INV-001",
		Number:    "INV-2024-001",
		Version:   1, // Old version
		Client:    invoice.Client,
		Date:      invoice.Date,
		DueDate:   invoice.DueDate,
		Status:    models.StatusPaid,
		CreatedAt: invoice.CreatedAt,
		UpdatedAt: invoice.UpdatedAt,
	}

	err = suite.storage.UpdateInvoice(suite.ctx, outdatedInvoice)
	require.Error(t, err)
	var versionErr storageTypes.ErrVersionMismatch
	assert.ErrorAs(t, err, &versionErr)
	assert.Equal(t, "invoice", versionErr.Resource)
	assert.Equal(t, "INV-001", versionErr.ID)
	assert.Equal(t, 1, versionErr.ExpectedVersion)
	assert.Equal(t, 2, versionErr.ActualVersion)

	// Test updating non-existent invoice
	nonExistent := &models.Invoice{
		ID:        "INV-999",
		Number:    "INV-2024-999",
		Version:   1,
		Client:    invoice.Client,
		Date:      invoice.Date,
		DueDate:   invoice.DueDate,
		Status:    models.StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = suite.storage.UpdateInvoice(suite.ctx, nonExistent)
	require.Error(t, err)
	var notFoundErr storageTypes.ErrNotFound
	assert.ErrorAs(t, err, &notFoundErr)

	// Test with nil invoice
	err = suite.storage.UpdateInvoice(suite.ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invoice cannot be nil")

	// Test with invalid invoice
	invalidInvoice := &models.Invoice{
		ID:      "INV-001",
		Version: 2,
		// Missing required fields - will fail validation
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = suite.storage.UpdateInvoice(suite.ctx, invalidInvoice)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid invoice")
}

func (suite *JSONStorageTestSuite) TestDeleteInvoice() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Create test invoice
	invoice := &models.Invoice{
		ID:     "INV-001",
		Number: "INV-2024-001",
		Client: models.Client{
			ID:        "CLIENT-001",
			Name:      "Test Client",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Version:   1,
		Date:      time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
		Status:    models.StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save invoice
	err = suite.storage.CreateInvoice(suite.ctx, invoice)
	require.NoError(t, err)

	// Verify file exists
	invoicePath := filepath.Join(suite.tempDir, "invoices", "INV-001.json")
	_, err = os.Stat(invoicePath)
	require.NoError(t, err)

	// Delete invoice
	err = suite.storage.DeleteInvoice(suite.ctx, "INV-001")
	require.NoError(t, err)

	// Verify file was deleted
	_, err = os.Stat(invoicePath)
	assert.True(t, os.IsNotExist(err))

	// Try to delete again
	err = suite.storage.DeleteInvoice(suite.ctx, "INV-001")
	require.Error(t, err)
	var notFoundErr storageTypes.ErrNotFound
	assert.ErrorAs(t, err, &notFoundErr)

	// Test with empty ID
	err = suite.storage.DeleteInvoice(suite.ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invoice ID cannot be empty")

	// Test with whitespace ID
	err = suite.storage.DeleteInvoice(suite.ctx, "   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invoice ID cannot be empty")

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = suite.storage.DeleteInvoice(ctx, "INV-001")
	assert.Equal(t, context.Canceled, err)
}

func (suite *JSONStorageTestSuite) TestExistsInvoice() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Test non-existent invoice
	exists, err := suite.storage.ExistsInvoice(suite.ctx, "INV-001")
	require.NoError(t, err)
	assert.False(t, exists)

	// Create test invoice
	invoice := &models.Invoice{
		ID:     "INV-001",
		Number: "INV-2024-001",
		Client: models.Client{
			ID:        "CLIENT-001",
			Name:      "Test Client",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Version:   1,
		Date:      time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
		Status:    models.StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save invoice
	err = suite.storage.CreateInvoice(suite.ctx, invoice)
	require.NoError(t, err)

	// Test existing invoice
	exists, err = suite.storage.ExistsInvoice(suite.ctx, "INV-001")
	require.NoError(t, err)
	assert.True(t, exists)

	// Delete invoice
	err = suite.storage.DeleteInvoice(suite.ctx, "INV-001")
	require.NoError(t, err)

	// Test after deletion
	exists, err = suite.storage.ExistsInvoice(suite.ctx, "INV-001")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	exists, err = suite.storage.ExistsInvoice(ctx, "INV-001")
	assert.Equal(t, context.Canceled, err)
	assert.False(t, exists)
}

func (suite *JSONStorageTestSuite) TestListInvoices() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Test empty list
	result, err := suite.storage.ListInvoices(suite.ctx, models.InvoiceFilter{})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Invoices)
	assert.Equal(t, int64(0), result.TotalCount)
	assert.False(t, result.HasMore)

	// Create test invoices
	now := time.Now()
	client1 := models.Client{
		ID:        "CLIENT-001",
		Name:      "Client One",
		Email:     "client1@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}
	client2 := models.Client{
		ID:        "CLIENT-002",
		Name:      "Client Two",
		Email:     "client2@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	invoices := []*models.Invoice{
		{
			ID:        "INV-001",
			Number:    "INV-2024-001",
			Client:    client1,
			Version:   1,
			Date:      now.AddDate(0, -2, 0),
			DueDate:   now.AddDate(0, -1, 0),
			Status:    models.StatusPaid,
			Total:     1000.0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "INV-002",
			Number:    "INV-2024-002",
			Client:    client1,
			Version:   1,
			Date:      now.AddDate(0, -1, 0),
			DueDate:   now,
			Status:    models.StatusSent,
			Total:     2000.0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "INV-003",
			Number:    "INV-2024-003",
			Client:    client2,
			Version:   1,
			Date:      now,
			DueDate:   now.AddDate(0, 1, 0),
			Status:    models.StatusDraft,
			Total:     3000.0,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// Save invoices
	for _, inv := range invoices {
		err = suite.storage.CreateInvoice(suite.ctx, inv)
		require.NoError(t, err)
	}

	// Test list all
	result, err = suite.storage.ListInvoices(suite.ctx, models.InvoiceFilter{})
	require.NoError(t, err)
	assert.Len(t, result.Invoices, 3)
	assert.Equal(t, int64(3), result.TotalCount)
	assert.False(t, result.HasMore)

	// Verify default sorting (date descending)
	assert.Equal(t, models.InvoiceID("INV-003"), result.Invoices[0].ID)
	assert.Equal(t, models.InvoiceID("INV-002"), result.Invoices[1].ID)
	assert.Equal(t, models.InvoiceID("INV-001"), result.Invoices[2].ID)

	// Test filter by status
	result, err = suite.storage.ListInvoices(suite.ctx, models.InvoiceFilter{
		Status: models.StatusPaid,
	})
	require.NoError(t, err)
	assert.Len(t, result.Invoices, 1)
	assert.Equal(t, models.InvoiceID("INV-001"), result.Invoices[0].ID)

	// Test filter by client
	result, err = suite.storage.ListInvoices(suite.ctx, models.InvoiceFilter{
		ClientID: "CLIENT-001",
	})
	require.NoError(t, err)
	assert.Len(t, result.Invoices, 2)

	// Test date range filter
	result, err = suite.storage.ListInvoices(suite.ctx, models.InvoiceFilter{
		DateFrom: now.AddDate(0, -2, 0),
		DateTo:   now.AddDate(0, -1, 0),
	})
	require.NoError(t, err)
	assert.Len(t, result.Invoices, 2)

	// Test amount range filter
	result, err = suite.storage.ListInvoices(suite.ctx, models.InvoiceFilter{
		AmountMin: 1500.0,
		AmountMax: 2500.0,
	})
	require.NoError(t, err)
	assert.Len(t, result.Invoices, 1)
	assert.Equal(t, models.InvoiceID("INV-002"), result.Invoices[0].ID)

	// Test pagination
	result, err = suite.storage.ListInvoices(suite.ctx, models.InvoiceFilter{
		Limit:  2,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.Len(t, result.Invoices, 2)
	assert.Equal(t, int64(3), result.TotalCount)
	assert.True(t, result.HasMore)
	assert.Equal(t, 2, result.NextOffset)

	// Test offset
	result, err = suite.storage.ListInvoices(suite.ctx, models.InvoiceFilter{
		Limit:  2,
		Offset: 2,
	})
	require.NoError(t, err)
	assert.Len(t, result.Invoices, 1)
	assert.Equal(t, int64(3), result.TotalCount)
	assert.False(t, result.HasMore)

	// Test invalid filter
	result, err = suite.storage.ListInvoices(suite.ctx, models.InvoiceFilter{
		Status: "invalid-status",
	})
	require.Error(t, err)
	assert.Nil(t, result)
	var filterErr storageTypes.ErrInvalidFilter
	assert.ErrorAs(t, err, &filterErr)

	// Test with corrupted invoice file
	corruptPath := filepath.Join(suite.tempDir, "invoices", "CORRUPT.json")
	err = os.WriteFile(corruptPath, []byte("invalid json"), 0o644)
	require.NoError(t, err)

	// Should skip corrupted files
	result, err = suite.storage.ListInvoices(suite.ctx, models.InvoiceFilter{})
	require.NoError(t, err)
	assert.Len(t, result.Invoices, 3) // Only valid invoices
}

func (suite *JSONStorageTestSuite) TestCountInvoices() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Test empty count
	count, err := suite.storage.CountInvoices(suite.ctx, models.InvoiceFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Create test invoices
	now := time.Now()
	for i := 1; i <= 5; i++ {
		invoice := &models.Invoice{
			ID:     models.InvoiceID(fmt.Sprintf("INV-%03d", i)),
			Number: fmt.Sprintf("INV-2024-%03d", i),
			Client: models.Client{
				ID:        "CLIENT-001",
				Name:      "Test Client",
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Version:   1,
			Date:      now,
			DueDate:   now.AddDate(0, 0, 30),
			Status:    models.StatusDraft,
			Total:     float64(i * 1000),
			CreatedAt: now,
			UpdatedAt: now,
		}

		if i%2 == 0 {
			invoice.Status = models.StatusPaid
		}

		err = suite.storage.CreateInvoice(suite.ctx, invoice)
		require.NoError(t, err)
	}

	// Test count all
	count, err = suite.storage.CountInvoices(suite.ctx, models.InvoiceFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)

	// Test count with filter
	count, err = suite.storage.CountInvoices(suite.ctx, models.InvoiceFilter{
		Status: models.StatusPaid,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Test count with amount filter
	count, err = suite.storage.CountInvoices(suite.ctx, models.InvoiceFilter{
		AmountMin: 3000,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count, err = suite.storage.CountInvoices(ctx, models.InvoiceFilter{})
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, int64(0), count)
}

func (suite *JSONStorageTestSuite) TestConcurrentAccess() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Test concurrent creates
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			invoice := &models.Invoice{
				ID:     models.InvoiceID(fmt.Sprintf("INV-%03d", id)),
				Number: fmt.Sprintf("INV-2024-%03d", id),
				Client: models.Client{
					ID:        "CLIENT-001",
					Name:      "Test Client",
					Email:     "test@example.com",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Version:   1,
				Date:      time.Now(),
				DueDate:   time.Now().AddDate(0, 0, 30),
				Status:    models.StatusDraft,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := suite.storage.CreateInvoice(suite.ctx, invoice); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent create error: %v", err)
	}

	// Verify all invoices were created
	result, err := suite.storage.ListInvoices(suite.ctx, models.InvoiceFilter{})
	require.NoError(t, err)
	assert.Len(t, result.Invoices, 10)

	// Test concurrent reads
	wg = sync.WaitGroup{}
	errors = make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			invoiceID := models.InvoiceID(fmt.Sprintf("INV-%03d", id))
			if _, err := suite.storage.GetInvoice(suite.ctx, invoiceID); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent read error: %v", err)
	}

	// Test concurrent updates with optimistic locking
	invoice, err := suite.storage.GetInvoice(suite.ctx, "INV-001")
	require.NoError(t, err)

	wg = sync.WaitGroup{}
	successCount := 0
	var successMu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(version int) {
			defer wg.Done()

			// Each goroutine tries to update with the same version
			updateInvoice := *invoice
			updateInvoice.Description = fmt.Sprintf("Update %d", version)

			if err := suite.storage.UpdateInvoice(suite.ctx, &updateInvoice); err == nil {
				successMu.Lock()
				successCount++
				successMu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Only one update should succeed due to optimistic locking
	assert.Equal(t, 1, successCount)
}

func (suite *JSONStorageTestSuite) TestAtomicWrites() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Create a large invoice to test atomic writes
	workItems := make([]models.WorkItem, 100)
	for i := 0; i < 100; i++ {
		workItems[i] = models.WorkItem{
			ID:          fmt.Sprintf("ITEM-%03d", i),
			Date:        time.Now(),
			Hours:       8.0,
			Rate:        100.0,
			Description: fmt.Sprintf("Work item %d with a very long description to increase file size", i),
			Total:       800.0,
			CreatedAt:   time.Now(),
		}
	}

	invoice := &models.Invoice{
		ID:     "INV-LARGE",
		Number: "INV-2024-LARGE",
		Client: models.Client{
			ID:        "CLIENT-001",
			Name:      "Test Client",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Version:     1,
		Date:        time.Now(),
		DueDate:     time.Now().AddDate(0, 0, 30),
		Status:      models.StatusDraft,
		WorkItems:   workItems,
		Description: strings.Repeat("Long description ", 1000),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save invoice
	err = suite.storage.CreateInvoice(suite.ctx, invoice)
	require.NoError(t, err)

	// Verify temp file doesn't exist
	invoicePath := filepath.Join(suite.tempDir, "invoices", "INV-LARGE.json")
	tempPath := invoicePath + ".tmp"
	_, err = os.Stat(tempPath)
	assert.True(t, os.IsNotExist(err), "Temp file should not exist after successful write")

	// Verify final file exists and is valid
	var savedInvoice models.Invoice
	data, err := os.ReadFile(invoicePath)
	require.NoError(t, err)
	err = json.Unmarshal(data, &savedInvoice)
	require.NoError(t, err)
	assert.Equal(t, invoice.ID, savedInvoice.ID)
	assert.Len(t, savedInvoice.WorkItems, 100)
}

func (suite *JSONStorageTestSuite) TestWriteJSONFileError() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Make invoices directory read-only
	err = os.Chmod(suite.storage.invoicesDir, 0o444)
	require.NoError(t, err)
	defer os.Chmod(suite.storage.invoicesDir, 0o755) // Restore permissions

	// Try to create invoice
	invoice := &models.Invoice{
		ID:     "INV-001",
		Number: "INV-2024-001",
		Client: models.Client{
			ID:        "CLIENT-001",
			Name:      "Test Client",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Version:   1,
		Date:      time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
		Status:    models.StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = suite.storage.CreateInvoice(suite.ctx, invoice)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to")
}

func (suite *JSONStorageTestSuite) TestReadJSONFileError() {
	t := suite.T()

	// Initialize storage
	err := suite.storage.Initialize(suite.ctx)
	require.NoError(t, err)

	// Create invalid JSON file
	invalidPath := filepath.Join(suite.storage.invoicesDir, "INVALID.json")
	err = os.WriteFile(invalidPath, []byte("{invalid json}"), 0o644)
	require.NoError(t, err)

	// Try to read it
	var data map[string]interface{}
	err = suite.storage.readJSONFile(suite.ctx, invalidPath, &data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode JSON")
}

func (suite *JSONStorageTestSuite) TestMatchesFilter() {
	t := suite.T()

	now := time.Now()
	invoice := &models.Invoice{
		ID:     "INV-001",
		Number: "INV-2024-001",
		Client: models.Client{
			ID: "CLIENT-001",
		},
		Date:    now,
		DueDate: now.AddDate(0, 0, 30),
		Status:  models.StatusSent,
		Total:   1500.0,
	}

	tests := []struct {
		name     string
		filter   models.InvoiceFilter
		expected bool
	}{
		{
			name:     "EmptyFilter",
			filter:   models.InvoiceFilter{},
			expected: true,
		},
		{
			name: "MatchingStatus",
			filter: models.InvoiceFilter{
				Status: models.StatusSent,
			},
			expected: true,
		},
		{
			name: "NonMatchingStatus",
			filter: models.InvoiceFilter{
				Status: models.StatusPaid,
			},
			expected: false,
		},
		{
			name: "MatchingClientID",
			filter: models.InvoiceFilter{
				ClientID: "CLIENT-001",
			},
			expected: true,
		},
		{
			name: "NonMatchingClientID",
			filter: models.InvoiceFilter{
				ClientID: "CLIENT-002",
			},
			expected: false,
		},
		{
			name: "DateInRange",
			filter: models.InvoiceFilter{
				DateFrom: now.AddDate(0, 0, -1),
				DateTo:   now.AddDate(0, 0, 1),
			},
			expected: true,
		},
		{
			name: "DateBeforeRange",
			filter: models.InvoiceFilter{
				DateFrom: now.AddDate(0, 0, 1),
			},
			expected: false,
		},
		{
			name: "DateAfterRange",
			filter: models.InvoiceFilter{
				DateTo: now.AddDate(0, 0, -1),
			},
			expected: false,
		},
		{
			name: "DueDateInRange",
			filter: models.InvoiceFilter{
				DueDateFrom: now.AddDate(0, 0, 29),
				DueDateTo:   now.AddDate(0, 0, 31),
			},
			expected: true,
		},
		{
			name: "AmountInRange",
			filter: models.InvoiceFilter{
				AmountMin: 1000.0,
				AmountMax: 2000.0,
			},
			expected: true,
		},
		{
			name: "AmountBelowRange",
			filter: models.InvoiceFilter{
				AmountMin: 2000.0,
			},
			expected: false,
		},
		{
			name: "AmountAboveRange",
			filter: models.InvoiceFilter{
				AmountMax: 1000.0,
			},
			expected: false,
		},
		{
			name: "MultipleMatchingFilters",
			filter: models.InvoiceFilter{
				Status:    models.StatusSent,
				ClientID:  "CLIENT-001",
				AmountMin: 1000.0,
				AmountMax: 2000.0,
			},
			expected: true,
		},
		{
			name: "OneNonMatchingFilter",
			filter: models.InvoiceFilter{
				Status:    models.StatusPaid, // Non-matching
				ClientID:  "CLIENT-001",
				AmountMin: 1000.0,
				AmountMax: 2000.0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.storage.matchesFilter(invoice, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}
