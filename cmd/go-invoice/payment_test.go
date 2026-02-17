package main

import (
	"testing"

	"github.com/mrz1836/go-invoice/internal/blockchain"
	"github.com/mrz1836/go-invoice/internal/cli"
	"github.com/mrz1836/go-invoice/internal/config"
	"github.com/mrz1836/go-invoice/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBlockchainProvider(t *testing.T) {
	app := &App{
		logger: cli.NewLogger(false),
	}

	tests := []struct {
		name          string
		paymentMethod string
		testnet       bool
		apiKey        string
		expectedType  string
		expectedName  string
		expectError   bool
	}{
		{
			name:          "USDC mainnet provider",
			paymentMethod: "USDC",
			testnet:       false,
			apiKey:        "test-key",
			expectedType:  "*blockchain.EtherscanProvider",
			expectedName:  "etherscan",
		},
		{
			name:          "USDC testnet provider",
			paymentMethod: "USDC",
			testnet:       true,
			apiKey:        "test-key",
			expectedType:  "*blockchain.EtherscanProvider",
			expectedName:  "etherscan-sepolia",
		},
		{
			name:          "USDC without API key",
			paymentMethod: "USDC",
			testnet:       false,
			apiKey:        "",
			expectedType:  "*blockchain.EtherscanProvider",
			expectedName:  "etherscan",
		},
		{
			name:          "BSV provider",
			paymentMethod: "BSV",
			testnet:       false,
			expectedType:  "*blockchain.BSVProvider",
		},
		{
			name:          "unsupported payment method",
			paymentMethod: "UNKNOWN",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := app.createBlockchainProvider(tt.paymentMethod, tt.testnet, tt.apiKey)

			if tt.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrUnsupportedPaymentMethod)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, provider)

			if tt.expectedName != "" {
				assert.Equal(t, tt.expectedName, provider.Name())
			}
		})
	}
}

func TestGetInvoicePaymentAddress(t *testing.T) {
	app := &App{
		logger: cli.NewLogger(false),
	}

	usdcAddress := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1"
	bsvAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	globalUSDCAddress := "0xGLOBALUSDC"
	globalBSVAddress := "1GLOBALBSV"

	cfg := &config.Config{
		Business: config.BusinessConfig{
			CryptoPayments: config.CryptoPayments{
				USDCAddress: globalUSDCAddress,
				BSVAddress:  globalBSVAddress,
			},
		},
	}

	tests := []struct {
		name            string
		invoice         *models.Invoice
		method          models.PaymentMethod
		expectedAddress string
		expectError     bool
		expectedError   error
	}{
		{
			name: "USDC with invoice override",
			invoice: &models.Invoice{
				USDCAddressOverride: &usdcAddress,
			},
			method:          models.PaymentMethodUSDC,
			expectedAddress: usdcAddress,
		},
		{
			name:            "USDC with global config",
			invoice:         &models.Invoice{},
			method:          models.PaymentMethodUSDC,
			expectedAddress: globalUSDCAddress,
		},
		{
			name: "USDC with empty override uses global",
			invoice: &models.Invoice{
				USDCAddressOverride: func() *string {
					s := ""
					return &s
				}(),
			},
			method:          models.PaymentMethodUSDC,
			expectedAddress: globalUSDCAddress,
		},
		{
			name: "BSV with invoice override",
			invoice: &models.Invoice{
				BSVAddressOverride: &bsvAddress,
			},
			method:          models.PaymentMethodBSV,
			expectedAddress: bsvAddress,
		},
		{
			name:            "BSV with global config",
			invoice:         &models.Invoice{},
			method:          models.PaymentMethodBSV,
			expectedAddress: globalBSVAddress,
		},
		{
			name:          "ACH method not supported",
			invoice:       &models.Invoice{},
			method:        models.PaymentMethodACH,
			expectError:   true,
			expectedError: ErrUnsupportedPaymentMethod,
		},
		{
			name:          "Wire method not supported",
			invoice:       &models.Invoice{},
			method:        models.PaymentMethodWire,
			expectError:   true,
			expectedError: ErrUnsupportedPaymentMethod,
		},
		{
			name:          "Other method not supported",
			invoice:       &models.Invoice{},
			method:        models.PaymentMethodOther,
			expectError:   true,
			expectedError: ErrUnsupportedPaymentMethod,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := app.getInvoicePaymentAddress(tt.invoice, cfg, tt.method)

			if tt.expectError {
				require.Error(t, err)
				if tt.expectedError != nil {
					require.ErrorIs(t, err, tt.expectedError)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedAddress, address)
		})
	}
}

func TestGetInvoicePaymentAddress_NoAddressConfigured(t *testing.T) {
	app := &App{
		logger: cli.NewLogger(false),
	}

	cfg := &config.Config{
		Business: config.BusinessConfig{
			CryptoPayments: config.CryptoPayments{
				// No addresses configured
			},
		},
	}

	tests := []struct {
		name          string
		method        models.PaymentMethod
		expectedError error
	}{
		{
			name:          "USDC with no address",
			method:        models.PaymentMethodUSDC,
			expectedError: ErrNoUSDCAddress,
		},
		{
			name:          "BSV with no address",
			method:        models.PaymentMethodBSV,
			expectedError: ErrNoBSVAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice := &models.Invoice{}
			_, err := app.getInvoicePaymentAddress(invoice, cfg, tt.method)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestGetInvoicePaymentAddress_PriorityOrder(t *testing.T) {
	app := &App{
		logger: cli.NewLogger(false),
	}

	invoiceSpecific := "0xINVOICESPECIFIC"
	globalConfig := "0xGLOBAL"

	cfg := &config.Config{
		Business: config.BusinessConfig{
			CryptoPayments: config.CryptoPayments{
				USDCAddress: globalConfig,
			},
		},
	}

	// Invoice-specific override should take priority
	invoice := &models.Invoice{
		USDCAddressOverride: &invoiceSpecific,
	}

	address, err := app.getInvoicePaymentAddress(invoice, cfg, models.PaymentMethodUSDC)

	require.NoError(t, err)
	assert.Equal(t, invoiceSpecific, address, "Invoice-specific address should take priority over global config")
}

func TestBlockchainProviderSupportedTokens(t *testing.T) {
	app := &App{
		logger: cli.NewLogger(false),
	}

	usdcProvider, err := app.createBlockchainProvider("USDC", false, "test-key")
	require.NoError(t, err)
	assert.Equal(t, []blockchain.TokenType{blockchain.TokenTypeUSDC}, usdcProvider.SupportedTokens())

	bsvProvider, err := app.createBlockchainProvider("BSV", false, "")
	require.NoError(t, err)
	assert.Equal(t, []blockchain.TokenType{blockchain.TokenTypeBSV}, bsvProvider.SupportedTokens())
}

func TestEtherscanProviderAPIKeyConfiguration(t *testing.T) {
	app := &App{
		logger: cli.NewLogger(false),
	}

	tests := []struct {
		name    string
		apiKey  string
		testnet bool
	}{
		{
			name:    "mainnet with API key",
			apiKey:  "my-api-key-123",
			testnet: false,
		},
		{
			name:    "testnet with API key",
			apiKey:  "my-testnet-key",
			testnet: true,
		},
		{
			name:    "mainnet without API key",
			apiKey:  "",
			testnet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := app.createBlockchainProvider("USDC", tt.testnet, tt.apiKey)
			require.NoError(t, err)

			etherscanProvider, ok := provider.(*blockchain.EtherscanProvider)
			require.True(t, ok, "Provider should be EtherscanProvider")

			// Verify the provider was created (we can't directly access private fields,
			// but we can verify it doesn't error)
			assert.NotNil(t, etherscanProvider)
		})
	}
}
