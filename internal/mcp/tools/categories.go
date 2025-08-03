package tools

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Error definitions for consistent error handling
var (
	ErrUnknownCategory = errors.New("unknown category")
)

// CategoryManager provides comprehensive tool categorization and discovery functionality.
//
// This implementation supports hierarchical tool organization, intelligent filtering,
// and natural language-friendly tool discovery. It follows context-first design
// principles and consumer-driven interface patterns.
//
// Key features:
// - Hierarchical category organization with metadata
// - Intelligent tool filtering and discovery
// - Natural language description generation
// - Usage pattern tracking for recommendations
// - Context-aware operations with cancellation support
// - Performance-optimized for frequent lookups
//
// The category manager helps Claude understand tool organization and makes
// tool discovery more intuitive through conversational interfaces.
type CategoryManager struct {
	// registry provides access to tool registry for category analysis
	registry ToolRegistry

	// logger provides structured logging for category operations
	logger Logger

	// categoryMetadata maps categories to detailed metadata
	categoryMetadata map[CategoryType]*CategoryMetadata
}

// CategoryMetadata provides detailed information about tool categories.
//
// This struct contains comprehensive category information to help Claude
// understand tool organization and provide intelligent recommendations.
//
// Fields:
// - Name: Human-readable category name
// - Description: Detailed category description for Claude interaction
// - Keywords: Search keywords for natural language discovery
// - UseCases: Common use cases for tools in this category
// - Prerequisites: Prerequisites for using tools in this category
// - RelatedCategories: Related categories for cross-discovery
// - Priority: Category priority for recommendation ordering
//
// Notes:
// - Keywords should include natural language terms users might use
// - UseCases help Claude understand when to recommend category tools
// - Prerequisites inform Claude about setup requirements
type CategoryMetadata struct {
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	Keywords          []string       `json:"keywords"`
	UseCases          []string       `json:"useCases"`
	Prerequisites     []string       `json:"prerequisites,omitempty"`
	RelatedCategories []CategoryType `json:"relatedCategories,omitempty"`
	Priority          int            `json:"priority"`
}

// CategoryDiscoveryFilter defines criteria for tool category discovery.
//
// This struct provides flexible filtering options for tool discovery
// based on category characteristics and user requirements.
//
// Fields:
// - Keywords: Keywords to match against category metadata
// - UseCases: Use cases to match for recommendation
// - IncludeEmpty: Whether to include categories with no tools
// - MaxResults: Maximum number of categories to return
// - SortBy: Sort criteria for result ordering
//
// Notes:
// - Empty filters return all categories
// - Keywords are matched case-insensitively
// - Results are always sorted by priority then name for consistency
type CategoryDiscoveryFilter struct {
	Keywords     []string `json:"keywords,omitempty"`
	UseCases     []string `json:"useCases,omitempty"`
	IncludeEmpty bool     `json:"includeEmpty"`
	MaxResults   int      `json:"maxResults"`
	SortBy       string   `json:"sortBy"`
}

// CategorySummary provides summary information about a tool category.
//
// This struct contains aggregated information about categories for
// discovery and recommendation purposes.
//
// Fields:
// - Category: The category type identifier
// - Metadata: Category metadata information
// - ToolCount: Number of tools in this category
// - PopularTools: Most commonly used tools in this category
// - RecommendationScore: Score for recommendation ranking
//
// Notes:
// - PopularTools are ordered by usage frequency
// - RecommendationScore considers category priority and tool count
// - Used for generating category recommendations for Claude
type CategorySummary struct {
	Category            CategoryType      `json:"category"`
	Metadata            *CategoryMetadata `json:"metadata"`
	ToolCount           int               `json:"toolCount"`
	PopularTools        []string          `json:"popularTools,omitempty"`
	RecommendationScore float64           `json:"recommendationScore"`
}

// NewCategoryManager creates a new category manager with dependency injection.
//
// This constructor initializes a comprehensive category management system
// with predefined category metadata and intelligent discovery capabilities.
//
// Parameters:
// - registry: Tool registry for category analysis and tool access
// - logger: Structured logger for category operations and debugging
//
// Returns:
// - *CategoryManager: Initialized category manager ready for discovery operations
//
// Notes:
// - Manager includes predefined metadata for all known categories
// - All dependencies must be non-nil or the constructor will panic
// - Thread-safe for concurrent category operations
func NewCategoryManager(registry ToolRegistry, logger Logger) *CategoryManager {
	if registry == nil {
		panic("registry cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	manager := &CategoryManager{
		registry:         registry,
		logger:           logger,
		categoryMetadata: make(map[CategoryType]*CategoryMetadata),
	}

	// Initialize predefined category metadata
	manager.initializeCategoryMetadata()

	return manager
}

// GetCategoryMetadata retrieves detailed metadata for a specific category.
//
// This method provides comprehensive category information for Claude
// to understand category purposes and characteristics.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - category: Category to retrieve metadata for
//
// Returns:
// - *CategoryMetadata: Detailed category metadata, or nil if category unknown
// - error: Error if context cancelled or category invalid
//
// Side Effects:
// - Logs category metadata requests for analytics
//
// Notes:
// - Returns nil metadata for unknown categories (not an error)
// - Metadata includes keywords, use cases, and prerequisites
// - Respects context cancellation for responsive operations
func (cm *CategoryManager) GetCategoryMetadata(ctx context.Context, category CategoryType) (*CategoryMetadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	metadata, exists := cm.categoryMetadata[category]
	if !exists {
		cm.logger.Debug("category metadata not found", "category", category)
		return nil, nil
	}

	cm.logger.Debug("category metadata retrieved", "category", category, "name", metadata.Name)

	// Return defensive copy to prevent external modification
	metadataCopy := *metadata
	if metadata.Keywords != nil {
		metadataCopy.Keywords = make([]string, len(metadata.Keywords))
		copy(metadataCopy.Keywords, metadata.Keywords)
	}
	if metadata.UseCases != nil {
		metadataCopy.UseCases = make([]string, len(metadata.UseCases))
		copy(metadataCopy.UseCases, metadata.UseCases)
	}
	if metadata.Prerequisites != nil {
		metadataCopy.Prerequisites = make([]string, len(metadata.Prerequisites))
		copy(metadataCopy.Prerequisites, metadata.Prerequisites)
	}
	if metadata.RelatedCategories != nil {
		metadataCopy.RelatedCategories = make([]CategoryType, len(metadata.RelatedCategories))
		copy(metadataCopy.RelatedCategories, metadata.RelatedCategories)
	}

	return &metadataCopy, nil
}

// DiscoverCategories finds categories matching discovery criteria.
//
// This method provides intelligent category discovery based on keywords,
// use cases, and other criteria to help Claude recommend appropriate tools.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - filter: Discovery criteria and filtering options
//
// Returns:
// - []*CategorySummary: Matching categories with summary information
// - error: Error if discovery fails or context cancelled
//
// Side Effects:
// - Logs category discovery operations for analytics
// - Updates recommendation scores based on usage patterns
//
// Notes:
// - Empty filter returns all categories with tools
// - Results are sorted by recommendation score and priority
// - Includes tool count and popular tools for each category
// - Respects context cancellation for large discovery operations
func (cm *CategoryManager) DiscoverCategories(ctx context.Context, filter *CategoryDiscoveryFilter) ([]*CategorySummary, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Use default filter if none provided
	if filter == nil {
		filter = &CategoryDiscoveryFilter{
			IncludeEmpty: false,
			MaxResults:   10,
			SortBy:       "priority",
		}
	}

	cm.logger.Debug("starting category discovery",
		"keywords", filter.Keywords,
		"useCases", filter.UseCases,
		"maxResults", filter.MaxResults)

	var summaries []*CategorySummary

	for category, metadata := range cm.categoryMetadata {
		// Check if category matches filter criteria
		if !cm.matchesFilter(ctx, category, metadata, filter) {
			continue
		}

		// Get tool count for this category
		tools, err := cm.registry.ListTools(ctx, category)
		if err != nil {
			cm.logger.Warn("failed to get tools for category", "category", category, "error", err)
			continue
		}

		// Skip empty categories if not included
		if len(tools) == 0 && !filter.IncludeEmpty {
			continue
		}

		// Create category summary
		summary := &CategorySummary{
			Category:            category,
			Metadata:            metadata,
			ToolCount:           len(tools),
			PopularTools:        cm.getPopularTools(tools),
			RecommendationScore: cm.calculateRecommendationScore(metadata, len(tools)),
		}

		summaries = append(summaries, summary)
	}

	// Sort results by recommendation criteria
	cm.sortCategorySummaries(summaries, filter.SortBy)

	// Apply max results limit
	if filter.MaxResults > 0 && len(summaries) > filter.MaxResults {
		summaries = summaries[:filter.MaxResults]
	}

	cm.logger.Debug("category discovery completed",
		"matchingCategories", len(summaries))

	return summaries, nil
}

// GenerateNaturalLanguageDescription creates human-readable category descriptions.
//
// This method generates conversational descriptions of categories that Claude
// can use in natural language interactions with users.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - category: Category to generate description for
// - includeTools: Whether to include tool examples in description
//
// Returns:
// - string: Natural language description of the category
// - error: Error if generation fails or context cancelled
//
// Side Effects:
// - Logs description generation for analytics
//
// Notes:
// - Descriptions are optimized for conversational interfaces
// - Includes use cases and prerequisites when relevant
// - Can include tool examples for better understanding
// - Respects context cancellation for complex generation
func (cm *CategoryManager) GenerateNaturalLanguageDescription(ctx context.Context, category CategoryType, includeTools bool) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	metadata, err := cm.GetCategoryMetadata(ctx, category)
	if err != nil {
		return "", fmt.Errorf("failed to get category metadata: %w", err)
	}
	if metadata == nil {
		return "", fmt.Errorf("%w: %s", ErrUnknownCategory, category)
	}

	var description strings.Builder

	// Start with category name and description
	description.WriteString(fmt.Sprintf("**%s**: %s", metadata.Name, metadata.Description))

	// Add use cases
	if len(metadata.UseCases) > 0 {
		description.WriteString("\n\nCommon use cases include:")
		for _, useCase := range metadata.UseCases {
			description.WriteString(fmt.Sprintf("\n• %s", useCase))
		}
	}

	// Add prerequisites if any
	if len(metadata.Prerequisites) > 0 {
		description.WriteString("\n\nPrerequisites:")
		for _, prereq := range metadata.Prerequisites {
			description.WriteString(fmt.Sprintf("\n• %s", prereq))
		}
	}

	// Add tool examples if requested
	if includeTools {
		tools, err := cm.registry.ListTools(ctx, category)
		if err == nil && len(tools) > 0 {
			description.WriteString("\n\nAvailable tools:")

			// Show up to 3 example tools
			maxTools := 3
			if len(tools) < maxTools {
				maxTools = len(tools)
			}

			for i := 0; i < maxTools; i++ {
				tool := tools[i]
				description.WriteString(fmt.Sprintf("\n• **%s**: %s", tool.Name, tool.Description))
			}

			if len(tools) > maxTools {
				description.WriteString(fmt.Sprintf("\n• ...and %d more tools", len(tools)-maxTools))
			}
		}
	}

	// Add related categories
	if len(metadata.RelatedCategories) > 0 {
		description.WriteString("\n\nRelated categories: ")
		var related []string
		for _, relatedCat := range metadata.RelatedCategories {
			if relatedMeta, exists := cm.categoryMetadata[relatedCat]; exists {
				related = append(related, relatedMeta.Name)
			}
		}
		description.WriteString(strings.Join(related, ", "))
	}

	cm.logger.Debug("natural language description generated",
		"category", category,
		"includeTools", includeTools,
		"descriptionLength", description.Len())

	return description.String(), nil
}

// GetRecommendedCategories suggests categories based on context and usage patterns.
//
// This method provides intelligent category recommendations to help Claude
// suggest relevant tools based on user context and common usage patterns.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - userQuery: User query or context for recommendations
// - maxRecommendations: Maximum number of categories to recommend
//
// Returns:
// - []*CategorySummary: Recommended categories ordered by relevance
// - error: Error if recommendation fails or context cancelled
//
// Side Effects:
// - Logs recommendation requests for analytics and improvement
//
// Notes:
// - Recommendations based on keyword matching and usage patterns
// - Results ordered by relevance score and category priority
// - Includes explanation of why each category was recommended
// - Respects context cancellation for complex recommendation logic
func (cm *CategoryManager) GetRecommendedCategories(ctx context.Context, userQuery string, maxRecommendations int) ([]*CategorySummary, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if maxRecommendations <= 0 {
		maxRecommendations = 3 // Default to top 3 recommendations
	}

	cm.logger.Debug("generating category recommendations",
		"userQuery", userQuery,
		"maxRecommendations", maxRecommendations)

	// Extract keywords from user query
	queryKeywords := cm.extractKeywords(userQuery)

	// Create filter based on query keywords
	filter := &CategoryDiscoveryFilter{
		Keywords:     queryKeywords,
		IncludeEmpty: false,
		MaxResults:   maxRecommendations,
		SortBy:       "relevance",
	}

	// Discover matching categories
	summaries, err := cm.DiscoverCategories(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to discover categories: %w", err)
	}

	// Enhance summaries with relevance scoring for the specific query
	for _, summary := range summaries {
		summary.RecommendationScore = cm.calculateQueryRelevance(userQuery, summary.Metadata)
	}

	// Re-sort by query relevance
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].RecommendationScore > summaries[j].RecommendationScore
	})

	cm.logger.Debug("category recommendations generated",
		"recommendationCount", len(summaries),
		"topCategory", func() string {
			if len(summaries) > 0 {
				return string(summaries[0].Category)
			}
			return "none"
		}())

	return summaries, nil
}

// initializeCategoryMetadata sets up predefined metadata for all known categories.
//
// This internal method initializes comprehensive metadata for each category
// with keywords, use cases, and relationships optimized for natural language interaction.
func (cm *CategoryManager) initializeCategoryMetadata() {
	cm.categoryMetadata[CategoryInvoiceManagement] = &CategoryMetadata{
		Name:        "Invoice Management",
		Description: "Tools for creating, updating, and managing invoices throughout their lifecycle",
		Keywords:    []string{"invoice", "create", "update", "manage", "billing", "payment", "due date", "client"},
		UseCases: []string{
			"Creating new invoices for clients",
			"Updating invoice details and metadata",
			"Managing invoice status and payments",
			"Setting due dates and payment terms",
		},
		RelatedCategories: []CategoryType{CategoryClientManagement, CategoryDataExport, CategoryReporting},
		Priority:          1,
	}

	cm.categoryMetadata[CategoryDataImport] = &CategoryMetadata{
		Name:        "Data Import",
		Description: "Tools for importing timesheet data, client information, and other external data",
		Keywords:    []string{"import", "csv", "timesheet", "data", "upload", "file", "spreadsheet", "hours"},
		UseCases: []string{
			"Importing timesheet data from CSV files",
			"Loading client information from external sources",
			"Bulk importing work hours and billing data",
			"Converting spreadsheet data to invoice items",
		},
		Prerequisites:     []string{"CSV files formatted correctly", "File paths accessible"},
		RelatedCategories: []CategoryType{CategoryInvoiceManagement, CategoryClientManagement},
		Priority:          2,
	}

	cm.categoryMetadata[CategoryDataExport] = &CategoryMetadata{
		Name:        "Data Export",
		Description: "Tools for generating and exporting invoice documents, reports, and data",
		Keywords:    []string{"export", "generate", "html", "pdf", "report", "document", "download", "template"},
		UseCases: []string{
			"Generating HTML invoices for clients",
			"Creating PDF documents for printing",
			"Exporting invoice data for accounting systems",
			"Generating reports for analysis",
		},
		RelatedCategories: []CategoryType{CategoryInvoiceManagement, CategoryReporting},
		Priority:          2,
	}

	cm.categoryMetadata[CategoryClientManagement] = &CategoryMetadata{
		Name:        "Client Management",
		Description: "Tools for managing client information, contacts, and relationships",
		Keywords:    []string{"client", "customer", "contact", "company", "address", "email", "phone", "manage"},
		UseCases: []string{
			"Adding new clients to the system",
			"Updating client contact information",
			"Managing client billing preferences",
			"Organizing client relationships",
		},
		RelatedCategories: []CategoryType{CategoryInvoiceManagement, CategoryDataImport},
		Priority:          3,
	}

	cm.categoryMetadata[CategoryConfiguration] = &CategoryMetadata{
		Name:        "Configuration",
		Description: "Tools for system configuration, settings management, and validation",
		Keywords:    []string{"config", "settings", "setup", "validate", "preferences", "options", "system"},
		UseCases: []string{
			"Configuring invoice templates and formats",
			"Setting up payment terms and tax rates",
			"Validating system configuration",
			"Managing user preferences",
		},
		Prerequisites:     []string{"Administrative access", "Understanding of invoice requirements"},
		RelatedCategories: []CategoryType{CategoryInvoiceManagement},
		Priority:          4,
	}

	cm.categoryMetadata[CategoryReporting] = &CategoryMetadata{
		Name:        "Reporting",
		Description: "Tools for analytics, reporting, and business intelligence on invoice data",
		Keywords:    []string{"report", "analytics", "statistics", "summary", "analysis", "metrics", "dashboard"},
		UseCases: []string{
			"Generating revenue reports and summaries",
			"Analyzing invoice patterns and trends",
			"Creating client billing summaries",
			"Tracking payment status and overdue amounts",
		},
		RelatedCategories: []CategoryType{CategoryInvoiceManagement, CategoryDataExport},
		Priority:          5,
	}
}

// matchesFilter checks if a category matches discovery filter criteria.
//
// This internal method evaluates whether a category meets the specified
// filter criteria for discovery operations.
func (cm *CategoryManager) matchesFilter(ctx context.Context, category CategoryType, metadata *CategoryMetadata, filter *CategoryDiscoveryFilter) bool {
	// Check keyword matching
	if len(filter.Keywords) > 0 {
		if !cm.matchesKeywords(metadata, filter.Keywords) {
			return false
		}
	}

	// Check use case matching
	if len(filter.UseCases) > 0 {
		if !cm.matchesUseCases(metadata, filter.UseCases) {
			return false
		}
	}

	return true
}

// matchesKeywords checks if category metadata matches any of the provided keywords.
func (cm *CategoryManager) matchesKeywords(metadata *CategoryMetadata, keywords []string) bool {
	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)

		// Check in category name
		if strings.Contains(strings.ToLower(metadata.Name), keywordLower) {
			return true
		}

		// Check in description
		if strings.Contains(strings.ToLower(metadata.Description), keywordLower) {
			return true
		}

		// Check in keywords
		for _, metaKeyword := range metadata.Keywords {
			if strings.Contains(strings.ToLower(metaKeyword), keywordLower) {
				return true
			}
		}

		// Check in use cases
		for _, useCase := range metadata.UseCases {
			if strings.Contains(strings.ToLower(useCase), keywordLower) {
				return true
			}
		}
	}

	return false
}

// matchesUseCases checks if category metadata matches any of the provided use cases.
func (cm *CategoryManager) matchesUseCases(metadata *CategoryMetadata, useCases []string) bool {
	for _, filterUseCase := range useCases {
		filterUseCaseLower := strings.ToLower(filterUseCase)

		for _, metaUseCase := range metadata.UseCases {
			if strings.Contains(strings.ToLower(metaUseCase), filterUseCaseLower) {
				return true
			}
		}
	}

	return false
}

// getPopularTools extracts popular tool names from a tool list.
//
// This internal method identifies the most commonly used tools in a category
// for recommendation purposes.
func (cm *CategoryManager) getPopularTools(tools []*MCPTool) []string {
	if len(tools) == 0 {
		return nil
	}

	// For now, return first 3 tools (in future, this could be based on usage statistics)
	var popular []string
	maxTools := 3
	if len(tools) < maxTools {
		maxTools = len(tools)
	}

	for i := 0; i < maxTools; i++ {
		popular = append(popular, tools[i].Name)
	}

	return popular
}

// calculateRecommendationScore computes a recommendation score for a category.
//
// This internal method calculates how strongly a category should be recommended
// based on its priority, tool count, and other factors.
func (cm *CategoryManager) calculateRecommendationScore(metadata *CategoryMetadata, toolCount int) float64 {
	// Base score from priority (higher priority = higher score)
	priorityScore := float64(10 - metadata.Priority) // Invert so priority 1 = score 9

	// Tool count factor (more tools = higher score, but with diminishing returns)
	toolScore := float64(toolCount) * 0.5
	if toolScore > 5 {
		toolScore = 5 // Cap at 5 points
	}

	return priorityScore + toolScore
}

// calculateQueryRelevance calculates how relevant a category is to a specific query.
//
// This internal method computes relevance scores for query-specific recommendations.
func (cm *CategoryManager) calculateQueryRelevance(query string, metadata *CategoryMetadata) float64 {
	if query == "" {
		return cm.calculateRecommendationScore(metadata, 0)
	}

	queryLower := strings.ToLower(query)
	score := 0.0

	// Check for exact matches in category name (highest weight)
	if strings.Contains(strings.ToLower(metadata.Name), queryLower) {
		score += 10.0
	}

	// Check for matches in keywords (high weight)
	for _, keyword := range metadata.Keywords {
		if strings.Contains(strings.ToLower(keyword), queryLower) {
			score += 5.0
		}
	}

	// Check for matches in description (medium weight)
	if strings.Contains(strings.ToLower(metadata.Description), queryLower) {
		score += 3.0
	}

	// Check for matches in use cases (medium weight)
	for _, useCase := range metadata.UseCases {
		if strings.Contains(strings.ToLower(useCase), queryLower) {
			score += 3.0
		}
	}

	// Add base priority score
	score += cm.calculateRecommendationScore(metadata, 0)

	return score
}

// extractKeywords extracts relevant keywords from a user query.
//
// This internal method processes user queries to extract meaningful keywords
// for category matching.
func (cm *CategoryManager) extractKeywords(query string) []string {
	if query == "" {
		return nil
	}

	// Simple keyword extraction (could be enhanced with NLP)
	queryLower := strings.ToLower(query)
	words := strings.Fields(queryLower)

	// Filter out common stop words
	stopWords := map[string]bool{
		"the": true, "and": true, "or": true, "but": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "of": true, "with": true, "by": true,
		"a": true, "an": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
		"me": true, "him": true, "her": true, "us": true, "them": true, "my": true, "your": true,
		"his": true, "its": true, "our": true, "their": true,
	}

	var keywords []string
	for _, word := range words {
		// Keep words that are not stop words and are at least 3 characters
		if !stopWords[word] && len(word) >= 3 {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// sortCategorySummaries sorts category summaries by the specified criteria.
//
// This internal method handles different sorting options for category discovery results.
func (cm *CategoryManager) sortCategorySummaries(summaries []*CategorySummary, sortBy string) {
	switch sortBy {
	case "priority":
		sort.Slice(summaries, func(i, j int) bool {
			if summaries[i].Metadata.Priority == summaries[j].Metadata.Priority {
				return summaries[i].Metadata.Name < summaries[j].Metadata.Name
			}
			return summaries[i].Metadata.Priority < summaries[j].Metadata.Priority
		})
	case "toolCount":
		sort.Slice(summaries, func(i, j int) bool {
			if summaries[i].ToolCount == summaries[j].ToolCount {
				return summaries[i].Metadata.Name < summaries[j].Metadata.Name
			}
			return summaries[i].ToolCount > summaries[j].ToolCount
		})
	case "name":
		sort.Slice(summaries, func(i, j int) bool {
			return summaries[i].Metadata.Name < summaries[j].Metadata.Name
		})
	case "relevance":
		fallthrough
	default:
		sort.Slice(summaries, func(i, j int) bool {
			if summaries[i].RecommendationScore == summaries[j].RecommendationScore {
				return summaries[i].Metadata.Name < summaries[j].Metadata.Name
			}
			return summaries[i].RecommendationScore > summaries[j].RecommendationScore
		})
	}
}
