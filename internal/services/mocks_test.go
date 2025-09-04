package services

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mrz/go-invoice/internal/models"
	"github.com/mrz/go-invoice/internal/storage"
)

// MockInvoiceStorage is a mock implementation of the invoice storage interface
type MockInvoiceStorage struct {
	mock.Mock
}

func (m *MockInvoiceStorage) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockInvoiceStorage) GetInvoice(ctx context.Context, id models.InvoiceID) (*models.Invoice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

func (m *MockInvoiceStorage) UpdateInvoice(ctx context.Context, invoice *models.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockInvoiceStorage) DeleteInvoice(ctx context.Context, id models.InvoiceID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInvoiceStorage) ListInvoices(ctx context.Context, filter models.InvoiceFilter) (*storage.InvoiceListResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.InvoiceListResult), args.Error(1)
}

func (m *MockInvoiceStorage) CountInvoices(ctx context.Context, filter models.InvoiceFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockInvoiceStorage) ExistsInvoice(ctx context.Context, id models.InvoiceID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(bool), args.Error(1)
}

// MockClientStorage is a mock implementation of the client storage interface
type MockClientStorage struct {
	mock.Mock
}

func (m *MockClientStorage) CreateClient(ctx context.Context, client *models.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientStorage) GetClient(ctx context.Context, id models.ClientID) (*models.Client, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientStorage) UpdateClient(ctx context.Context, client *models.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientStorage) DeleteClient(ctx context.Context, id models.ClientID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientStorage) ListClients(ctx context.Context, activeOnly bool, limit, offset int) (*storage.ClientListResult, error) {
	args := m.Called(ctx, activeOnly, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.ClientListResult), args.Error(1)
}

func (m *MockClientStorage) FindClientByEmail(ctx context.Context, email string) (*models.Client, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientStorage) ExistsClient(ctx context.Context, id models.ClientID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(bool), args.Error(1)
}

// MockIDGenerator is a mock implementation of the ID generator interface
type MockIDGenerator struct {
	mock.Mock
}

func (m *MockIDGenerator) GenerateClientID(ctx context.Context) (models.ClientID, error) {
	args := m.Called(ctx)
	return args.Get(0).(models.ClientID), args.Error(1)
}

func (m *MockIDGenerator) GenerateInvoiceID(ctx context.Context) (models.InvoiceID, error) {
	args := m.Called(ctx)
	return args.Get(0).(models.InvoiceID), args.Error(1)
}

func (m *MockIDGenerator) GenerateWorkItemID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.Get(0).(string), args.Error(1)
}
