package blockchain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMockProvider(t *testing.T) {
	provider := NewMockProvider()

	assert.NotNil(t, provider)
	assert.Equal(t, "mock", provider.Name())
	assert.Equal(t, []TokenType{TokenTypeUSDC, TokenTypeBSV}, provider.SupportedTokens())
	assert.NotNil(t, provider.balances)
	assert.NotNil(t, provider.transactions)
}

const testAddress = "0xtest123"

func TestMockProvider_SetBalance(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress

	// Set USDC balance
	provider.SetBalance(address, TokenTypeUSDC, 100.50)

	// Verify balance was set
	ctx := context.Background()
	result, err := provider.GetBalance(ctx, address, TokenTypeUSDC)

	require.NoError(t, err)
	assert.InDelta(t, 100.50, result.Balance, 0.01)
	assert.Equal(t, address, result.Address)
	assert.Equal(t, TokenTypeUSDC, result.Token)
	assert.Equal(t, "mock", result.Provider)
}

func TestMockProvider_SetBalance_MultipleTokens(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress

	// Set balances for different tokens
	provider.SetBalance(address, TokenTypeUSDC, 100.00)
	provider.SetBalance(address, TokenTypeBSV, 0.5)

	ctx := context.Background()

	// Check USDC balance
	usdcResult, err := provider.GetBalance(ctx, address, TokenTypeUSDC)
	require.NoError(t, err)
	assert.InDelta(t, 100.00, usdcResult.Balance, 0.01)

	// Check BSV balance
	bsvResult, err := provider.GetBalance(ctx, address, TokenTypeBSV)
	require.NoError(t, err)
	assert.InDelta(t, 0.5, bsvResult.Balance, 0.01)
}

func TestMockProvider_GetBalance_DefaultZero(t *testing.T) {
	provider := NewMockProvider()
	ctx := context.Background()

	// Query balance for address with no configured balance
	result, err := provider.GetBalance(ctx, "0xunknown", TokenTypeUSDC)

	require.NoError(t, err)
	assert.InDelta(t, 0.0, result.Balance, 0.01)
}

func TestMockProvider_GetBalance_WithError(t *testing.T) {
	provider := NewMockProvider()
	provider.SetBalanceError(ErrNetworkTimeout)

	ctx := context.Background()
	_, err := provider.GetBalance(ctx, "0xtest", TokenTypeUSDC)

	require.Error(t, err)
	assert.Equal(t, ErrNetworkTimeout, err)
}

func TestMockProvider_GetBalance_ContextCancellation(t *testing.T) {
	provider := NewMockProvider()
	provider.SetBalance("0xtest", TokenTypeUSDC, 100.00)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := provider.GetBalance(ctx, "0xtest", TokenTypeUSDC)

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestMockProvider_AddTransaction(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress

	tx := Transaction{
		Hash:        "0xabc123",
		From:        "0xsender",
		To:          address,
		Amount:      100.00,
		Token:       TokenTypeUSDC,
		BlockNumber: 12345,
		Timestamp:   time.Now(),
		Confirmed:   true,
	}

	provider.AddTransaction(address, tx)

	// Retrieve transactions
	ctx := context.Background()
	query := TransactionQuery{
		Address: address,
		Token:   TokenTypeUSDC,
	}

	result, err := provider.GetTransactions(ctx, query)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, tx.Hash, result[0].Hash)
	assert.InDelta(t, tx.Amount, result[0].Amount, 0.01)
	assert.Equal(t, tx.From, result[0].From)
	assert.Equal(t, tx.To, result[0].To)
}

func TestMockProvider_GetTransactions_FilterByToken(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress

	// Add transactions for different tokens
	provider.AddTransaction(address, Transaction{
		Hash:   "0x1",
		To:     address,
		Amount: 100.00,
		Token:  TokenTypeUSDC,
	})
	provider.AddTransaction(address, Transaction{
		Hash:   "0x2",
		To:     address,
		Amount: 0.5,
		Token:  TokenTypeBSV,
	})

	ctx := context.Background()

	// Query USDC transactions
	usdcQuery := TransactionQuery{
		Address: address,
		Token:   TokenTypeUSDC,
	}
	usdcTxs, err := provider.GetTransactions(ctx, usdcQuery)
	require.NoError(t, err)
	assert.Len(t, usdcTxs, 1)
	assert.Equal(t, "0x1", usdcTxs[0].Hash)

	// Query BSV transactions
	bsvQuery := TransactionQuery{
		Address: address,
		Token:   TokenTypeBSV,
	}
	bsvTxs, err := provider.GetTransactions(ctx, bsvQuery)
	require.NoError(t, err)
	assert.Len(t, bsvTxs, 1)
	assert.Equal(t, "0x2", bsvTxs[0].Hash)
}

func TestMockProvider_GetTransactions_FilterByTime(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress
	now := time.Now()

	// Add transactions at different times
	provider.AddTransaction(address, Transaction{
		Hash:      "0x1",
		To:        address,
		Amount:    50.00,
		Token:     TokenTypeUSDC,
		Timestamp: now.Add(-2 * time.Hour),
	})
	provider.AddTransaction(address, Transaction{
		Hash:      "0x2",
		To:        address,
		Amount:    100.00,
		Token:     TokenTypeUSDC,
		Timestamp: now.Add(-1 * time.Hour),
	})
	provider.AddTransaction(address, Transaction{
		Hash:      "0x3",
		To:        address,
		Amount:    75.00,
		Token:     TokenTypeUSDC,
		Timestamp: now.Add(1 * time.Hour),
	})

	ctx := context.Background()

	// Filter by start time
	startTime := now.Add(-90 * time.Minute)
	query := TransactionQuery{
		Address:   address,
		Token:     TokenTypeUSDC,
		StartTime: &startTime,
	}
	result, err := provider.GetTransactions(ctx, query)
	require.NoError(t, err)
	assert.Len(t, result, 2) // Should get 0x2 and 0x3

	// Filter by end time
	endTime := now
	query = TransactionQuery{
		Address: address,
		Token:   TokenTypeUSDC,
		EndTime: &endTime,
	}
	result, err = provider.GetTransactions(ctx, query)
	require.NoError(t, err)
	assert.Len(t, result, 2) // Should get 0x1 and 0x2

	// Filter by both
	query = TransactionQuery{
		Address:   address,
		Token:     TokenTypeUSDC,
		StartTime: &startTime,
		EndTime:   &endTime,
	}
	result, err = provider.GetTransactions(ctx, query)
	require.NoError(t, err)
	assert.Len(t, result, 1) // Should only get 0x2
	assert.Equal(t, "0x2", result[0].Hash)
}

func TestMockProvider_GetTransactions_FilterByMinAmount(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress

	// Add transactions with different amounts
	provider.AddTransaction(address, Transaction{
		Hash:   "0x1",
		To:     address,
		Amount: 25.00,
		Token:  TokenTypeUSDC,
	})
	provider.AddTransaction(address, Transaction{
		Hash:   "0x2",
		To:     address,
		Amount: 75.00,
		Token:  TokenTypeUSDC,
	})
	provider.AddTransaction(address, Transaction{
		Hash:   "0x3",
		To:     address,
		Amount: 150.00,
		Token:  TokenTypeUSDC,
	})

	ctx := context.Background()

	// Filter by minimum amount
	minAmount := 50.0
	query := TransactionQuery{
		Address:   address,
		Token:     TokenTypeUSDC,
		MinAmount: &minAmount,
	}

	result, err := provider.GetTransactions(ctx, query)
	require.NoError(t, err)
	assert.Len(t, result, 2) // Should get 0x2 and 0x3

	for _, tx := range result {
		assert.GreaterOrEqual(t, tx.Amount, minAmount)
	}
}

func TestMockProvider_GetTransactions_NoMatch(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress

	// Add a transaction
	provider.AddTransaction(address, Transaction{
		Hash:   "0x1",
		To:     address,
		Amount: 100.00,
		Token:  TokenTypeUSDC,
	})

	ctx := context.Background()

	// Query for different address
	query := TransactionQuery{
		Address: "0xdifferent",
		Token:   TokenTypeUSDC,
	}

	result, err := provider.GetTransactions(ctx, query)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestMockProvider_GetTransactions_WithError(t *testing.T) {
	provider := NewMockProvider()
	provider.SetTransactionError(ErrNetworkTimeout)

	ctx := context.Background()
	query := TransactionQuery{
		Address: "0xtest",
		Token:   TokenTypeUSDC,
	}

	_, err := provider.GetTransactions(ctx, query)
	require.Error(t, err)
	assert.Equal(t, ErrNetworkTimeout, err)
}

func TestMockProvider_GetTransactions_ContextCancellation(t *testing.T) {
	provider := NewMockProvider()
	provider.AddTransaction("0xtest", Transaction{
		Hash:   "0x1",
		To:     "0xtest",
		Amount: 100.00,
		Token:  TokenTypeUSDC,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	query := TransactionQuery{
		Address: "0xtest",
		Token:   TokenTypeUSDC,
	}

	_, err := provider.GetTransactions(ctx, query)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestMockProvider_Reset(t *testing.T) {
	provider := NewMockProvider()

	// Configure some data
	provider.SetBalance("0xtest", TokenTypeUSDC, 100.00)
	provider.AddTransaction("0xtest", Transaction{
		Hash:   "0x1",
		To:     "0xtest",
		Amount: 100.00,
		Token:  TokenTypeUSDC,
	})
	provider.SetBalanceError(ErrNetworkTimeout)
	provider.SetTransactionError(ErrNetworkTimeout)

	// Reset
	provider.Reset()

	ctx := context.Background()

	// Verify balance is cleared
	result, err := provider.GetBalance(ctx, "0xtest", TokenTypeUSDC)
	require.NoError(t, err)
	assert.InDelta(t, 0.0, result.Balance, 0.01)

	// Verify transactions are cleared
	query := TransactionQuery{
		Address: "0xtest",
		Token:   TokenTypeUSDC,
	}
	txs, err := provider.GetTransactions(ctx, query)
	require.NoError(t, err)
	assert.Empty(t, txs)
}

func TestMockProvider_MockPaymentScenario_PaymentFound(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress
	amount := 100.00

	err := provider.MockPaymentScenario("payment_found", address, amount, TokenTypeUSDC)
	require.NoError(t, err)

	ctx := context.Background()

	// Check balance
	balance, err := provider.GetBalance(ctx, address, TokenTypeUSDC)
	require.NoError(t, err)
	assert.InDelta(t, amount, balance.Balance, 0.01)

	// Check transaction
	query := TransactionQuery{
		Address: address,
		Token:   TokenTypeUSDC,
	}
	txs, err := provider.GetTransactions(ctx, query)
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, "0xmocktxhash123", txs[0].Hash)
	assert.InDelta(t, amount, txs[0].Amount, 0.01)
	assert.True(t, txs[0].Confirmed)
}

func TestMockProvider_MockPaymentScenario_PaymentNotFound(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress

	err := provider.MockPaymentScenario("payment_not_found", address, 100.00, TokenTypeUSDC)
	require.NoError(t, err)

	ctx := context.Background()

	// Check balance is zero
	balance, err := provider.GetBalance(ctx, address, TokenTypeUSDC)
	require.NoError(t, err)
	assert.InDelta(t, 0.0, balance.Balance, 0.01)

	// No transactions
	query := TransactionQuery{
		Address: address,
		Token:   TokenTypeUSDC,
	}
	txs, err := provider.GetTransactions(ctx, query)
	require.NoError(t, err)
	assert.Empty(t, txs)
}

func TestMockProvider_MockPaymentScenario_PartialPayment(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress
	amount := 100.00

	err := provider.MockPaymentScenario("partial_payment", address, amount, TokenTypeUSDC)
	require.NoError(t, err)

	ctx := context.Background()

	// Check balance is 50% of expected
	balance, err := provider.GetBalance(ctx, address, TokenTypeUSDC)
	require.NoError(t, err)
	assert.InDelta(t, 50.00, balance.Balance, 0.01)

	// Check transaction amount
	query := TransactionQuery{
		Address: address,
		Token:   TokenTypeUSDC,
	}
	txs, err := provider.GetTransactions(ctx, query)
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.InDelta(t, 50.00, txs[0].Amount, 0.01)
}

func TestMockProvider_MockPaymentScenario_Overpayment(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress
	amount := 100.00

	err := provider.MockPaymentScenario("overpayment", address, amount, TokenTypeUSDC)
	require.NoError(t, err)

	ctx := context.Background()

	// Check balance is 150% of expected
	balance, err := provider.GetBalance(ctx, address, TokenTypeUSDC)
	require.NoError(t, err)
	assert.InDelta(t, 150.00, balance.Balance, 0.01)

	// Check transaction amount
	query := TransactionQuery{
		Address: address,
		Token:   TokenTypeUSDC,
	}
	txs, err := provider.GetTransactions(ctx, query)
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.InDelta(t, 150.00, txs[0].Amount, 0.01)
}

func TestMockProvider_MockPaymentScenario_NetworkError(t *testing.T) {
	provider := NewMockProvider()
	address := testAddress

	err := provider.MockPaymentScenario("network_error", address, 100.00, TokenTypeUSDC)
	require.NoError(t, err)

	ctx := context.Background()

	// Balance query should fail
	_, err = provider.GetBalance(ctx, address, TokenTypeUSDC)
	require.Error(t, err)
	assert.Equal(t, ErrNetworkTimeout, err)

	// Transaction query should fail
	query := TransactionQuery{
		Address: address,
		Token:   TokenTypeUSDC,
	}
	_, err = provider.GetTransactions(ctx, query)
	require.Error(t, err)
	assert.Equal(t, ErrNetworkTimeout, err)
}

func TestMockProvider_MockPaymentScenario_UnknownScenario(t *testing.T) {
	provider := NewMockProvider()

	err := provider.MockPaymentScenario("unknown_scenario", "0xtest", 100.00, TokenTypeUSDC)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownScenario)
}

func TestMockProvider_MockPaymentScenario_BSVToken(t *testing.T) {
	provider := NewMockProvider()
	address := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
	amount := 0.5

	err := provider.MockPaymentScenario("payment_found", address, amount, TokenTypeBSV)
	require.NoError(t, err)

	ctx := context.Background()

	// Check BSV balance
	balance, err := provider.GetBalance(ctx, address, TokenTypeBSV)
	require.NoError(t, err)
	assert.InDelta(t, amount, balance.Balance, 0.01)

	// Check BSV transaction
	query := TransactionQuery{
		Address: address,
		Token:   TokenTypeBSV,
	}
	txs, err := provider.GetTransactions(ctx, query)
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, TokenTypeBSV, txs[0].Token)
}
