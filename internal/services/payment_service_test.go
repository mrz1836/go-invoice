package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-invoice/internal/blockchain"
	"github.com/mrz1836/go-invoice/internal/models"
)

func TestPaymentService_VerifyPayment(t *testing.T) {
	ctx := context.Background()

	// Create test invoice
	usdcAddress := "0x1234567890abcdef"
	testClient := models.Client{
		ID:    "CLIENT-001",
		Name:  "Test Client",
		Email: "test@example.com",
	}

	testInvoice := &models.Invoice{
		ID:                  "INV-001",
		Number:              "INV-2024-001",
		Total:               100.00,
		Client:              testClient,
		Status:              models.StatusSent,
		USDCAddressOverride: &usdcAddress,
		CreatedAt:           time.Now().AddDate(0, 0, -7),
		DueDate:             time.Now().AddDate(0, 0, 7),
	}

	tests := []struct {
		name             string
		setupMock        func(*blockchain.MockProvider)
		expectedStatus   models.PaymentStatus
		expectedReceived float64
		expectError      bool
		expectTxHash     bool
	}{
		{
			name: "payment found - exact amount",
			setupMock: func(m *blockchain.MockProvider) {
				err := m.MockPaymentScenario("payment_found", usdcAddress, 100.00, blockchain.TokenTypeUSDC)
				require.NoError(t, err)
			},
			expectedStatus:   models.PaymentStatusVerified,
			expectedReceived: 100.00,
			expectTxHash:     true,
		},
		{
			name: "payment not found",
			setupMock: func(m *blockchain.MockProvider) {
				err := m.MockPaymentScenario("payment_not_found", usdcAddress, 100.00, blockchain.TokenTypeUSDC)
				require.NoError(t, err)
			},
			expectedStatus:   models.PaymentStatusNotFound,
			expectedReceived: 0.00,
			expectTxHash:     false,
		},
		{
			name: "partial payment",
			setupMock: func(m *blockchain.MockProvider) {
				err := m.MockPaymentScenario("partial_payment", usdcAddress, 100.00, blockchain.TokenTypeUSDC)
				require.NoError(t, err)
			},
			expectedStatus:   models.PaymentStatusPartial,
			expectedReceived: 50.00,
			expectTxHash:     false,
		},
		{
			name: "overpayment",
			setupMock: func(m *blockchain.MockProvider) {
				err := m.MockPaymentScenario("overpayment", usdcAddress, 100.00, blockchain.TokenTypeUSDC)
				require.NoError(t, err)
			},
			expectedStatus:   models.PaymentStatusOverpaid,
			expectedReceived: 150.00,
			expectTxHash:     true,
		},
		{
			name: "network error",
			setupMock: func(m *blockchain.MockProvider) {
				err := m.MockPaymentScenario("network_error", usdcAddress, 100.00, blockchain.TokenTypeUSDC)
				require.NoError(t, err)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockProvider := blockchain.NewMockProvider()
			tt.setupMock(mockProvider)

			logger := &SimpleTestLogger{}
			mockStorage := new(MockInvoiceStorage)
			service := NewPaymentService(mockStorage, logger)

			config := PaymentVerificationConfig{
				PaymentMethod:      models.PaymentMethodUSDC,
				DefaultUSDCAddress: usdcAddress,
			}

			// Execute
			result, err := service.VerifyPayment(ctx, testInvoice, mockProvider, config)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.Status)
			// Use InDelta for float comparisons, with absolute delta to handle zero values
			assert.InDelta(t, tt.expectedReceived, result.ReceivedAmount, 0.01)
			assert.InDelta(t, 100.00, result.ExpectedAmount, 0.01)
			assert.Equal(t, usdcAddress, result.WalletAddress)
			assert.Equal(t, models.PaymentMethodUSDC, result.Method)

			if tt.expectTxHash {
				assert.NotEmpty(t, result.TransactionHash)
			} else {
				assert.Empty(t, result.TransactionHash)
			}
		})
	}
}

func TestPaymentService_MarkInvoiceAsPaid(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		initialStatus string
		expectUpdate  bool
		expectError   bool
	}{
		{
			name:          "mark sent invoice as paid",
			initialStatus: models.StatusSent,
			expectUpdate:  true,
		},
		{
			name:          "mark draft invoice as paid",
			initialStatus: models.StatusDraft,
			expectUpdate:  true,
		},
		{
			name:          "already paid invoice (idempotent)",
			initialStatus: models.StatusPaid,
			expectUpdate:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := &SimpleTestLogger{}
			mockStorage := new(MockInvoiceStorage)

			testClient := models.Client{
				ID:    "CLIENT-001",
				Name:  "Test Client",
				Email: "test@example.com",
			}

			invoice := &models.Invoice{
				ID:          "INV-001",
				Number:      "INV-2024-001",
				Status:      tt.initialStatus,
				Total:       100.00,
				Client:      testClient,
				Description: "Test invoice",
				Version:     1,
			}

			// Setup mock expectations
			mockStorage.On("GetInvoice", ctx, models.InvoiceID("INV-001")).Return(invoice, nil)
			if tt.expectUpdate {
				mockStorage.On("UpdateInvoice", ctx, mock.Anything).Return(nil)
			}

			service := NewPaymentService(mockStorage, logger)

			verification := &models.PaymentVerification{
				InvoiceID:       "INV-001",
				Status:          models.PaymentStatusVerified,
				Method:          models.PaymentMethodUSDC,
				ExpectedAmount:  100.00,
				ReceivedAmount:  100.00,
				TransactionHash: "0xabcdef",
			}

			// Execute
			err := service.MarkInvoiceAsPaid(ctx, "INV-001", verification)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify mock expectations
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestPaymentService_AddressOverrides(t *testing.T) {
	ctx := context.Background()

	// Test invoice-specific address override
	invoiceAddress := "0xINVOICESPECIFIC"
	globalAddress := "0xGLOBALADDRESS"

	invoice := &models.Invoice{
		ID:                  "INV-001",
		Number:              "INV-2024-001",
		Total:               100.00,
		USDCAddressOverride: &invoiceAddress,
		CreatedAt:           time.Now(),
		DueDate:             time.Now().AddDate(0, 0, 30),
	}

	mockProvider := blockchain.NewMockProvider()
	mockProvider.SetBalance(invoiceAddress, blockchain.TokenTypeUSDC, 100.00)

	logger := &SimpleTestLogger{}
	mockStorage := new(MockInvoiceStorage)
	service := NewPaymentService(mockStorage, logger)

	config := PaymentVerificationConfig{
		PaymentMethod:      models.PaymentMethodUSDC,
		DefaultUSDCAddress: globalAddress, // Should be overridden by invoice address
	}

	result, err := service.VerifyPayment(ctx, invoice, mockProvider, config)

	require.NoError(t, err)
	assert.Equal(t, invoiceAddress, result.WalletAddress, "should use invoice-specific address")
	assert.Equal(t, models.PaymentStatusVerified, result.Status)
}

func TestPaymentService_NoAddressConfigured(t *testing.T) {
	ctx := context.Background()

	invoice := &models.Invoice{
		ID:        "INV-001",
		Number:    "INV-2024-001",
		Total:     100.00,
		CreatedAt: time.Now(),
		DueDate:   time.Now().AddDate(0, 0, 30),
	}

	mockProvider := blockchain.NewMockProvider()
	logger := &SimpleTestLogger{}
	mockStorage := new(MockInvoiceStorage)
	service := NewPaymentService(mockStorage, logger)

	config := PaymentVerificationConfig{
		PaymentMethod: models.PaymentMethodUSDC,
		// No addresses configured
	}

	_, err := service.VerifyPayment(ctx, invoice, mockProvider, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no USDC address configured")
}

func TestPaymentService_GetPaymentAddress_BSV(t *testing.T) {
	ctx := context.Background()

	bsvAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	globalBSVAddress := "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"

	tests := []struct {
		name              string
		invoice           *models.Invoice
		config            PaymentVerificationConfig
		expectedAddress   string
		expectError       bool
		expectedErrorType error
	}{
		{
			name: "BSV with invoice-specific address",
			invoice: &models.Invoice{
				ID:                 "INV-001",
				BSVAddressOverride: &bsvAddress,
			},
			config: PaymentVerificationConfig{
				PaymentMethod:     models.PaymentMethodBSV,
				DefaultBSVAddress: globalBSVAddress,
			},
			expectedAddress: bsvAddress,
		},
		{
			name: "BSV with global config address",
			invoice: &models.Invoice{
				ID: "INV-001",
			},
			config: PaymentVerificationConfig{
				PaymentMethod:     models.PaymentMethodBSV,
				DefaultBSVAddress: globalBSVAddress,
			},
			expectedAddress: globalBSVAddress,
		},
		{
			name: "BSV with no address configured",
			invoice: &models.Invoice{
				ID: "INV-001",
			},
			config: PaymentVerificationConfig{
				PaymentMethod: models.PaymentMethodBSV,
			},
			expectError:       true,
			expectedErrorType: ErrNoBSVAddress,
		},
		{
			name: "BSV with empty override falls back to global",
			invoice: &models.Invoice{
				ID: "INV-001",
				BSVAddressOverride: func() *string {
					s := ""
					return &s
				}(),
			},
			config: PaymentVerificationConfig{
				PaymentMethod:     models.PaymentMethodBSV,
				DefaultBSVAddress: globalBSVAddress,
			},
			expectedAddress: globalBSVAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &SimpleTestLogger{}
			mockStorage := new(MockInvoiceStorage)
			service := NewPaymentService(mockStorage, logger)

			// Use reflection or make getPaymentAddress exported for testing
			// For now, we'll test through VerifyPayment which calls it
			mockProvider := blockchain.NewMockProvider()
			if !tt.expectError {
				mockProvider.SetBalance(tt.expectedAddress, blockchain.TokenTypeBSV, 100.00)
			}

			result, err := service.VerifyPayment(ctx, tt.invoice, mockProvider, tt.config)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErrorType)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedAddress, result.WalletAddress)
		})
	}
}

func TestPaymentService_UnsupportedPaymentMethods(t *testing.T) {
	ctx := context.Background()

	invoice := &models.Invoice{
		ID:     "INV-001",
		Total:  100.00,
		Status: models.StatusSent,
	}

	tests := []struct {
		name          string
		paymentMethod models.PaymentMethod
		expectError   bool
	}{
		{
			name:          "ACH payment method",
			paymentMethod: models.PaymentMethodACH,
			expectError:   true,
		},
		{
			name:          "Wire payment method",
			paymentMethod: models.PaymentMethodWire,
			expectError:   true,
		},
		{
			name:          "Other payment method",
			paymentMethod: models.PaymentMethodOther,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &SimpleTestLogger{}
			mockStorage := new(MockInvoiceStorage)
			service := NewPaymentService(mockStorage, logger)
			mockProvider := blockchain.NewMockProvider()

			config := PaymentVerificationConfig{
				PaymentMethod: tt.paymentMethod,
			}

			_, err := service.VerifyPayment(ctx, invoice, mockProvider, config)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrUnsupportedPaymentMethodForBlockchain)
			}
		})
	}
}

func TestPaymentService_BuildPaymentNotes(t *testing.T) {
	logger := &SimpleTestLogger{}
	mockStorage := new(MockInvoiceStorage)
	service := NewPaymentService(mockStorage, logger)

	confirmedTime := time.Now()

	tests := []struct {
		name             string
		verification     *models.PaymentVerification
		shouldInclude    []string
		shouldNotInclude []string
	}{
		{
			name: "complete payment notes with all fields",
			verification: &models.PaymentVerification{
				Method:          models.PaymentMethodUSDC,
				ReceivedAmount:  100.00,
				ExpectedAmount:  100.00,
				Currency:        "USDC",
				WalletAddress:   "0x123",
				TransactionHash: "0xabc",
				ConfirmedAt:     &confirmedTime,
				VerifiedAt:      time.Now(),
				VerifiedBy:      "etherscan",
				Status:          models.PaymentStatusVerified,
			},
			shouldInclude: []string{"USDC", "100.00", "0x123", "0xabc", "Confirmed:", "Verified:"},
		},
		{
			name: "payment without transaction hash",
			verification: &models.PaymentVerification{
				Method:         models.PaymentMethodUSDC,
				ReceivedAmount: 100.00,
				Currency:       "USDC",
				WalletAddress:  "0x123",
				Status:         models.PaymentStatusVerified,
			},
			shouldInclude:    []string{"USDC", "100.00", "0x123"},
			shouldNotInclude: []string{"Transaction:"},
		},
		{
			name: "overpayment note",
			verification: &models.PaymentVerification{
				Method:          models.PaymentMethodUSDC,
				ReceivedAmount:  150.00,
				ExpectedAmount:  100.00,
				Currency:        "USDC",
				WalletAddress:   "0x123",
				TransactionHash: "0xabc",
				Status:          models.PaymentStatusOverpaid,
			},
			shouldInclude: []string{"Overpaid by 50.00"},
		},
		{
			name: "payment without confirmation",
			verification: &models.PaymentVerification{
				Method:          models.PaymentMethodUSDC,
				ReceivedAmount:  100.00,
				Currency:        "USDC",
				WalletAddress:   "0x123",
				TransactionHash: "0xabc",
				ConfirmedAt:     nil,
				Status:          models.PaymentStatusVerified,
			},
			shouldNotInclude: []string{"Confirmed:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notes := service.buildPaymentNotes(tt.verification)

			for _, expected := range tt.shouldInclude {
				assert.Contains(t, notes, expected, "notes should include: %s", expected)
			}

			for _, notExpected := range tt.shouldNotInclude {
				assert.NotContains(t, notes, notExpected, "notes should not include: %s", notExpected)
			}
		})
	}
}

func TestPaymentService_DeterminePaymentStatus_EdgeCases(t *testing.T) {
	logger := &SimpleTestLogger{}
	mockStorage := new(MockInvoiceStorage)
	service := NewPaymentService(mockStorage, logger)

	tests := []struct {
		name           string
		verification   *models.PaymentVerification
		expectedStatus models.PaymentStatus
	}{
		{
			name: "exact amount match",
			verification: &models.PaymentVerification{
				ExpectedAmount: 100.00,
				ReceivedAmount: 100.00,
			},
			expectedStatus: models.PaymentStatusVerified,
		},
		{
			name: "within 1% tolerance - slightly under",
			verification: &models.PaymentVerification{
				ExpectedAmount: 100.00,
				ReceivedAmount: 99.50, // Within 1% tolerance
			},
			expectedStatus: models.PaymentStatusVerified,
		},
		{
			name: "within 1% tolerance - slightly over",
			verification: &models.PaymentVerification{
				ExpectedAmount: 100.00,
				ReceivedAmount: 100.50, // Within 1% tolerance
			},
			expectedStatus: models.PaymentStatusVerified,
		},
		{
			name: "overpaid by more than 1%",
			verification: &models.PaymentVerification{
				ExpectedAmount: 100.00,
				ReceivedAmount: 105.00, // More than 1% over
			},
			expectedStatus: models.PaymentStatusOverpaid,
		},
		{
			name: "underpaid beyond tolerance",
			verification: &models.PaymentVerification{
				ExpectedAmount: 100.00,
				ReceivedAmount: 95.00, // More than 1% under
			},
			expectedStatus: models.PaymentStatusPartial,
		},
		{
			name: "zero received",
			verification: &models.PaymentVerification{
				ExpectedAmount: 100.00,
				ReceivedAmount: 0.00,
			},
			expectedStatus: models.PaymentStatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := service.determinePaymentStatus(tt.verification)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestPaymentService_ContextCancellation(t *testing.T) {
	logger := &SimpleTestLogger{}
	mockStorage := new(MockInvoiceStorage)
	service := NewPaymentService(mockStorage, logger)

	invoice := &models.Invoice{
		ID:    "INV-001",
		Total: 100.00,
	}

	config := PaymentVerificationConfig{
		PaymentMethod:      models.PaymentMethodUSDC,
		DefaultUSDCAddress: "0x123",
	}

	t.Run("VerifyPayment with canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		mockProvider := blockchain.NewMockProvider()
		_, err := service.VerifyPayment(ctx, invoice, mockProvider, config)

		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("MarkInvoiceAsPaid with canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		verification := &models.PaymentVerification{
			InvoiceID: "INV-001",
			Status:    models.PaymentStatusVerified,
		}

		err := service.MarkInvoiceAsPaid(ctx, "INV-001", verification)
		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}

func TestPaymentService_MarkInvoiceAsPaid_WithPaymentNotes(t *testing.T) {
	ctx := context.Background()
	logger := &SimpleTestLogger{}
	mockStorage := new(MockInvoiceStorage)

	invoice := &models.Invoice{
		ID:          "INV-001",
		Number:      "INV-2024-001",
		Status:      models.StatusSent,
		Total:       100.00,
		Description: "Original description",
		Version:     1,
	}

	mockStorage.On("GetInvoice", ctx, models.InvoiceID("INV-001")).Return(invoice, nil)
	mockStorage.On("UpdateInvoice", ctx, mock.MatchedBy(func(inv *models.Invoice) bool {
		// Verify description was updated with payment details
		return assert.Contains(t, inv.Description, "Original description") &&
			assert.Contains(t, inv.Description, "Payment Details:") &&
			assert.Contains(t, inv.Description, "0xabcdef")
	})).Return(nil)

	service := NewPaymentService(mockStorage, logger)

	verification := &models.PaymentVerification{
		InvoiceID:       "INV-001",
		Status:          models.PaymentStatusVerified,
		Method:          models.PaymentMethodUSDC,
		ReceivedAmount:  100.00,
		TransactionHash: "0xabcdef",
	}

	err := service.MarkInvoiceAsPaid(ctx, "INV-001", verification)
	require.NoError(t, err)

	mockStorage.AssertExpectations(t)
}

func TestPaymentService_GetRelevantTransactions_TimeWindow(t *testing.T) {
	ctx := context.Background()
	logger := &SimpleTestLogger{}
	mockStorage := new(MockInvoiceStorage)
	service := NewPaymentService(mockStorage, logger)

	createdAt := time.Now().AddDate(0, 0, -10)
	dueDate := time.Now().AddDate(0, 0, 5)

	invoice := &models.Invoice{
		ID:        "INV-001",
		Total:     100.00,
		CreatedAt: createdAt,
		DueDate:   dueDate,
	}

	mockProvider := blockchain.NewMockProvider()

	// Add transactions at different times
	mockProvider.AddTransaction("0xtest", blockchain.Transaction{
		Hash:      "0x1",
		To:        "0xtest",
		Amount:    100.00,
		Token:     blockchain.TokenTypeUSDC,
		Timestamp: createdAt.Add(-1 * time.Hour), // Before invoice created - should be excluded
	})
	mockProvider.AddTransaction("0xtest", blockchain.Transaction{
		Hash:      "0x2",
		To:        "0xtest",
		Amount:    100.00,
		Token:     blockchain.TokenTypeUSDC,
		Timestamp: createdAt.Add(1 * time.Hour), // Within window
	})
	mockProvider.AddTransaction("0xtest", blockchain.Transaction{
		Hash:      "0x3",
		To:        "0xtest",
		Amount:    100.00,
		Token:     blockchain.TokenTypeUSDC,
		Timestamp: dueDate.AddDate(0, 0, 31), // After 30 day window - should be excluded
	})

	// Call through VerifyPayment which uses getRelevantTransactions internally
	mockProvider.SetBalance("0xtest", blockchain.TokenTypeUSDC, 100.00)

	config := PaymentVerificationConfig{
		PaymentMethod:      models.PaymentMethodUSDC,
		DefaultUSDCAddress: "0xtest",
	}

	invoice.USDCAddressOverride = func() *string { s := "0xtest"; return &s }()

	result, err := service.VerifyPayment(ctx, invoice, mockProvider, config)
	require.NoError(t, err)

	// Should use transaction within the window
	assert.Equal(t, "0x2", result.TransactionHash)
}

// SimpleTestLogger is a simple test logger that discards all output
type SimpleTestLogger struct{}

func (l *SimpleTestLogger) Info(msg string, fields ...any)  {}
func (l *SimpleTestLogger) Error(msg string, fields ...any) {}
func (l *SimpleTestLogger) Debug(msg string, fields ...any) {}
