package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEtherscanProvider(t *testing.T) {
	tests := []struct {
		name            string
		apiKey          string
		testnet         bool
		expectedChainID int
		expectedName    string
		expectedAddress string
	}{
		{
			name:            "mainnet provider",
			apiKey:          "test-api-key",
			testnet:         false,
			expectedChainID: EtherscanMainnetChainID,
			expectedName:    "etherscan",
			expectedAddress: USDCMainnetContract,
		},
		{
			name:            "sepolia testnet provider",
			apiKey:          "test-api-key",
			testnet:         true,
			expectedChainID: EtherscanSepoliaChainID,
			expectedName:    "etherscan-sepolia",
			expectedAddress: USDCSepoliaContract,
		},
		{
			name:            "provider without api key",
			apiKey:          "",
			testnet:         false,
			expectedChainID: EtherscanMainnetChainID,
			expectedName:    "etherscan",
			expectedAddress: USDCMainnetContract,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewEtherscanProvider(tt.apiKey, tt.testnet)

			assert.Equal(t, tt.apiKey, provider.apiKey)
			assert.Equal(t, tt.expectedChainID, provider.chainID)
			assert.Equal(t, tt.expectedName, provider.Name())
			assert.Equal(t, tt.expectedAddress, provider.usdcContract)
			assert.Equal(t, []TokenType{TokenTypeUSDC}, provider.SupportedTokens())
		})
	}
}

func TestEtherscanProvider_GetBalance(t *testing.T) {
	tests := []struct {
		name           string
		token          TokenType
		apiResponse    EtherscanResponse
		statusCode     int
		expectedError  string
		expectedAmount float64
		testnet        bool
	}{
		{
			name:  "successful balance query - mainnet",
			token: TokenTypeUSDC,
			apiResponse: EtherscanResponse{
				Status:  "1",
				Message: "OK",
				Result:  "100000000", // 100 USDC (6 decimals)
			},
			statusCode:     http.StatusOK,
			expectedAmount: 100.00,
			testnet:        false,
		},
		{
			name:  "successful balance query - testnet",
			token: TokenTypeUSDC,
			apiResponse: EtherscanResponse{
				Status:  "1",
				Message: "OK",
				Result:  "250500000", // 250.50 USDC
			},
			statusCode:     http.StatusOK,
			expectedAmount: 250.50,
			testnet:        true,
		},
		{
			name:  "zero balance",
			token: TokenTypeUSDC,
			apiResponse: EtherscanResponse{
				Status:  "1",
				Message: "OK",
				Result:  "0",
			},
			statusCode:     http.StatusOK,
			expectedAmount: 0.00,
		},
		{
			name:  "decimal precision test",
			token: TokenTypeUSDC,
			apiResponse: EtherscanResponse{
				Status:  "1",
				Message: "OK",
				Result:  "123456", // 0.123456 USDC
			},
			statusCode:     http.StatusOK,
			expectedAmount: 0.123456,
		},
		{
			name:  "large balance",
			token: TokenTypeUSDC,
			apiResponse: EtherscanResponse{
				Status:  "1",
				Message: "OK",
				Result:  "1000000000000", // 1,000,000 USDC
			},
			statusCode:     http.StatusOK,
			expectedAmount: 1000000.00,
		},
		{
			name:          "unsupported token type",
			token:         TokenTypeBSV,
			expectedError: "etherscan provider only supports USDC",
		},
		{
			name:  "api error - invalid key",
			token: TokenTypeUSDC,
			apiResponse: EtherscanResponse{
				Status:  "0",
				Message: "NOTOK",
				Result:  "",
			},
			statusCode:    http.StatusOK,
			expectedError: "invalid API key or rate limit exceeded",
		},
		{
			name:  "api error - rate limit",
			token: TokenTypeUSDC,
			apiResponse: EtherscanResponse{
				Status:  "0",
				Message: "Max rate limit reached",
				Result:  "",
			},
			statusCode:    http.StatusOK,
			expectedError: "Max rate limit reached",
		},
		{
			name:          "http error - 500",
			token:         TokenTypeUSDC,
			statusCode:    http.StatusInternalServerError,
			expectedError: "etherscan API returned non-200 status: 500",
		},
		{
			name:          "http error - 403",
			token:         TokenTypeUSDC,
			statusCode:    http.StatusForbidden,
			expectedError: "etherscan API returned non-200 status: 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request parameters
				assert.Equal(t, "GET", r.Method)
				query := r.URL.Query()
				assert.Equal(t, "account", query.Get("module"))
				assert.Equal(t, "tokenbalance", query.Get("action"))
				assert.Equal(t, "latest", query.Get("tag"))

				if tt.testnet {
					assert.Equal(t, "11155111", query.Get("chainid"))
					assert.Equal(t, USDCSepoliaContract, query.Get("contractaddress"))
				} else {
					assert.Equal(t, "1", query.Get("chainid"))
					assert.Equal(t, USDCMainnetContract, query.Get("contractaddress"))
				}

				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_ = json.NewEncoder(w).Encode(tt.apiResponse)
				}
			}))
			defer server.Close()

			// Create provider with test server URL
			provider := NewEtherscanProvider("test-api-key", tt.testnet)
			provider.apiURL = server.URL

			// Execute
			ctx := context.Background()
			result, err := provider.GetBalance(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", tt.token)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.InDelta(t, tt.expectedAmount, result.Balance, 0.000001)
			assert.Equal(t, TokenTypeUSDC, result.Token)
			assert.Equal(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", result.Address)
			if tt.testnet {
				assert.Equal(t, "etherscan-sepolia", result.Provider)
			} else {
				assert.Equal(t, "etherscan", result.Provider)
			}
		})
	}
}

func TestEtherscanProvider_GetBalance_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(EtherscanResponse{
			Status:  "1",
			Message: "OK",
			Result:  "100000000",
		})
	}))
	defer server.Close()

	provider := NewEtherscanProvider("test-key", false)
	provider.apiURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := provider.GetBalance(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", TokenTypeUSDC)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestEtherscanProvider_GetTransactions(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	endTime := now.Add(24 * time.Hour)
	minAmount := 50.0

	tests := []struct {
		name            string
		query           TransactionQuery
		apiResponse     EtherscanTxListResponse
		statusCode      int
		expectedError   string
		expectedTxCount int
		validateFirstTx func(*testing.T, Transaction)
	}{
		{
			name: "successful transaction query",
			query: TransactionQuery{
				Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
				Token:   TokenTypeUSDC,
			},
			apiResponse: EtherscanTxListResponse{
				Status:  "1",
				Message: "OK",
				Result: []EtherscanTxResult{
					{
						BlockNumber: "12345678",
						TimeStamp:   "1640000000",
						Hash:        "0xabc123",
						From:        "0xsender",
						To:          "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
						Value:       "100000000", // 100 USDC
					},
				},
			},
			statusCode:      http.StatusOK,
			expectedTxCount: 1,
			validateFirstTx: func(t *testing.T, tx Transaction) {
				assert.Equal(t, "0xabc123", tx.Hash)
				assert.Equal(t, "0xsender", tx.From)
				assert.Equal(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", tx.To)
				assert.InDelta(t, 100.00, tx.Amount, 0.01)
				assert.Equal(t, TokenTypeUSDC, tx.Token)
				assert.Equal(t, int64(12345678), tx.BlockNumber)
				assert.True(t, tx.Confirmed)
			},
		},
		{
			name: "filter by minimum amount",
			query: TransactionQuery{
				Address:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
				Token:     TokenTypeUSDC,
				MinAmount: &minAmount,
			},
			apiResponse: EtherscanTxListResponse{
				Status:  "1",
				Message: "OK",
				Result: []EtherscanTxResult{
					{
						BlockNumber: "1",
						TimeStamp:   "1640000000",
						Hash:        "0x1",
						From:        "0xsender",
						To:          "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
						Value:       "25000000", // 25 USDC - filtered out
					},
					{
						BlockNumber: "2",
						TimeStamp:   "1640000100",
						Hash:        "0x2",
						From:        "0xsender",
						To:          "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
						Value:       "75000000", // 75 USDC - included
					},
				},
			},
			statusCode:      http.StatusOK,
			expectedTxCount: 1,
			validateFirstTx: func(t *testing.T, tx Transaction) {
				assert.Equal(t, "0x2", tx.Hash)
				assert.InDelta(t, 75.00, tx.Amount, 0.01)
			},
		},
		{
			name: "filter by time range",
			query: TransactionQuery{
				Address:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
				Token:     TokenTypeUSDC,
				StartTime: &startTime,
				EndTime:   &endTime,
			},
			apiResponse: EtherscanTxListResponse{
				Status:  "1",
				Message: "OK",
				Result: []EtherscanTxResult{
					{
						BlockNumber: "1",
						TimeStamp:   "1000000000", // Old - filtered out
						Hash:        "0x1",
						From:        "0xsender",
						To:          "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
						Value:       "50000000",
					},
					{
						BlockNumber: "2",
						TimeStamp:   fmt.Sprintf("%d", time.Now().Unix()),
						Hash:        "0x2",
						From:        "0xsender",
						To:          "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
						Value:       "75000000",
					},
				},
			},
			statusCode:      http.StatusOK,
			expectedTxCount: 1,
			validateFirstTx: func(t *testing.T, tx Transaction) {
				assert.Equal(t, "0x2", tx.Hash)
			},
		},
		{
			name: "filter outgoing transactions",
			query: TransactionQuery{
				Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
				Token:   TokenTypeUSDC,
			},
			apiResponse: EtherscanTxListResponse{
				Status:  "1",
				Message: "OK",
				Result: []EtherscanTxResult{
					{
						BlockNumber: "1",
						TimeStamp:   "1640000000",
						Hash:        "0x1",
						From:        "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
						To:          "0xrecipient", // Outgoing - filtered
						Value:       "50000000",
					},
					{
						BlockNumber: "2",
						TimeStamp:   "1640000100",
						Hash:        "0x2",
						From:        "0xsender",
						To:          "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", // Incoming
						Value:       "75000000",
					},
				},
			},
			statusCode:      http.StatusOK,
			expectedTxCount: 1,
			validateFirstTx: func(t *testing.T, tx Transaction) {
				assert.Equal(t, "0x2", tx.Hash)
			},
		},
		{
			name: "no transactions found",
			query: TransactionQuery{
				Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
				Token:   TokenTypeUSDC,
			},
			apiResponse: EtherscanTxListResponse{
				Status:  "0",
				Message: "No transactions found",
				Result:  []EtherscanTxResult{},
			},
			statusCode:      http.StatusOK,
			expectedTxCount: 0,
		},
		{
			name: "unsupported token",
			query: TransactionQuery{
				Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
				Token:   TokenTypeBSV,
			},
			expectedError: "etherscan provider only supports USDC",
		},
		{
			name: "api error",
			query: TransactionQuery{
				Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
				Token:   TokenTypeUSDC,
			},
			apiResponse: EtherscanTxListResponse{
				Status:  "0",
				Message: "NOTOK",
				Result:  []EtherscanTxResult{},
			},
			statusCode:    http.StatusOK,
			expectedError: "invalid API key or rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_ = json.NewEncoder(w).Encode(tt.apiResponse)
				}
			}))
			defer server.Close()

			// Create provider
			provider := NewEtherscanProvider("test-key", false)
			provider.apiURL = server.URL

			// Execute
			ctx := context.Background()
			result, err := provider.GetTransactions(ctx, tt.query)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result, tt.expectedTxCount)

			if tt.expectedTxCount > 0 && tt.validateFirstTx != nil {
				tt.validateFirstTx(t, result[0])
			}
		})
	}
}

func TestEtherscanProvider_GetTransactions_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider := NewEtherscanProvider("test-key", false)
	provider.apiURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	query := TransactionQuery{
		Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
		Token:   TokenTypeUSDC,
	}

	_, err := provider.GetTransactions(ctx, query)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestEtherscanProvider_APIKeyHandling(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		expectKeyParam bool
	}{
		{
			name:           "with api key",
			apiKey:         "my-api-key-123",
			expectKeyParam: true,
		},
		{
			name:           "without api key",
			apiKey:         "",
			expectKeyParam: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()
				if tt.expectKeyParam {
					assert.Equal(t, tt.apiKey, query.Get("apikey"))
				} else {
					assert.Empty(t, query.Get("apikey"))
				}

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(EtherscanResponse{
					Status:  "1",
					Message: "OK",
					Result:  "0",
				})
			}))
			defer server.Close()

			provider := NewEtherscanProvider(tt.apiKey, false)
			provider.apiURL = server.URL

			ctx := context.Background()
			_, _ = provider.GetBalance(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", TokenTypeUSDC)
		})
	}
}

func TestEtherscanProvider_NoAPIKeyErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(EtherscanResponse{
			Status:  "0",
			Message: "",
			Result:  "",
		})
	}))
	defer server.Close()

	provider := NewEtherscanProvider("", false) // No API key
	provider.apiURL = server.URL

	ctx := context.Background()
	_, err := provider.GetBalance(ctx, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1", TokenTypeUSDC)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing or invalid API key")
	assert.Contains(t, err.Error(), "ETHERSCAN_API_KEY")
}
