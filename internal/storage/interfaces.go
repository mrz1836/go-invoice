package storage

import (
	"context"

	"github.com/mrz/go-invoice/internal/models"
)

// InvoiceStorage defines the interface for invoice persistence operations
// This interface is consumer-driven and focuses on core CRUD operations
type InvoiceStorage interface {
	// CreateInvoice stores a new invoice
	// Returns ConflictError if invoice with same ID already exists
	CreateInvoice(ctx context.Context, invoice *models.Invoice) error

	// GetInvoice retrieves an invoice by ID
	// Returns NotFoundError if invoice doesn't exist
	GetInvoice(ctx context.Context, id models.InvoiceID) (*models.Invoice, error)

	// UpdateInvoice updates an existing invoice with optimistic locking
	// Returns NotFoundError if invoice doesn't exist
	// Returns VersionMismatchError if version doesn't match (optimistic locking)
	UpdateInvoice(ctx context.Context, invoice *models.Invoice) error

	// DeleteInvoice removes an invoice by ID
	// Returns NotFoundError if invoice doesn't exist
	DeleteInvoice(ctx context.Context, id models.InvoiceID) error

	// ListInvoices retrieves invoices based on filter criteria with pagination
	// Returns InvalidFilterError if filter parameters are invalid
	ListInvoices(ctx context.Context, filter models.InvoiceFilter) (*InvoiceListResult, error)

	// ExistsInvoice checks if an invoice exists without loading the full data
	ExistsInvoice(ctx context.Context, id models.InvoiceID) (bool, error)

	// CountInvoices returns the total count of invoices matching the filter
	CountInvoices(ctx context.Context, filter models.InvoiceFilter) (int64, error)
}

// ClientStorage defines the interface for client persistence operations
// Consumer-driven interface focusing on client management needs
type ClientStorage interface {
	// CreateClient stores a new client
	// Returns ConflictError if client with same ID already exists
	CreateClient(ctx context.Context, client *models.Client) error

	// GetClient retrieves a client by ID
	// Returns NotFoundError if client doesn't exist
	GetClient(ctx context.Context, id models.ClientID) (*models.Client, error)

	// UpdateClient updates an existing client
	// Returns NotFoundError if client doesn't exist
	UpdateClient(ctx context.Context, client *models.Client) error

	// DeleteClient removes a client by ID (soft delete - marks as inactive)
	// Returns NotFoundError if client doesn't exist
	DeleteClient(ctx context.Context, id models.ClientID) error

	// ListClients retrieves all active clients with pagination
	ListClients(ctx context.Context, activeOnly bool, limit, offset int) (*ClientListResult, error)

	// FindClientByEmail finds a client by email address
	// Returns NotFoundError if no client with that email exists
	FindClientByEmail(ctx context.Context, email string) (*models.Client, error)

	// ExistsClient checks if a client exists without loading the full data
	ExistsClient(ctx context.Context, id models.ClientID) (bool, error)
}

// StorageInitializer defines the interface for storage system initialization
// Consumer-driven interface for setup and configuration operations
type StorageInitializer interface {
	// Initialize sets up the storage system (creates directories, indexes, etc.)
	// Returns StorageUnavailableError if initialization fails
	Initialize(ctx context.Context) error

	// IsInitialized checks if the storage system is properly initialized
	IsInitialized(ctx context.Context) (bool, error)

	// GetStorageInfo returns information about the storage system
	GetStorageInfo(ctx context.Context) (*StorageInfo, error)

	// Validate performs integrity checks on the storage system
	// Returns errors if corruption or inconsistencies are found
	Validate(ctx context.Context) error
}

// StorageHealthMonitor defines the interface for monitoring storage health
// Consumer-driven interface for health checks and performance monitoring
type StorageHealthMonitor interface {
	// GetStats returns current storage statistics and health information
	GetStats(ctx context.Context) (*StorageStats, error)

	// HealthCheck performs a comprehensive health check
	// Returns error if storage is unhealthy
	HealthCheck(ctx context.Context) error

	// GetPerformanceMetrics returns recent performance metrics
	GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error)
}

// BackupManager defines the interface for backup and restore operations
// Consumer-driven interface for data protection operations
type BackupManager interface {
	// CreateBackup creates a backup with the specified options
	// Returns path to the created backup file
	CreateBackup(ctx context.Context, options BackupOptions) (string, error)

	// RestoreBackup restores data from a backup file
	// Returns summary of restored data
	RestoreBackup(ctx context.Context, options RestoreOptions) (*RestoreResult, error)

	// ListBackups returns available backup files
	ListBackups(ctx context.Context) ([]*BackupInfo, error)

	// DeleteBackup removes a backup file
	DeleteBackup(ctx context.Context, backupPath string) error

	// ValidateBackup checks if a backup file is valid and complete
	ValidateBackup(ctx context.Context, backupPath string) error
}

// TransactionManager defines the interface for transaction support
// Consumer-driven interface for atomic operations across multiple resources
type TransactionManager interface {
	// BeginTransaction starts a new transaction
	BeginTransaction(ctx context.Context) (Transaction, error)
}

// Transaction defines the interface for individual transactions
// Provides atomic operations with rollback capability
type Transaction interface {
	// Commit commits all changes made within the transaction
	Commit(ctx context.Context) error

	// Rollback rolls back all changes made within the transaction
	Rollback(ctx context.Context) error

	// InvoiceStorage returns the invoice storage within this transaction
	InvoiceStorage() InvoiceStorage

	// ClientStorage returns the client storage within this transaction
	ClientStorage() ClientStorage
}

// StorageInfo represents information about the storage system
type StorageInfo struct {
	Type             string `json:"type"`              // "json", "sqlite", etc.
	Version          string `json:"version"`           // Storage format version
	Path             string `json:"path"`              // Storage location
	Initialized      bool   `json:"initialized"`       // Whether storage is initialized
	ReadOnly         bool   `json:"read_only"`         // Whether storage is read-only
	SupportsBackups  bool   `json:"supports_backups"`  // Whether backups are supported
	SupportsIndexing bool   `json:"supports_indexing"` // Whether indexing is supported
}

// BackupInfo represents information about a backup
type BackupInfo struct {
	Path         string `json:"path"`
	CreatedAt    string `json:"created_at"`
	SizeBytes    int64  `json:"size_bytes"`
	InvoiceCount int64  `json:"invoice_count"`
	ClientCount  int64  `json:"client_count"`
	Compressed   bool   `json:"compressed"`
	Encrypted    bool   `json:"encrypted"`
}

// RestoreResult represents the result of a restore operation
type RestoreResult struct {
	InvoicesRestored  int64    `json:"invoices_restored"`
	ClientsRestored   int64    `json:"clients_restored"`
	ErrorsEncountered int64    `json:"errors_encountered"`
	Warnings          []string `json:"warnings,omitempty"`
	DryRun            bool     `json:"dry_run"`
}

// SearchOptions represents options for search operations
type SearchOptions struct {
	Query      string   `json:"query"`       // Text to search for
	Fields     []string `json:"fields"`      // Fields to search in
	Limit      int      `json:"limit"`       // Maximum results to return
	Offset     int      `json:"offset"`      // Number of results to skip
	SortBy     string   `json:"sort_by"`     // Field to sort by
	SortOrder  string   `json:"sort_order"`  // "asc" or "desc"
	Fuzzy      bool     `json:"fuzzy"`       // Enable fuzzy matching
	MatchScore float64  `json:"match_score"` // Minimum match score for fuzzy search
}

// SearchResult represents a search result with relevance scoring
type SearchResult struct {
	ID           string                 `json:"id"`
	ResourceType string                 `json:"resource_type"` // "invoice" or "client"
	Score        float64                `json:"score"`         // Relevance score
	Matches      map[string]interface{} `json:"matches"`       // Matching field highlights
	Resource     interface{}            `json:"resource"`      // The actual resource (Invoice or Client)
}

// SearchResults represents the complete search results
type SearchResults struct {
	Results         []*SearchResult `json:"results"`
	TotalCount      int64           `json:"total_count"`
	Query           string          `json:"query"`
	ExecutionTimeMs float64         `json:"execution_time_ms"`
}

// IndexManager defines the interface for search indexing operations
// Consumer-driven interface for search functionality
type IndexManager interface {
	// RebuildIndex rebuilds the search index from scratch
	RebuildIndex(ctx context.Context) error

	// UpdateIndex updates the index for specific resources
	UpdateIndex(ctx context.Context, resourceType string, resources []interface{}) error

	// Search performs a search across indexed resources
	Search(ctx context.Context, options SearchOptions) (*SearchResults, error)

	// GetIndexStats returns statistics about the search index
	GetIndexStats(ctx context.Context) (*IndexStats, error)
}

// IndexStats represents statistics about the search index
type IndexStats struct {
	IndexedInvoices    int64   `json:"indexed_invoices"`
	IndexedClients     int64   `json:"indexed_clients"`
	IndexSizeBytes     int64   `json:"index_size_bytes"`
	LastUpdated        string  `json:"last_updated"`
	AverageQueryTimeMs float64 `json:"average_query_time_ms"`
}
