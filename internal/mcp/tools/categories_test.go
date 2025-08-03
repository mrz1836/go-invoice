package tools

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// CategoriesTestSuite provides comprehensive tests for the category system
type CategoriesTestSuite struct {
	suite.Suite

	// Test context managed per test method
}

func (s *CategoriesTestSuite) SetupTest() {
	// Test setup if needed
}

func (s *CategoriesTestSuite) TestCategoryConstants() {
	s.Run("CategoryValues", func() {
		expectedCategories := map[CategoryType]string{
			CategoryInvoiceManagement: "invoice_management",
			CategoryDataImport:        "data_import",
			CategoryDataExport:        "data_export",
			CategoryClientManagement:  "client_management",
			CategoryConfiguration:     "configuration",
			CategoryReporting:         "reporting",
		}

		for category, expectedValue := range expectedCategories {
			s.Equal(expectedValue, string(category), "Category %s should have correct string value", category)
			s.NotEmpty(string(category), "Category should not be empty")
			// Verify categories use lowercase with underscores for multi-word names
			categoryStr := string(category)
			s.Equal(strings.ToLower(categoryStr), categoryStr, "Category should be lowercase")

			// Multi-word categories should use underscores
			if strings.Contains(expectedValue, " ") || len(strings.Fields(expectedValue)) > 1 {
				s.Contains(categoryStr, "_", "Multi-word category should use snake_case format")
			}
		}
	})

	s.Run("CategoryUniqueness", func() {
		categories := []CategoryType{
			CategoryInvoiceManagement,
			CategoryDataImport,
			CategoryDataExport,
			CategoryClientManagement,
			CategoryConfiguration,
			CategoryReporting,
		}

		categorySet := make(map[string]bool)
		for _, category := range categories {
			categoryStr := string(category)
			s.False(categorySet[categoryStr], "Category %s should be unique", categoryStr)
			categorySet[categoryStr] = true
		}

		s.Equal(6, len(categorySet), "Should have exactly 6 unique categories")
	})

	s.Run("CategoryNamingConventions", func() {
		categories := []CategoryType{
			CategoryInvoiceManagement,
			CategoryDataImport,
			CategoryDataExport,
			CategoryClientManagement,
			CategoryConfiguration,
			CategoryReporting,
		}

		for _, category := range categories {
			categoryStr := string(category)

			// Should use snake_case
			s.NotContains(categoryStr, " ", "Category should not contain spaces")
			s.NotContains(categoryStr, "-", "Category should not contain hyphens")
			s.Equal(categoryStr, strings.ToLower(categoryStr), "Category should be lowercase")

			// Should be descriptive
			s.GreaterOrEqual(len(categoryStr), 5, "Category name should be descriptive")
			s.LessOrEqual(len(categoryStr), 30, "Category name should not be too long")
		}
	})
}

func (s *CategoriesTestSuite) TestCategoryGrouping() {
	s.Run("BusinessFunctionGrouping", func() {
		// Test that categories group related business functions logically

		// Core business operations
		coreCategories := []CategoryType{
			CategoryInvoiceManagement,
			CategoryClientManagement,
		}
		for _, category := range coreCategories {
			s.Contains(string(category), "management", "Core business categories should contain 'management'")
		}

		// Data operations
		dataCategories := []CategoryType{
			CategoryDataImport,
			CategoryDataExport,
		}
		for _, category := range dataCategories {
			s.Contains(string(category), "data", "Data operation categories should contain 'data'")
		}

		// System operations
		systemCategories := []CategoryType{
			CategoryConfiguration,
			CategoryReporting,
		}
		for _, category := range systemCategories {
			categoryStr := string(category)
			s.True(
				category == CategoryConfiguration || category == CategoryReporting,
				"System category %s should be configuration or reporting", categoryStr,
			)
		}
	})

	s.Run("CategoryCoverage", func() {
		// Ensure categories cover all major functional areas of the invoice system

		functionalAreas := map[string]CategoryType{
			"invoice_operations":   CategoryInvoiceManagement,
			"client_operations":    CategoryClientManagement,
			"data_input":           CategoryDataImport,
			"data_output":          CategoryDataExport,
			"system_configuration": CategoryConfiguration,
			"business_reporting":   CategoryReporting,
		}

		s.Equal(6, len(functionalAreas), "Should cover 6 major functional areas")

		for area, category := range functionalAreas {
			s.NotEmpty(string(category), "Functional area %s should have non-empty category", area)
		}
	})
}

func (s *CategoriesTestSuite) TestCategoryUsagePatterns() {
	s.Run("CategoryAsMapKey", func() {
		// Test that categories work properly as map keys
		categoryMap := make(map[CategoryType]int)

		categories := []CategoryType{
			CategoryInvoiceManagement,
			CategoryDataImport,
			CategoryDataExport,
			CategoryClientManagement,
			CategoryConfiguration,
			CategoryReporting,
		}

		for i, category := range categories {
			categoryMap[category] = i
		}

		s.Equal(6, len(categoryMap), "All categories should be stored as unique keys")

		for i, category := range categories {
			value, exists := categoryMap[category]
			s.True(exists, "Category %s should exist in map", category)
			s.Equal(i, value, "Category %s should have correct value", category)
		}
	})

	s.Run("CategoryComparison", func() {
		// Test category comparison operations
		s.Equal(CategoryInvoiceManagement, CategoryInvoiceManagement, "Same categories should be equal")
		s.NotEqual(CategoryInvoiceManagement, CategoryDataImport, "Different categories should not be equal")

		// Test string comparison
		s.Equal("invoice_management", string(CategoryInvoiceManagement))
		s.NotEqual("invoice_management", string(CategoryDataImport))
	})

	s.Run("CategoryConversion", func() {
		// Test conversion between CategoryType and string
		testCases := []struct {
			category CategoryType
			str      string
		}{
			{CategoryInvoiceManagement, "invoice_management"},
			{CategoryDataImport, "data_import"},
			{CategoryDataExport, "data_export"},
			{CategoryClientManagement, "client_management"},
			{CategoryConfiguration, "configuration"},
			{CategoryReporting, "reporting"},
		}

		for _, tc := range testCases {
			s.Equal(tc.str, string(tc.category), "Category to string conversion should work")
			s.Equal(tc.category, CategoryType(tc.str), "String to category conversion should work")
		}
	})
}

func (s *CategoriesTestSuite) TestCategoryValidation() {
	s.Run("ValidCategoryValidation", func() {
		validCategories := []CategoryType{
			CategoryInvoiceManagement,
			CategoryDataImport,
			CategoryDataExport,
			CategoryClientManagement,
			CategoryConfiguration,
			CategoryReporting,
		}

		for _, category := range validCategories {
			// This would be tested by the registry's isValidCategory method
			s.NotEmpty(string(category), "Valid category should not be empty")
			s.True(s.isValidCategoryFormat(category), "Category should follow valid format")
		}
	})

	s.Run("InvalidCategoryValidation", func() {
		invalidCategories := []CategoryType{
			CategoryType(""),
			CategoryType("invalid_category"),
			CategoryType("UPPERCASE_CATEGORY"),
			CategoryType("category with spaces"),
			CategoryType("category-with-hyphens"),
		}

		for _, category := range invalidCategories {
			if category == CategoryType("") {
				s.Empty(string(category), "Empty category should be empty")
			} else {
				s.False(s.isKnownCategory(category), "Invalid category should not be recognized")
			}
		}
	})
}

func (s *CategoriesTestSuite) TestCategoryContextUsage() {
	s.Run("CategoryInToolDefinition", func() {
		// Test how categories are used in tool definitions
		for _, category := range s.getAllValidCategories() {
			// Simulate tool creation with category
			tool := &MCPTool{
				Name:        "test_tool",
				Description: "Test tool",
				InputSchema: map[string]interface{}{"type": "object"},
				Category:    category,
				CLICommand:  "test",
				Version:     "1.0.0",
				Timeout:     10000000000, // 10 seconds in nanoseconds
			}

			s.Equal(category, tool.Category, "Tool should retain assigned category")
			s.True(s.isKnownCategory(tool.Category), "Tool category should be valid")
		}
	})

	s.Run("CategoryFiltering", func() {
		// Test category filtering scenarios
		allCategories := s.getAllValidCategories()

		// Test individual category filtering
		for _, category := range allCategories {
			s.True(s.categoryMatchesFilter(category, category), "Category should match itself")
		}

		// Test empty filter (should match all)
		emptyFilter := CategoryType("")
		for _, category := range allCategories {
			s.True(s.categoryMatchesFilter(category, emptyFilter), "All categories should match empty filter")
		}
	})
}

func (s *CategoriesTestSuite) TestCategoryMetadata() {
	s.Run("CategoryDescriptions", func() {
		// Test that each category has a clear purpose
		categoryPurposes := map[CategoryType]string{
			CategoryInvoiceManagement: "creating, updating, and managing invoices",
			CategoryDataImport:        "importing timesheet and client data",
			CategoryDataExport:        "generating and exporting invoice documents",
			CategoryClientManagement:  "managing client information",
			CategoryConfiguration:     "system configuration and validation",
			CategoryReporting:         "analytics and reporting",
		}

		for category, purpose := range categoryPurposes {
			s.NotEmpty(purpose, "Category %s should have a clear purpose", category)
			s.GreaterOrEqual(len(purpose), 20, "Category purpose should be descriptive")
		}
	})

	s.Run("CategoryToolCount", func() {
		// Test expected tool distribution across categories
		expectedToolCounts := map[CategoryType]int{
			CategoryInvoiceManagement: 7, // Based on invoice_tools.go
			CategoryClientManagement:  4, // Expected client management tools
			CategoryDataImport:        3, // Expected import tools
			CategoryDataExport:        2, // Expected export/generation tools
			CategoryConfiguration:     3, // Expected configuration tools
			CategoryReporting:         1, // Expected reporting tools
		}

		for category, expectedCount := range expectedToolCounts {
			s.Greater(expectedCount, 0, "Category %s should have at least one tool", category)
			s.LessOrEqual(expectedCount, 10, "Category %s should not have too many tools", category)
		}
	})
}

func (s *CategoriesTestSuite) TestCategoryEvolution() {
	s.Run("CategoryExtensibility", func() {
		// Test that the category system can be extended
		currentCategoryCount := 6
		s.Equal(currentCategoryCount, len(s.getAllValidCategories()), "Should have expected number of categories")

		// Future categories that might be added
		potentialFutureCategories := []string{
			"workflow_automation",
			"integration_management",
			"audit_logging",
			"security_management",
			"template_management",
		}

		for _, futureCategory := range potentialFutureCategories {
			// Verify they follow naming conventions
			s.Contains(futureCategory, "_", "Future category should use snake_case")
			s.Equal(futureCategory, strings.ToLower(futureCategory), "Future category should be lowercase")
			s.GreaterOrEqual(len(futureCategory), 5, "Future category should be descriptive")
		}
	})

	s.Run("CategoryBackwardCompatibility", func() {
		// Test that existing categories maintain their values
		// This ensures backward compatibility when categories are referenced by string value

		legacyMappings := map[string]CategoryType{
			"invoice_management": CategoryInvoiceManagement,
			"data_import":        CategoryDataImport,
			"data_export":        CategoryDataExport,
			"client_management":  CategoryClientManagement,
			"configuration":      CategoryConfiguration,
			"reporting":          CategoryReporting,
		}

		for legacyStr, expectedCategory := range legacyMappings {
			s.Equal(legacyStr, string(expectedCategory), "Legacy category string should match")
			s.Equal(expectedCategory, CategoryType(legacyStr), "Legacy string should convert to category")
		}
	})
}

// Helper methods for testing

func (s *CategoriesTestSuite) getAllValidCategories() []CategoryType {
	return []CategoryType{
		CategoryInvoiceManagement,
		CategoryDataImport,
		CategoryDataExport,
		CategoryClientManagement,
		CategoryConfiguration,
		CategoryReporting,
	}
}

func (s *CategoriesTestSuite) isKnownCategory(category CategoryType) bool {
	validCategories := s.getAllValidCategories()
	for _, validCategory := range validCategories {
		if category == validCategory {
			return true
		}
	}
	return false
}

func (s *CategoriesTestSuite) isValidCategoryFormat(category CategoryType) bool {
	categoryStr := string(category)

	// Check format requirements
	if len(categoryStr) == 0 {
		return false
	}

	// Should be lowercase
	if categoryStr != strings.ToLower(categoryStr) {
		return false
	}

	// Should not contain spaces or hyphens (allow single words or snake_case)
	if strings.Contains(categoryStr, " ") || strings.Contains(categoryStr, "-") {
		return false
	}

	// Only alphanumeric characters and underscores allowed
	for _, char := range categoryStr {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

func (s *CategoriesTestSuite) categoryMatchesFilter(category CategoryType, filter CategoryType) bool {
	// Empty filter matches all categories
	if filter == CategoryType("") {
		return true
	}

	// Otherwise, exact match required
	return category == filter
}

// TestCategoriesTestSuite runs the complete categories test suite
func TestCategoriesTestSuite(t *testing.T) {
	suite.Run(t, new(CategoriesTestSuite))
}

// Unit tests for specific category behaviors
func TestCategoryType_Behaviors(t *testing.T) {
	t.Run("CategoryAsString", func(t *testing.T) {
		category := CategoryInvoiceManagement
		str := string(category)
		assert.Equal(t, "invoice_management", str)
		assert.NotEmpty(t, str)
	})

	t.Run("CategoryComparison", func(t *testing.T) {
		assert.Equal(t, CategoryInvoiceManagement, CategoryInvoiceManagement)
		assert.NotEqual(t, CategoryInvoiceManagement, CategoryDataImport)

		// Test with type conversion
		assert.Equal(t, CategoryInvoiceManagement, CategoryType("invoice_management"))
		assert.NotEqual(t, CategoryInvoiceManagement, CategoryType("data_import"))
	})

	t.Run("CategoryInMap", func(t *testing.T) {
		categoryMap := map[CategoryType]string{
			CategoryInvoiceManagement: "Invoice tools",
			CategoryDataImport:        "Import tools",
			CategoryDataExport:        "Export tools",
		}

		assert.Equal(t, "Invoice tools", categoryMap[CategoryInvoiceManagement])
		assert.Equal(t, "Import tools", categoryMap[CategoryDataImport])
		assert.Equal(t, "Export tools", categoryMap[CategoryDataExport])
		assert.Equal(t, "", categoryMap[CategoryConfiguration]) // Not in map
	})

	t.Run("CategoryInSlice", func(t *testing.T) {
		categories := []CategoryType{
			CategoryInvoiceManagement,
			CategoryDataImport,
			CategoryConfiguration,
		}

		assert.Len(t, categories, 3)
		assert.Contains(t, categories, CategoryInvoiceManagement)
		assert.Contains(t, categories, CategoryDataImport)
		assert.NotContains(t, categories, CategoryReporting)
	})
}

func TestCategoryValidation_EdgeCases(t *testing.T) {
	t.Run("EmptyCategory", func(t *testing.T) {
		empty := CategoryType("")
		assert.Empty(t, string(empty))
		assert.Equal(t, "", string(empty))
	})

	t.Run("WhitespaceCategory", func(t *testing.T) {
		whitespace := CategoryType("   ")
		assert.Equal(t, "   ", string(whitespace))
		assert.NotEqual(t, "", string(whitespace))
	})

	t.Run("UnicodeCategory", func(t *testing.T) {
		unicode := CategoryType("配置_管理")
		assert.Equal(t, "配置_管理", string(unicode))
		assert.NotEmpty(t, string(unicode))
	})

	t.Run("LongCategory", func(t *testing.T) {
		long := CategoryType("very_long_category_name_that_exceeds_normal_length_expectations")
		assert.Greater(t, len(string(long)), 50)
		assert.NotEmpty(t, string(long))
	})
}

// Benchmark tests for category operations
func BenchmarkCategoryType_String(b *testing.B) {
	category := CategoryInvoiceManagement
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = string(category)
	}
}

func BenchmarkCategoryType_Comparison(b *testing.B) {
	category1 := CategoryInvoiceManagement
	category2 := CategoryDataImport
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = category1 == category2
	}
}

func BenchmarkCategoryType_MapLookup(b *testing.B) {
	categoryMap := map[CategoryType]string{
		CategoryInvoiceManagement: "Invoice tools",
		CategoryDataImport:        "Import tools",
		CategoryDataExport:        "Export tools",
		CategoryClientManagement:  "Client tools",
		CategoryConfiguration:     "Config tools",
		CategoryReporting:         "Report tools",
	}

	category := CategoryInvoiceManagement
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = categoryMap[category]
	}
}
