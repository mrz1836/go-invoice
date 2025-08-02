// Package storage provides data persistence interfaces and implementations for the go-invoice application.
package storage

import (
	"fmt"

	"github.com/mrz/go-invoice/internal/models"
)

// Storage error types for comprehensive error handling

// NotFoundError indicates that a requested resource was not found
type NotFoundError struct {
	Resource string
	ID       string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID '%s' not found", e.Resource, e.ID)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string) NotFoundError {
	return NotFoundError{Resource: resource, ID: id}
}

// ConflictError indicates that a resource already exists or conflicts with another
type ConflictError struct {
	Resource string
	ID       string
	Message  string
}

func (e ConflictError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s with ID '%s' conflicts: %s", e.Resource, e.ID, e.Message)
	}
	return fmt.Sprintf("%s with ID '%s' already exists", e.Resource, e.ID)
}

// NewConflictError creates a new conflict error
func NewConflictError(resource, id, message string) ConflictError {
	return ConflictError{Resource: resource, ID: id, Message: message}
}

// VersionMismatchError indicates an optimistic locking failure
type VersionMismatchError struct {
	Resource        string
	ID              string
	ExpectedVersion int
	ActualVersion   int
}

func (e VersionMismatchError) Error() string {
	return fmt.Sprintf("%s with ID '%s' version mismatch: expected %d, got %d",
		e.Resource, e.ID, e.ExpectedVersion, e.ActualVersion)
}

// NewVersionMismatchError creates a new version mismatch error
func NewVersionMismatchError(resource, id string, expected, actual int) VersionMismatchError {
	return VersionMismatchError{
		Resource:        resource,
		ID:              id,
		ExpectedVersion: expected,
		ActualVersion:   actual,
	}
}

// CorruptedError indicates that stored data is corrupted or invalid
type CorruptedError struct {
	Resource string
	ID       string
	Message  string
}

func (e CorruptedError) Error() string {
	return fmt.Sprintf("%s with ID '%s' is corrupted: %s", e.Resource, e.ID, e.Message)
}

// NewCorruptedError creates a new corrupted data error
func NewCorruptedError(resource, id, message string) CorruptedError {
	return CorruptedError{Resource: resource, ID: id, Message: message}
}

// StorageUnavailableError indicates that the storage system is unavailable
type StorageUnavailableError struct {
	Message string
	Cause   error
}

func (e StorageUnavailableError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("storage unavailable: %s (caused by: %v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("storage unavailable: %s", e.Message)
}

// Unwrap returns the underlying cause error
func (e StorageUnavailableError) Unwrap() error {
	return e.Cause
}

// NewStorageUnavailableError creates a new storage unavailable error
func NewStorageUnavailableError(message string, cause error) StorageUnavailableError {
	return StorageUnavailableError{Message: message, Cause: cause}
}

// InvalidFilterError indicates that a filter or query parameter is invalid
type InvalidFilterError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e InvalidFilterError) Error() string {
	return fmt.Sprintf("invalid filter for field '%s' with value '%v': %s", e.Field, e.Value, e.Message)
}

// NewInvalidFilterError creates a new invalid filter error
func NewInvalidFilterError(field string, value interface{}, message string) InvalidFilterError {
	return InvalidFilterError{Field: field, Value: value, Message: message}
}

// PermissionError indicates insufficient permissions for the operation
type PermissionError struct {
	Operation string
	Resource  string
	Message   string
}

func (e PermissionError) Error() string {
	return fmt.Sprintf("permission denied for %s on %s: %s", e.Operation, e.Resource, e.Message)
}

// NewPermissionError creates a new permission error
func NewPermissionError(operation, resource, message string) PermissionError {
	return PermissionError{Operation: operation, Resource: resource, Message: message}
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	_, ok := err.(NotFoundError)
	return ok
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	_, ok := err.(ConflictError)
	return ok
}

// IsVersionMismatch checks if an error is a version mismatch error
func IsVersionMismatch(err error) bool {
	_, ok := err.(VersionMismatchError)
	return ok
}

// IsCorrupted checks if an error is a corrupted data error
func IsCorrupted(err error) bool {
	_, ok := err.(CorruptedError)
	return ok
}

// IsStorageUnavailable checks if an error is a storage unavailable error
func IsStorageUnavailable(err error) bool {
	_, ok := err.(StorageUnavailableError)
	return ok
}

// IsInvalidFilter checks if an error is an invalid filter error
func IsInvalidFilter(err error) bool {
	_, ok := err.(InvalidFilterError)
	return ok
}

// IsPermission checks if an error is a permission error
func IsPermission(err error) bool {
	_, ok := err.(PermissionError)
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
