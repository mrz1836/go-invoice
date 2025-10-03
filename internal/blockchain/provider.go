package blockchain

import (
	"context"
	"time"
)

// TokenType represents the type of cryptocurrency token
type TokenType string

const (
	// TokenTypeUSDC represents USDC stablecoin on Ethereum
	TokenTypeUSDC TokenType = "USDC"
	// TokenTypeBSV represents Bitcoin SV
	TokenTypeBSV TokenType = "BSV"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	Hash        string    // Transaction hash/ID
	From        string    // Sender address
	To          string    // Recipient address
	Amount      float64   // Amount in human-readable units (e.g., 100.50 USDC, not wei)
	Token       TokenType // Token type
	BlockNumber int64     // Block number
	Timestamp   time.Time // Transaction timestamp
	Confirmed   bool      // Whether transaction is confirmed
}

// BalanceResult represents the balance query result
type BalanceResult struct {
	Address  string    // Wallet address
	Balance  float64   // Current balance in human-readable units
	Token    TokenType // Token type
	AsOf     time.Time // Time of balance check
	Provider string    // Provider name (e.g., "etherscan", "whatsonchain")
}

// TransactionQuery represents parameters for querying transactions
type TransactionQuery struct {
	Address   string     // Wallet address to query
	Token     TokenType  // Token type to filter
	StartTime *time.Time // Optional: Only transactions after this time
	EndTime   *time.Time // Optional: Only transactions before this time
	MinAmount *float64   // Optional: Minimum transaction amount
}

// Provider defines the interface for blockchain data providers
// This abstraction allows for:
// - Offline testing with mock implementations
// - Multiple provider backends (Etherscan, Alchemy, Infura, etc.)
// - Different blockchain support (Ethereum, BSV, etc.)
type Provider interface {
	// GetBalance returns the current balance for an address
	GetBalance(ctx context.Context, address string, token TokenType) (*BalanceResult, error)

	// GetTransactions returns transactions for an address matching the query criteria
	// This is useful for verifying specific payments and getting transaction details
	GetTransactions(ctx context.Context, query TransactionQuery) ([]Transaction, error)

	// Name returns the provider name (e.g., "etherscan", "mock")
	Name() string

	// SupportedTokens returns the list of token types this provider supports
	SupportedTokens() []TokenType
}
