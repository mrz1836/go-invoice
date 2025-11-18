package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mrz/go-invoice/internal/models"
	"github.com/mrz/go-invoice/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Static test errors for err113 compliance
var (
	ErrTestIDGeneration       = errors.New("id generation failed")
	ErrTestStorage            = errors.New("storage error")
	ErrTestDatabase           = errors.New("database error")
	ErrTestDatabaseConnection = errors.New("database connection failed")
	ErrTestUpdateFailed       = errors.New("update failed")
)

// Test constants for goconst compliance
const (
	testClientID = "CLIENT-001"
)

// ClientServiceTestSuite tests for the ClientService
type ClientServiceTestSuite struct {
	suite.Suite

	ctx            context.Context //nolint:containedctx // Test suite context is acceptable
	cancelFunc     context.CancelFunc
	service        *ClientService
	clientStorage  *MockClientStorage
	invoiceStorage *MockInvoiceStorage
	logger         *MockLogger
	idGen          *MockIDGenerator
}

func (suite *ClientServiceTestSuite) SetupTest() {
	suite.ctx, suite.cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)

	suite.clientStorage = new(MockClientStorage)
	suite.invoiceStorage = new(MockInvoiceStorage)
	suite.logger = new(MockLogger)
	suite.idGen = new(MockIDGenerator)

	suite.service = NewClientService(
		suite.clientStorage,
		suite.invoiceStorage,
		suite.logger,
		suite.idGen,
	)
}

func (suite *ClientServiceTestSuite) TearDownTest() {
	suite.cancelFunc()
	suite.clientStorage.AssertExpectations(suite.T())
	suite.invoiceStorage.AssertExpectations(suite.T())
	suite.idGen.AssertExpectations(suite.T())
}

func TestClientServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ClientServiceTestSuite))
}

func (suite *ClientServiceTestSuite) TestCreateClient() {
	t := suite.T()

	request := models.CreateClientRequest{
		Name:    "Test Client",
		Email:   "test@example.com",
		Phone:   "+1234567890",
		Address: "123 Test St",
		TaxID:   "TAX-123",
	}

	// Success case
	suite.Run("Success", func() {
		suite.clientStorage.On("FindClientByEmail", suite.ctx, "test@example.com").Return(nil, storage.NewNotFoundError("client", "email:test@example.com")).Once()
		suite.idGen.On("GenerateClientID", suite.ctx).Return(models.ClientID(testClientID), nil).Once()
		suite.clientStorage.On("CreateClient", suite.ctx, mock.AnythingOfType("*models.Client")).Return(nil).Once()

		client, err := suite.service.CreateClient(suite.ctx, request)

		require.NoError(t, err)
		require.NotNil(t, client)
		assert.Equal(t, models.ClientID(testClientID), client.ID)
		assert.Equal(t, "Test Client", client.Name)
		assert.Equal(t, "test@example.com", client.Email)
		assert.Equal(t, "+1234567890", client.Phone)
		assert.Equal(t, "123 Test St", client.Address)
		assert.Equal(t, "TAX-123", client.TaxID)
		assert.True(t, client.Active)
	})

	// Duplicate email
	suite.Run("DuplicateEmail", func() {
		existingClient := &models.Client{
			ID:    "CLIENT-999",
			Email: "test@example.com",
		}

		suite.clientStorage.On("FindClientByEmail", suite.ctx, "test@example.com").Return(existingClient, nil).Once()

		client, err := suite.service.CreateClient(suite.ctx, request)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "client with email already exists: test@example.com")
	})

	// Invalid request
	suite.Run("InvalidRequest", func() {
		invalidRequest := models.CreateClientRequest{
			Name: "", // Missing required field
		}

		client, err := suite.service.CreateClient(suite.ctx, invalidRequest)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "invalid create client request")
	})

	// Context cancellation
	suite.Run("ContextCancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		client, err := suite.service.CreateClient(ctx, request)

		assert.Equal(t, context.Canceled, err)
		assert.Nil(t, client)
	})

	// ID generation failure
	suite.Run("IDGenerationFailure", func() {
		suite.clientStorage.On("FindClientByEmail", suite.ctx, "test@example.com").Return(nil, storage.NewNotFoundError("client", "email:test@example.com")).Once()
		suite.idGen.On("GenerateClientID", suite.ctx).Return(models.ClientID(""), ErrTestIDGeneration).Once()

		client, err := suite.service.CreateClient(suite.ctx, request)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to generate client ID")
	})

	// Request validation failure
	suite.Run("RequestValidationFailure", func() {
		invalidRequest := models.CreateClientRequest{
			Name:  "Test Client",
			Email: "invalid-email", // Invalid email format
		}

		client, err := suite.service.CreateClient(suite.ctx, invalidRequest)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "invalid create client request")
	})

	// Storage error
	suite.Run("StorageError", func() {
		suite.clientStorage.On("FindClientByEmail", suite.ctx, "test@example.com").Return(nil, storage.NewNotFoundError("client", "email:test@example.com")).Once()
		suite.idGen.On("GenerateClientID", suite.ctx).Return(models.ClientID(testClientID), nil).Once()
		suite.clientStorage.On("CreateClient", suite.ctx, mock.AnythingOfType("*models.Client")).Return(ErrTestStorage).Once()

		client, err := suite.service.CreateClient(suite.ctx, request)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to store client")
	})

	// Email lookup error during validation
	suite.Run("EmailLookupError", func() {
		suite.clientStorage.On("FindClientByEmail", suite.ctx, "test@example.com").Return(nil, ErrTestDatabase).Once()

		client, err := suite.service.CreateClient(suite.ctx, request)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "database error")
	})
}

func (suite *ClientServiceTestSuite) TestGetClient() {
	t := suite.T()

	testClient := &models.Client{
		ID:        testClientID,
		Name:      "Test Client",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Success case
	suite.Run("Success", func() {
		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID(testClientID)).Return(testClient, nil).Once()

		client, err := suite.service.GetClient(suite.ctx, testClientID)

		require.NoError(t, err)
		require.NotNil(t, client)
		assert.Equal(t, testClient.ID, client.ID)
		assert.Equal(t, testClient.Name, client.Name)
	})

	// Not found
	suite.Run("NotFound", func() {
		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID("CLIENT-999")).Return(nil, storage.NewNotFoundError("client", "CLIENT-999")).Once()

		client, err := suite.service.GetClient(suite.ctx, "CLIENT-999")

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "client with ID 'CLIENT-999' not found")
	})

	// Empty client ID
	suite.Run("EmptyClientID", func() {
		client, err := suite.service.GetClient(suite.ctx, "")

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "client ID cannot be empty")
	})

	// Whitespace-only client ID
	suite.Run("WhitespaceClientID", func() {
		client, err := suite.service.GetClient(suite.ctx, "   ")

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "client ID cannot be empty")
	})

	// Context cancellation
	suite.Run("ContextCancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		client, err := suite.service.GetClient(ctx, testClientID)

		assert.Equal(t, context.Canceled, err)
		assert.Nil(t, client)
	})

	// Storage error
	suite.Run("StorageError", func() {
		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID(testClientID)).Return(nil, ErrTestDatabaseConnection).Once()

		client, err := suite.service.GetClient(suite.ctx, testClientID)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to retrieve client")
	})
}

func (suite *ClientServiceTestSuite) TestUpdateClient() {
	t := suite.T()

	// Success case - update all fields
	suite.Run("Success", func() {
		existingClient := &models.Client{
			ID:        testClientID,
			Name:      "Old Name",
			Email:     "old@example.com",
			Phone:     "+1111111111",
			Address:   "Old Address",
			TaxID:     "OLD-TAX",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Clone and update
		updatedClient := *existingClient
		updatedClient.Name = "New Name"
		updatedClient.Email = "new@example.com"
		updatedClient.Phone = "+2222222222"
		updatedClient.Address = "New Address"
		updatedClient.TaxID = "NEW-TAX"

		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID(testClientID)).Return(existingClient, nil).Once()
		suite.clientStorage.On("FindClientByEmail", suite.ctx, "new@example.com").Return(nil, storage.NewNotFoundError("client", "new@example.com")).Once()
		suite.clientStorage.On("UpdateClient", suite.ctx, &updatedClient).Return(nil).Once()

		client, err := suite.service.UpdateClient(suite.ctx, &updatedClient)

		require.NoError(t, err)
		require.NotNil(t, client)
		assert.Equal(t, "New Name", client.Name)
		assert.Equal(t, "new@example.com", client.Email)
	})

	// Storage error
	suite.Run("StorageError", func() {
		existingClient := &models.Client{
			ID:    testClientID,
			Name:  "Test Client",
			Email: "old@example.com",
		}
		client := &models.Client{
			ID:        testClientID,
			Name:      "Test Client",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID(testClientID)).Return(existingClient, nil).Once()
		suite.clientStorage.On("FindClientByEmail", suite.ctx, "test@example.com").Return(nil, storage.NewNotFoundError("client", "test@example.com")).Once()
		suite.clientStorage.On("UpdateClient", suite.ctx, client).Return(errors.New("storage error")).Once() //nolint:err113 // Test mock error

		updatedClient, err := suite.service.UpdateClient(suite.ctx, client)

		require.Error(t, err)
		assert.Nil(t, updatedClient)
		assert.Contains(t, err.Error(), "failed to update client")
	})

	// Nil client
	suite.Run("NilClient", func() {
		client, err := suite.service.UpdateClient(suite.ctx, nil)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "client cannot be nil")
	})
}

func (suite *ClientServiceTestSuite) TestDeleteClient() {
	t := suite.T()

	// Success - no active invoices
	suite.Run("Success", func() {
		// Mock ListInvoices for each status check - three calls for each status
		emptyResult := &storage.InvoiceListResult{Invoices: []*models.Invoice{}, TotalCount: 0}
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusDraft && filter.Limit == 1
		})).Return(emptyResult, nil).Once()
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusSent && filter.Limit == 1
		})).Return(emptyResult, nil).Once()
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusOverdue && filter.Limit == 1
		})).Return(emptyResult, nil).Once()
		suite.clientStorage.On("DeleteClient", suite.ctx, models.ClientID(testClientID)).Return(nil).Once()

		err := suite.service.DeleteClient(suite.ctx, testClientID)

		require.NoError(t, err)
	})

	// Has active invoices
	suite.Run("HasActiveInvoices", func() {
		// Mock ListInvoices to return invoices for first status check (draft)
		activeResult := &storage.InvoiceListResult{
			Invoices: []*models.Invoice{{
				ID:     "INV-001",
				Status: models.StatusDraft,
				Client: models.Client{ID: testClientID},
			}},
			TotalCount: 5,
		}
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusDraft && filter.Limit == 1
		})).Return(activeResult, nil).Once()

		err := suite.service.DeleteClient(suite.ctx, testClientID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete client with active invoices")
	})
}

func (suite *ClientServiceTestSuite) TestListClients() {
	t := suite.T()

	clients := []*models.Client{
		{ID: testClientID, Name: "Client A", Active: true},
		{ID: "CLIENT-002", Name: "Client B", Active: true},
		{ID: "CLIENT-003", Name: "Client C", Active: false},
	}

	// Success - active only
	suite.Run("ActiveOnly", func() {
		result := &storage.ClientListResult{
			Clients:    clients[:2],
			TotalCount: 2,
			HasMore:    false,
		}

		suite.clientStorage.On("ListClients", suite.ctx, true, 10, 0).Return(result, nil).Once()

		listResult, err := suite.service.ListClients(suite.ctx, true, 10, 0)

		require.NoError(t, err)
		require.NotNil(t, listResult)
		assert.Len(t, listResult.Clients, 2)
		assert.Equal(t, int64(2), listResult.TotalCount)
	})
}

func (suite *ClientServiceTestSuite) TestFindClientByEmail() {
	t := suite.T()

	testClient := &models.Client{
		ID:     testClientID,
		Name:   "Test Client",
		Email:  "test@example.com",
		Active: true,
	}

	// Success case
	suite.Run("Success", func() {
		suite.clientStorage.On("FindClientByEmail", suite.ctx, "test@example.com").Return(testClient, nil).Once()

		client, err := suite.service.FindClientByEmail(suite.ctx, "test@example.com")

		require.NoError(t, err)
		require.NotNil(t, client)
		assert.Equal(t, testClient.ID, client.ID)
		assert.Equal(t, testClient.Email, client.Email)
	})

	// Not found
	suite.Run("NotFound", func() {
		suite.clientStorage.On("FindClientByEmail", suite.ctx, "notfound@example.com").Return(nil, storage.NewNotFoundError("client", "email:notfound@example.com")).Once()

		client, err := suite.service.FindClientByEmail(suite.ctx, "notfound@example.com")

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to find client by email")
	})
}

func (suite *ClientServiceTestSuite) TestActivateClient() {
	t := suite.T()

	// Success - activate inactive client
	suite.Run("Success", func() {
		inactiveClient := &models.Client{
			ID:        testClientID,
			Name:      "Test Client",
			Active:    false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID(testClientID)).Return(inactiveClient, nil).Once()
		suite.clientStorage.On("UpdateClient", suite.ctx, mock.AnythingOfType("*models.Client")).Return(nil).Once()

		client, err := suite.service.ActivateClient(suite.ctx, testClientID)

		require.NoError(t, err)
		require.NotNil(t, client)
		assert.True(t, client.Active)
	})

	// Context cancellation
	suite.Run("ContextCancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		client, err := suite.service.ActivateClient(ctx, testClientID)

		assert.Equal(t, context.Canceled, err)
		assert.Nil(t, client)
	})

	// Client not found
	suite.Run("ClientNotFound", func() {
		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID("CLIENT-999")).Return(nil, storage.NewNotFoundError("client", "CLIENT-999")).Once()

		client, err := suite.service.ActivateClient(suite.ctx, "CLIENT-999")

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to retrieve client")
	})

	// Storage error on get
	suite.Run("StorageErrorOnGet", func() {
		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID(testClientID)).Return(nil, ErrTestDatabase).Once()

		client, err := suite.service.ActivateClient(suite.ctx, testClientID)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to retrieve client")
	})

	// Storage error on update
	suite.Run("StorageErrorOnUpdate", func() {
		inactiveClient := &models.Client{
			ID:        testClientID,
			Name:      "Test Client",
			Active:    false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID(testClientID)).Return(inactiveClient, nil).Once()
		suite.clientStorage.On("UpdateClient", suite.ctx, mock.AnythingOfType("*models.Client")).Return(ErrTestUpdateFailed).Once()

		client, err := suite.service.ActivateClient(suite.ctx, testClientID)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to update client")
	})

	// Already active client
	suite.Run("AlreadyActive", func() {
		activeClient := &models.Client{
			ID:        testClientID,
			Name:      "Test Client",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID(testClientID)).Return(activeClient, nil).Once()

		client, err := suite.service.ActivateClient(suite.ctx, testClientID)

		require.NoError(t, err)
		require.NotNil(t, client)
		assert.True(t, client.Active) // Should remain active
	})
}

func (suite *ClientServiceTestSuite) TestDeactivateClient() {
	t := suite.T()

	// Success - deactivate active client
	suite.Run("Success", func() {
		activeClient := &models.Client{
			ID:        testClientID,
			Name:      "Test Client",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock ListInvoices for active invoice check (no active invoices) - three calls for each status
		emptyResult := &storage.InvoiceListResult{Invoices: []*models.Invoice{}, TotalCount: 0}
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusDraft && filter.Limit == 1
		})).Return(emptyResult, nil).Once()
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusSent && filter.Limit == 1
		})).Return(emptyResult, nil).Once()
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusOverdue && filter.Limit == 1
		})).Return(emptyResult, nil).Once()

		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID(testClientID)).Return(activeClient, nil).Once()
		suite.clientStorage.On("UpdateClient", suite.ctx, mock.AnythingOfType("*models.Client")).Return(nil).Once()

		client, err := suite.service.DeactivateClient(suite.ctx, testClientID)

		require.NoError(t, err)
		require.NotNil(t, client)
		assert.False(t, client.Active)
	})

	// Context cancellation
	suite.Run("ContextCancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		client, err := suite.service.DeactivateClient(ctx, testClientID)

		assert.Equal(t, context.Canceled, err)
		assert.Nil(t, client)
	})

	// Client not found
	suite.Run("ClientNotFound", func() {
		// First it checks for active invoices, then tries to get the client
		emptyResult := &storage.InvoiceListResult{Invoices: []*models.Invoice{}, TotalCount: 0}
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == "CLIENT-999" && filter.Status == models.StatusDraft && filter.Limit == 1
		})).Return(emptyResult, nil).Once()
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == "CLIENT-999" && filter.Status == models.StatusSent && filter.Limit == 1
		})).Return(emptyResult, nil).Once()
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == "CLIENT-999" && filter.Status == models.StatusOverdue && filter.Limit == 1
		})).Return(emptyResult, nil).Once()

		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID("CLIENT-999")).Return(nil, storage.NewNotFoundError("client", "CLIENT-999")).Once()

		client, err := suite.service.DeactivateClient(suite.ctx, "CLIENT-999")

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to retrieve client")
	})

	// Has active invoices - should fail
	suite.Run("HasActiveInvoices", func() {
		// Mock ListInvoices to return active invoices for first status check (draft)
		activeInvoices := []*models.Invoice{
			{ID: "INV-001", Status: models.StatusDraft},
		}
		invoiceResult := &storage.InvoiceListResult{Invoices: activeInvoices, TotalCount: 1}

		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusDraft && filter.Limit == 1
		})).Return(invoiceResult, nil).Once()

		client, err := suite.service.DeactivateClient(suite.ctx, testClientID)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "cannot deactivate client with active invoices")
	})

	// Storage error on active invoice check
	suite.Run("ActiveInvoiceCheckError", func() {
		// Error occurs on first status check (draft)
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusDraft && filter.Limit == 1
		})).Return(nil, ErrTestDatabase).Once()

		client, err := suite.service.DeactivateClient(suite.ctx, testClientID)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to check client invoices")
	})

	// Storage error on update
	suite.Run("StorageErrorOnUpdate", func() {
		activeClient := &models.Client{
			ID:        testClientID,
			Name:      "Test Client",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock ListInvoices for active invoice check (no active invoices) - three calls for each status
		emptyResult := &storage.InvoiceListResult{Invoices: []*models.Invoice{}, TotalCount: 0}
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusDraft && filter.Limit == 1
		})).Return(emptyResult, nil).Once()
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusSent && filter.Limit == 1
		})).Return(emptyResult, nil).Once()
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID && filter.Status == models.StatusOverdue && filter.Limit == 1
		})).Return(emptyResult, nil).Once()

		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID(testClientID)).Return(activeClient, nil).Once()
		suite.clientStorage.On("UpdateClient", suite.ctx, mock.AnythingOfType("*models.Client")).Return(ErrTestUpdateFailed).Once()

		client, err := suite.service.DeactivateClient(suite.ctx, testClientID)

		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to update client")
	})
}

func (suite *ClientServiceTestSuite) TestGetClientWithInvoices() {
	t := suite.T()

	client := &models.Client{
		ID:        testClientID,
		Name:      "Test Client",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	invoices := []*models.Invoice{
		{ID: "INV-001", Number: "INV-2024-001", Status: models.StatusDraft},
		{ID: "INV-002", Number: "INV-2024-002", Status: models.StatusSent},
		{ID: "INV-003", Number: "INV-2024-003", Status: models.StatusPaid},
	}

	// Success case
	suite.Run("Success", func() {
		result := &storage.InvoiceListResult{
			Invoices:   invoices,
			TotalCount: 3,
		}

		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID(testClientID)).Return(client, nil).Once()
		suite.invoiceStorage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.ClientID == testClientID
		})).Return(result, nil).Once()

		clientWithInvoices, err := suite.service.GetClientWithInvoices(suite.ctx, testClientID)

		require.NoError(t, err)
		require.NotNil(t, clientWithInvoices)
		assert.Equal(t, client.ID, clientWithInvoices.Client.ID)
		assert.Len(t, clientWithInvoices.Invoices, 3)
		assert.Equal(t, 3, clientWithInvoices.TotalInvoices)
	})
}

func (suite *ClientServiceTestSuite) TestGetClientStatistics() {
	t := suite.T()

	activeClients := []*models.Client{
		{ID: testClientID, Name: "Client A", Active: true},
		{ID: "CLIENT-002", Name: "Client B", Active: true},
	}

	inactiveClients := []*models.Client{
		{ID: "CLIENT-003", Name: "Client C", Active: false},
	}

	// Success case
	suite.Run("Success", func() {
		// Mock call to get all clients
		suite.clientStorage.On("ListClients", suite.ctx, false, -1, 0).Return(&storage.ClientListResult{
			Clients:    append(activeClients, inactiveClients...),
			TotalCount: 3,
		}, nil).Once()

		stats, err := suite.service.GetClientStatistics(suite.ctx)

		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, int64(3), stats.TotalClients)
		assert.Equal(t, int64(2), stats.ActiveClients)
		assert.Equal(t, int64(1), stats.InactiveClients)
	})
}
