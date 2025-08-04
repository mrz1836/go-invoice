package tools

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestToolSearchIndex tests the ToolSearchIndex structure
func TestToolSearchIndex(t *testing.T) {
	t.Run("BasicStructure", func(t *testing.T) {
		index := &ToolSearchIndex{
			NameIndex:        make(map[string][]*MCPTool),
			DescriptionIndex: make(map[string][]*MCPTool),
			CategoryIndex:    make(map[CategoryType][]*MCPTool),
			TagIndex:         make(map[string][]*MCPTool),
			FullTextIndex:    make(map[string][]ToolSearchResult),
		}

		assert.NotNil(t, index.NameIndex, "Name index should be initialized")
		assert.NotNil(t, index.DescriptionIndex, "Description index should be initialized")
		assert.NotNil(t, index.CategoryIndex, "Category index should be initialized")
		assert.NotNil(t, index.TagIndex, "Tag index should be initialized")
		assert.NotNil(t, index.FullTextIndex, "Full text index should be initialized")
		assert.Empty(t, index.NameIndex, "Name index should start empty")
		assert.Empty(t, index.DescriptionIndex, "Description index should start empty")
		assert.Empty(t, index.CategoryIndex, "Category index should start empty")
		assert.Empty(t, index.TagIndex, "Tag index should start empty")
		assert.Empty(t, index.FullTextIndex, "Full text index should start empty")
	})

	t.Run("IndexOperations", func(t *testing.T) {
		index := &ToolSearchIndex{
			NameIndex:        make(map[string][]*MCPTool),
			DescriptionIndex: make(map[string][]*MCPTool),
			CategoryIndex:    make(map[CategoryType][]*MCPTool),
			TagIndex:         make(map[string][]*MCPTool),
			FullTextIndex:    make(map[string][]ToolSearchResult),
		}

		// Create test tool
		tool := &MCPTool{
			Name:        "test_tool",
			Description: "A test tool for testing",
			Category:    CategoryInvoiceManagement,
			InputSchema: map[string]interface{}{"type": "object"},
			CLICommand:  "test",
			Version:     "1.0.0",
			Timeout:     30 * time.Second,
		}

		// Test adding to name index
		index.NameIndex["test"] = []*MCPTool{tool}
		assert.Len(t, index.NameIndex["test"], 1, "Name index should have one tool")
		assert.Equal(t, tool, index.NameIndex["test"][0], "Tool should be in name index")

		// Test adding to category index
		index.CategoryIndex[CategoryInvoiceManagement] = []*MCPTool{tool}
		assert.Len(t, index.CategoryIndex[CategoryInvoiceManagement], 1, "Category index should have one tool")
		assert.Equal(t, tool, index.CategoryIndex[CategoryInvoiceManagement][0], "Tool should be in category index")
	})
}

// TestToolSearchResult tests the ToolSearchResult structure
func TestToolSearchResult(t *testing.T) {
	tool := &MCPTool{
		Name:        "result_test_tool",
		Description: "Tool for result testing",
		Category:    CategoryClientManagement,
		InputSchema: map[string]interface{}{"type": "object"},
		CLICommand:  "result",
		Version:     "1.0.0",
		Timeout:     15 * time.Second,
	}

	t.Run("BasicStructure", func(t *testing.T) {
		result := ToolSearchResult{
			Tool:           tool,
			RelevanceScore: 0.85,
			MatchContext:   "Test match context",
			MatchedFields:  []string{"name", "description"},
			CategoryMatch:  true,
		}

		assert.Equal(t, tool, result.Tool, "Tool should be assigned correctly")
		assert.InDelta(t, 0.85, result.RelevanceScore, 0.001, "Relevance score should be set")
		assert.Equal(t, "Test match context", result.MatchContext, "Match context should be set")
		assert.Equal(t, []string{"name", "description"}, result.MatchedFields, "Matched fields should be set")
		assert.True(t, result.CategoryMatch, "Category match should be true")
	})

	t.Run("RelevanceScoreValidation", func(t *testing.T) {
		tests := []struct {
			name  string
			score float64
			valid bool
		}{
			{"ValidScore", 0.5, true},
			{"MinScore", 0.0, true},
			{"MaxScore", 1.0, true},
			{"InvalidNegative", -0.1, false},
			{"InvalidHigh", 1.1, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := ToolSearchResult{Tool: tool, RelevanceScore: tt.score}

				if tt.valid {
					assert.GreaterOrEqual(t, result.RelevanceScore, 0.0, "Valid score should be >= 0")
					assert.LessOrEqual(t, result.RelevanceScore, 1.0, "Valid score should be <= 1")
				} else {
					assert.True(t, result.RelevanceScore < 0.0 || result.RelevanceScore > 1.0,
						"Invalid score should be outside 0-1 range")
				}
			})
		}
	})
}

// TestToolSearchCriteria tests the ToolSearchCriteria structure
func TestToolSearchCriteria(t *testing.T) {
	t.Run("BasicStructure", func(t *testing.T) {
		criteria := &ToolSearchCriteria{
			Query:             "test search",
			Categories:        []CategoryType{CategoryInvoiceManagement, CategoryDataImport},
			Tags:              []string{"important", "test"},
			IncludeExamples:   true,
			MaxResults:        20,
			MinRelevanceScore: 0.3,
			SortBy:            "relevance",
			SortOrder:         "desc",
		}

		assert.Equal(t, "test search", criteria.Query, "Query should be set")
		assert.Len(t, criteria.Categories, 2, "Should have 2 categories")
		assert.Contains(t, criteria.Categories, CategoryInvoiceManagement, "Should contain invoice category")
		assert.Contains(t, criteria.Categories, CategoryDataImport, "Should contain import category")
		assert.Equal(t, []string{"important", "test"}, criteria.Tags, "Tags should be set")
		assert.True(t, criteria.IncludeExamples, "Include examples should be true")
		assert.Equal(t, 20, criteria.MaxResults, "Max results should be 20")
		assert.InDelta(t, 0.3, criteria.MinRelevanceScore, 0.001, "Min relevance score should be 0.3")
		assert.Equal(t, "relevance", criteria.SortBy, "Sort by should be relevance")
		assert.Equal(t, "desc", criteria.SortOrder, "Sort order should be desc")
	})

	t.Run("EmptyStructure", func(t *testing.T) {
		criteria := &ToolSearchCriteria{}

		assert.Empty(t, criteria.Query, "Empty query should be empty string")
		assert.Empty(t, criteria.Categories, "Empty categories should be nil/empty")
		assert.Empty(t, criteria.Tags, "Empty tags should be nil/empty")
		assert.False(t, criteria.IncludeExamples, "Include examples should default to false")
		assert.Equal(t, 0, criteria.MaxResults, "Max results should default to 0")
		assert.InDelta(t, 0.0, criteria.MinRelevanceScore, 0.001, "Min relevance score should default to 0")
		assert.Empty(t, criteria.SortBy, "Sort by should default to empty")
		assert.Empty(t, criteria.SortOrder, "Sort order should default to empty")
	})
}

// TestCategoryDiscoveryResult tests the CategoryDiscoveryResult structure
func TestCategoryDiscoveryResult(t *testing.T) {
	tools := []*MCPTool{
		{
			Name:        "discovery_tool_1",
			Description: "First discovery tool",
			Category:    CategoryConfiguration,
			InputSchema: map[string]interface{}{"type": "object"},
			CLICommand:  "discover1",
			Version:     "1.0.0",
			Timeout:     25 * time.Second,
		},
		{
			Name:        "discovery_tool_2",
			Description: "Second discovery tool",
			Category:    CategoryConfiguration,
			InputSchema: map[string]interface{}{"type": "object"},
			CLICommand:  "discover2",
			Version:     "1.0.0",
			Timeout:     25 * time.Second,
		},
	}

	result := &CategoryDiscoveryResult{
		Category:          CategoryConfiguration,
		ToolCount:         len(tools),
		Tools:             tools,
		RelatedCategories: []CategoryType{CategoryInvoiceManagement, CategoryClientManagement},
		RecommendedTools:  tools[:1],
	}

	assert.Equal(t, CategoryConfiguration, result.Category, "Category should be set correctly")
	assert.Equal(t, 2, result.ToolCount, "Tool count should match tools length")
	assert.Len(t, result.Tools, 2, "Should have 2 tools")
	assert.Equal(t, tools, result.Tools, "Tools should match input")
	assert.Len(t, result.RelatedCategories, 2, "Should have 2 related categories")
	assert.Contains(t, result.RelatedCategories, CategoryInvoiceManagement, "Should contain invoice category")
	assert.Contains(t, result.RelatedCategories, CategoryClientManagement, "Should contain client category")
	assert.Len(t, result.RecommendedTools, 1, "Should have 1 recommended tool")
	assert.Equal(t, tools[0], result.RecommendedTools[0], "Recommended tool should be first tool")
}

// TestToolRecommendation tests the ToolRecommendation structure
func TestToolRecommendation(t *testing.T) {
	tool := &MCPTool{
		Name:        "recommendation_tool",
		Description: "Tool for recommendation testing",
		Category:    CategoryDataExport,
		InputSchema: map[string]interface{}{"type": "object"},
		CLICommand:  "recommend",
		Version:     "1.0.0",
		Timeout:     40 * time.Second,
	}

	recommendation := ToolRecommendation{
		Tool:       tool,
		Confidence: 0.92,
		Rationale:  "This tool is highly recommended for export workflows",
		UseCase:    "Data export and report generation",
	}

	assert.Equal(t, tool, recommendation.Tool, "Tool should be assigned correctly")
	assert.InDelta(t, 0.92, recommendation.Confidence, 0.001, "Confidence should be 0.92")
	assert.Equal(t, "This tool is highly recommended for export workflows", recommendation.Rationale, "Rationale should be set")
	assert.Equal(t, "Data export and report generation", recommendation.UseCase, "Use case should be set")
}

// TestDiscoveryConstants tests discovery error constants
func TestDiscoveryConstants(t *testing.T) {
	t.Run("ErrorConstants", func(t *testing.T) {
		require.Error(t, ErrRegistryNil, "ErrRegistryNil should be defined")
		require.Error(t, ErrSearchCriteriaNil, "ErrSearchCriteriaNil should be defined")

		assert.Contains(t, ErrRegistryNil.Error(), "registry", "Error should mention registry")
		assert.Contains(t, ErrSearchCriteriaNil.Error(), "search criteria", "Error should mention search criteria")

		// Test that errors are actually error types
		err1 := ErrRegistryNil
		err2 := ErrSearchCriteriaNil
		require.Error(t, err1, "ErrRegistryNil should implement error interface")
		assert.Error(t, err2, "ErrSearchCriteriaNil should implement error interface")
	})
}

// TestMinFunction tests the min helper function used in discovery
func TestMinFunction(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"FirstSmaller", 3, 7, 3},
		{"SecondSmaller", 9, 4, 4},
		{"Equal", 5, 5, 5},
		{"ZeroFirst", 0, 10, 0},
		{"ZeroSecond", 15, 0, 0},
		{"Negative", -3, 2, -3},
		{"BothNegative", -8, -2, -8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result, "min(%d, %d) should be %d", tt.a, tt.b, tt.expected)
		})
	}
}

// TestMinIntFunction tests the minInt helper function
func TestMinIntFunction(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"FirstSmaller", 2, 8, 2},
		{"SecondSmaller", 12, 1, 1},
		{"Equal", 6, 6, 6},
		{"Large", 1000, 999, 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := minInt(tt.a, tt.b)
			assert.Equal(t, tt.expected, result, "minInt(%d, %d) should be %d", tt.a, tt.b, tt.expected)
		})
	}
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	t.Run("ImmediateCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		select {
		case <-ctx.Done():
			assert.Equal(t, context.Canceled, ctx.Err(), "Context should be canceled")
		default:
			t.Fatal("Context should be canceled")
		}
	})

	t.Run("TimeoutCancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		time.Sleep(5 * time.Millisecond) // Wait for timeout

		select {
		case <-ctx.Done():
			assert.Contains(t, ctx.Err().Error(), "deadline exceeded", "Context should timeout")
		default:
			t.Fatal("Context should have timed out")
		}
	})

	t.Run("ValidContext", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		select {
		case <-ctx.Done():
			t.Fatal("Context should not be canceled yet")
		default:
			// Context is still valid
			assert.NoError(t, ctx.Err(), "Context should not have error yet")
		}
	})
}

// TestToolSearchResultJSON tests JSON handling of search results
func TestToolSearchResultJSON(t *testing.T) {
	tool := &MCPTool{
		Name:        "json_test_tool",
		Description: "Tool for JSON testing",
		Category:    CategoryReporting,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"format": map[string]interface{}{
					"type": "string",
				},
			},
		},
		CLICommand: "json-test",
		Version:    "1.0.0",
		Timeout:    20 * time.Second,
	}

	result := ToolSearchResult{
		Tool:           tool,
		RelevanceScore: 0.75,
		MatchContext:   "JSON serialization test",
		MatchedFields:  []string{"name", "category", "description"},
		CategoryMatch:  false,
	}

	// Test that all fields are accessible (simulating JSON marshaling concerns)
	assert.NotNil(t, result.Tool, "Tool should not be nil")
	assert.Equal(t, "json_test_tool", result.Tool.Name, "Tool name should be accessible")
	assert.Equal(t, CategoryReporting, result.Tool.Category, "Tool category should be accessible")
	assert.InDelta(t, 0.75, result.RelevanceScore, 0.001, "Relevance score should be accessible")
	assert.Equal(t, "JSON serialization test", result.MatchContext, "Match context should be accessible")
	assert.Equal(t, []string{"name", "category", "description"}, result.MatchedFields, "Matched fields should be accessible")
	assert.False(t, result.CategoryMatch, "Category match should be accessible")
}

// DiscoveryServiceTestSuite provides comprehensive tests for ToolDiscoveryService
type DiscoveryServiceTestSuite struct {
	suite.Suite

	mockRegistry *MockToolRegistry
	mockLogger   *MockLogger
	service      *ToolDiscoveryService
}

// getContext returns a background context for testing
func (s *DiscoveryServiceTestSuite) getContext() context.Context {
	return context.Background()
}

func (s *DiscoveryServiceTestSuite) SetupTest() {
	s.mockRegistry = NewMockToolRegistry()
	s.mockLogger = NewMockLogger()

	// Create sample tools for testing
	tools := []*MCPTool{
		{
			Name:        "create_invoice",
			Description: "Create a new invoice for clients",
			Category:    CategoryInvoiceManagement,
			InputSchema: map[string]interface{}{"type": "object"},
			CLICommand:  "create-invoice",
			Version:     "1.0.0",
			Timeout:     30 * time.Second,
			HelpText:    "Creates invoices with line items",
			Examples: []MCPToolExample{
				{Description: "Basic invoice", UseCase: "Invoice creation"},
			},
		},
		{
			Name:        "import_csv",
			Description: "Import timesheet data from CSV files",
			Category:    CategoryDataImport,
			InputSchema: map[string]interface{}{"type": "object"},
			CLICommand:  "import-csv",
			Version:     "1.0.0",
			Timeout:     45 * time.Second,
			HelpText:    "Imports CSV data for processing",
		},
		{
			Name:        "generate_report",
			Description: "Generate comprehensive reports",
			Category:    CategoryReporting,
			InputSchema: map[string]interface{}{"type": "object"},
			CLICommand:  "generate-report",
			Version:     "1.0.0",
			Timeout:     60 * time.Second,
		},
	}

	// Set up mock registry responses
	s.mockRegistry.SetListToolsResponse("", tools) // All tools
	s.mockRegistry.SetListToolsResponse(CategoryInvoiceManagement, tools[:1])
	s.mockRegistry.SetListToolsResponse(CategoryDataImport, tools[1:2])
	s.mockRegistry.SetListToolsResponse(CategoryReporting, tools[2:3])

	// Initialize service
	var err error
	s.service, err = NewToolDiscoveryService(s.getContext(), s.mockRegistry, s.mockLogger)
	s.Require().NoError(err)
	s.NotNil(s.service)
}

// TestNewToolDiscoveryService tests service construction
func (s *DiscoveryServiceTestSuite) TestNewToolDiscoveryService() {
	s.Run("ValidConstruction", func() {
		registry := NewMockToolRegistry()
		logger := NewMockLogger()

		service, err := NewToolDiscoveryService(s.getContext(), registry, logger)

		s.Require().NoError(err)
		s.NotNil(service)
		s.Equal(registry, service.registry)
		s.Equal(logger, service.logger)
		s.NotNil(service.searchIndex)
	})

	s.Run("NilRegistry", func() {
		logger := NewMockLogger()

		service, err := NewToolDiscoveryService(s.getContext(), nil, logger)

		s.Require().Error(err)
		s.Nil(service)
		s.Equal(ErrRegistryNil, err)
	})

	s.Run("NilLogger", func() {
		registry := NewMockToolRegistry()

		service, err := NewToolDiscoveryService(s.getContext(), registry, nil)

		s.Require().Error(err)
		s.Nil(service)
		s.Contains(err.Error(), "logger")
	})

	s.Run("ContextCancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		registry := NewMockToolRegistry()
		logger := NewMockLogger()

		service, err := NewToolDiscoveryService(ctx, registry, logger)

		s.Require().Error(err)
		s.Nil(service)
		s.Equal(context.Canceled, err)
	})
}

// TestSearchTools tests the main search functionality
func (s *DiscoveryServiceTestSuite) TestSearchTools() {
	s.Run("QuerySearch", func() {
		criteria := &ToolSearchCriteria{
			Query:      "invoice",
			MaxResults: 10,
		}

		results, err := s.service.SearchTools(s.getContext(), criteria)

		s.Require().NoError(err)
		s.NotEmpty(results)

		// Should find invoice-related tools
		found := false
		for _, result := range results {
			if result.Tool.Name == "create_invoice" {
				found = true
				s.Greater(result.RelevanceScore, 0.0)
				s.NotEmpty(result.MatchContext)
				break
			}
		}
		s.True(found, "Should find create_invoice tool")
	})

	s.Run("CategorySearch", func() {
		criteria := &ToolSearchCriteria{
			Categories: []CategoryType{CategoryDataImport},
			MaxResults: 5,
		}

		results, err := s.service.SearchTools(s.getContext(), criteria)

		s.Require().NoError(err)
		s.NotEmpty(results)

		// All results should be from the specified category
		for _, result := range results {
			s.Equal(CategoryDataImport, result.Tool.Category)
			s.True(result.CategoryMatch)
		}
	})

	s.Run("CombinedSearch", func() {
		criteria := &ToolSearchCriteria{
			Query:      "data",
			Categories: []CategoryType{CategoryDataImport, CategoryReporting},
			MaxResults: 10,
			SortBy:     "relevance",
			SortOrder:  "desc",
		}

		results, err := s.service.SearchTools(s.getContext(), criteria)

		s.Require().NoError(err)
		s.NotEmpty(results)
	})

	s.Run("EmptyResults", func() {
		criteria := &ToolSearchCriteria{
			Query: "nonexistent_tool_name",
		}

		results, err := s.service.SearchTools(s.getContext(), criteria)

		s.Require().NoError(err)
		s.Empty(results)
	})

	s.Run("NilCriteria", func() {
		results, err := s.service.SearchTools(s.getContext(), nil)

		s.Require().Error(err)
		s.Nil(results)
		s.Equal(ErrSearchCriteriaNil, err)
	})

	s.Run("ContextCancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		criteria := &ToolSearchCriteria{Query: "test"}

		results, err := s.service.SearchTools(ctx, criteria)

		s.Require().Error(err)
		s.Nil(results)
		s.Equal(context.Canceled, err)
	})
}

// TestDiscoverToolsByCategory tests category-based discovery
func (s *DiscoveryServiceTestSuite) TestDiscoverToolsByCategory() {
	s.Run("ValidCategory", func() {
		result, err := s.service.DiscoverToolsByCategory(s.getContext(), CategoryInvoiceManagement)

		s.Require().NoError(err)
		s.NotNil(result)
		s.Equal(CategoryInvoiceManagement, result.Category)
		s.Equal(1, result.ToolCount)
		s.Len(result.Tools, 1)
		s.Equal("create_invoice", result.Tools[0].Name)
	})

	s.Run("EmptyCategory", func() {
		s.mockRegistry.SetListToolsResponse(CategoryConfiguration, []*MCPTool{})

		result, err := s.service.DiscoverToolsByCategory(s.getContext(), CategoryConfiguration)

		s.Require().NoError(err)
		s.NotNil(result)
		s.Equal(CategoryConfiguration, result.Category)
		s.Equal(0, result.ToolCount)
		s.Empty(result.Tools)
	})

	s.Run("ContextCancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		result, err := s.service.DiscoverToolsByCategory(ctx, CategoryInvoiceManagement)

		s.Require().Error(err)
		s.Nil(result)
		s.Equal(context.Canceled, err)
	})
}

// TestGetToolRecommendations tests recommendation functionality
func (s *DiscoveryServiceTestSuite) TestGetToolRecommendations() {
	s.Run("BasicRecommendations", func() {
		recommendations, err := s.service.GetToolRecommendations(s.getContext(), "create invoice for client", 3)

		s.Require().NoError(err)
		s.NotEmpty(recommendations)
	})

	s.Run("EmptyContext", func() {
		recommendations, err := s.service.GetToolRecommendations(s.getContext(), "", 3)

		s.Require().NoError(err)
		// Should still return some recommendations
		s.LessOrEqual(len(recommendations), 3)
	})

	s.Run("ZeroLimit", func() {
		recommendations, err := s.service.GetToolRecommendations(s.getContext(), "test", 0)

		s.Require().NoError(err)
		s.LessOrEqual(len(recommendations), 5) // Should default to 5
	})

	s.Run("ContextCancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		recommendations, err := s.service.GetToolRecommendations(ctx, "test", 3)

		s.Require().Error(err)
		s.Nil(recommendations)
		s.Equal(context.Canceled, err)
	})
}

func TestDiscoveryServiceTestSuite(t *testing.T) {
	suite.Run(t, new(DiscoveryServiceTestSuite))
}

// TestToolStructureValidation tests tool structure validation for discovery
func TestToolStructureValidation(t *testing.T) {
	t.Run("ValidTool", func(t *testing.T) {
		tool := &MCPTool{
			Name:        "valid_tool",
			Description: "A valid tool for testing",
			Category:    CategoryInvoiceManagement,
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"param": map[string]interface{}{
						"type": "string",
					},
				},
			},
			CLICommand: "valid-tool",
			Version:    "1.0.0",
			Timeout:    30 * time.Second,
		}

		// Validate structure
		assert.NotEmpty(t, tool.Name, "Tool should have a name")
		assert.NotEmpty(t, tool.Description, "Tool should have a description")
		assert.NotNil(t, tool.InputSchema, "Tool should have input schema")
		assert.NotEmpty(t, tool.CLICommand, "Tool should have CLI command")
		assert.NotEmpty(t, tool.Version, "Tool should have version")
		assert.Greater(t, tool.Timeout, time.Duration(0), "Tool should have positive timeout")

		// Validate schema structure
		schemaType, hasType := tool.InputSchema["type"]
		assert.True(t, hasType, "Schema should have type")
		assert.Equal(t, "object", schemaType, "Schema type should be object")
	})

	t.Run("ToolWithExamples", func(t *testing.T) {
		tool := &MCPTool{
			Name:        "example_tool",
			Description: "Tool with examples",
			Category:    CategoryClientManagement,
			InputSchema: map[string]interface{}{"type": "object"},
			Examples: []MCPToolExample{
				{
					Description:    "Basic usage example",
					Input:          map[string]interface{}{"name": "test"},
					ExpectedOutput: "Success",
					UseCase:        "Testing",
				},
			},
			CLICommand: "example-tool",
			Version:    "1.0.0",
			Timeout:    25 * time.Second,
		}

		assert.Len(t, tool.Examples, 1, "Tool should have one example")
		example := tool.Examples[0]
		assert.NotEmpty(t, example.Description, "Example should have description")
		assert.NotNil(t, example.Input, "Example should have input")
		assert.NotEmpty(t, example.ExpectedOutput, "Example should have expected output")
		assert.NotEmpty(t, example.UseCase, "Example should have use case")
	})
}
