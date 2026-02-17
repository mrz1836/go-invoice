package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/go-invoice/internal/models"
	"github.com/mrz1836/go-invoice/internal/storage"
)

var errConnectionTimeout = errors.New("connection timeout")

// InvoiceServiceTestSuite tests for the InvoiceService
type InvoiceServiceTestSuite struct {
	suite.Suite

	ctx           context.Context //nolint:containedctx // Test suite context is acceptable
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

	// Version handling - ensures version is NOT incremented by service layer
	suite.Run("VersionNotIncrementedByServiceLayer", func() {
		invoiceWithVersion := &models.Invoice{
			ID:        "INV-VERSION-TEST",
			Number:    "INV-2024-002",
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

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-VERSION-TEST")).Return(invoiceWithVersion, nil).Once()
		suite.idGen.On("GenerateWorkItemID", suite.ctx).Return("WORK-002", nil).Once()

		// Verify that the invoice passed to UpdateInvoice still has version 1
		// (not incremented to 2 by AddWorkItem)
		suite.storage.On("UpdateInvoice", suite.ctx, mock.MatchedBy(func(inv *models.Invoice) bool {
			// The version should still be 1 because AddWorkItemWithoutVersionIncrement is used
			// Storage layer will handle the increment
			return inv.Version == 1
		})).Return(nil).Once()

		updatedInvoice, err := suite.service.AddWorkItemToInvoice(suite.ctx, "INV-VERSION-TEST", newWorkItem)

		require.NoError(t, err)
		require.NotNil(t, updatedInvoice)
		assert.Equal(t, 1, updatedInvoice.Version) // Version should still be 1
	})
}

func (suite *InvoiceServiceTestSuite) TestAddLineItemToInvoice() {
	t := suite.T()

	invoice := &models.Invoice{
		ID:        "INV-001",
		Number:    "INV-2024-001",
		Status:    models.StatusDraft,
		LineItems: []models.LineItem{},
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

	amount := 5000.0
	newLineItem := models.LineItem{
		Type:        models.LineItemTypeFixed,
		Date:        time.Now(),
		Description: "Repository Maintenance",
		Amount:      &amount,
		Total:       5000.0,
	}

	// Success case
	suite.Run("Success", func() {
		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(invoice, nil).Once()
		suite.idGen.On("GenerateWorkItemID", suite.ctx).Return("LINE-001", nil).Once()
		suite.storage.On("UpdateInvoice", suite.ctx, mock.AnythingOfType("*models.Invoice")).Return(nil).Once()

		updatedInvoice, err := suite.service.AddLineItemToInvoice(suite.ctx, "INV-001", newLineItem)

		require.NoError(t, err)
		require.NotNil(t, updatedInvoice)
		assert.Len(t, updatedInvoice.LineItems, 1)
		assert.Equal(t, "LINE-001", updatedInvoice.LineItems[0].ID)
		assert.Equal(t, models.LineItemTypeFixed, updatedInvoice.LineItems[0].Type)
		require.NotNil(t, updatedInvoice.LineItems[0].Amount)
		assert.InEpsilon(t, 5000.0, *updatedInvoice.LineItems[0].Amount, 1e-9)
	})

	// Version handling - ensures version is NOT incremented by service layer
	// This is a regression test for the bug where AddLineItem() was incrementing version
	suite.Run("VersionNotIncrementedByServiceLayer", func() {
		invoiceWithVersion := &models.Invoice{
			ID:        "INV-VERSION-TEST",
			Number:    "INV-2024-002",
			Status:    models.StatusDraft,
			LineItems: []models.LineItem{},
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

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-VERSION-TEST")).Return(invoiceWithVersion, nil).Once()
		suite.idGen.On("GenerateWorkItemID", suite.ctx).Return("LINE-002", nil).Once()

		// Verify that the invoice passed to UpdateInvoice still has version 1
		// (not incremented to 2 by AddLineItem)
		suite.storage.On("UpdateInvoice", suite.ctx, mock.MatchedBy(func(inv *models.Invoice) bool {
			// The version should still be 1 because AddLineItemWithoutVersionIncrement is used
			// Storage layer will handle the increment
			return inv.Version == 1
		})).Return(nil).Once()

		updatedInvoice, err := suite.service.AddLineItemToInvoice(suite.ctx, "INV-VERSION-TEST", newLineItem)

		require.NoError(t, err)
		require.NotNil(t, updatedInvoice)
		assert.Equal(t, 1, updatedInvoice.Version) // Version should still be 1
	})

	// Invoice not found
	suite.Run("InvoiceNotFound", func() {
		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-MISSING")).Return(nil, storage.NewNotFoundError("invoice", "INV-MISSING")).Once()

		updatedInvoice, err := suite.service.AddLineItemToInvoice(suite.ctx, "INV-MISSING", newLineItem)

		require.Error(t, err)
		assert.Nil(t, updatedInvoice)
		assert.Contains(t, err.Error(), "failed to retrieve invoice")
	})

	// Cannot add to non-draft invoice
	suite.Run("CannotAddToNonDraft", func() {
		sentInvoice := &models.Invoice{
			ID:        "INV-SENT",
			Number:    "INV-2024-003",
			Status:    models.StatusSent,
			LineItems: []models.LineItem{},
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

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-SENT")).Return(sentInvoice, nil).Once()

		updatedInvoice, err := suite.service.AddLineItemToInvoice(suite.ctx, "INV-SENT", newLineItem)

		require.Error(t, err)
		assert.Nil(t, updatedInvoice)
		assert.Contains(t, err.Error(), "can only add work items to draft invoices")
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
		// SentAt field doesn't exist in models.Invoice, checking status is sufficient
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
		// PaidAt field doesn't exist in models.Invoice, checking status is sufficient
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

func (suite *InvoiceServiceTestSuite) TestRemoveWorkItemFromInvoice() {
	t := suite.T()

	// Create test invoice with multiple work items
	invoiceWithWorkItems := &models.Invoice{
		ID:      "INV-001",
		Number:  "INV-2024-001",
		Status:  models.StatusDraft,
		Version: 1,
		Client: models.Client{
			ID:        "CLIENT-001",
			Name:      "Test Client",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		WorkItems: []models.WorkItem{
			{
				ID:          "WORK-001",
				Date:        time.Now(),
				Hours:       8.0,
				Rate:        100.0,
				Description: "First work item",
				Total:       800.0,
				CreatedAt:   time.Now(),
			},
			{
				ID:          "WORK-002",
				Date:        time.Now(),
				Hours:       4.0,
				Rate:        150.0,
				Description: "Second work item",
				Total:       600.0,
				CreatedAt:   time.Now(),
			},
		},
		Date:      time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Success case - remove existing work item
	suite.Run("Success", func() {
		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(invoiceWithWorkItems, nil).Once()
		suite.storage.On("UpdateInvoice", suite.ctx, mock.AnythingOfType("*models.Invoice")).Return(nil).Once()

		updatedInvoice, err := suite.service.RemoveWorkItemFromInvoice(suite.ctx, "INV-001", "WORK-001")

		require.NoError(t, err)
		require.NotNil(t, updatedInvoice)
		assert.Len(t, updatedInvoice.WorkItems, 1) // Should have 1 work item left
		assert.Equal(t, "WORK-002", updatedInvoice.WorkItems[0].ID)
	})

	// Invoice not found
	suite.Run("InvoiceNotFound", func() {
		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-999")).Return(nil, storage.NewNotFoundError("invoice", "INV-999")).Once()

		updatedInvoice, err := suite.service.RemoveWorkItemFromInvoice(suite.ctx, "INV-999", "WORK-001")

		require.Error(t, err)
		assert.Nil(t, updatedInvoice)
		assert.Contains(t, err.Error(), "failed to retrieve invoice")
	})

	// Cannot remove from non-draft invoice
	suite.Run("CannotRemoveFromNonDraft", func() {
		sentInvoice := &models.Invoice{
			ID:      "INV-001",
			Status:  models.StatusSent,
			Version: 1,
			WorkItems: []models.WorkItem{
				{ID: "WORK-001", Hours: 8, Rate: 100, Total: 800},
			},
		}

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(sentInvoice, nil).Once()

		updatedInvoice, err := suite.service.RemoveWorkItemFromInvoice(suite.ctx, "INV-001", "WORK-001")

		require.Error(t, err)
		assert.Nil(t, updatedInvoice)
		assert.Contains(t, err.Error(), "can only remove work items from draft invoices")
		assert.Contains(t, err.Error(), "current status: sent")
	})

	// Work item not found
	suite.Run("WorkItemNotFound", func() {
		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(invoiceWithWorkItems, nil).Once()

		updatedInvoice, err := suite.service.RemoveWorkItemFromInvoice(suite.ctx, "INV-001", "WORK-999")

		require.Error(t, err)
		assert.Nil(t, updatedInvoice)
		assert.Contains(t, err.Error(), "failed to remove work item")
	})

	// Remove last work item
	suite.Run("RemoveLastWorkItem", func() {
		invoiceWithOneItem := &models.Invoice{
			ID:      "INV-001",
			Status:  models.StatusDraft,
			Version: 1,
			WorkItems: []models.WorkItem{
				{ID: "WORK-001", Hours: 8, Rate: 100, Total: 800},
			},
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

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-001")).Return(invoiceWithOneItem, nil).Once()
		suite.storage.On("UpdateInvoice", suite.ctx, mock.AnythingOfType("*models.Invoice")).Return(nil).Once()

		updatedInvoice, err := suite.service.RemoveWorkItemFromInvoice(suite.ctx, "INV-001", "WORK-001")

		require.NoError(t, err)
		require.NotNil(t, updatedInvoice)
		assert.Empty(t, updatedInvoice.WorkItems) // Should have no work items left
	})

	// Storage update failure
	suite.Run("StorageUpdateFailure", func() {
		// Create a copy so we can modify it without affecting other tests
		testInvoice := &models.Invoice{
			ID:      "INV-STORAGE-FAIL",
			Number:  "INV-2024-001",
			Status:  models.StatusDraft,
			Version: 1,
			Client: models.Client{
				ID:        "CLIENT-001",
				Name:      "Test Client",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			WorkItems: []models.WorkItem{
				{
					ID:          "WORK-001",
					Date:        time.Now(),
					Hours:       8.0,
					Rate:        100.0,
					Description: "First work item",
					Total:       800.0,
					CreatedAt:   time.Now(),
				},
			},
			Date:      time.Now(),
			DueDate:   time.Now().AddDate(0, 0, 30),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		suite.storage.On("GetInvoice", suite.ctx, models.InvoiceID("INV-STORAGE-FAIL")).Return(testInvoice, nil).Once()
		suite.storage.On("UpdateInvoice", suite.ctx, mock.AnythingOfType("*models.Invoice")).Return(storage.NewStorageUnavailableError("database connection lost", errConnectionTimeout)).Once()

		updatedInvoice, err := suite.service.RemoveWorkItemFromInvoice(suite.ctx, "INV-STORAGE-FAIL", "WORK-001")

		require.Error(t, err)
		assert.Nil(t, updatedInvoice)
		assert.Contains(t, err.Error(), "failed to update invoice after removing work item")
	})

	// Context cancellation
	suite.Run("ContextCanceled", func() {
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		updatedInvoice, err := suite.service.RemoveWorkItemFromInvoice(canceledCtx, "INV-001", "WORK-001")

		require.Error(t, err)
		assert.Nil(t, updatedInvoice)
		assert.Equal(t, context.Canceled, err)
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

// TestGetInvoiceByNumber tests the GetInvoiceByNumber method
func (suite *InvoiceServiceTestSuite) TestGetInvoiceByNumber() {
	t := suite.T()

	testInvoice := &models.Invoice{
		ID:        "INV-001",
		Number:    "INV-2024-001",
		Status:    models.StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	suite.Run("Success", func() {
		suite.storage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return true
		})).Return(&storage.InvoiceListResult{
			Invoices: []*models.Invoice{testInvoice},
		}, nil).Once()

		invoice, err := suite.service.GetInvoiceByNumber(suite.ctx, "INV-2024-001")

		require.NoError(t, err)
		require.NotNil(t, invoice)
		assert.Equal(t, "INV-2024-001", invoice.Number)
	})

	suite.Run("EmptyNumber", func() {
		invoice, err := suite.service.GetInvoiceByNumber(suite.ctx, "")

		require.Error(t, err)
		assert.Nil(t, invoice)
		assert.ErrorIs(t, err, ErrInvoiceNumberEmpty)
	})

	suite.Run("WhitespaceOnlyNumber", func() {
		invoice, err := suite.service.GetInvoiceByNumber(suite.ctx, "   ")

		require.Error(t, err)
		assert.Nil(t, invoice)
		assert.ErrorIs(t, err, ErrInvoiceNumberEmpty)
	})

	suite.Run("NotFound", func() {
		suite.storage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return true
		})).Return(&storage.InvoiceListResult{
			Invoices: []*models.Invoice{testInvoice},
		}, nil).Once()

		invoice, err := suite.service.GetInvoiceByNumber(suite.ctx, "NONEXISTENT")

		require.Error(t, err)
		assert.Nil(t, invoice)
		assert.ErrorIs(t, err, ErrInvoiceNumberNotFound)
	})

	suite.Run("StorageError", func() {
		suite.storage.On("ListInvoices", suite.ctx, mock.MatchedBy(func(filter models.InvoiceFilter) bool {
			return true
		})).Return(nil, errConnectionTimeout).Once()

		invoice, err := suite.service.GetInvoiceByNumber(suite.ctx, "INV-2024-001")

		require.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "failed to search for invoice")
	})

	suite.Run("ContextCanceled", func() {
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		invoice, err := suite.service.GetInvoiceByNumber(canceledCtx, "INV-2024-001")

		require.Error(t, err)
		assert.Nil(t, invoice)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

// Helper function
func ptrString(s string) *string {
	return &s
}
