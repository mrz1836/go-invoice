package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// OutputParserTestSuite tests the DefaultOutputParser implementation
type OutputParserTestSuite struct {
	suite.Suite

	parser *DefaultOutputParser
	logger *MockLogger
}

func (suite *OutputParserTestSuite) SetupTest() {
	suite.logger = new(MockLogger)

	// Setup logger expectations
	suite.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	suite.parser = NewDefaultOutputParser(suite.logger)
}

// TestNewDefaultOutputParser tests the constructor
func (suite *OutputParserTestSuite) TestNewDefaultOutputParser() {
	parser := NewDefaultOutputParser(suite.logger)
	suite.NotNil(parser)
	suite.Equal(suite.logger, parser.logger)
}

// TestNewDefaultOutputParserPanicsWithNilLogger tests constructor panics without logger
func (suite *OutputParserTestSuite) TestNewDefaultOutputParserPanicsWithNilLogger() {
	suite.Panics(func() {
		NewDefaultOutputParser(nil)
	})
}

// TestParseJSON tests JSON parsing
func (suite *OutputParserTestSuite) TestParseJSON() {
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		keys    []string
	}{
		{
			name:    "ValidJSON",
			input:   `{"name": "John", "age": 30}`,
			wantErr: false,
			keys:    []string{"name", "age"},
		},
		{
			name:    "EmptyObject",
			input:   `{}`,
			wantErr: false,
			keys:    []string{},
		},
		{
			name:    "NestedJSON",
			input:   `{"user": {"name": "John"}, "active": true}`,
			wantErr: false,
			keys:    []string{"user", "active"},
		},
		{
			name:    "EmptyInput",
			input:   "",
			wantErr: true,
		},
		{
			name:    "WhitespaceOnly",
			input:   "   \n\t  ",
			wantErr: true,
		},
		{
			name:    "InvalidJSON",
			input:   `{invalid json}`,
			wantErr: true,
		},
		{
			name:    "JSONWithPrefixText",
			input:   "Some prefix text\n{\"key\": \"value\"}",
			wantErr: false,
			keys:    []string{"key"},
		},
		{
			name:    "JSONWithSuffixText",
			input:   "{\"key\": \"value\"}\nSome suffix text",
			wantErr: false,
			keys:    []string{"key"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := suite.parser.ParseJSON(ctx, tt.input)
			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.Require().NoError(err)
				suite.NotNil(result)
				for _, key := range tt.keys {
					suite.Contains(result, key)
				}
			}
		})
	}
}

// TestParseJSONContextCancellation tests context cancellation
func (suite *OutputParserTestSuite) TestParseJSONContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := suite.parser.ParseJSON(ctx, `{"key": "value"}`)
	suite.Equal(context.Canceled, err)
	suite.Nil(result)
}

// TestParseTable tests table parsing
func (suite *OutputParserTestSuite) TestParseTable() {
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		rows    int
	}{
		{
			name: "PipeSeparatedTable",
			input: `Name | Email | Status
------|-------|-------
John | john@example.com | Active
Jane | jane@example.com | Inactive`,
			wantErr: false,
			rows:    2,
		},
		{
			name:    "TabSeparatedTable",
			input:   "Name\tEmail\tStatus\nJohn\tjohn@example.com\tActive\nJane\tjane@example.com\tInactive",
			wantErr: false,
			rows:    2,
		},
		{
			name:    "SpaceSeparatedTable",
			input:   "Name Email Status\nJohn john@example.com Active\nJane jane@example.com Inactive",
			wantErr: false,
			rows:    2,
		},
		{
			name:    "EmptyInput",
			input:   "",
			wantErr: true,
		},
		{
			name:    "OnlyHeader",
			input:   "Name | Email | Status",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := suite.parser.ParseTable(ctx, tt.input)
			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Len(result, tt.rows)
			}
		})
	}
}

// TestParseTableContextCancellation tests context cancellation
func (suite *OutputParserTestSuite) TestParseTableContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := suite.parser.ParseTable(ctx, "Name | Email\nJohn | john@example.com")
	suite.Equal(context.Canceled, err)
	suite.Nil(result)
}

// TestParseKeyValue tests key-value parsing
func (suite *OutputParserTestSuite) TestParseKeyValue() {
	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		keys    []string
	}{
		{
			name:    "ColonSeparated",
			input:   "Name: John\nAge: 30\nEmail: john@example.com",
			wantErr: false,
			keys:    []string{"Name", "Age", "Email"},
		},
		{
			name:    "EqualsSeparated",
			input:   "NAME=John\nAGE=30",
			wantErr: false,
			keys:    []string{"NAME", "AGE"},
		},
		{
			name:    "MixedSeparators",
			input:   "Name: John\nAGE=30",
			wantErr: false,
			keys:    []string{"Name", "AGE"},
		},
		{
			name:    "EmptyInput",
			input:   "",
			wantErr: true,
		},
		{
			name:    "SpaceSeparatedValues",
			input:   "Just some text\nWithout separators",
			wantErr: false, // Space separator pattern matches these
			keys:    []string{"Just", "Without"},
		},
		{
			name:    "WithEmptyLines",
			input:   "Name: John\n\nAge: 30",
			wantErr: false,
			keys:    []string{"Name", "Age"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := suite.parser.ParseKeyValue(ctx, tt.input)
			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.Require().NoError(err)
				suite.NotNil(result)
				for _, key := range tt.keys {
					suite.Contains(result, key)
				}
			}
		})
	}
}

// TestParseKeyValueContextCancellation tests context cancellation
func (suite *OutputParserTestSuite) TestParseKeyValueContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := suite.parser.ParseKeyValue(ctx, "Key: Value")
	suite.Equal(context.Canceled, err)
	suite.Nil(result)
}

// TestExtractError tests error extraction
func (suite *OutputParserTestSuite) TestExtractError() {
	ctx := context.Background()

	tests := []struct {
		name     string
		stdout   string
		stderr   string
		exitCode int
		wantErr  bool
		wantNil  bool
	}{
		{
			name:     "ExitCodeZero",
			stdout:   "Success",
			stderr:   "",
			exitCode: 0,
			wantErr:  false,
			wantNil:  true,
		},
		{
			name:     "ErrorInStderr",
			stdout:   "",
			stderr:   "Error: something went wrong",
			exitCode: 1,
			wantErr:  true,
			wantNil:  false,
		},
		{
			name:     "FatalInStderr",
			stdout:   "",
			stderr:   "Fatal: critical failure",
			exitCode: 1,
			wantErr:  true,
			wantNil:  false,
		},
		{
			name:     "ErrorInStdout",
			stdout:   "Error: operation failed",
			stderr:   "",
			exitCode: 1,
			wantErr:  true,
			wantNil:  false,
		},
		{
			name:     "GenericStderr",
			stdout:   "",
			stderr:   "Some generic error message",
			exitCode: 1,
			wantErr:  true,
			wantNil:  false,
		},
		{
			name:     "ExitCodeOnly",
			stdout:   "",
			stderr:   "",
			exitCode: 127,
			wantErr:  true,
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.parser.ExtractError(ctx, tt.stdout, tt.stderr, tt.exitCode)
			if tt.wantNil {
				suite.Require().NoError(err)
			} else if tt.wantErr {
				suite.Error(err)
			}
		})
	}
}

// TestExtractErrorContextCancellation tests context cancellation
func (suite *OutputParserTestSuite) TestExtractErrorContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := suite.parser.ExtractError(ctx, "", "Error", 1)
	suite.Equal(context.Canceled, err)
}

// TestExtractJSON tests JSON extraction from mixed output
func (suite *OutputParserTestSuite) TestExtractJSON() {
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{
			name:    "CleanJSON",
			input:   `{"key": "value"}`,
			wantLen: 16,
		},
		{
			name:    "JSONWithPrefix",
			input:   `Some text {"key": "value"}`,
			wantLen: 16,
		},
		{
			name:    "JSONWithSuffix",
			input:   `{"key": "value"} more text`,
			wantLen: 16,
		},
		{
			name:    "JSONArray",
			input:   `[1, 2, 3]`,
			wantLen: 9,
		},
		{
			name:    "NoJSON",
			input:   `Just plain text`,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.parser.extractJSON(tt.input)
			suite.Len(result, tt.wantLen)
		})
	}
}

// TestIsSeparatorLine tests separator line detection
func (suite *OutputParserTestSuite) TestIsSeparatorLine() {
	tests := []struct {
		name  string
		line  string
		isSep bool
	}{
		{"DashLine", "-------------------", true},
		{"EqualLine", "===================", true},
		{"MixedSeparator", "---|---|---", true},
		{"PipeAndDash", "| --- | --- |", true},
		{"EmptyLine", "", false},
		{"TextLine", "Name Email Status", false},
		{"DataLine", "John | john@example.com", false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.parser.isSeparatorLine(tt.line)
			suite.Equal(tt.isSep, result)
		})
	}
}

// TestSplitTableRow tests table row splitting
func (suite *OutputParserTestSuite) TestSplitTableRow() {
	tests := []struct {
		name      string
		row       string
		separator string
		wantLen   int
	}{
		{"PipeSeparated", "| A | B | C |", "|", 3},
		{"PipeNoBorders", "A | B | C", "|", 3},
		{"TabSeparated", "A\tB\tC", "\t", 3},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.parser.splitTableRow(tt.row, tt.separator)
			suite.Len(result, tt.wantLen)
		})
	}
}

// TestLooksLikeFixedWidthTable tests fixed-width table detection
func (suite *OutputParserTestSuite) TestLooksLikeFixedWidthTable() {
	tests := []struct {
		name   string
		lines  []string
		result bool
	}{
		{
			name:   "FixedWidthTable",
			lines:  []string{"Name     Email", "------   -----", "John     john@example.com"},
			result: true,
		},
		{
			name:   "TooFewLines",
			lines:  []string{"Name     Email", "------   -----"},
			result: false,
		},
		{
			name:   "NoSeparatorLine",
			lines:  []string{"Name     Email", "John     john@example.com", "Jane     jane@example.com"},
			result: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.parser.looksLikeFixedWidthTable(tt.lines)
			suite.Equal(tt.result, result)
		})
	}
}

// TestParseQuotedFields tests quoted field parsing
func (suite *OutputParserTestSuite) TestParseQuotedFields() {
	tests := []struct {
		name     string
		line     string
		expected int
		fields   []string
	}{
		{
			name:     "SimpleFields",
			line:     "a b c",
			expected: 3,
			fields:   []string{"a", "b", "c"},
		},
		{
			name:     "QuotedField",
			line:     `a "b c" d`,
			expected: 3,
			fields:   []string{"a", "b c", "d"},
		},
		{
			name:     "SingleQuotes",
			line:     `a 'b c' d`,
			expected: 3,
			fields:   []string{"a", "b c", "d"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := suite.parser.parseQuotedFields(tt.line, tt.expected)
			suite.Len(result, tt.expected)
			suite.Equal(tt.fields, result)
		})
	}
}

func TestOutputParserTestSuite(t *testing.T) {
	suite.Run(t, new(OutputParserTestSuite))
}

// TestParserErrorVariables tests that error variables are properly defined
func TestParserErrorVariables(t *testing.T) {
	tests := []struct {
		err     error
		name    string
		wantMsg string
	}{
		{ErrInvalidJSON, "ErrInvalidJSON", "invalid JSON output"},
		{ErrInvalidTable, "ErrInvalidTable", "invalid table format"},
		{ErrInvalidKeyValue, "ErrInvalidKeyValue", "invalid key-value format"},
		{ErrParsingFailed, "ErrParsingFailed", "output parsing failed"},
		{ErrNoDataFound, "ErrNoDataFound", "no data found in output"},
		{ErrUnexpectedFormat, "ErrUnexpectedFormat", "unexpected output format"},
		{ErrCommandExitCode, "ErrCommandExitCode", "command failed with exit code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.wantMsg)
			}
		})
	}
}
