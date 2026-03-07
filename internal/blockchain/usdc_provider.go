package blockchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	// ErrUnsupportedToken is returned when an unsupported token is requested
	ErrUnsupportedToken = errors.New("etherscan provider only supports USDC")
	// ErrEtherscanAPIStatus is returned when Etherscan API returns non-200 status
	ErrEtherscanAPIStatus = errors.New("etherscan API returned non-200 status")
	// ErrEtherscanAPIError is returned when Etherscan API returns an error
	ErrEtherscanAPIError = errors.New("etherscan API error")
)

const (
	// EtherscanV2URL is the Etherscan V2 API endpoint (supports multichain via chainid parameter)
	EtherscanV2URL = "https://api.etherscan.io/v2/api"

	// EtherscanMainnetChainID is the chain ID for Ethereum mainnet
	EtherscanMainnetChainID = 1
	// EtherscanSepoliaChainID is the chain ID for Sepolia testnet
	EtherscanSepoliaChainID = 11155111

	// USDCMainnetContract is the USDC token contract address on Ethereum mainnet
	USDCMainnetContract = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
	// USDCSepoliaContract is the USDC token contract address on Sepolia testnet
	USDCSepoliaContract = "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238"

	// USDCDecimals is the number of decimal places for USDC (6 decimals, not 18 like most ERC20)
	USDCDecimals = 6

	// DefaultRateLimit is the default rate limit for free Etherscan API (5 calls/sec)
	DefaultRateLimit = 5 * time.Second / 5
)

// EtherscanProvider implements the Provider interface for USDC on Ethereum via Etherscan API
type EtherscanProvider struct {
	apiURL       string
	apiKey       string
	chainID      int
	usdcContract string
	testnet      bool
	httpClient   *http.Client
	rateLimit    time.Duration
}

// NewEtherscanProvider creates a new Etherscan provider for USDC
func NewEtherscanProvider(apiKey string, testnet bool) *EtherscanProvider {
	chainID := EtherscanMainnetChainID
	usdcContract := USDCMainnetContract

	if testnet {
		chainID = EtherscanSepoliaChainID
		usdcContract = USDCSepoliaContract
	}

	return &EtherscanProvider{
		apiURL:       EtherscanV2URL,
		apiKey:       apiKey,
		chainID:      chainID,
		usdcContract: usdcContract,
		testnet:      testnet,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimit: DefaultRateLimit,
	}
}

// GetBalance returns the USDC balance for an Ethereum address
func (e *EtherscanProvider) GetBalance(ctx context.Context, address string, token TokenType) (*BalanceResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if token != TokenTypeUSDC {
		return nil, fmt.Errorf("%w, got %s", ErrUnsupportedToken, token)
	}

	// Build API request for V2 API
	params := url.Values{}
	params.Set("chainid", strconv.Itoa(e.chainID))
	params.Set("module", "account")
	params.Set("action", "tokenbalance")
	params.Set("contractaddress", e.usdcContract)
	params.Set("address", address)
	params.Set("tag", "latest")
	if e.apiKey != "" {
		params.Set("apikey", e.apiKey)
	}

	reqURL := fmt.Sprintf("%s?%s", e.apiURL, params.Encode())

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch balance: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrEtherscanAPIStatus, resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp EtherscanResponse
	if unmarshalErr := json.Unmarshal(body, &apiResp); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse response: %w", unmarshalErr)
	}

	if apiResp.Status != "1" {
		// Provide helpful error message based on common API errors
		errMsg := apiResp.Message
		if errMsg == "NOTOK" || errMsg == "" {
			if e.apiKey == "" {
				return nil, fmt.Errorf("%w: missing or invalid API key (consider setting ETHERSCAN_API_KEY in config or using --etherscan-api-key flag)", ErrEtherscanAPIError)
			}
			return nil, fmt.Errorf("%w: invalid API key or rate limit exceeded", ErrEtherscanAPIError)
		}
		return nil, fmt.Errorf("%w: %s", ErrEtherscanAPIError, errMsg)
	}

	// Convert balance from smallest units (6 decimals for USDC)
	balanceWei, err := strconv.ParseInt(apiResp.Result, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}

	balance := float64(balanceWei) / 1e6 // USDC has 6 decimals

	return &BalanceResult{
		Address:  address,
		Balance:  balance,
		Token:    TokenTypeUSDC,
		AsOf:     time.Now(),
		Provider: e.Name(),
	}, nil
}

// GetTransactions returns USDC transactions for an address
func (e *EtherscanProvider) GetTransactions(ctx context.Context, query TransactionQuery) ([]Transaction, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if query.Token != TokenTypeUSDC {
		return nil, fmt.Errorf("%w, got %s", ErrUnsupportedToken, query.Token)
	}

	// Build API request for token transfers (V2 API)
	params := url.Values{}
	params.Set("chainid", strconv.Itoa(e.chainID))
	params.Set("module", "account")
	params.Set("action", "tokentx")
	params.Set("contractaddress", e.usdcContract)
	params.Set("address", query.Address)
	params.Set("sort", "asc")

	// Add optional time filters using block numbers
	// Etherscan doesn't support direct timestamp filtering,
	// so we get all transactions and filter in code
	if query.StartTime != nil {
		// Could convert timestamp to approximate block number, but for simplicity
		// we'll filter after fetching
		params.Set("startblock", "0")
	}

	if e.apiKey != "" {
		params.Set("apikey", e.apiKey)
	}

	reqURL := fmt.Sprintf("%s?%s", e.apiURL, params.Encode())

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrEtherscanAPIStatus, resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp EtherscanTxListResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Status != "1" {
		// Status "0" with message "No transactions found" is not an error
		if apiResp.Message == "No transactions found" {
			return []Transaction{}, nil
		}
		// Provide helpful error message based on common API errors
		errMsg := apiResp.Message
		if errMsg == "NOTOK" || errMsg == "" {
			if e.apiKey == "" {
				return nil, fmt.Errorf("%w: missing or invalid API key (consider setting ETHERSCAN_API_KEY in config or using --etherscan-api-key flag)", ErrEtherscanAPIError)
			}
			return nil, fmt.Errorf("%w: invalid API key or rate limit exceeded", ErrEtherscanAPIError)
		}
		return nil, fmt.Errorf("%w: %s", ErrEtherscanAPIError, errMsg)
	}

	// Convert and filter transactions
	transactions := make([]Transaction, 0, len(apiResp.Result))
	for _, tx := range apiResp.Result {
		// Only include incoming transactions (to the query address)
		if tx.To != query.Address {
			continue
		}

		// Parse timestamp
		timestampInt, err := strconv.ParseInt(tx.TimeStamp, 10, 64)
		if err != nil {
			continue
		}
		timestamp := time.Unix(timestampInt, 0)

		// Filter by start time
		if query.StartTime != nil && timestamp.Before(*query.StartTime) {
			continue
		}

		// Filter by end time
		if query.EndTime != nil && timestamp.After(*query.EndTime) {
			continue
		}

		// Parse amount (USDC has 6 decimals)
		amountWei, err := strconv.ParseInt(tx.Value, 10, 64)
		if err != nil {
			continue
		}
		amount := float64(amountWei) / 1e6

		// Filter by minimum amount
		if query.MinAmount != nil && amount < *query.MinAmount {
			continue
		}

		// Parse block number
		blockNumber, _ := strconv.ParseInt(tx.BlockNumber, 10, 64)

		transactions = append(transactions, Transaction{
			Hash:        tx.Hash,
			From:        tx.From,
			To:          tx.To,
			Amount:      amount,
			Token:       TokenTypeUSDC,
			BlockNumber: blockNumber,
			Timestamp:   timestamp,
			Confirmed:   true, // Etherscan only returns confirmed transactions
		})
	}

	return transactions, nil
}

// Name returns the provider name
func (e *EtherscanProvider) Name() string {
	if e.testnet {
		return "etherscan-sepolia"
	}
	return "etherscan"
}

// SupportedTokens returns the list of supported tokens
func (e *EtherscanProvider) SupportedTokens() []TokenType {
	return []TokenType{TokenTypeUSDC}
}

// EtherscanResponse represents the Etherscan API response for balance queries
type EtherscanResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

// EtherscanTxListResponse represents the Etherscan API response for transaction lists
type EtherscanTxListResponse struct {
	Status  string              `json:"status"`
	Message string              `json:"message"`
	Result  []EtherscanTxResult `json:"result"`
}

// EtherscanTxResult represents a single transaction in the Etherscan response
type EtherscanTxResult struct {
	BlockNumber string `json:"blockNumber"`
	TimeStamp   string `json:"timeStamp"`
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
}
