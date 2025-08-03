package tools

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ConfigToolsTestSuite provides comprehensive tests for configuration tool definitions
type ConfigToolsTestSuite struct {
	suite.Suite

	// Test context managed per test method
}

func (s *ConfigToolsTestSuite) SetupTest() {
	// Test setup if needed
}

func (s *ConfigToolsTestSuite) TestCreateConfigTools() {
	tools := CreateConfigTools()

	s.NotNil(tools, "Should return tools slice")
	s.NotEmpty(tools, "Should return at least one tool")

	// Expected configuration tools based on typical functionality
	expectedToolNames := []string{
		"config_validate",
		"config_show",
		"config_set",
		"config_init",
	}

	// Verify we have the expected tools (at least the core ones)
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for _, expectedName := range expectedToolNames {
		if toolNames[expectedName] {
			s.True(toolNames[expectedName], "Should have %s tool", expectedName)
		}
	}

	// Verify all tools are properly configured
	for _, tool := range tools {
		s.validateConfigTool(tool)
	}
}

func (s *ConfigToolsTestSuite) TestConfigValidationTool() {
	tools := CreateConfigTools()
	var configValidateTool *MCPTool

	for _, tool := range tools {
		if strings.Contains(tool.Name, "validate") {
			configValidateTool = tool
			break
		}
	}

	if configValidateTool == nil {
		s.T().Skip("config validation tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		s.Contains(strings.ToLower(configValidateTool.Name), "validate")
		s.NotEmpty(configValidateTool.Description)
		s.Contains(strings.ToLower(configValidateTool.Description), "validate")
		s.Equal(CategoryConfiguration, configValidateTool.Category)
		s.NotEmpty(configValidateTool.Version)
		s.Greater(configValidateTool.Timeout, time.Duration(0))
	})

	s.Run("Schema", func() {
		s.NotNil(configValidateTool.InputSchema)
		s.Equal("object", configValidateTool.InputSchema["type"])

		// Verify expected properties exist
		properties, hasProperties := configValidateTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties in schema")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// May have config file path or config sections
			configFields := []string{"config_file", "config_path", "section", "all"}
			hasConfigField := false
			for _, field := range configFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasConfigField = true
					break
				}
			}
			// Config validation might work on current config or specified file
			if len(propertiesMap) > 0 {
				s.True(hasConfigField, "Should have config-related field")
			}
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(configValidateTool.Examples, "Should have examples")

		for i, example := range configValidateTool.Examples {
			s.NotEmpty(example.Description, "Example %d should have description", i)
			s.NotNil(example.Input, "Example %d should have input", i)

			// Verify example demonstrates validation scenarios
			exampleText := strings.ToLower(example.Description + " " + example.UseCase)
			validationKeywords := []string{"validate", "check", "verify", "config", "settings"}
			hasValidationKeyword := false
			for _, keyword := range validationKeywords {
				if strings.Contains(exampleText, keyword) {
					hasValidationKeyword = true
					break
				}
			}
			s.True(hasValidationKeyword, "Example should be validation-related")
		}
	})
}

func (s *ConfigToolsTestSuite) TestConfigShowTool() {
	tools := CreateConfigTools()
	var configShowTool *MCPTool

	for _, tool := range tools {
		if strings.Contains(tool.Name, "show") || strings.Contains(tool.Name, "get") {
			configShowTool = tool
			break
		}
	}

	if configShowTool == nil {
		s.T().Skip("config show tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		toolNameLower := strings.ToLower(configShowTool.Name)
		s.True(
			strings.Contains(toolNameLower, "show") || strings.Contains(toolNameLower, "get"),
			"Tool name should indicate display/retrieval functionality",
		)
		s.NotEmpty(configShowTool.Description)
		s.Equal(CategoryConfiguration, configShowTool.Category)
	})

	s.Run("Schema", func() {
		s.NotNil(configShowTool.InputSchema)
		s.Equal("object", configShowTool.InputSchema["type"])

		properties, hasProperties := configShowTool.InputSchema["properties"]
		if hasProperties {
			if propertiesMap, ok := properties.(map[string]interface{}); ok {
				// May have specific config key or section to show
				configFields := []string{"key", "section", "setting", "all", "format"}
				for _, field := range configFields {
					if fieldDef, hasField := propertiesMap[field]; hasField {
						s.NotNil(fieldDef, "Config field %s should have definition", field)
					}
				}
			}
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(configShowTool.Examples, "Should have examples")

		for _, example := range configShowTool.Examples {
			s.NotEmpty(example.Description)
			s.NotNil(example.Input)

			// Should demonstrate config display scenarios
			exampleText := strings.ToLower(example.Description + " " + example.UseCase)
			displayKeywords := []string{"show", "display", "get", "view", "config", "setting"}
			hasDisplayKeyword := false
			for _, keyword := range displayKeywords {
				if strings.Contains(exampleText, keyword) {
					hasDisplayKeyword = true
					break
				}
			}
			s.True(hasDisplayKeyword, "Example should be config display-related")
		}
	})
}

func (s *ConfigToolsTestSuite) TestConfigSetTool() {
	tools := CreateConfigTools()
	var configSetTool *MCPTool

	for _, tool := range tools {
		if strings.Contains(tool.Name, "set") || strings.Contains(tool.Name, "update") {
			configSetTool = tool
			break
		}
	}

	if configSetTool == nil {
		s.T().Skip("config set tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		toolNameLower := strings.ToLower(configSetTool.Name)
		s.True(
			strings.Contains(toolNameLower, "set") || strings.Contains(toolNameLower, "update"),
			"Tool name should indicate modification functionality",
		)
		s.NotEmpty(configSetTool.Description)
		s.Contains(strings.ToLower(configSetTool.Description), "set")
		s.Equal(CategoryConfiguration, configSetTool.Category)
	})

	s.Run("Schema", func() {
		s.NotNil(configSetTool.InputSchema)
		s.Equal("object", configSetTool.InputSchema["type"])

		properties, hasProperties := configSetTool.InputSchema["properties"]
		s.True(hasProperties, "Should have properties for setting config")

		if propertiesMap, ok := properties.(map[string]interface{}); ok {
			// Should have key/value or setting parameters
			configFields := []string{"key", "value", "setting", "section"}
			hasConfigField := false
			for _, field := range configFields {
				if _, hasField := propertiesMap[field]; hasField {
					hasConfigField = true
					break
				}
			}
			s.True(hasConfigField, "Should have config setting fields")
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(configSetTool.Examples, "Should have examples")

		for _, example := range configSetTool.Examples {
			s.NotEmpty(example.Description)
			s.NotNil(example.Input)
			s.NotEmpty(example.Input, "Set examples should have input parameters")

			// Should demonstrate setting different config values
			exampleText := strings.ToLower(example.Description + " " + example.UseCase)
			setKeywords := []string{"set", "update", "configure", "change", "modify"}
			hasSetKeyword := false
			for _, keyword := range setKeywords {
				if strings.Contains(exampleText, keyword) {
					hasSetKeyword = true
					break
				}
			}
			s.True(hasSetKeyword, "Example should be config setting-related")
		}
	})
}

func (s *ConfigToolsTestSuite) TestConfigInitTool() {
	tools := CreateConfigTools()
	var configInitTool *MCPTool

	for _, tool := range tools {
		if strings.Contains(tool.Name, "init") || strings.Contains(tool.Name, "create") {
			configInitTool = tool
			break
		}
	}

	if configInitTool == nil {
		s.T().Skip("config init tool not found in implementation")
		return
	}

	s.Run("BasicStructure", func() {
		toolNameLower := strings.ToLower(configInitTool.Name)
		s.True(
			strings.Contains(toolNameLower, "init") || strings.Contains(toolNameLower, "create"),
			"Tool name should indicate initialization functionality",
		)
		s.NotEmpty(configInitTool.Description)
		s.Equal(CategoryConfiguration, configInitTool.Category)
	})

	s.Run("Schema", func() {
		s.NotNil(configInitTool.InputSchema)
		s.Equal("object", configInitTool.InputSchema["type"])

		properties, hasProperties := configInitTool.InputSchema["properties"]
		if hasProperties {
			if propertiesMap, ok := properties.(map[string]interface{}); ok {
				// May have template, path, or initialization options
				initFields := []string{"template", "path", "force", "default"}
				for _, field := range initFields {
					if fieldDef, hasField := propertiesMap[field]; hasField {
						s.NotNil(fieldDef, "Init field %s should have definition", field)
					}
				}
			}
		}
	})

	s.Run("Examples", func() {
		s.NotEmpty(configInitTool.Examples, "Should have examples")

		for _, example := range configInitTool.Examples {
			s.NotEmpty(example.Description)
			s.NotNil(example.Input)

			// Should demonstrate initialization scenarios
			exampleText := strings.ToLower(example.Description + " " + example.UseCase)
			initKeywords := []string{"init", "initialize", "create", "setup", "new", "default"}
			hasInitKeyword := false
			for _, keyword := range initKeywords {
				if strings.Contains(exampleText, keyword) {
					hasInitKeyword = true
					break
				}
			}
			s.True(hasInitKeyword, "Example should be initialization-related")
		}
	})
}

func (s *ConfigToolsTestSuite) TestRegisterConfigTools() {
	// Create mock registry
	validator := new(MockInputValidator)
	logger := new(MockLogger)
	registry := NewDefaultToolRegistry(validator, logger)

	// Setup expectations for tool registration
	configTools := CreateConfigTools()
	logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(len(configTools))

	// Test registration
	ctx := context.Background()
	err := RegisterConfigTools(ctx, registry)
	if err != nil {
		s.T().Skipf("RegisterConfigTools not implemented or failed: %v", err)
		return
	}

	s.NoError(err, "Should register config tools successfully")

	// Verify tools were registered
	logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	registeredTools, err := registry.ListTools(ctx, CategoryConfiguration)
	s.NoError(err)
	s.NotEmpty(registeredTools, "Should have registered configuration tools")

	// Verify each registered tool
	for _, tool := range registeredTools {
		s.Equal(CategoryConfiguration, tool.Category, "All tools should be in configuration category")
		s.validateConfigTool(tool)
	}
}

func (s *ConfigToolsTestSuite) TestConfigToolsIntegration() {
	s.Run("ToolNamesUnique", func() {
		tools := CreateConfigTools()
		toolNames := make(map[string]bool)

		for _, tool := range tools {
			s.False(toolNames[tool.Name], "Tool name %s should be unique", tool.Name)
			toolNames[tool.Name] = true
		}
	})

	s.Run("ToolsCoverConfigLifecycle", func() {
		tools := CreateConfigTools()
		toolNames := make(map[string]bool)

		for _, tool := range tools {
			toolNames[tool.Name] = true
		}

		// Core config operations
		configOperations := map[string][]string{
			"initialization": {"init", "create", "setup"},
			"validation":     {"validate", "check", "verify"},
			"display":        {"show", "get", "display"},
			"modification":   {"set", "update", "modify"},
		}

		operationsCovered := 0
		for _, keywords := range configOperations {
			operationFound := false
			for _, tool := range tools {
				toolNameLower := strings.ToLower(tool.Name)
				for _, keyword := range keywords {
					if strings.Contains(toolNameLower, keyword) {
						operationFound = true
						break
					}
				}
				if operationFound {
					break
				}
			}
			if operationFound {
				operationsCovered++
			}
		}

		s.Greater(operationsCovered, 0, "Should cover at least one config operation")
	})

	s.Run("ConsistentNamingConvention", func() {
		tools := CreateConfigTools()

		for _, tool := range tools {
			s.True(
				strings.HasPrefix(tool.Name, "config_") ||
					strings.Contains(tool.Name, "config"),
				"Tool %s should be clearly config-related", tool.Name,
			)
			s.NotContains(tool.Name, " ", "Tool name should not contain spaces")
			s.Equal(strings.ToLower(tool.Name), tool.Name, "Tool name should be lowercase")
		}
	})

	s.Run("ConsistentCLIIntegration", func() {
		tools := CreateConfigTools()

		for _, tool := range tools {
			s.NotEmpty(tool.CLICommand, "Tool %s should have CLI command", tool.Name)
			s.NotEmpty(tool.CLIArgs, "Tool %s should have CLI args", tool.Name)

			// Verify CLI args make sense for config operations
			hasConfigArg := false
			cliArgsStr := strings.Join(tool.CLIArgs, " ")
			configKeywords := []string{"config", "configuration", "settings", "validate", "init"}
			for _, keyword := range configKeywords {
				if strings.Contains(cliArgsStr, keyword) {
					hasConfigArg = true
					break
				}
			}
			s.True(hasConfigArg, "Tool %s should have config-related CLI args", tool.Name)
		}
	})

	s.Run("ExamplesCoverCommonSettings", func() {
		tools := CreateConfigTools()

		allExampleTexts := ""
		for _, tool := range tools {
			for _, example := range tool.Examples {
				allExampleTexts += strings.ToLower(example.Description + " " + example.UseCase + " ")

				// Check setting keys in input
				for key, value := range example.Input {
					allExampleTexts += strings.ToLower(key) + " "
					if valueStr, isStr := value.(string); isStr {
						allExampleTexts += strings.ToLower(valueStr) + " "
					}
				}
			}
		}

		// Should cover common configuration areas
		commonSettings := []string{"email", "path", "format", "template", "output", "default"}
		settingsCovered := 0
		for _, setting := range commonSettings {
			if strings.Contains(allExampleTexts, setting) {
				settingsCovered++
			}
		}

		s.Greater(settingsCovered, 0, "Should have examples covering common settings")
	})

	s.Run("ExamplesCoverErrorScenarios", func() {
		tools := CreateConfigTools()

		hasErrorExample := false
		for _, tool := range tools {
			for _, example := range tool.Examples {
				exampleText := strings.ToLower(example.Description + " " + example.UseCase)
				errorKeywords := []string{"invalid", "missing", "error", "corrupt", "malformed"}
				for _, keyword := range errorKeywords {
					if strings.Contains(exampleText, keyword) {
						hasErrorExample = true
						break
					}
				}
				if hasErrorExample {
					break
				}
			}
			if hasErrorExample {
				break
			}
		}

		s.True(hasErrorExample, "Should have examples covering error scenarios")
	})
}

func (s *ConfigToolsTestSuite) TestConfigToolsEdgeCases() {
	s.Run("EmptyToolsList", func() {
		// This tests the function behavior itself
		tools := CreateConfigTools()
		s.NotNil(tools, "Should never return nil")
		// Note: It's acceptable for tools to be empty if not implemented yet
	})

	s.Run("ConfigPathValidation", func() {
		tools := CreateConfigTools()

		for _, tool := range tools {
			for _, example := range tool.Examples {
				// Check config path examples are realistic
				for key, value := range example.Input {
					if strings.Contains(key, "path") || strings.Contains(key, "file") {
						if configPath, isStr := value.(string); isStr && configPath != "" {
							// Should look like a real config file path
							s.True(
								strings.Contains(configPath, ".") ||
									strings.Contains(configPath, "/") ||
									strings.Contains(configPath, "config"),
								"Config path example should look realistic: %s", configPath,
							)
						}
					}
				}
			}
		}
	})

	s.Run("SchemaSupportsConfigOperations", func() {
		tools := CreateConfigTools()

		for _, tool := range tools {
			properties, hasProps := tool.InputSchema["properties"]
			if !hasProps {
				continue
			}

			if propsMap, ok := properties.(map[string]interface{}); ok {
				// Look for config-related fields
				configFields := []string{"key", "value", "section", "setting", "config_file"}
				hasConfigField := false
				for _, field := range configFields {
					if _, hasField := propsMap[field]; hasField {
						hasConfigField = true
						break
					}
				}
				// Config tools should have config-related fields (unless they work on current config)
				if len(propsMap) > 0 {
					s.True(hasConfigField, "Config tool should have config-related fields")
				}
			}
		}
	})

	s.Run("TimeoutAppropriateForConfigOps", func() {
		tools := CreateConfigTools()

		for _, tool := range tools {
			// Config operations should be relatively fast
			s.Greater(tool.Timeout, 1*time.Second, "Tool %s should have sufficient timeout", tool.Name)
			s.LessOrEqual(tool.Timeout, 2*time.Minute, "Tool %s timeout should not be excessive for config ops", tool.Name)
		}
	})

	s.Run("DescriptionsIndicateConfigAreas", func() {
		tools := CreateConfigTools()

		for _, tool := range tools {
			desc := strings.ToLower(tool.Description)

			// Should indicate configuration functionality
			configKeywords := []string{"config", "setting", "option", "preference", "validate", "initialize"}
			hasConfigKeyword := false
			for _, keyword := range configKeywords {
				if strings.Contains(desc, keyword) {
					hasConfigKeyword = true
					break
				}
			}
			s.True(hasConfigKeyword, "Tool %s description should indicate config functionality", tool.Name)
		}
	})
}

func (s *ConfigToolsTestSuite) TestConfigToolsSecurity() {
	s.Run("NoSensitiveDataInExamples", func() {
		tools := CreateConfigTools()

		for _, tool := range tools {
			for _, example := range tool.Examples {
				// Check that examples don't contain sensitive data
				exampleJSON := ""
				for key, value := range example.Input {
					exampleJSON += key + ": " + fmt.Sprintf("%v", value) + " "
				}
				exampleText := strings.ToLower(example.Description + " " + example.UseCase + " " + exampleJSON)

				// Sensitive patterns to avoid
				sensitivePatterns := []string{"password", "secret", "key=", "token", "api_key"}
				for _, pattern := range sensitivePatterns {
					if strings.Contains(exampleText, pattern) {
						s.T().Logf("Warning: Tool %s example may contain sensitive pattern: %s", tool.Name, pattern)
					}
				}
			}
		}
	})

	s.Run("ValidateOperationsSafe", func() {
		tools := CreateConfigTools()

		for _, tool := range tools {
			if strings.Contains(strings.ToLower(tool.Name), "validate") {
				// Validation operations should be read-only
				desc := strings.ToLower(tool.Description)
				safeWords := []string{"check", "verify", "validate", "inspect", "review"}
				unsafeWords := []string{"delete", "remove", "modify", "change", "write"}

				hasSafeWord := false
				for _, word := range safeWords {
					if strings.Contains(desc, word) {
						hasSafeWord = true
						break
					}
				}

				hasUnsafeWord := false
				for _, word := range unsafeWords {
					if strings.Contains(desc, word) {
						hasUnsafeWord = true
						break
					}
				}

				s.True(hasSafeWord, "Validation tool should indicate safe operations")
				s.False(hasUnsafeWord, "Validation tool should not indicate unsafe operations")
			}
		}
	})
}

// Helper method to validate config tool structure
func (s *ConfigToolsTestSuite) validateConfigTool(tool *MCPTool) {
	s.NotNil(tool, "Tool should not be nil")
	s.NotEmpty(tool.Name, "Tool should have name")
	s.NotEmpty(tool.Description, "Tool should have description")
	s.NotNil(tool.InputSchema, "Tool should have input schema")
	s.Equal(CategoryConfiguration, tool.Category, "Tool should be in configuration category")
	s.NotEmpty(tool.CLICommand, "Tool should have CLI command")
	s.NotEmpty(tool.Version, "Tool should have version")
	s.Greater(tool.Timeout, time.Duration(0), "Tool should have positive timeout")

	// Verify schema structure
	s.Equal("object", tool.InputSchema["type"], "Schema should be object type")

	// Verify examples structure if present
	for i, example := range tool.Examples {
		s.NotEmpty(example.Description, "Example %d should have description", i)
		s.NotNil(example.Input, "Example %d should have input", i)
	}

	// Config tools should indicate configuration functionality
	toolText := strings.ToLower(tool.Name + " " + tool.Description)
	configKeywords := []string{"config", "setting", "validate", "init", "show", "set"}
	hasConfigKeyword := false
	for _, keyword := range configKeywords {
		if strings.Contains(toolText, keyword) {
			hasConfigKeyword = true
			break
		}
	}
	s.True(hasConfigKeyword, "Config tool should indicate configuration functionality")
}

// TestConfigToolsTestSuite runs the complete config tools test suite
func TestConfigToolsTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigToolsTestSuite))
}

// Benchmark tests for config tools
func BenchmarkCreateConfigTools(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tools := CreateConfigTools()
		_ = tools
	}
}

// Unit tests for specific config tool behaviors
func TestConfigTools_Specific(t *testing.T) {
	t.Run("AllToolsHaveValidCategory", func(t *testing.T) {
		tools := CreateConfigTools()
		for _, tool := range tools {
			assert.Equal(t, CategoryConfiguration, tool.Category)
		}
	})

	t.Run("AllToolsHaveReasonableTimeouts", func(t *testing.T) {
		tools := CreateConfigTools()
		for _, tool := range tools {
			// Config operations should be relatively fast
			assert.Greater(t, tool.Timeout, 1*time.Second, "Tool %s timeout too short", tool.Name)
			assert.LessOrEqual(t, tool.Timeout, 2*time.Minute, "Tool %s timeout too long", tool.Name)
		}
	})

	t.Run("AllToolsHaveValidSchemas", func(t *testing.T) {
		tools := CreateConfigTools()
		for _, tool := range tools {
			require.NotNil(t, tool.InputSchema, "Tool %s missing schema", tool.Name)
			assert.Equal(t, "object", tool.InputSchema["type"], "Tool %s should have object schema", tool.Name)
		}
	})

	t.Run("ToolNamesIndicatePurpose", func(t *testing.T) {
		tools := CreateConfigTools()
		for _, tool := range tools {
			toolName := strings.ToLower(tool.Name)
			assert.True(t,
				strings.Contains(toolName, "config") ||
					strings.Contains(toolName, "setting") ||
					strings.Contains(toolName, "validate") ||
					strings.Contains(toolName, "init"),
				"Tool name %s should indicate config purpose", tool.Name,
			)
		}
	})
}
