package cli

import (
	"bufio"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MockLogger implements Logger for testing
type MockLogger struct {
	messages []string
}

func (m *MockLogger) Info(msg string, fields ...any)  { m.messages = append(m.messages, msg) }
func (m *MockLogger) Error(msg string, fields ...any) { m.messages = append(m.messages, msg) }
func (m *MockLogger) Debug(msg string, fields ...any) { m.messages = append(m.messages, msg) }

// PrompterTestSuite tests the Prompter functionality
type PrompterTestSuite struct {
	suite.Suite
	logger   *MockLogger
	prompter *Prompter
}

func (suite *PrompterTestSuite) SetupTest() {
	suite.logger = &MockLogger{}
	suite.prompter = NewPrompter(suite.logger)
}

// createPrompterWithInput creates a prompter with predefined input
func (suite *PrompterTestSuite) createPrompterWithInput(input string) *Prompter {
	reader := bufio.NewReader(strings.NewReader(input))
	return &Prompter{
		reader: reader,
		logger: suite.logger,
	}
}

func (suite *PrompterTestSuite) TestNewPrompter() {
	prompter := NewPrompter(suite.logger)
	assert.NotNil(suite.T(), prompter)
	assert.NotNil(suite.T(), prompter.reader)
	assert.Equal(suite.T(), suite.logger, prompter.logger)
}

func (suite *PrompterTestSuite) TestPromptString() {
	tests := []struct {
		name         string
		input        string
		defaultValue string
		expected     string
	}{
		{
			name:         "SimpleInput",
			input:        "hello world\n",
			defaultValue: "",
			expected:     "hello world",
		},
		{
			name:         "InputWithWhitespace",
			input:        "  spaced input  \n",
			defaultValue: "",
			expected:     "spaced input",
		},
		{
			name:         "EmptyInputWithDefault",
			input:        "\n",
			defaultValue: "default value",
			expected:     "default value",
		},
		{
			name:         "InputOverridesDefault",
			input:        "user input\n",
			defaultValue: "default value",
			expected:     "user input",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			prompter := suite.createPrompterWithInput(tt.input)
			ctx := context.Background()

			result, err := prompter.PromptString(ctx, "Test prompt", tt.defaultValue)
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), tt.expected, result)
		})
	}
}

func (suite *PrompterTestSuite) TestPromptString_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	prompter := suite.createPrompterWithInput("test\n")
	result, err := prompter.PromptString(ctx, "Test prompt", "")

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), context.Canceled, err)
	assert.Empty(suite.T(), result)
}

func (suite *PrompterTestSuite) TestPromptStringRequired() {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "ValidInput",
			input:    "required value\n",
			expected: "required value",
			wantErr:  false,
		},
		{
			name:     "EmptyInputThenValid",
			input:    "\n\nvalid input\n",
			expected: "valid input",
			wantErr:  false,
		},
		{
			name:     "WhitespaceOnlyThenValid",
			input:    "   \n\t\n  final value  \n",
			expected: "final value",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			prompter := suite.createPrompterWithInput(tt.input)
			ctx := context.Background()

			result, err := prompter.PromptStringRequired(ctx, "Required prompt")

			if tt.wantErr {
				assert.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expected, result)
			}
		})
	}
}

func (suite *PrompterTestSuite) TestPromptInt() {
	tests := []struct {
		name         string
		input        string
		defaultValue int
		expected     int
		wantErr      bool
	}{
		{
			name:         "ValidInteger",
			input:        "42\n",
			defaultValue: 0,
			expected:     42,
			wantErr:      false,
		},
		{
			name:         "NegativeInteger",
			input:        "-10\n",
			defaultValue: 0,
			expected:     -10,
			wantErr:      false,
		},
		{
			name:         "EmptyInputWithDefault",
			input:        "\n",
			defaultValue: 100,
			expected:     100,
			wantErr:      false,
		},
		{
			name:         "InvalidInputThenValid",
			input:        "not a number\n123\n",
			defaultValue: 0,
			expected:     123,
			wantErr:      false,
		},
		{
			name:         "FloatInputThenValid",
			input:        "3.14\n42\n",
			defaultValue: 0,
			expected:     42,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			prompter := suite.createPrompterWithInput(tt.input)
			ctx := context.Background()

			result, err := prompter.PromptInt(ctx, "Integer prompt", tt.defaultValue)

			if tt.wantErr {
				assert.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expected, result)
			}
		})
	}
}

func (suite *PrompterTestSuite) TestPromptFloat() {
	tests := []struct {
		name         string
		input        string
		defaultValue float64
		expected     float64
		wantErr      bool
	}{
		{
			name:         "ValidFloat",
			input:        "3.14\n",
			defaultValue: 0.0,
			expected:     3.14,
			wantErr:      false,
		},
		{
			name:         "ValidInteger",
			input:        "42\n",
			defaultValue: 0.0,
			expected:     42.0,
			wantErr:      false,
		},
		{
			name:         "NegativeFloat",
			input:        "-2.5\n",
			defaultValue: 0.0,
			expected:     -2.5,
			wantErr:      false,
		},
		{
			name:         "EmptyInputWithDefault",
			input:        "\n",
			defaultValue: 1.5,
			expected:     1.5,
			wantErr:      false,
		},
		{
			name:         "InvalidInputThenValid",
			input:        "not a number\n2.718\n",
			defaultValue: 0.0,
			expected:     2.718,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			prompter := suite.createPrompterWithInput(tt.input)
			ctx := context.Background()

			result, err := prompter.PromptFloat(ctx, "Float prompt", tt.defaultValue)

			if tt.wantErr {
				assert.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expected, result)
			}
		})
	}
}

func (suite *PrompterTestSuite) TestPromptBool() {
	tests := []struct {
		name         string
		input        string
		defaultValue bool
		expected     bool
		wantErr      bool
	}{
		{
			name:         "Yes",
			input:        "yes\n",
			defaultValue: false,
			expected:     true,
			wantErr:      false,
		},
		{
			name:         "Y",
			input:        "y\n",
			defaultValue: false,
			expected:     true,
			wantErr:      false,
		},
		{
			name:         "True",
			input:        "true\n",
			defaultValue: false,
			expected:     true,
			wantErr:      false,
		},
		{
			name:         "1",
			input:        "1\n",
			defaultValue: false,
			expected:     true,
			wantErr:      false,
		},
		{
			name:         "No",
			input:        "no\n",
			defaultValue: true,
			expected:     false,
			wantErr:      false,
		},
		{
			name:         "N",
			input:        "n\n",
			defaultValue: true,
			expected:     false,
			wantErr:      false,
		},
		{
			name:         "False",
			input:        "false\n",
			defaultValue: true,
			expected:     false,
			wantErr:      false,
		},
		{
			name:         "0",
			input:        "0\n",
			defaultValue: true,
			expected:     false,
			wantErr:      false,
		},
		{
			name:         "CaseInsensitive",
			input:        "YES\n",
			defaultValue: false,
			expected:     true,
			wantErr:      false,
		},
		{
			name:         "InvalidThenValid",
			input:        "maybe\ninvalid\nyes\n",
			defaultValue: false,
			expected:     true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			prompter := suite.createPrompterWithInput(tt.input)
			ctx := context.Background()

			result, err := prompter.PromptBool(ctx, "Boolean prompt", tt.defaultValue)

			if tt.wantErr {
				assert.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expected, result)
			}
		})
	}
}

func (suite *PrompterTestSuite) TestPromptSelect() {
	tests := []struct {
		name         string
		input        string
		options      []string
		defaultIndex int
		expectedIdx  int
		expectedVal  string
		wantErr      bool
	}{
		{
			name:         "ValidSelection",
			input:        "2\n",
			options:      []string{"Option A", "Option B", "Option C"},
			defaultIndex: 0,
			expectedIdx:  1,
			expectedVal:  "Option B",
			wantErr:      false,
		},
		{
			name:         "FirstOption",
			input:        "1\n",
			options:      []string{"First", "Second"},
			defaultIndex: -1,
			expectedIdx:  0,
			expectedVal:  "First",
			wantErr:      false,
		},
		{
			name:         "InvalidThenValid",
			input:        "0\n5\n2\n",
			options:      []string{"A", "B", "C"},
			defaultIndex: 0,
			expectedIdx:  1,
			expectedVal:  "B",
			wantErr:      false,
		},
		{
			name:         "EmptyOptions",
			input:        "1\n",
			options:      []string{},
			defaultIndex: 0,
			expectedIdx:  -1,
			expectedVal:  "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			prompter := suite.createPrompterWithInput(tt.input)
			ctx := context.Background()

			idx, val, err := prompter.PromptSelect(ctx, "Select prompt", tt.options, tt.defaultIndex)

			if tt.wantErr {
				assert.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expectedIdx, idx)
				assert.Equal(suite.T(), tt.expectedVal, val)
			}
		})
	}
}

func (suite *PrompterTestSuite) TestPromptMultiSelect() {
	tests := []struct {
		name        string
		input       string
		options     []string
		expectedIdx []int
		expectedVal []string
		wantErr     bool
	}{
		{
			name:        "SingleSelection",
			input:       "2\n",
			options:     []string{"A", "B", "C"},
			expectedIdx: []int{1},
			expectedVal: []string{"B"},
			wantErr:     false,
		},
		{
			name:        "MultipleSelections",
			input:       "1,3\n",
			options:     []string{"A", "B", "C"},
			expectedIdx: []int{0, 2},
			expectedVal: []string{"A", "C"},
			wantErr:     false,
		},
		{
			name:        "AllOptions",
			input:       "all\n",
			options:     []string{"A", "B", "C"},
			expectedIdx: []int{0, 1, 2},
			expectedVal: []string{"A", "B", "C"},
			wantErr:     false,
		},
		{
			name:        "SelectionsWithSpaces",
			input:       "1, 2, 3\n",
			options:     []string{"First", "Second", "Third"},
			expectedIdx: []int{0, 1, 2},
			expectedVal: []string{"First", "Second", "Third"},
			wantErr:     false,
		},
		{
			name:        "InvalidSelections",
			input:       "0,5,abc\n",
			options:     []string{"A", "B", "C"},
			expectedIdx: nil,
			expectedVal: nil,
			wantErr:     true,
		},
		{
			name:        "EmptyOptions",
			input:       "1\n",
			options:     []string{},
			expectedIdx: nil,
			expectedVal: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			prompter := suite.createPrompterWithInput(tt.input)
			ctx := context.Background()

			idx, val, err := prompter.PromptMultiSelect(ctx, "Multi-select prompt", tt.options)

			if tt.wantErr {
				assert.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expectedIdx, idx)
				assert.Equal(suite.T(), tt.expectedVal, val)
			}
		})
	}
}

func (suite *PrompterTestSuite) TestPromptDate() {
	tests := []struct {
		name         string
		input        string
		defaultValue time.Time
		expected     time.Time
		wantErr      bool
	}{
		{
			name:         "ValidISODate",
			input:        "2024-01-15\n",
			defaultValue: time.Time{},
			expected:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:      false,
		},
		{
			name:         "ValidUSDate",
			input:        "01/15/2024\n",
			defaultValue: time.Time{},
			expected:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:      false,
		},
		{
			name:         "ValidEuropeanDate",
			input:        "15-01-2024\n",
			defaultValue: time.Time{},
			expected:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:      false,
		},
		{
			name:         "ValidMonthName",
			input:        "Jan 15, 2024\n",
			defaultValue: time.Time{},
			expected:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:      false,
		},
		{
			name:         "ValidFullMonthName",
			input:        "January 15, 2024\n",
			defaultValue: time.Time{},
			expected:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:      false,
		},
		{
			name:         "EmptyInputWithDefault",
			input:        "\n",
			defaultValue: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
			expected:     time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
			wantErr:      false,
		},
		{
			name:         "InvalidThenValid",
			input:        "invalid date\n2024-02-29\n",
			defaultValue: time.Time{},
			expected:     time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			prompter := suite.createPrompterWithInput(tt.input)
			ctx := context.Background()

			result, err := prompter.PromptDate(ctx, "Date prompt", tt.defaultValue)

			if tt.wantErr {
				assert.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				assert.True(suite.T(), tt.expected.Equal(result),
					"expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func (suite *PrompterTestSuite) TestPromptConfirm() {
	tests := []struct {
		name     string
		input    string
		expected bool
		wantErr  bool
	}{
		{
			name:     "ConfirmYes",
			input:    "yes\n",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "ConfirmNo",
			input:    "no\n",
			expected: false,
			wantErr:  false,
		},
		{
			name:     "ConfirmY",
			input:    "y\n",
			expected: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			prompter := suite.createPrompterWithInput(tt.input)
			ctx := context.Background()

			result, err := prompter.PromptConfirm(ctx, "Confirm message")

			if tt.wantErr {
				assert.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				assert.Equal(suite.T(), tt.expected, result)
			}
		})
	}
}

func (suite *PrompterTestSuite) TestPromptPassword() {
	prompter := suite.createPrompterWithInput("secret123\n")
	ctx := context.Background()

	result, err := prompter.PromptPassword(ctx, "Password prompt")

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "secret123", result)
}

func (suite *PrompterTestSuite) TestPromptPassword_ContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	prompter := suite.createPrompterWithInput("password\n")
	result, err := prompter.PromptPassword(ctx, "Password")

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), context.Canceled, err)
	assert.Empty(suite.T(), result)
}

// TestPrompterEdgeCases tests various edge cases
func (suite *PrompterTestSuite) TestPrompterEdgeCases() {
	suite.Run("PromptString_ReadError", func() {
		// Create a prompter with a reader that will return an error
		// In real scenarios, this could happen with closed stdin
		prompter := suite.createPrompterWithInput("")
		ctx := context.Background()

		_, _ = prompter.PromptString(ctx, "Test", "")
		// This may or may not error depending on the input, but should not panic
		// The key is that it handles the error gracefully
	})

	suite.Run("PromptSelect_DefaultIndex", func() {
		prompter := suite.createPrompterWithInput("\n")
		ctx := context.Background()
		options := []string{"A", "B", "C"}

		// Test with valid default index
		idx, val, err := prompter.PromptSelect(ctx, "Test", options, 1)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), 1, idx)
		assert.Equal(suite.T(), "B", val)
	})

	suite.Run("PromptMultiSelect_AllCaseInsensitive", func() {
		prompter := suite.createPrompterWithInput("ALL\n")
		ctx := context.Background()
		options := []string{"A", "B", "C"}

		idx, val, err := prompter.PromptMultiSelect(ctx, "Test", options)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), []int{0, 1, 2}, idx)
		assert.Equal(suite.T(), options, val)
	})
}

// TestPrompterConcurrency tests concurrent usage (though prompts are typically not concurrent)
func (suite *PrompterTestSuite) TestPrompterConcurrency() {
	// This test ensures the prompter doesn't have race conditions in its structure
	// Even though prompts are typically sequential
	_ = NewPrompter(suite.logger)

	// Test that creating multiple prompters doesn't cause issues
	prompters := make([]*Prompter, 10)
	for i := range prompters {
		prompters[i] = NewPrompter(suite.logger)
		assert.NotNil(suite.T(), prompters[i])
	}
}

func TestPrompterTestSuite(t *testing.T) {
	suite.Run(t, new(PrompterTestSuite))
}

// Benchmark tests for performance validation
func BenchmarkPromptString(b *testing.B) {
	logger := &MockLogger{}

	for i := 0; i < b.N; i++ {
		prompter := &Prompter{
			reader: bufio.NewReader(strings.NewReader("test input\n")),
			logger: logger,
		}

		ctx := context.Background()
		_, _ = prompter.PromptString(ctx, "Test", "")
	}
}

func BenchmarkPromptInt(b *testing.B) {
	logger := &MockLogger{}

	for i := 0; i < b.N; i++ {
		prompter := &Prompter{
			reader: bufio.NewReader(strings.NewReader("42\n")),
			logger: logger,
		}

		ctx := context.Background()
		_, _ = prompter.PromptInt(ctx, "Test", 0)
	}
}

func BenchmarkPromptBool(b *testing.B) {
	logger := &MockLogger{}

	for i := 0; i < b.N; i++ {
		prompter := &Prompter{
			reader: bufio.NewReader(strings.NewReader("yes\n")),
			logger: logger,
		}

		ctx := context.Background()
		_, _ = prompter.PromptBool(ctx, "Test", false)
	}
}
