package blockchain

import (
	"context"
	"errors"
)

// ErrBSVNotImplemented is returned when BSV provider is called but not yet implemented
var ErrBSVNotImplemented = errors.New("BSV provider not yet implemented - see TODOs in bsv_provider.go for implementation guidance")

// BSVProvider implements the Provider interface for Bitcoin SV
// TODO: Implement full BSV support with the following considerations:
//
// BSV Payment Verification Challenges:
// 1. Amount Conversion: BSV amounts are in satoshis, not USD
//   - Need real-time exchange rate (BSV/USD)
//   - Exchange rate can fluctuate between invoice creation and payment
//   - Solution: Use fuzzy matching with ±5% tolerance
//
// 2. Time Window: Payment might not happen exactly on invoice date
//   - Check transactions within ±7 days of invoice date
//   - Or within invoice creation to 30 days after due date
//
// 3. API Provider Options:
//   - WhatsOnChain API (recommended, free tier available)
//   - Blockchair API
//   - Custom BSV node (requires infrastructure)
//
// 4. Transaction Verification:
//   - Verify transaction has sufficient confirmations (6+ recommended)
//   - Check transaction is to the correct address
//   - Convert satoshi amount to USD using historical rate at tx time
//
// Implementation Example:
//
//	type BSVProvider struct {
//		apiURL         string
//		testnet        bool
//		httpClient     *http.Client
//		exchangeRateAPI string // e.g., CoinGecko, CoinMarketCap
//	}
//
//	func (b *BSVProvider) GetBalance(ctx, address, token) {
//		// 1. Get address balance in satoshis from WhatsOnChain
//		// 2. Get current BSV/USD exchange rate
//		// 3. Convert satoshis to USD
//		// 4. Return BalanceResult
//	}
//
//	func (b *BSVProvider) GetTransactions(ctx, query) {
//		// 1. Get transactions from WhatsOnChain
//		// 2. Filter by time window (invoice date ±7 days)
//		// 3. For each transaction:
//		//    - Get historical exchange rate at tx timestamp
//		//    - Convert satoshis to USD
//		//    - Check if amount matches invoice (±5% tolerance)
//		// 4. Return matching transactions
//	}
//
// Reference APIs:
// - WhatsOnChain: https://developers.whatsonchain.com/
// - Exchange Rates: https://api.coingecko.com/api/v3/simple/price?ids=bitcoin-sv&vs_currencies=usd
type BSVProvider struct {
	apiURL  string
	testnet bool
}

// NewBSVProvider creates a new BSV provider (STUB - Not implemented)
func NewBSVProvider(testnet bool) *BSVProvider {
	return &BSVProvider{
		apiURL:  "https://api.whatsonchain.com/v1/bsv/main", // mainnet
		testnet: testnet,
	}
}

// GetBalance returns the BSV balance for an address (STUB - Not implemented)
func (b *BSVProvider) GetBalance(ctx context.Context, address string, token TokenType) (*BalanceResult, error) {
	return nil, ErrBSVNotImplemented
}

// GetTransactions returns BSV transactions for an address (STUB - Not implemented)
func (b *BSVProvider) GetTransactions(ctx context.Context, query TransactionQuery) ([]Transaction, error) {
	return nil, ErrBSVNotImplemented
}

// Name returns the provider name
func (b *BSVProvider) Name() string {
	if b.testnet {
		return "whatsonchain-testnet"
	}
	return "whatsonchain"
}

// SupportedTokens returns the list of supported tokens
func (b *BSVProvider) SupportedTokens() []TokenType {
	return []TokenType{TokenTypeBSV}
}

// TODO: Implement helper functions for BSV provider:
//
// func (b *BSVProvider) getExchangeRate(ctx context.Context, timestamp time.Time) (float64, error)
// - Get BSV/USD exchange rate at specific timestamp
// - Use CoinGecko or similar API
// - Cache rates to avoid excessive API calls
//
// func (b *BSVProvider) satoshiToUSD(satoshis int64, exchangeRate float64) float64
// - Convert satoshis to USD
// - 1 BSV = 100,000,000 satoshis
//
// func (b *BSVProvider) matchesInvoiceAmount(txAmountUSD, invoiceAmount float64) bool
// - Check if transaction amount matches invoice with tolerance
// - Use ±5% tolerance for exchange rate fluctuations
//
// func (b *BSVProvider) isWithinTimeWindow(txTime, invoiceDate time.Time) bool
// - Check if transaction is within acceptable time window
// - Invoice creation to 30 days after due date
