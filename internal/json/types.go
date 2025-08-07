package json

import "time"

// JSONImport represents the structured JSON import format with optional metadata
type JSONImport struct {
	Metadata  *ImportMetadata `json:"metadata,omitempty"`
	WorkItems []WorkItemJSON  `json:"work_items"`
}

// ImportMetadata contains optional metadata about the import
type ImportMetadata struct {
	Client      string    `json:"client,omitempty"`
	Period      string    `json:"period,omitempty"`
	Description string    `json:"description,omitempty"`
	Exported    time.Time `json:"exported,omitempty"`
	Currency    string    `json:"currency,omitempty"`
	TotalHours  float64   `json:"total_hours,omitempty"`
	TotalAmount float64   `json:"total_amount,omitempty"`
}

// WorkItemJSON represents a single work item in JSON format
type WorkItemJSON struct {
	Date        string   `json:"date"`
	Hours       float64  `json:"hours"`
	Rate        float64  `json:"rate"`
	Description string   `json:"description"`
	Project     string   `json:"project,omitempty"`  // Optional project field
	Category    string   `json:"category,omitempty"` // Optional category field
	Billable    *bool    `json:"billable,omitempty"` // Optional billable flag
	Tags        []string `json:"tags,omitempty"`     // Optional tags
}

// SimpleWorkItemJSON represents the simple array format for work items
type SimpleWorkItemJSON struct {
	Date        string  `json:"date"`
	Hours       float64 `json:"hours"`
	Rate        float64 `json:"rate"`
	Description string  `json:"description"`
}
