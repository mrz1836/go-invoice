package storage

import (
	"fmt"

	"github.com/mrz/go-invoice/internal/models"
)

// Storage error types for comprehensive error handling

// ErrNotFound indicates that a requested resource was not found
type ErrNotFound struct {
	Resource string
	ID       string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s with ID '%s' not found", e.Resource, e.ID)
}

// NewErrNotFound creates a new not found error
func NewErrNotFound(resource, id string) ErrNotFound {
	return ErrNotFound{Resource: resource, ID: id}
}

// ErrConflict indicates that a resource already exists or conflicts with another
type ErrConflict struct {
	Resource string
	ID       string
	Message  string
}

func (e ErrConflict) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s with ID '%s' conflicts: %s", e.Resource, e.ID, e.Message)
	}
	return fmt.Sprintf("%s with ID '%s' already exists", e.Resource, e.ID)
}

// NewErrConflict creates a new conflict error
func NewErrConflict(resource, id, message string) ErrConflict {
	return ErrConflict{Resource: resource, ID: id, Message: message}
}

// ErrVersionMismatch indicates an optimistic locking failure
type ErrVersionMismatch struct {
	Resource        string
	ID              string
	ExpectedVersion int
	ActualVersion   int
}

func (e ErrVersionMismatch) Error() string {
	return fmt.Sprintf("%s with ID '%s' version mismatch: expected %d, got %d",
		e.Resource, e.ID, e.ExpectedVersion, e.ActualVersion)
}

// NewErrVersionMismatch creates a new version mismatch error
func NewErrVersionMismatch(resource, id string, expected, actual int) ErrVersionMismatch {
	return ErrVersionMismatch{
		Resource:        resource,
		ID:              id,
		ExpectedVersion: expected,
		ActualVersion:   actual,
	}
}

// ErrCorrupted indicates that stored data is corrupted or invalid
type ErrCorrupted struct {
	Resource string
	ID       string
	Message  string
}

func (e ErrCorrupted) Error() string {
	return fmt.Sprintf("%s with ID '%s' is corrupted: %s", e.Resource, e.ID, e.Message)
}

// NewErrCorrupted creates a new corrupted data error
func NewErrCorrupted(resource, id, message string) ErrCorrupted {
	return ErrCorrupted{Resource: resource, ID: id, Message: message}
}

// ErrStorageUnavailable indicates that the storage system is unavailable
type ErrStorageUnavailable struct {
	Message string
	Cause   error
}

func (e ErrStorageUnavailable) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("storage unavailable: %s (caused by: %v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("storage unavailable: %s", e.Message)
}

// Unwrap returns the underlying cause error
func (e ErrStorageUnavailable) Unwrap() error {
	return e.Cause
}

// NewErrStorageUnavailable creates a new storage unavailable error
func NewErrStorageUnavailable(message string, cause error) ErrStorageUnavailable {
	return ErrStorageUnavailable{Message: message, Cause: cause}
}

// ErrInvalidFilter indicates that a filter or query parameter is invalid
type ErrInvalidFilter struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ErrInvalidFilter) Error() string {
	return fmt.Sprintf("invalid filter for field '%s' with value '%v': %s", e.Field, e.Value, e.Message)
}

// NewErrInvalidFilter creates a new invalid filter error
func NewErrInvalidFilter(field string, value interface{}, message string) ErrInvalidFilter {
	return ErrInvalidFilter{Field: field, Value: value, Message: message}
}

// ErrPermission indicates insufficient permissions for the operation
type ErrPermission struct {
	Operation string
	Resource  string
	Message   string
}

func (e ErrPermission) Error() string {
	return fmt.Sprintf("permission denied for %s on %s: %s", e.Operation, e.Resource, e.Message)
}

// NewErrPermission creates a new permission error
func NewErrPermission(operation, resource, message string) ErrPermission {
	return ErrPermission{Operation: operation, Resource: resource, Message: message}
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	_, ok := err.(ErrNotFound)
	return ok
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	_, ok := err.(ErrConflict)
	return ok
}

// IsVersionMismatch checks if an error is a version mismatch error
func IsVersionMismatch(err error) bool {
	_, ok := err.(ErrVersionMismatch)
	return ok
}

// IsCorrupted checks if an error is a corrupted data error
func IsCorrupted(err error) bool {
	_, ok := err.(ErrCorrupted)
	return ok
}

// IsStorageUnavailable checks if an error is a storage unavailable error
func IsStorageUnavailable(err error) bool {
	_, ok := err.(ErrStorageUnavailable)
	return ok
}

// IsInvalidFilter checks if an error is an invalid filter error
func IsInvalidFilter(err error) bool {
	_, ok := err.(ErrInvalidFilter)
	return ok
}

// IsPermission checks if an error is a permission error
func IsPermission(err error) bool {
	_, ok := err.(ErrPermission)
	return ok
}

// StorageStats represents storage statistics and health information
type StorageStats struct {
	TotalInvoices      int64               `json:"total_invoices"`
	TotalClients       int64               `json:"total_clients"`
	DiskUsageBytes     int64               `json:"disk_usage_bytes"`
	LastBackupTime     *string             `json:"last_backup_time,omitempty"`
	HealthStatus       string              `json:"health_status"` // "healthy", "degraded", "unhealthy"
	ErrorCount         int64               `json:"error_count"`   // Number of errors in last 24h
	PerformanceMetrics *PerformanceMetrics `json:"performance_metrics,omitempty"`
}

// PerformanceMetrics represents performance statistics
type PerformanceMetrics struct {
	AvgReadTimeMs  float64 `json:"avg_read_time_ms"`
	AvgWriteTimeMs float64 `json:"avg_write_time_ms"`
	TotalReads     int64   `json:"total_reads"`
	TotalWrites    int64   `json:"total_writes"`
}

// HealthStatus constants
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
)

// BackupOptions represents options for backup operations
type BackupOptions struct {
	IncludeClients   bool   `json:"include_clients"`
	IncludeInvoices  bool   `json:"include_invoices"`
	CompressionLevel int    `json:"compression_level"` // 0-9, 0 = no compression
	DestinationPath  string `json:"destination_path"`
	Encryption       bool   `json:"encryption"`
}

// RestoreOptions represents options for restore operations
type RestoreOptions struct {
	SourcePath        string `json:"source_path"`
	OverwriteExisting bool   `json:"overwrite_existing"`
	ValidateData      bool   `json:"validate_data"`
	DryRun            bool   `json:"dry_run"`
}

// InvoiceListResult represents the result of a list operation with pagination
type InvoiceListResult struct {
	Invoices   []*models.Invoice `json:"invoices"`
	TotalCount int64             `json:"total_count"`
	HasMore    bool              `json:"has_more"`
	NextOffset int               `json:"next_offset,omitempty"`
}

// ClientListResult represents the result of a client list operation with pagination
type ClientListResult struct {
	Clients    []*models.Client `json:"clients"`
	TotalCount int64            `json:"total_count"`
	HasMore    bool             `json:"has_more"`
	NextOffset int              `json:"next_offset,omitempty"`
}
