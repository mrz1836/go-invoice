package csv

import (
	"context"

	"github.com/mrz/go-invoice/internal/models"
)

// ParseOptions defines options for CSV parsing
type ParseOptions struct {
	Format          string `json:"format"`            // CSV format: "standard", "excel", "tab", etc.
	ContinueOnError bool   `json:"continue_on_error"` // Continue parsing even if some rows fail
	SkipEmptyRows   bool   `json:"skip_empty_rows"`   // Skip rows that are completely empty
	DateFormat      string `json:"date_format"`       // Preferred date format for parsing
}

// ParseResult represents the result of CSV parsing operation
type ParseResult struct {
	WorkItems   []models.WorkItem `json:"work_items"`   // Successfully parsed work items
	TotalRows   int               `json:"total_rows"`   // Total number of data rows processed
	SuccessRows int               `json:"success_rows"` // Number of successfully parsed rows
	ErrorRows   int               `json:"error_rows"`   // Number of rows that failed parsing
	Errors      []ParseError      `json:"errors"`       // Detailed error information
	HeaderMap   map[string]int    `json:"header_map"`   // Mapping of field names to column indices
	Format      string            `json:"format"`       // Detected or specified format
}

// ParseError represents an error that occurred during CSV parsing
type ParseError struct {
	Line       int      `json:"line"`       // Line number where error occurred (1-based)
	Column     string   `json:"column"`     // Column name where error occurred
	Value      string   `json:"value"`      // The problematic value
	Message    string   `json:"message"`    // Error message
	Suggestion string   `json:"suggestion"` // Suggested fix for the error
	Row        []string `json:"row"`        // The entire row that caused the error
}

// FormatInfo contains information about detected CSV format
type FormatInfo struct {
	Name      string `json:"name"`       // Format name: "standard", "excel", "tab", etc.
	Delimiter rune   `json:"delimiter"`  // Field delimiter character
	HasHeader bool   `json:"has_header"` // Whether the first row contains headers
	Encoding  string `json:"encoding"`   // Text encoding (UTF-8, etc.)
}

// CSVValidator defines validation interface for CSV parsing
// Consumer-driven interface defined at point of use
type CSVValidator interface { //nolint:revive // Keeping existing exported type name for API compatibility
	ValidateWorkItem(ctx context.Context, item *models.WorkItem) error
	ValidateRow(ctx context.Context, row []string, lineNum int) error
	ValidateBatch(ctx context.Context, items []models.WorkItem) error
}

// ValidationRule represents a single validation rule
type ValidationRule struct {
	Name         string                                                     `json:"name"`
	Description  string                                                     `json:"description"`
	Validator    func(ctx context.Context, item *models.WorkItem) error     `json:"-"`
	RowValidator func(ctx context.Context, row []string, lineNum int) error `json:"-"`
}

// ImportRequest represents a request to import CSV data
type ImportRequest struct {
	Reader      interface{}  `json:"-"`           // io.Reader containing CSV data
	Options     ParseOptions `json:"options"`     // Parsing options
	ClientID    string       `json:"client_id"`   // Client ID for new invoice (optional)
	InvoiceID   string       `json:"invoice_id"`  // Existing invoice ID to append to (optional)
	DryRun      bool         `json:"dry_run"`     // Validate only, don't persist
	Interactive bool         `json:"interactive"` // Enable interactive mode for ambiguous data
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	ParseResult    *ParseResult    `json:"parse_result"`     // Parsing results
	InvoiceID      string          `json:"invoice_id"`       // ID of created/updated invoice
	WorkItemsAdded int             `json:"work_items_added"` // Number of work items successfully added
	TotalAmount    float64         `json:"total_amount"`     // Total amount of imported work items
	Warnings       []ImportWarning `json:"warnings"`         // Non-fatal warnings
	DryRun         bool            `json:"dry_run"`          // Whether this was a dry run
}

// ImportWarning represents a non-fatal warning during import
type ImportWarning struct {
	Type    string `json:"type"`    // Warning type: "duplicate", "date_range", etc.
	Message string `json:"message"` // Warning message
	Line    int    `json:"line"`    // Line number (if applicable)
}

// ValidateImportRequest represents a request to validate CSV data without importing
type ValidateImportRequest struct {
	Reader  interface{}  `json:"-"`       // io.Reader containing CSV data
	Options ParseOptions `json:"options"` // Parsing options
}

// ValidationResult represents the result of CSV validation
type ValidationResult struct {
	Valid          bool            `json:"valid"`           // Whether the CSV is valid
	ParseResult    *ParseResult    `json:"parse_result"`    // Parsing results including errors
	Warnings       []ImportWarning `json:"warnings"`        // Non-fatal warnings
	Suggestions    []string        `json:"suggestions"`     // Suggested improvements
	EstimatedTotal float64         `json:"estimated_total"` // Estimated total amount if imported
}

// BatchImportRequest represents a request to import multiple CSV files
type BatchImportRequest struct {
	Requests []ImportRequest `json:"requests"` // Multiple import requests
	Options  BatchOptions    `json:"options"`  // Batch processing options
}

// BatchOptions defines options for batch import operations
type BatchOptions struct {
	ContinueOnError bool `json:"continue_on_error"` // Continue processing even if some imports fail
	MaxConcurrency  int  `json:"max_concurrency"`   // Maximum number of concurrent import operations
	ProgressReport  bool `json:"progress_report"`   // Enable progress reporting
}

// BatchResult represents the result of a batch import operation
type BatchResult struct {
	TotalRequests   int            `json:"total_requests"`   // Total number of import requests
	SuccessRequests int            `json:"success_requests"` // Number of successful imports
	FailedRequests  int            `json:"failed_requests"`  // Number of failed imports
	Results         []ImportResult `json:"results"`          // Individual import results
	TotalWorkItems  int            `json:"total_work_items"` // Total work items imported across all files
	TotalAmount     float64        `json:"total_amount"`     // Total amount across all imports
}

// ProgressReport represents progress information for long-running operations
type ProgressReport struct {
	Operation     string  `json:"operation"`      // Current operation description
	TotalRows     int     `json:"total_rows"`     // Total rows to process
	ProcessedRows int     `json:"processed_rows"` // Rows processed so far
	SuccessRows   int     `json:"success_rows"`   // Successfully processed rows
	ErrorRows     int     `json:"error_rows"`     // Rows with errors
	Percentage    float64 `json:"percentage"`     // Completion percentage (0-100)
}

// FormatDetectionResult represents the result of format detection
type FormatDetectionResult struct {
	PrimaryFormat    *FormatInfo   `json:"primary_format"`    // Most likely format
	AlternateFormats []*FormatInfo `json:"alternate_formats"` // Other possible formats
	Confidence       float64       `json:"confidence"`        // Confidence score (0-1)
	Suggestions      []string      `json:"suggestions"`       // Suggestions for format issues
}

// DuplicateWorkItem represents a potential duplicate work item
type DuplicateWorkItem struct {
	ImportedItem *models.WorkItem `json:"imported_item"` // Work item from CSV
	ExistingItem *models.WorkItem `json:"existing_item"` // Existing work item in system
	MatchScore   float64          `json:"match_score"`   // Similarity score (0-1)
	MatchReasons []string         `json:"match_reasons"` // Reasons why items might be duplicates
}

// DuplicateDetectionResult represents results of duplicate detection
type DuplicateDetectionResult struct {
	PotentialDuplicates []DuplicateWorkItem `json:"potential_duplicates"` // Potential duplicate pairs
	TotalChecked        int                 `json:"total_checked"`        // Total items checked
	DuplicatesFound     int                 `json:"duplicates_found"`     // Number of potential duplicates
	Confidence          float64             `json:"confidence"`           // Overall confidence in detection
}

// AggregationRules defines rules for aggregating work items
type AggregationRules struct {
	GroupByDate        bool    `json:"group_by_date"`        // Group work items by date
	GroupByDescription bool    `json:"group_by_description"` // Group work items by description
	MergeThreshold     float64 `json:"merge_threshold"`      // Minimum hours to create separate entries
	RoundingPrecision  int     `json:"rounding_precision"`   // Decimal places for hour rounding
}

// AggregationResult represents the result of work item aggregation
type AggregationResult struct {
	OriginalItems   []models.WorkItem `json:"original_items"`   // Original work items
	AggregatedItems []models.WorkItem `json:"aggregated_items"` // Aggregated work items
	ReductionCount  int               `json:"reduction_count"`  // Number of items reduced through aggregation
	TotalHours      float64           `json:"total_hours"`      // Total hours (should be same before/after)
	TotalAmount     float64           `json:"total_amount"`     // Total amount (should be same before/after)
}
