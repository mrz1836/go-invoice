package config

import "time"

// Config represents the complete application configuration
type Config struct {
	Business BusinessConfig `json:"business" validate:"required"`
	Invoice  InvoiceConfig  `json:"invoice" validate:"required"`
	Storage  StorageConfig  `json:"storage" validate:"required"`
}

// BusinessConfig contains business information for invoices
type BusinessConfig struct {
	Name           string         `json:"name" validate:"required"`
	Address        string         `json:"address" validate:"required"`
	Phone          string         `json:"phone,omitempty"`
	Email          string         `json:"email" validate:"required,email"`
	TaxID          string         `json:"tax_id,omitempty"`
	VATID          string         `json:"vat_id,omitempty"`
	Website        string         `json:"website,omitempty"`
	PaymentTerms   string         `json:"payment_terms" validate:"required"`
	BankDetails    BankDetails    `json:"bank_details,omitempty"`
	CryptoPayments CryptoPayments `json:"crypto_payments,omitempty"`
}

// BankDetails contains banking information for payments
type BankDetails struct {
	Name                string `json:"name,omitempty"`
	AccountNumber       string `json:"account_number,omitempty"`
	RoutingNumber       string `json:"routing_number,omitempty"`
	IBAN                string `json:"iban,omitempty"`
	SWIFT               string `json:"swift,omitempty"`
	PaymentInstructions string `json:"payment_instructions,omitempty"`
	ACHEnabled          bool   `json:"ach_enabled"`
}

// CryptoPayments contains cryptocurrency payment addresses
type CryptoPayments struct {
	USDCAddress     string `json:"usdc_address,omitempty"`
	USDCEnabled     bool   `json:"usdc_enabled"`
	BSVAddress      string `json:"bsv_address,omitempty"`
	BSVEnabled      bool   `json:"bsv_enabled"`
	EtherscanAPIKey string `json:"etherscan_api_key,omitempty"`
}

// InvoiceConfig contains invoice generation settings
type InvoiceConfig struct {
	Prefix         string  `json:"prefix" validate:"required"`
	StartNumber    int     `json:"start_number" validate:"min=1"`
	Footer         string  `json:"footer,omitempty"`
	Currency       string  `json:"currency" validate:"required"`
	VATRate        float64 `json:"vat_rate" validate:"min=0,max=1"`
	DefaultDueDays int     `json:"default_due_days" validate:"min=0"`
}

// StorageConfig contains storage location settings
type StorageConfig struct {
	DataDir        string        `json:"data_dir" validate:"required"`
	BackupDir      string        `json:"backup_dir,omitempty"`
	RetentionDays  int           `json:"retention_days" validate:"min=0"`
	AutoBackup     bool          `json:"auto_backup"`
	BackupInterval time.Duration `json:"backup_interval,omitempty"`
}

// LoadConfigRequest represents the configuration loading request.
type LoadConfigRequest struct {
	Path   string `json:"path" validate:"required"`
	Strict bool   `json:"strict"`
}

// ValidateConfigRequest represents the configuration validation request
type ValidateConfigRequest struct {
	Config *Config `json:"config" validate:"required"`
}
