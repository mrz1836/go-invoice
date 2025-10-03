package blockchain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrUnknownScenario is returned when an unknown mock scenario is requested
	ErrUnknownScenario = errors.New("unknown scenario")
	// ErrNetworkTimeout is a mock network error for testing
	ErrNetworkTimeout = errors.New("network error: connection timeout")
)

// MockProvider is a mock blockchain provider for testing
// It allows configuring responses for different scenarios without requiring internet
type MockProvider struct {
	name             string
	supportedTokens  []TokenType
	balances         map[string]map[TokenType]float64 // address -> token -> balance
	transactions     map[string][]Transaction         // address -> transactions
	balanceError     error
	transactionError error
}

// NewMockProvider creates a new mock provider with default configuration
func NewMockProvider() *MockProvider {
	return &MockProvider{
		name:            "mock",
		supportedTokens: []TokenType{TokenTypeUSDC, TokenTypeBSV},
		balances:        make(map[string]map[TokenType]float64),
		transactions:    make(map[string][]Transaction),
	}
}

// SetBalance configures the balance to return for an address and token
func (m *MockProvider) SetBalance(address string, token TokenType, balance float64) {
	if m.balances[address] == nil {
		m.balances[address] = make(map[TokenType]float64)
	}
	m.balances[address][token] = balance
}

// AddTransaction adds a transaction to the mock provider's history
func (m *MockProvider) AddTransaction(address string, tx Transaction) {
	m.transactions[address] = append(m.transactions[address], tx)
}

// SetBalanceError configures an error to return from GetBalance
func (m *MockProvider) SetBalanceError(err error) {
	m.balanceError = err
}

// SetTransactionError configures an error to return from GetTransactions
func (m *MockProvider) SetTransactionError(err error) {
	m.transactionError = err
}

// GetBalance returns the configured balance for an address
func (m *MockProvider) GetBalance(ctx context.Context, address string, token TokenType) (*BalanceResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if m.balanceError != nil {
		return nil, m.balanceError
	}

	balance := 0.0
	if m.balances[address] != nil {
		balance = m.balances[address][token]
	}

	return &BalanceResult{
		Address:  address,
		Balance:  balance,
		Token:    token,
		AsOf:     time.Now(),
		Provider: m.name,
	}, nil
}

// GetTransactions returns configured transactions for an address
func (m *MockProvider) GetTransactions(ctx context.Context, query TransactionQuery) ([]Transaction, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if m.transactionError != nil {
		return nil, m.transactionError
	}

	txs := m.transactions[query.Address]
	if txs == nil {
		return []Transaction{}, nil
	}

	// Filter transactions based on query criteria
	filtered := make([]Transaction, 0, len(txs))
	for _, tx := range txs {
		// Filter by token type
		if tx.Token != query.Token {
			continue
		}

		// Filter by start time
		if query.StartTime != nil && tx.Timestamp.Before(*query.StartTime) {
			continue
		}

		// Filter by end time
		if query.EndTime != nil && tx.Timestamp.After(*query.EndTime) {
			continue
		}

		// Filter by minimum amount
		if query.MinAmount != nil && tx.Amount < *query.MinAmount {
			continue
		}

		filtered = append(filtered, tx)
	}

	return filtered, nil
}

// Name returns the provider name
func (m *MockProvider) Name() string {
	return m.name
}

// SupportedTokens returns the list of supported tokens
func (m *MockProvider) SupportedTokens() []TokenType {
	return m.supportedTokens
}

// Reset clears all configured balances and transactions
func (m *MockProvider) Reset() {
	m.balances = make(map[string]map[TokenType]float64)
	m.transactions = make(map[string][]Transaction)
	m.balanceError = nil
	m.transactionError = nil
}

// MockPaymentScenario configures the mock provider for common test scenarios
func (m *MockProvider) MockPaymentScenario(scenario, address string, amount float64, token TokenType) error {
	switch scenario {
	case "payment_found":
		// Balance equals the expected amount
		m.SetBalance(address, token, amount)
		m.AddTransaction(address, Transaction{
			Hash:        "0xmocktxhash123",
			From:        "0xsenderaddress",
			To:          address,
			Amount:      amount,
			Token:       token,
			BlockNumber: 12345678,
			Timestamp:   time.Now().Add(-1 * time.Hour),
			Confirmed:   true,
		})

	case "payment_not_found":
		// No balance, no transactions
		m.SetBalance(address, token, 0)

	case "partial_payment":
		// Balance is less than expected
		partialAmount := amount * 0.5
		m.SetBalance(address, token, partialAmount)
		m.AddTransaction(address, Transaction{
			Hash:        "0xmocktxhash456",
			From:        "0xsenderaddress",
			To:          address,
			Amount:      partialAmount,
			Token:       token,
			BlockNumber: 12345679,
			Timestamp:   time.Now().Add(-30 * time.Minute),
			Confirmed:   true,
		})

	case "overpayment":
		// Balance is more than expected
		overpayAmount := amount * 1.5
		m.SetBalance(address, token, overpayAmount)
		m.AddTransaction(address, Transaction{
			Hash:        "0xmocktxhash789",
			From:        "0xsenderaddress",
			To:          address,
			Amount:      overpayAmount,
			Token:       token,
			BlockNumber: 12345680,
			Timestamp:   time.Now().Add(-2 * time.Hour),
			Confirmed:   true,
		})

	case "network_error":
		m.SetBalanceError(ErrNetworkTimeout)
		m.SetTransactionError(ErrNetworkTimeout)

	default:
		return fmt.Errorf("%w: %s", ErrUnknownScenario, scenario)
	}

	return nil
}
