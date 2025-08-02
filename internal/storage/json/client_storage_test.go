package json

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz/go-invoice/internal/models"
	storageTypes "github.com/mrz/go-invoice/internal/storage"
)

// ClientStorage wraps JSONStorage for client-specific operations
type ClientStorage struct {
	*JSONStorage
}

// ClientStorageTestSuite tests client storage operations
type ClientStorageTestSuite struct {
	suite.Suite
	ctx        context.Context
	cancelFunc context.CancelFunc
	storage    *ClientStorage
	tempDir    string
	logger     *MockLogger
}

func (suite *ClientStorageTestSuite) SetupTest() {
	suite.ctx, suite.cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)

	// Create temporary directory for tests
	tempDir, err := os.MkdirTemp("", "client-storage-test-*")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	// Create mock logger
	suite.logger = &MockLogger{}

	// Create storage instance
	jsonStorage := NewJSONStorage(suite.tempDir, suite.logger)
	suite.storage = &ClientStorage{JSONStorage: jsonStorage}

	// Initialize storage
	err = jsonStorage.Initialize(suite.ctx)
	suite.Require().NoError(err)
}

func (suite *ClientStorageTestSuite) TearDownTest() {
	suite.cancelFunc()

	// Clean up temporary directory
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

func TestClientStorageTestSuite(t *testing.T) {
	suite.Run(t, new(ClientStorageTestSuite))
}

func (suite *ClientStorageTestSuite) TestCreateClient() {
	t := suite.T()

	// Create test client
	client := &models.Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test successful creation
	err := suite.storage.CreateClient(suite.ctx, client)
	require.NoError(t, err)

	// Verify file was created
	clientPath := filepath.Join(suite.tempDir, "clients", "CLIENT-001.json")
	_, err = os.Stat(clientPath)
	require.NoError(t, err)

	// Verify file content
	var savedClient models.Client
	data, err := os.ReadFile(clientPath)
	require.NoError(t, err)
	err = json.Unmarshal(data, &savedClient)
	require.NoError(t, err)

	assert.Equal(t, client.ID, savedClient.ID)
	assert.Equal(t, client.Name, savedClient.Name)
	assert.Equal(t, client.Email, savedClient.Email)
	assert.True(t, savedClient.Active)

	// Test duplicate creation
	err = suite.storage.CreateClient(suite.ctx, client)
	require.Error(t, err)
	var conflictErr storageTypes.ConflictError
	assert.ErrorAs(t, err, &conflictErr)
	assert.Equal(t, "client", conflictErr.Resource)
	assert.Equal(t, "CLIENT-001", conflictErr.ID)

	// Test with nil client
	err = suite.storage.CreateClient(suite.ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client cannot be nil")

	// Test with invalid client
	invalidClient := &models.Client{
		ID: "", // Invalid - empty ID
	}
	err = suite.storage.CreateClient(suite.ctx, invalidClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid client")

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = suite.storage.CreateClient(ctx, client)
	assert.Equal(t, context.Canceled, err)
}

func (suite *ClientStorageTestSuite) TestGetClient() {
	t := suite.T()

	// Create test client
	client := &models.Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		Phone:     "+1234567890",
		Address:   "123 Test St",
		Active:    true,
		CreatedAt: time.Now().Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}

	// Save client
	err := suite.storage.CreateClient(suite.ctx, client)
	require.NoError(t, err)

	// Test successful retrieval
	retrieved, err := suite.storage.GetClient(suite.ctx, "CLIENT-001")
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, client.ID, retrieved.ID)
	assert.Equal(t, client.Name, retrieved.Name)
	assert.Equal(t, client.Email, retrieved.Email)
	assert.Equal(t, client.Phone, retrieved.Phone)
	assert.Equal(t, client.Address, retrieved.Address)
	assert.True(t, retrieved.Active)
	assert.WithinDuration(t, client.CreatedAt, retrieved.CreatedAt, time.Second)

	// Test non-existent client
	retrieved, err = suite.storage.GetClient(suite.ctx, "CLIENT-999")
	require.Error(t, err)
	assert.Nil(t, retrieved)
	var notFoundErr storageTypes.NotFoundError
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, "client", notFoundErr.Resource)
	assert.Equal(t, "CLIENT-999", notFoundErr.ID)

	// Test with empty ID
	retrieved, err = suite.storage.GetClient(suite.ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client ID cannot be empty")

	// Test with whitespace ID
	retrieved, err = suite.storage.GetClient(suite.ctx, "   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client ID cannot be empty")

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	retrieved, err = suite.storage.GetClient(ctx, "CLIENT-001")
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, retrieved)
}

func (suite *ClientStorageTestSuite) TestUpdateClient() {
	t := suite.T()

	// Create test client
	client := &models.Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save client
	err := suite.storage.CreateClient(suite.ctx, client)
	require.NoError(t, err)

	// Update client
	client.Name = "Updated Client"
	client.Email = "updated@example.com"
	client.Phone = "+9876543210"

	err = suite.storage.UpdateClient(suite.ctx, client)
	require.NoError(t, err)

	// Verify UpdatedAt was changed
	assert.True(t, client.UpdatedAt.After(time.Now().Add(-time.Second)))

	// Verify changes were saved
	retrieved, err := suite.storage.GetClient(suite.ctx, "CLIENT-001")
	require.NoError(t, err)
	assert.Equal(t, "Updated Client", retrieved.Name)
	assert.Equal(t, "updated@example.com", retrieved.Email)
	assert.Equal(t, "+9876543210", retrieved.Phone)
	assert.WithinDuration(t, time.Now(), retrieved.UpdatedAt, time.Second)

	// Test concurrent update (no optimistic locking in client storage)
	// Multiple updates should succeed
	concurrentClient := &models.Client{
		ID:        "CLIENT-001",
		Name:      "Concurrent Update",
		Email:     "concurrent@example.com",
		Active:    true,
		CreatedAt: client.CreatedAt,
		UpdatedAt: time.Now(),
	}

	err = suite.storage.UpdateClient(suite.ctx, concurrentClient)
	require.NoError(t, err)

	// Verify last update wins
	retrieved, err = suite.storage.GetClient(suite.ctx, "CLIENT-001")
	require.NoError(t, err)
	assert.Equal(t, "Concurrent Update", retrieved.Name)
	assert.Equal(t, "concurrent@example.com", retrieved.Email)

	// Test updating non-existent client
	nonExistent := &models.Client{
		ID:        "CLIENT-999",
		Name:      "Non-existent",
		Email:     "nonexistent@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = suite.storage.UpdateClient(suite.ctx, nonExistent)
	require.Error(t, err)
	var notFoundErr storageTypes.NotFoundError
	assert.ErrorAs(t, err, &notFoundErr)

	// Test with nil client
	err = suite.storage.UpdateClient(suite.ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client cannot be nil")

	// Test with invalid client
	invalidClient := &models.Client{
		ID: "CLIENT-001",
		// Missing required fields
	}
	err = suite.storage.UpdateClient(suite.ctx, invalidClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid client")
}

func (suite *ClientStorageTestSuite) TestDeleteClient() {
	t := suite.T()

	// Create test client
	client := &models.Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save client
	err := suite.storage.CreateClient(suite.ctx, client)
	require.NoError(t, err)

	// Verify file exists
	clientPath := filepath.Join(suite.tempDir, "clients", "CLIENT-001.json")
	_, err = os.Stat(clientPath)
	require.NoError(t, err)

	// Delete client (soft delete)
	err = suite.storage.DeleteClient(suite.ctx, "CLIENT-001")
	require.NoError(t, err)

	// Verify file still exists (soft delete)
	_, err = os.Stat(clientPath)
	require.NoError(t, err)

	// Verify client is marked as inactive
	retrieved, err := suite.storage.GetClient(suite.ctx, "CLIENT-001")
	require.NoError(t, err)
	assert.False(t, retrieved.Active)

	// Try to delete non-existent client
	err = suite.storage.DeleteClient(suite.ctx, "CLIENT-999")
	require.Error(t, err)
	var notFoundErr storageTypes.NotFoundError
	assert.ErrorAs(t, err, &notFoundErr)

	// Test with empty ID
	err = suite.storage.DeleteClient(suite.ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client ID cannot be empty")

	// Test with whitespace ID
	err = suite.storage.DeleteClient(suite.ctx, "   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client ID cannot be empty")

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = suite.storage.DeleteClient(ctx, "CLIENT-001")
	assert.Equal(t, context.Canceled, err)
}

func (suite *ClientStorageTestSuite) TestListClients() {
	t := suite.T()

	// Test empty list
	result, err := suite.storage.ListClients(suite.ctx, true, 10, 0)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Clients)
	assert.Equal(t, int64(0), result.TotalCount)
	assert.False(t, result.HasMore)

	// Create test clients
	now := time.Now()
	clients := []*models.Client{
		{
			ID:        "CLIENT-001",
			Name:      "Alpha Client",
			Email:     "alpha@example.com",
			Active:    true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "CLIENT-002",
			Name:      "Beta Client",
			Email:     "beta@example.com",
			Active:    true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "CLIENT-003",
			Name:      "Gamma Client",
			Email:     "gamma@example.com",
			Active:    false, // Inactive
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "CLIENT-004",
			Name:      "Delta Client",
			Email:     "delta@example.com",
			Active:    true,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// Save clients
	for _, client := range clients {
		err = suite.storage.CreateClient(suite.ctx, client)
		require.NoError(t, err)
	}

	// Test list all active clients
	result, err = suite.storage.ListClients(suite.ctx, true, 10, 0)
	require.NoError(t, err)
	assert.Len(t, result.Clients, 3) // Only active clients
	assert.Equal(t, int64(3), result.TotalCount)
	assert.False(t, result.HasMore)

	// Verify sorting (by name ascending)
	assert.Equal(t, "Alpha Client", result.Clients[0].Name)
	assert.Equal(t, "Beta Client", result.Clients[1].Name)
	assert.Equal(t, "Delta Client", result.Clients[2].Name)

	// Test list all clients (including inactive)
	result, err = suite.storage.ListClients(suite.ctx, false, 10, 0)
	require.NoError(t, err)
	assert.Len(t, result.Clients, 4) // All clients
	assert.Equal(t, int64(4), result.TotalCount)

	// Test pagination
	result, err = suite.storage.ListClients(suite.ctx, true, 2, 0)
	require.NoError(t, err)
	assert.Len(t, result.Clients, 2)
	assert.Equal(t, int64(3), result.TotalCount)
	assert.True(t, result.HasMore)
	assert.Equal(t, 2, result.NextOffset)

	// Test offset
	result, err = suite.storage.ListClients(suite.ctx, true, 2, 2)
	require.NoError(t, err)
	assert.Len(t, result.Clients, 1) // Only one remaining
	assert.Equal(t, int64(3), result.TotalCount)
	assert.False(t, result.HasMore)

	// Test negative limit
	result, err = suite.storage.ListClients(suite.ctx, true, -1, 0)
	require.NoError(t, err)
	assert.Len(t, result.Clients, 3) // All active clients

	// Test large offset
	result, err = suite.storage.ListClients(suite.ctx, true, 10, 100)
	require.NoError(t, err)
	assert.Empty(t, result.Clients)
	assert.Equal(t, int64(3), result.TotalCount)
	assert.False(t, result.HasMore)

	// Test with corrupted client file
	corruptPath := filepath.Join(suite.tempDir, "clients", "CORRUPT.json")
	err = os.WriteFile(corruptPath, []byte("invalid json"), 0o644)
	require.NoError(t, err)

	// Should skip corrupted files
	result, err = suite.storage.ListClients(suite.ctx, true, 10, 0)
	require.NoError(t, err)
	assert.Len(t, result.Clients, 3) // Only valid active clients

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err = suite.storage.ListClients(ctx, true, 10, 0)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, result)
}

func (suite *ClientStorageTestSuite) TestFindClientByEmail() {
	t := suite.T()

	// Test with non-existent email
	client, err := suite.storage.FindClientByEmail(suite.ctx, "nonexistent@example.com")
	require.Error(t, err)
	assert.Nil(t, client)
	var notFoundErr storageTypes.NotFoundError
	assert.ErrorAs(t, err, &notFoundErr)

	// Create test clients
	clients := []*models.Client{
		{
			ID:        "CLIENT-001",
			Name:      "Test Client 1",
			Email:     "test1@example.com",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "CLIENT-002",
			Name:      "Test Client 2",
			Email:     "test2@example.com",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "CLIENT-003",
			Name:      "Test Client 3",
			Email:     "test3@example.com",
			Active:    false, // Inactive
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Save clients
	for _, c := range clients {
		err = suite.storage.CreateClient(suite.ctx, c)
		require.NoError(t, err)
	}

	// Test finding active client
	client, err = suite.storage.FindClientByEmail(suite.ctx, "test1@example.com")
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "CLIENT-001", string(client.ID))
	assert.Equal(t, "Test Client 1", client.Name)
	assert.Equal(t, "test1@example.com", client.Email)

	// Test finding another active client
	client, err = suite.storage.FindClientByEmail(suite.ctx, "test2@example.com")
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "CLIENT-002", string(client.ID))

	// Test finding inactive client (should still find it)
	client, err = suite.storage.FindClientByEmail(suite.ctx, "test3@example.com")
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "CLIENT-003", string(client.ID))
	assert.False(t, client.Active)

	// Test with empty email
	client, err = suite.storage.FindClientByEmail(suite.ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email cannot be empty")

	// Test with whitespace email
	client, err = suite.storage.FindClientByEmail(suite.ctx, "   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email cannot be empty")

	// Test case sensitivity
	client, err = suite.storage.FindClientByEmail(suite.ctx, "TEST1@EXAMPLE.COM")
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "test1@example.com", client.Email)

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err = suite.storage.FindClientByEmail(ctx, "test1@example.com")
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, client)
}

func (suite *ClientStorageTestSuite) TestExistsClient() {
	t := suite.T()

	// Test non-existent client
	exists, err := suite.storage.ExistsClient(suite.ctx, "CLIENT-001")
	require.NoError(t, err)
	assert.False(t, exists)

	// Create test client
	client := &models.Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save client
	err = suite.storage.CreateClient(suite.ctx, client)
	require.NoError(t, err)

	// Test existing client
	exists, err = suite.storage.ExistsClient(suite.ctx, "CLIENT-001")
	require.NoError(t, err)
	assert.True(t, exists)

	// Soft delete client
	err = suite.storage.DeleteClient(suite.ctx, "CLIENT-001")
	require.NoError(t, err)

	// Test after soft deletion (should still exist)
	exists, err = suite.storage.ExistsClient(suite.ctx, "CLIENT-001")
	require.NoError(t, err)
	assert.True(t, exists)

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	exists, err = suite.storage.ExistsClient(ctx, "CLIENT-001")
	assert.Equal(t, context.Canceled, err)
	assert.False(t, exists)
}

func (suite *ClientStorageTestSuite) TestConcurrentClientAccess() {
	t := suite.T()

	// Test concurrent creates
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			client := &models.Client{
				ID:        models.ClientID(fmt.Sprintf("CLIENT-%03d", id)),
				Name:      fmt.Sprintf("Client %d", id),
				Email:     fmt.Sprintf("client%d@example.com", id),
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := suite.storage.CreateClient(suite.ctx, client); err != nil {
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

	// Verify all clients were created
	result, err := suite.storage.ListClients(suite.ctx, true, 100, 0)
	require.NoError(t, err)
	assert.Len(t, result.Clients, 10)

	// Test concurrent reads
	wg = sync.WaitGroup{}
	errors = make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			clientID := models.ClientID(fmt.Sprintf("CLIENT-%03d", id))
			if _, err := suite.storage.GetClient(suite.ctx, clientID); err != nil {
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

	// Test concurrent updates (no optimistic locking in client storage)
	client, err := suite.storage.GetClient(suite.ctx, "CLIENT-001")
	require.NoError(t, err)

	wg = sync.WaitGroup{}
	successCount := 0
	var successMu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(version int) {
			defer wg.Done()

			// Each goroutine tries to update
			updateClient := *client
			updateClient.Name = fmt.Sprintf("Updated Client %d", version)

			if err := suite.storage.UpdateClient(suite.ctx, &updateClient); err == nil {
				successMu.Lock()
				successCount++
				successMu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// All updates should succeed (no optimistic locking)
	assert.Equal(t, 5, successCount)

	// Test concurrent email lookups
	wg = sync.WaitGroup{}
	errors = make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			email := fmt.Sprintf("client%d@example.com", id)
			if _, err := suite.storage.FindClientByEmail(suite.ctx, email); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent email lookup error: %v", err)
	}
}

func (suite *ClientStorageTestSuite) TestClientFileValidation() {
	t := suite.T()

	// Create a client with all fields populated
	client := &models.Client{
		ID:        "CLIENT-FULL",
		Name:      "Full Test Client",
		Email:     "full@example.com",
		Phone:     "+1234567890",
		Address:   "123 Full St, Test City, TC 12345",
		TaxID:     "TAX-123456",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save client
	err := suite.storage.CreateClient(suite.ctx, client)
	require.NoError(t, err)

	// Read back and verify all fields
	retrieved, err := suite.storage.GetClient(suite.ctx, "CLIENT-FULL")
	require.NoError(t, err)

	assert.Equal(t, client.ID, retrieved.ID)
	assert.Equal(t, client.Name, retrieved.Name)
	assert.Equal(t, client.Email, retrieved.Email)
	assert.Equal(t, client.Phone, retrieved.Phone)
	assert.Equal(t, client.Address, retrieved.Address)
	assert.Equal(t, client.TaxID, retrieved.TaxID)
	assert.Equal(t, client.Active, retrieved.Active)
	assert.WithinDuration(t, client.CreatedAt, retrieved.CreatedAt, time.Second)

	// Test that the JSON file is properly formatted
	clientPath := filepath.Join(suite.tempDir, "clients", "CLIENT-FULL.json")
	data, err := os.ReadFile(clientPath)
	require.NoError(t, err)

	// Verify JSON is indented
	assert.Contains(t, string(data), "\n  ")

	// Verify can unmarshal back
	var unmarshaledClient models.Client
	err = json.Unmarshal(data, &unmarshaledClient)
	require.NoError(t, err)
	assert.Equal(t, client.ID, unmarshaledClient.ID)
}

func (suite *ClientStorageTestSuite) TestClientStorageWithInvalidJSON() {
	t := suite.T()

	// Create invalid JSON files in clients directory
	invalidFiles := map[string]string{
		"INVALID1.json": "{invalid json}",
		"INVALID2.json": `{"id": "CLIENT-X", "name": }`, // Missing value
		"INVALID3.json": `[1, 2, 3]`,                    // Wrong type
		"INVALID4.json": `null`,                         // Null
		"INVALID5.json": `""`,                           // String instead of object
	}

	for filename, content := range invalidFiles {
		path := filepath.Join(suite.tempDir, "clients", filename)
		err := os.WriteFile(path, []byte(content), 0o644)
		require.NoError(t, err)
	}

	// Create some valid clients
	validClients := []models.Client{
		{
			ID:        "CLIENT-VALID-1",
			Name:      "Valid Client 1",
			Email:     "valid1@example.com",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "CLIENT-VALID-2",
			Name:      "Valid Client 2",
			Email:     "valid2@example.com",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, client := range validClients {
		err := suite.storage.CreateClient(suite.ctx, &client)
		require.NoError(t, err)
	}

	// List should only return valid clients
	result, err := suite.storage.ListClients(suite.ctx, true, 100, 0)
	require.NoError(t, err)
	assert.Len(t, result.Clients, 2)

	// Verify only valid clients are returned
	for _, client := range result.Clients {
		assert.Contains(t, []string{"CLIENT-VALID-1", "CLIENT-VALID-2"}, string(client.ID))
	}

	// Try to get an invalid client directly
	// Some might parse but fail validation, others might fail to parse
	// We just verify they don't show up in the list above
}

func (suite *ClientStorageTestSuite) TestEmailIndexConsistency() {
	t := suite.T()

	// Create client
	client := &models.Client{
		ID:        "CLIENT-EMAIL-1",
		Name:      "Email Test Client",
		Email:     "original@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := suite.storage.CreateClient(suite.ctx, client)
	require.NoError(t, err)

	// Find by original email
	found, err := suite.storage.FindClientByEmail(suite.ctx, "original@example.com")
	require.NoError(t, err)
	assert.Equal(t, client.ID, found.ID)

	// Update email
	client.Email = "updated@example.com"
	err = suite.storage.UpdateClient(suite.ctx, client)
	require.NoError(t, err)

	// Original email should not find anything
	_, err = suite.storage.FindClientByEmail(suite.ctx, "original@example.com")
	require.Error(t, err)
	var notFoundErr storageTypes.NotFoundError
	assert.ErrorAs(t, err, &notFoundErr)

	// New email should find the client
	found, err = suite.storage.FindClientByEmail(suite.ctx, "updated@example.com")
	require.NoError(t, err)
	assert.Equal(t, client.ID, found.ID)

	// Create another client with the old email (should work)
	newClient := &models.Client{
		ID:        "CLIENT-EMAIL-2",
		Name:      "New Client",
		Email:     "original@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = suite.storage.CreateClient(suite.ctx, newClient)
	require.NoError(t, err)

	// Find by original email should now find the new client
	found, err = suite.storage.FindClientByEmail(suite.ctx, "original@example.com")
	require.NoError(t, err)
	assert.Equal(t, newClient.ID, found.ID)
}
