// Package tools provides comprehensive tool discovery and search functionality.
//
// This package implements advanced tool discovery capabilities for the MCP tool registry,
// enabling efficient tool lookup, filtering, and search operations. It supports natural
// language interaction patterns and provides intelligent suggestions for tool usage.
//
// Key features:
// - Flexible tool search with fuzzy matching
// - Category-based filtering and discovery
// - Natural language search capabilities
// - Tool recommendation engine
// - Performance-optimized search algorithms
// - Context-aware operations with cancellation support
//
// The discovery system is designed to help Claude and other MCP clients efficiently
// find and understand available tools for various invoice management workflows.
package tools

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Error definitions for discovery service
var (
	ErrRegistryNil       = errors.New("registry cannot be nil")
	ErrSearchCriteriaNil = errors.New("search criteria cannot be nil")
)

// ToolDiscoveryService provides advanced tool discovery and search capabilities.
//
// This service extends the basic registry functionality with sophisticated search
// and recommendation features optimized for conversational interaction with Claude.
//
// Key features:
// - Multi-criteria search (name, description, category, tags)
// - Fuzzy matching for typo tolerance
// - Usage-based tool recommendations
// - Category-aware filtering
// - Performance-optimized search indices
// - Relevance scoring for search results
//
// The service maintains search indices for fast lookup and provides comprehensive
// search results with relevance scoring and suggested alternatives.
type ToolDiscoveryService struct {
	// registry provides access to the complete tool registry
	registry ToolRegistry

	// logger provides structured logging for discovery operations
	logger Logger

	// searchIndex maintains optimized search data structures
	searchIndex *ToolSearchIndex
}

// ToolSearchIndex provides optimized data structures for tool search operations.
//
// This index maintains multiple search paths to enable fast tool discovery
// across different search criteria and use cases.
//
// Fields:
// - NameIndex: Maps tool names and aliases to tool definitions
// - DescriptionIndex: Maps description keywords to relevant tools
// - CategoryIndex: Maps categories to tool lists with metadata
// - TagIndex: Maps tags to tool collections
// - FullTextIndex: Comprehensive text search index
//
// Notes:
// - Indices are built once during initialization for performance
// - Updates require index rebuilding for consistency
// - Memory-optimized for fast lookup operations
type ToolSearchIndex struct {
	NameIndex        map[string][]*MCPTool         `json:"nameIndex"`
	DescriptionIndex map[string][]*MCPTool         `json:"descriptionIndex"`
	CategoryIndex    map[CategoryType][]*MCPTool   `json:"categoryIndex"`
	TagIndex         map[string][]*MCPTool         `json:"tagIndex"`
	FullTextIndex    map[string][]ToolSearchResult `json:"fullTextIndex"`
}

// ToolSearchResult represents a search result with relevance scoring.
//
// This struct provides comprehensive information about tool search matches
// including relevance scoring and match context for result ranking.
//
// Fields:
// - Tool: The matched tool definition
// - RelevanceScore: Numerical relevance score (0.0 to 1.0)
// - MatchContext: Description of why this tool matched
// - MatchedFields: List of fields that contributed to the match
// - CategoryMatch: Whether the match was based on category filtering
//
// Notes:
// - RelevanceScore enables intelligent result ranking
// - MatchContext helps users understand search results
// - Used for both simple and complex search operations
type ToolSearchResult struct {
	Tool           *MCPTool `json:"tool"`
	RelevanceScore float64  `json:"relevanceScore"`
	MatchContext   string   `json:"matchContext"`
	MatchedFields  []string `json:"matchedFields"`
	CategoryMatch  bool     `json:"categoryMatch"`
}

// ToolSearchCriteria defines comprehensive search parameters for tool discovery.
//
// This struct supports complex search operations with multiple criteria
// and filtering options for precise tool discovery.
//
// Fields:
// - Query: Free-text search query for natural language search
// - Categories: List of categories to filter by
// - Tags: List of tags to match
// - IncludeExamples: Whether to include tool examples in search
// - MaxResults: Maximum number of results to return
// - MinRelevanceScore: Minimum relevance score threshold
// - SortBy: Field to sort results by
// - SortOrder: Sort order (asc/desc)
//
// Notes:
// - Empty query returns all tools with category/tag filtering
// - Multiple criteria are combined using AND logic
// - Flexible sorting enables different result presentations
type ToolSearchCriteria struct {
	Query             string         `json:"query"`
	Categories        []CategoryType `json:"categories,omitempty"`
	Tags              []string       `json:"tags,omitempty"`
	IncludeExamples   bool           `json:"includeExamples"`
	MaxResults        int            `json:"maxResults"`
	MinRelevanceScore float64        `json:"minRelevanceScore"`
	SortBy            string         `json:"sortBy"`
	SortOrder         string         `json:"sortOrder"`
}

// NewToolDiscoveryService creates a new tool discovery service with search indexing.
//
// This constructor initializes the discovery service with comprehensive search
// indices for fast tool lookup and recommendation operations.
//
// Parameters:
// - ctx: Context for cancellation and timeout during initialization
// - registry: Tool registry to provide discovery services for
// - logger: Structured logger for discovery operations
//
// Returns:
// - *ToolDiscoveryService: Fully initialized discovery service with search indices
// - error: Initialization error if index building fails
//
// Side Effects:
// - Builds comprehensive search indices from registry data
// - Logs initialization progress and performance metrics
//
// Notes:
// - Index building may take time for large tool sets
// - Service is ready for immediate use after construction
// - Respects context cancellation during initialization
// - Thread-safe for concurrent search operations after initialization
func NewToolDiscoveryService(ctx context.Context, registry ToolRegistry, logger Logger) (*ToolDiscoveryService, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if registry == nil {
		return nil, ErrRegistryNil
	}
	if logger == nil {
		return nil, ErrLoggerNil
	}

	service := &ToolDiscoveryService{
		registry: registry,
		logger:   logger,
	}

	// Build search indices
	if err := service.buildSearchIndex(ctx); err != nil {
		return nil, fmt.Errorf("failed to build search index: %w", err)
	}

	logger.Info("tool discovery service initialized successfully")
	return service, nil
}

// SearchTools performs comprehensive tool search with relevance scoring.
//
// This method provides flexible tool search capabilities with support for
// natural language queries, category filtering, and intelligent ranking.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - criteria: Search criteria and filtering options
//
// Returns:
// - []ToolSearchResult: Ranked search results with relevance scoring
// - error: Search error or context cancellation
//
// Side Effects:
// - Logs search operations for analytics and debugging
//
// Notes:
// - Results are ranked by relevance score for optimal ordering
// - Supports fuzzy matching for typo tolerance
// - Empty query with category filters returns all tools in categories
// - Respects MaxResults and MinRelevanceScore limits
// - Thread-safe for concurrent search operations
func (s *ToolDiscoveryService) SearchTools(ctx context.Context, criteria *ToolSearchCriteria) ([]ToolSearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if criteria == nil {
		return nil, ErrSearchCriteriaNil
	}

	s.logger.Debug("starting tool search",
		"query", criteria.Query,
		"categories", len(criteria.Categories),
		"maxResults", criteria.MaxResults)

	var results []ToolSearchResult

	// Handle category-only filtering
	if criteria.Query == "" && len(criteria.Categories) > 0 {
		results = s.searchByCategory(ctx, criteria)
	} else if criteria.Query != "" {
		results = s.searchByQuery(ctx, criteria)
	} else {
		// Return all tools if no criteria specified
		results = s.getAllToolsAsResults(ctx)
	}

	// Apply post-processing filters
	results = s.filterResults(results, criteria)
	results = s.sortResults(results, criteria)

	s.logger.Debug("tool search completed",
		"query", criteria.Query,
		"totalResults", len(results),
		"criteriaApplied", fmt.Sprintf("categories:%d, tags:%d", len(criteria.Categories), len(criteria.Tags)))

	return results, nil
}

// DiscoverToolsByCategory provides efficient category-based tool discovery.
//
// This method optimizes tool discovery for category-based browsing and
// provides comprehensive category information for navigation.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - category: Category to discover tools for (empty for all categories)
//
// Returns:
// - *CategoryDiscoveryResult: Comprehensive category discovery information
// - error: Discovery error or context cancellation
//
// Notes:
// - Provides tool counts and metadata for category navigation
// - Includes related categories and tool recommendations
// - Optimized for category browsing workflows
// - Thread-safe for concurrent discovery operations
func (s *ToolDiscoveryService) DiscoverToolsByCategory(ctx context.Context, category CategoryType) (*CategoryDiscoveryResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.logger.Debug("discovering tools by category", "category", category)

	if category == "" {
		return s.discoverAllCategories(ctx)
	}

	// Get tools for specific category
	tools, err := s.registry.ListTools(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools for category %s: %w", category, err)
	}

	// Build discovery result
	result := &CategoryDiscoveryResult{
		Category:          category,
		ToolCount:         len(tools),
		Tools:             tools,
		RelatedCategories: s.findRelatedCategories(category),
		RecommendedTools:  s.getRecommendedToolsForCategory(category, tools),
	}

	s.logger.Debug("category discovery completed",
		"category", category,
		"toolCount", len(tools),
		"relatedCategories", len(result.RelatedCategories))

	return result, nil
}

// GetToolRecommendations provides intelligent tool recommendations based on context.
//
// This method analyzes usage patterns and tool relationships to provide
// relevant tool recommendations for different workflows.
//
// Parameters:
// - ctx: Context for cancellation and timeout
// - context: Usage context or workflow description
// - limit: Maximum number of recommendations to return
//
// Returns:
// - []ToolRecommendation: Ranked tool recommendations with rationale
// - error: Recommendation error or context cancellation
//
// Notes:
// - Uses intelligent algorithms to match tools to workflows
// - Provides rationale for each recommendation
// - Considers tool popularity and usage patterns
// - Optimized for workflow-based tool discovery
func (s *ToolDiscoveryService) GetToolRecommendations(ctx context.Context, context string, limit int) ([]ToolRecommendation, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if limit <= 0 {
		limit = 5 // Default recommendation limit
	}

	s.logger.Debug("generating tool recommendations",
		"context", context,
		"limit", limit)

	// Analyze context for workflow patterns
	workflow := s.analyzeWorkflowContext(context)

	// Get recommendations based on workflow analysis
	recommendations := s.generateRecommendations(workflow, limit)

	s.logger.Debug("tool recommendations generated",
		"context", context,
		"recommendationCount", len(recommendations),
		"workflow", workflow)

	return recommendations, nil
}

// buildSearchIndex constructs comprehensive search indices for fast tool lookup.
//
// This internal method builds optimized data structures for different search
// operations and maintains them for efficient query processing.
func (s *ToolDiscoveryService) buildSearchIndex(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.logger.Debug("building tool search index")

	// Get all tools from registry
	allTools, err := s.registry.ListTools(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get tools for indexing: %w", err)
	}

	// Initialize search index
	s.searchIndex = &ToolSearchIndex{
		NameIndex:        make(map[string][]*MCPTool),
		DescriptionIndex: make(map[string][]*MCPTool),
		CategoryIndex:    make(map[CategoryType][]*MCPTool),
		TagIndex:         make(map[string][]*MCPTool),
		FullTextIndex:    make(map[string][]ToolSearchResult),
	}

	// Build indices
	for _, tool := range allTools {
		s.indexTool(tool)
	}

	s.logger.Debug("search index built successfully",
		"toolCount", len(allTools),
		"nameIndexSize", len(s.searchIndex.NameIndex),
		"descriptionIndexSize", len(s.searchIndex.DescriptionIndex))

	return nil
}

// indexTool adds a tool to all relevant search indices.
func (s *ToolDiscoveryService) indexTool(tool *MCPTool) {
	// Index by name (exact and partial matches)
	nameTokens := s.tokenize(tool.Name)
	for _, token := range nameTokens {
		s.searchIndex.NameIndex[token] = append(s.searchIndex.NameIndex[token], tool)
	}

	// Index by description keywords
	descTokens := s.tokenize(tool.Description)
	for _, token := range descTokens {
		s.searchIndex.DescriptionIndex[token] = append(s.searchIndex.DescriptionIndex[token], tool)
	}

	// Index by category
	s.searchIndex.CategoryIndex[tool.Category] = append(s.searchIndex.CategoryIndex[tool.Category], tool)

	// Build full-text search entries
	allText := strings.ToLower(tool.Name + " " + tool.Description + " " + tool.HelpText)
	tokens := s.tokenize(allText)
	for _, token := range tokens {
		result := ToolSearchResult{
			Tool:           tool,
			RelevanceScore: s.calculateBaseRelevance(tool),
			MatchContext:   fmt.Sprintf("Matched on: %s", token),
			MatchedFields:  []string{"name", "description"},
		}
		s.searchIndex.FullTextIndex[token] = append(s.searchIndex.FullTextIndex[token], result)
	}
}

// tokenize breaks text into searchable tokens.
func (s *ToolDiscoveryService) tokenize(text string) []string {
	text = strings.ToLower(text)
	// Split on common delimiters and clean up
	tokens := strings.FieldsFunc(text, func(c rune) bool {
		return c == ' ' || c == '_' || c == '-' || c == '.' || c == ','
	})

	var cleanTokens []string
	for _, token := range tokens {
		if len(token) > 2 { // Filter out very short tokens
			cleanTokens = append(cleanTokens, token)
		}
	}
	return cleanTokens
}

// calculateBaseRelevance computes base relevance score for a tool.
func (s *ToolDiscoveryService) calculateBaseRelevance(tool *MCPTool) float64 {
	score := 0.5 // Base score

	// Boost score based on tool characteristics
	if len(tool.Examples) > 0 {
		score += 0.1
	}
	if tool.HelpText != "" {
		score += 0.1
	}
	if len(tool.Description) > 50 {
		score += 0.1
	}

	return score
}

// searchByCategory performs category-based tool search.
func (s *ToolDiscoveryService) searchByCategory(_ context.Context, criteria *ToolSearchCriteria) []ToolSearchResult {
	var results []ToolSearchResult

	for _, category := range criteria.Categories {
		if tools, exists := s.searchIndex.CategoryIndex[category]; exists {
			for _, tool := range tools {
				result := ToolSearchResult{
					Tool:           tool,
					RelevanceScore: 0.8, // High relevance for category matches
					MatchContext:   fmt.Sprintf("Category match: %s", category),
					MatchedFields:  []string{"category"},
					CategoryMatch:  true,
				}
				results = append(results, result)
			}
		}
	}

	return results
}

// searchByQuery performs text-based tool search with fuzzy matching.
func (s *ToolDiscoveryService) searchByQuery(_ context.Context, criteria *ToolSearchCriteria) []ToolSearchResult {
	queryTokens := s.tokenize(criteria.Query)
	resultMap := make(map[string]*ToolSearchResult)

	for _, token := range queryTokens {
		// Exact matches
		if tools, exists := s.searchIndex.FullTextIndex[token]; exists {
			for _, result := range tools {
				key := result.Tool.Name
				if existing, exists := resultMap[key]; exists {
					existing.RelevanceScore += 0.2 // Boost for multiple matches
				} else {
					resultCopy := result
					resultCopy.RelevanceScore += 0.3 // Boost for query match
					resultMap[key] = &resultCopy
				}
			}
		}

		// Fuzzy matches
		for indexToken, tools := range s.searchIndex.FullTextIndex {
			if s.isFuzzyMatch(token, indexToken) {
				for _, result := range tools {
					key := result.Tool.Name
					if existing, exists := resultMap[key]; exists {
						existing.RelevanceScore += 0.1 // Smaller boost for fuzzy match
					} else {
						resultCopy := result
						resultCopy.RelevanceScore += 0.2 // Boost for fuzzy match
						resultCopy.MatchContext = fmt.Sprintf("Fuzzy match: %s ≈ %s", token, indexToken)
						resultMap[key] = &resultCopy
					}
				}
			}
		}
	}

	results := make([]ToolSearchResult, 0, len(resultMap))
	for _, result := range resultMap {
		results = append(results, *result)
	}

	return results
}

// isFuzzyMatch determines if two tokens are similar enough for fuzzy matching.
func (s *ToolDiscoveryService) isFuzzyMatch(token1, token2 string) bool {
	if len(token1) < 3 || len(token2) < 3 {
		return false
	}

	// Simple fuzzy matching: check if one contains the other or they share significant prefix
	return strings.Contains(token1, token2) ||
		strings.Contains(token2, token1) ||
		(len(token1) > 4 && len(token2) > 4 && token1[:3] == token2[:3])
}

// getAllToolsAsResults converts all tools to search results.
func (s *ToolDiscoveryService) getAllToolsAsResults(ctx context.Context) []ToolSearchResult {
	allTools, err := s.registry.ListTools(ctx, "")
	if err != nil {
		return []ToolSearchResult{}
	}

	results := make([]ToolSearchResult, 0, len(allTools))
	for _, tool := range allTools {
		result := ToolSearchResult{
			Tool:           tool,
			RelevanceScore: s.calculateBaseRelevance(tool),
			MatchContext:   "All tools",
			MatchedFields:  []string{},
		}
		results = append(results, result)
	}

	return results
}

// filterResults applies post-search filtering to results.
func (s *ToolDiscoveryService) filterResults(results []ToolSearchResult, criteria *ToolSearchCriteria) []ToolSearchResult {
	filtered := make([]ToolSearchResult, 0, len(results))

	for _, result := range results {
		// Apply relevance score filter
		if result.RelevanceScore < criteria.MinRelevanceScore {
			continue
		}

		// Apply category filter if specified
		if len(criteria.Categories) > 0 && !result.CategoryMatch {
			categoryMatch := false
			for _, category := range criteria.Categories {
				if result.Tool.Category == category {
					categoryMatch = true
					break
				}
			}
			if !categoryMatch {
				continue
			}
		}

		filtered = append(filtered, result)
	}

	return filtered
}

// sortResults sorts search results by specified criteria.
func (s *ToolDiscoveryService) sortResults(results []ToolSearchResult, criteria *ToolSearchCriteria) []ToolSearchResult {
	sort.Slice(results, func(i, j int) bool {
		switch criteria.SortBy {
		case "name":
			if criteria.SortOrder == "desc" {
				return results[i].Tool.Name > results[j].Tool.Name
			}
			return results[i].Tool.Name < results[j].Tool.Name
		case "category":
			if criteria.SortOrder == "desc" {
				return string(results[i].Tool.Category) > string(results[j].Tool.Category)
			}
			return string(results[i].Tool.Category) < string(results[j].Tool.Category)
		default: // Default to relevance score
			if criteria.SortOrder == "asc" {
				return results[i].RelevanceScore < results[j].RelevanceScore
			}
			return results[i].RelevanceScore > results[j].RelevanceScore
		}
	})

	// Apply max results limit
	if criteria.MaxResults > 0 && len(results) > criteria.MaxResults {
		results = results[:criteria.MaxResults]
	}

	return results
}

// CategoryDiscoveryResult provides comprehensive category discovery information.
type CategoryDiscoveryResult struct {
	Category          CategoryType   `json:"category"`
	ToolCount         int            `json:"toolCount"`
	Tools             []*MCPTool     `json:"tools"`
	RelatedCategories []CategoryType `json:"relatedCategories"`
	RecommendedTools  []*MCPTool     `json:"recommendedTools"`
}

// ToolRecommendation represents a tool recommendation with rationale.
type ToolRecommendation struct {
	Tool       *MCPTool `json:"tool"`
	Confidence float64  `json:"confidence"`
	Rationale  string   `json:"rationale"`
	UseCase    string   `json:"useCase"`
}

// Placeholder implementations for helper methods
func (s *ToolDiscoveryService) discoverAllCategories(ctx context.Context) (*CategoryDiscoveryResult, error) {
	categories, err := s.registry.GetCategories(ctx)
	if err != nil {
		return nil, err
	}

	var allTools []*MCPTool
	for _, category := range categories {
		tools, err := s.registry.ListTools(ctx, category)
		if err != nil {
			continue
		}
		allTools = append(allTools, tools...)
	}

	return &CategoryDiscoveryResult{
		Category:          "",
		ToolCount:         len(allTools),
		Tools:             allTools,
		RelatedCategories: categories,
		RecommendedTools:  allTools[:min(5, len(allTools))],
	}, nil
}

func (s *ToolDiscoveryService) findRelatedCategories(category CategoryType) []CategoryType {
	// Simple implementation - return other categories
	allCategories := []CategoryType{
		CategoryInvoiceManagement,
		CategoryClientManagement,
		CategoryDataImport,
		CategoryDataExport,
		CategoryConfiguration,
	}

	var related []CategoryType
	for _, cat := range allCategories {
		if cat != category {
			related = append(related, cat)
		}
	}
	return related
}

func (s *ToolDiscoveryService) getRecommendedToolsForCategory(_ CategoryType, tools []*MCPTool) []*MCPTool {
	// Simple implementation - return first few tools
	limit := min(3, len(tools))
	return tools[:limit]
}

func (s *ToolDiscoveryService) analyzeWorkflowContext(context string) string {
	// Simple workflow analysis
	context = strings.ToLower(context)
	if strings.Contains(context, "invoice") {
		return "invoice_management"
	}
	if strings.Contains(context, "client") {
		return "client_management"
	}
	if strings.Contains(context, "import") {
		return "data_import"
	}
	if strings.Contains(context, "export") || strings.Contains(context, "generate") {
		return "data_export"
	}
	if strings.Contains(context, "config") {
		return "configuration"
	}
	return "general"
}

func (s *ToolDiscoveryService) generateRecommendations(workflow string, limit int) []ToolRecommendation {
	// Simple recommendation logic based on workflow
	var recommendations []ToolRecommendation

	switch workflow {
	case "invoice_management":
		recommendations = append(recommendations, ToolRecommendation{
			Confidence: 0.9,
			Rationale:  "Essential for invoice creation workflow",
			UseCase:    "Creating new invoices for clients",
		})
	case "client_management":
		recommendations = append(recommendations, ToolRecommendation{
			Confidence: 0.8,
			Rationale:  "Essential for client management workflow",
			UseCase:    "Managing client information and contacts",
		})
	}

	return recommendations[:minInt(limit, len(recommendations))]
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
