package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz/go-invoice/internal/models"
	"github.com/mrz/go-invoice/internal/storage"
)

// InvoiceServiceTestSuite tests for the InvoiceService
type InvoiceServiceTestSuite struct {
	suite.Suite
	ctx           context.Context
	cancelFunc    context.CancelFunc
	service       *InvoiceService
	storage       *MockInvoiceStorage
	clientStorage *MockClientStorage
	logger        *MockLogger
	idGen         *MockIDGenerator
}

func (suite *InvoiceServiceTestSuite) SetupTest() {
	suite.ctx, suite.cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)

	suite.storage = new(MockInvoiceStorage)
	suite.clientStorage = new(MockClientStorage)
	suite.logger = new(MockLogger)
	suite.idGen = new(MockIDGenerator)

	suite.service = NewInvoiceService(
		suite.storage,
		suite.clientStorage,
		suite.logger,
		suite.idGen,
	)
}

func (suite *InvoiceServiceTestSuite) TearDownTest() {
	suite.cancelFunc()
	suite.storage.AssertExpectations(suite.T())
	suite.clientStorage.AssertExpectations(suite.T())
	suite.idGen.AssertExpectations(suite.T())
}

func TestInvoiceServiceTestSuite(t *testing.T) {
	suite.Run(t, new(InvoiceServiceTestSuite))
}

func (suite *InvoiceServiceTestSuite) TestCreateInvoice() {
	t := suite.T()

	// Test data
	client := &models.Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	request := models.CreateInvoiceRequest{
		Number:      "INV-2024-001",
		ClientID:    "CLIENT-001",
		Date:        time.Now(),
		DueDate:     time.Now().AddDate(0, 0, 30),
		Description: "Test Invoice",
		WorkItems: []models.WorkItem{
			{
				ID:          "WORK-001",
				Date:        time.Now(),
				Hours:       8.0,
				Rate:        100.0,
				Description: "Development work",
				Total:       800.0,
				CreatedAt:   time.Now(),
			},
		},
	}

	// Success case
	suite.Run("Success", func() {
		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID("CLIENT-001")).Return(client, nil).Once()
		suite.idGen.On("GenerateInvoiceID", suite.ctx).Return(models.InvoiceID("INV-001"), nil).Once()
		suite.storage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.Status == "" && filter.ClientID == "" && filter.Limit == 0
		})).Return(&storage.InvoiceListResult{Invoices: []*models.Invoice{}}, nil).Once()
		// Mock GenerateWorkItemID for each work item in request
		suite.idGen.On("GenerateWorkItemID", suite.ctx).Return("WORK-001", nil).Once()
		suite.storage.On("CreateInvoice", suite.ctx, mock.AnythingOfType("*models.Invoice")).Return(nil).Once()

		invoice, err := suite.service.CreateInvoice(suite.ctx, request)

		require.NoError(t, err)
		require.NotNil(t, invoice)
		assert.Equal(t, models.InvoiceID("INV-001"), invoice.ID)
		assert.Equal(t, "INV-2024-001", invoice.Number)
		assert.Equal(t, client.ID, invoice.Client.ID)
		assert.Equal(t, "Test Invoice", invoice.Description)
		assert.Len(t, invoice.WorkItems, 1)
		assert.Equal(t, models.StatusDraft, invoice.Status)
	})

	// Client not found
	suite.Run("ClientNotFound", func() {
		suite.clientStorage.On("GetClient", suite.ctx, models.ClientID("CLIENT-001")).Return(nil, storage.NewNotFoundError("client", "CLIENT-001")).Once()

		invoice, err := suite.service.CreateInvoice(suite.ctx, request)

		require.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "client not found: CLIENT-001")
	})
}

func (suite *InvoiceServiceTestSuite) TestGetInvoice() {
	t := suite.T()

	testInvoice := &models.Invoice{
		ID:        "INV-001",
		Number:    "INV-2024-001",
		Status:    models.StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	// Success case
	suite.Run("Success", func() {
		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(testInvoice, nil).Once()

		invoice, err := suite.service.GetInvoice(suite.ctx, "INV-001")

		require.NoError(t, err)
		require.NotNil(t, invoice)
		assert.Equal(t, testInvoice.ID, invoice.ID)
		assert.Equal(t, testInvoice.Number, invoice.Number)
	})

	// Not found
	suite.Run("NotFound", func() {
		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-999")).Return(nil, storage.NewNotFoundError("invoice", "INV-999")).Once()

		invoice, err := suite.service.GetInvoice(suite.ctx, "INV-999")

		require.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "invoice with ID 'INV-999' not found")
	})
}

func (suite *InvoiceServiceTestSuite) TestUpdateInvoice() {
	t := suite.T()

	existingInvoice := &models.Invoice{
		ID:        "INV-001",
		Number:    "INV-2024-001",
		Status:    models.StatusDraft,
		Total:     1000.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	updateRequest := models.UpdateInvoiceRequest{
		ID:          "INV-001",
		Description: ptrString("Updated description"),
	}

	// Success case
	suite.Run("Success", func() {
		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(existingInvoice, nil).Once()
		suite.storage.On("UpdateInvoice", suite.ctx, mock.AnythingOfType("*models.Invoice")).Return(nil).Once()

		updatedInvoice, err := suite.service.UpdateInvoice(suite.ctx, updateRequest)

		require.NoError(t, err)
		require.NotNil(t, updatedInvoice)
		assert.Equal(t, "Updated description", updatedInvoice.Description)
	})

	// Invoice not found
	suite.Run("InvoiceNotFound", func() {
		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(nil, storage.NewNotFoundError("invoice", "INV-001")).Once()

		updatedInvoice, err := suite.service.UpdateInvoice(suite.ctx, updateRequest)

		require.Error(t, err)
		assert.Nil(t, updatedInvoice)
		assert.Contains(t, err.Error(), "invoice with ID 'INV-001' not found")
	})
}

func (suite *InvoiceServiceTestSuite) TestDeleteInvoice() {
	t := suite.T()

	// Success - Draft invoice
	suite.Run("DeleteDraftInvoice", func() {
		draftInvoice := &models.Invoice{
			ID:     "INV-001",
			Status: models.StatusDraft,
		}

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(draftInvoice, nil).Once()
		suite.storage.On("DeleteInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(nil).Once()

		err := suite.service.DeleteInvoice(suite.ctx, "INV-001")

		require.NoError(t, err)
	})

	// Cannot delete paid invoice
	suite.Run("CannotDeletePaid", func() {
		paidInvoice := &models.Invoice{
			ID:     "INV-001",
			Number: "INV-2024-001",
			Status: models.StatusPaid,
		}

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(paidInvoice, nil).Once()

		err := suite.service.DeleteInvoice(suite.ctx, "INV-001")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete paid invoice")
	})
}

func (suite *InvoiceServiceTestSuite) TestListInvoices() {
	t := suite.T()

	invoices := []*models.Invoice{
		{ID: "INV-001", Number: "INV-2024-001", Status: models.StatusDraft},
		{ID: "INV-002", Number: "INV-2024-002", Status: models.StatusSent},
		{ID: "INV-003", Number: "INV-2024-003", Status: models.StatusPaid},
	}

	// Success case
	suite.Run("Success", func() {
		filter := models.InvoiceFilter{
			Status: models.StatusDraft,
			Limit:  10,
		}

		result := &storage.InvoiceListResult{
			Invoices:   invoices[:1],
			TotalCount: 1,
			HasMore:    false,
		}

		suite.storage.On("ListInvoices", suite.ctx, filter).Return(result, nil).Once()

		listResult, err := suite.service.ListInvoices(suite.ctx, filter)

		require.NoError(t, err)
		require.NotNil(t, listResult)
		assert.Len(t, listResult.Invoices, 1)
		assert.Equal(t, int64(1), listResult.TotalCount)
	})
}

func (suite *InvoiceServiceTestSuite) TestAddWorkItemToInvoice() {
	t := suite.T()

	invoice := &models.Invoice{
		ID:        "INV-001",
		Number:    "INV-2024-001",
		Status:    models.StatusDraft,
		WorkItems: []models.WorkItem{},
		Version:   1,
		Client: models.Client{
			ID:        "CLIENT-001",
			Name:      "Test Client",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Date:      time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newWorkItem := models.WorkItem{
		Date:        time.Now(),
		Hours:       8.0,
		Rate:        100.0,
		Description: "New work",
		Total:       800.0,
	}

	// Success case
	suite.Run("Success", func() {
		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(invoice, nil).Once()
		suite.idGen.On("GenerateWorkItemID", suite.ctx).Return("WORK-001", nil).Once()
		suite.storage.On("UpdateInvoice", suite.ctx, mock.AnythingOfType("*models.Invoice")).Return(nil).Once()

		updatedInvoice, err := suite.service.AddWorkItemToInvoice(suite.ctx, "INV-001", newWorkItem)

		require.NoError(t, err)
		require.NotNil(t, updatedInvoice)
		assert.Len(t, updatedInvoice.WorkItems, 1)
		assert.Equal(t, "WORK-001", updatedInvoice.WorkItems[0].ID)
		assert.InEpsilon(t, 800.0, updatedInvoice.WorkItems[0].Total, 1e-9)
	})
}

func (suite *InvoiceServiceTestSuite) TestSendInvoice() {
	t := suite.T()

	// Success case
	suite.Run("Success", func() {
		draftInvoice := &models.Invoice{
			ID:      "INV-001",
			Status:  models.StatusDraft,
			Version: 1,
			WorkItems: []models.WorkItem{
				{ID: "WORK-001", Hours: 8, Rate: 100, Total: 800},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(draftInvoice, nil).Once()
		suite.storage.On("UpdateInvoice", suite.ctx, mock.AnythingOfType("*models.Invoice")).Return(nil).Once()

		sentInvoice, err := suite.service.SendInvoice(suite.ctx, "INV-001")

		require.NoError(t, err)
		require.NotNil(t, sentInvoice)
		assert.Equal(t, models.StatusSent, sentInvoice.Status)
		// Note: SentAt field doesn't exist in models.Invoice, checking status is sufficient
	})

	// Cannot send non-draft invoice
	suite.Run("CannotSendNonDraft", func() {
		sentInvoice := &models.Invoice{
			ID:     "INV-001",
			Status: models.StatusSent,
			WorkItems: []models.WorkItem{
				{ID: "WORK-001", Hours: 8, Rate: 100, Total: 800},
			},
		}

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(sentInvoice, nil).Once()

		invoice, err := suite.service.SendInvoice(suite.ctx, "INV-001")

		require.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "can only send draft invoices")
	})
}

func (suite *InvoiceServiceTestSuite) TestMarkInvoicePaid() {
	t := suite.T()

	// Success case
	suite.Run("Success", func() {
		sentInvoice := &models.Invoice{
			ID:        "INV-001",
			Status:    models.StatusSent,
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(sentInvoice, nil).Once()
		suite.storage.On("UpdateInvoice", suite.ctx, mock.AnythingOfType("*models.Invoice")).Return(nil).Once()

		paidInvoice, err := suite.service.MarkInvoicePaid(suite.ctx, "INV-001")

		require.NoError(t, err)
		require.NotNil(t, paidInvoice)
		assert.Equal(t, models.StatusPaid, paidInvoice.Status)
		// Note: PaidAt field doesn't exist in models.Invoice, checking status is sufficient
	})

	// Cannot mark draft invoice as paid
	suite.Run("CannotMarkDraftPaid", func() {
		draftInvoice := &models.Invoice{
			ID:     "INV-001",
			Status: models.StatusDraft,
		}

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(draftInvoice, nil).Once()

		invoice, err := suite.service.MarkInvoicePaid(suite.ctx, "INV-001")

		require.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "can only mark sent or overdue invoices as paid")
	})
}

func (suite *InvoiceServiceTestSuite) TestGetOverdueInvoices() {
	t := suite.T()

	overdueInvoices := []*models.Invoice{
		{ID: "INV-001", Status: models.StatusSent, DueDate: time.Now().AddDate(0, 0, -5)},
		{ID: "INV-002", Status: models.StatusSent, DueDate: time.Now().AddDate(0, 0, -10)},
	}

	// Success case
	suite.Run("Success", func() {
		suite.storage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return filter.Status == models.StatusSent && !filter.DueDateTo.IsZero()
		})).Return(&storage.InvoiceListResult{
			Invoices: overdueInvoices,
		}, nil).Once()

		// Mock UpdateInvoice calls for each overdue invoice (status gets updated to overdue)
		suite.storage.On("UpdateInvoice", suite.ctx, mock.AnythingOfType("*models.Invoice")).Return(nil).Times(2)

		invoices, err := suite.service.GetOverdueInvoices(suite.ctx)

		require.NoError(t, err)
		assert.Len(t, invoices, 2)
		assert.Equal(t, models.StatusOverdue, invoices[0].Status)
	})
}

func (suite *InvoiceServiceTestSuite) TestGetInvoiceStatistics() {
	t := suite.T()

	// Success case
	suite.Run("Success", func() {
		// Mock single ListInvoices call with all invoices
		allInvoices := []*models.Invoice{
			{Status: models.StatusDraft, Total: 100.0},
			{Status: models.StatusDraft, Total: 200.0},
			{Status: models.StatusDraft, Total: 150.0},
			{Status: models.StatusDraft, Total: 250.0},
			{Status: models.StatusDraft, Total: 300.0},
			{Status: models.StatusSent, Total: 400.0},
			{Status: models.StatusSent, Total: 500.0},
			{Status: models.StatusSent, Total: 600.0},
			{Status: models.StatusPaid, Total: 1000.0},
			{Status: models.StatusPaid, Total: 1200.0},
			{Status: models.StatusPaid, Total: 800.0},
			{Status: models.StatusPaid, Total: 900.0},
			{Status: models.StatusPaid, Total: 700.0},
			{Status: models.StatusPaid, Total: 600.0},
			{Status: models.StatusPaid, Total: 500.0},
			{Status: models.StatusPaid, Total: 400.0},
			{Status: models.StatusPaid, Total: 300.0},
			{Status: models.StatusPaid, Total: 200.0},
			{Status: models.StatusOverdue, Total: 750.0},
			{Status: models.StatusOverdue, Total: 850.0},
			{Status: models.StatusVoided, Total: 50.0},
		}

		suite.storage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			// Empty filter to get all invoices
			return filter.Status == "" && filter.ClientID == "" && filter.Limit == 0
		})).Return(&storage.InvoiceListResult{
			Invoices:   allInvoices,
			TotalCount: int64(len(allInvoices)),
		}, nil).Once()

		stats, err := suite.service.GetInvoiceStatistics(suite.ctx)

		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.Equal(t, 21, stats.TotalInvoices)
		assert.Equal(t, 5, stats.DraftCount)
		assert.Equal(t, 3, stats.SentCount)
		assert.Equal(t, 10, stats.PaidCount)
		assert.Equal(t, 2, stats.OverdueCount)
		assert.Equal(t, 1, stats.VoidedCount)
	})
}

// Helper function
func ptrString(s string) *string {
	return &s
}
