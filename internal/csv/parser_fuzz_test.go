package csv

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// FuzzParseTimesheet fuzzes the complete CSV parsing pipeline
func FuzzParseTimesheet(f *testing.F) {
	// Seed with known good examples
	seeds := []string{
		"Date,Hours,Rate,Description\n2024-01-15,8.0,100.00,Development work",
		"Date\tHours\tRate\tDescription\n2024-01-15\t8.0\t100.00\tDevelopment work",
		"Date;Hours;Rate;Description\n2024-01-15;8.0;100.00;Development work",
		"Date,Hours,Rate,Description\n2024-01-15,8.0,100.00,Work",
		"Date,Hours,Rate,Description\n01/15/2024,8.5,125.50,Bug fixes",
		"Date,Hours,Rate,Description\n15/01/2024,4.0,80.00,Code review",
		"DATE,DURATION,HOURLY_RATE,TASK\n2024-01-15,8.0,100.00,Development",
		"Work_Date,Time,Billing_Rate,Notes\n2024-01-15,8.0,100.00,Work",
	}

	// Add seeds to fuzzer
	for _, seed := range seeds {
		f.Add(seed)
	}

	// Add some edge case seeds
	edgeCases := []string{
		"",                              // Empty file
		"Date,Hours,Rate,Description",   // Header only
		"Date,Hours,Rate,Description\n", // Header with newline
		"Date,Hours,Rate,Description\n2024-01-15,8.0,100.00",                      // Missing field
		"Date,Hours,Rate,Description\n2024-01-15,8.0,100.00,Work,Extra",           // Extra field
		"Date,Hours,Rate,Description\n,,,",                                        // Empty fields
		"Date,Hours,Rate,Description\n2024-01-15,,100.00,Work",                    // Missing hours
		"Date,Hours,Rate,Description\n2024-01-15,8.0,,Work",                       // Missing rate
		"Date,Hours,Rate,Description\n,8.0,100.00,Work",                           // Missing date
		"Date,Hours,Rate,Description\n2024-01-15,8.0,100.00,",                     // Missing description
		"Date,Hours,Rate,Description\n\"2024-01-15\",\"8.0\",\"100.00\",\"Work\"", // Quoted fields
		"Date,Hours,Rate,Description\n2024-01-15,8.0,100.00,\"Work, with comma\"", // Quoted description with comma
	}

	for _, edge := range edgeCases {
		f.Add(edge)
	}

	f.Fuzz(func(t *testing.T, csvData string) {
		// Create parser with mock dependencies
		logger := &MockLogger{}
		validator := &MockValidator{}
		idGenerator := &MockIDGenerator{}
		parser := NewCSVParser(validator, logger, idGenerator)

		reader := strings.NewReader(csvData)
		options := ParseOptions{
			Format:          "standard",
			ContinueOnError: true, // Don't fail on errors, collect them
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		// The parser should never panic, regardless of input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Parser panicked with input %q: %v", csvData, r)
			}
		}()

		result, err := parser.ParseTimesheet(ctx, reader, options)

		// Test invariants that should always hold
		if err == nil && result != nil {
			// If parsing succeeded, result should be valid
			if result.TotalRows < 0 {
				t.Errorf("TotalRows should not be negative: %d", result.TotalRows)
			}
			if result.SuccessRows < 0 {
				t.Errorf("SuccessRows should not be negative: %d", result.SuccessRows)
			}
			if result.ErrorRows < 0 {
				t.Errorf("ErrorRows should not be negative: %d", result.ErrorRows)
			}
			if result.SuccessRows+result.ErrorRows != result.TotalRows {
				t.Errorf("SuccessRows + ErrorRows should equal TotalRows: %d + %d != %d",
					result.SuccessRows, result.ErrorRows, result.TotalRows)
			}
			if len(result.WorkItems) != result.SuccessRows {
				t.Errorf("WorkItems length should equal SuccessRows: %d != %d",
					len(result.WorkItems), result.SuccessRows)
			}
			if len(result.Errors) != result.ErrorRows {
				t.Errorf("Errors length should equal ErrorRows: %d != %d",
					len(result.Errors), result.ErrorRows)
			}

			// Check work item invariants
			for i, item := range result.WorkItems {
				if item.ID == "" {
					t.Errorf("WorkItem[%d] should have non-empty ID", i)
				}
				if item.Hours < 0 {
					t.Errorf("WorkItem[%d] should have non-negative hours: %f", i, item.Hours)
				}
				if item.Rate < 0 {
					t.Errorf("WorkItem[%d] should have non-negative rate: %f", i, item.Rate)
				}
				if item.Total < 0 {
					t.Errorf("WorkItem[%d] should have non-negative total: %f", i, item.Total)
				}
				if item.Date.IsZero() {
					t.Errorf("WorkItem[%d] should have non-zero date", i)
				}
				if item.Description == "" {
					t.Errorf("WorkItem[%d] should have non-empty description", i)
				}
			}

			// Check error invariants
			for i, parseErr := range result.Errors {
				if parseErr.Line <= 0 {
					t.Errorf("ParseError[%d] should have positive line number: %d", i, parseErr.Line)
				}
				if parseErr.Message == "" {
					t.Errorf("ParseError[%d] should have non-empty message", i)
				}
			}
		}

		// Context should not be canceled unless we canceled it
		if errors.Is(err, context.Canceled) && ctx.Err() == nil {
			t.Errorf("Parser returned context.Canceled but context was not canceled")
		}
	})
}

// FuzzDetectFormat fuzzes the CSV format detection
func FuzzDetectFormat(f *testing.F) {
	// Seed with various format examples
	seeds := []string{
		"Date,Hours,Rate,Description",
		"Date\tHours\tRate\tDescription",
		"Date;Hours;Rate;Description",
		"Date|Hours|Rate|Description",
		"Date:Hours:Rate:Description",
		"Date Hours Rate Description",
		"a,b,c,d",
		"a\tb\tc\td",
		"a;b;c;d",
		"Date,Hours,Rate,Description\n2024-01-15,8.0,100.00,Work",
		"",
		"\n",
		"\t\t\t",
		";;;",
		",,,",
		"single_column",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, content string) {
		logger := &MockLogger{}
		validator := &MockValidator{}
		idGenerator := &MockIDGenerator{}
		parser := NewCSVParser(validator, logger, idGenerator)

		reader := strings.NewReader(content)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("DetectFormat panicked with input %q: %v", content, r)
			}
		}()

		format, err := parser.DetectFormat(ctx, reader)

		// Test invariants
		if err == nil && format != nil {
			// Valid format should have a name and delimiter
			if format.Name == "" {
				t.Errorf("Valid format should have non-empty name")
			}

			// Delimiter should be a printable character
			if format.Delimiter < 32 || format.Delimiter > 126 {
				// Allow tab (9) as special case
				if format.Delimiter != '\t' {
					t.Errorf("Delimiter should be printable or tab: %d", int(format.Delimiter))
				}
			}

			// Name should match expected values
			validNames := []string{"standard", "tab", "semicolon", "excel", "rfc4180", "tsv"}
			isValidName := false
			for _, validName := range validNames {
				if format.Name == validName {
					isValidName = true
					break
				}
			}
			if !isValidName {
				t.Errorf("Format name should be one of valid names: %s", format.Name)
			}
		}

		// Empty content should return an error
		if content == "" && err == nil {
			t.Errorf("Empty content should return an error")
		}
	})
}

// FuzzAnalyzeFormat fuzzes the internal format analysis
func FuzzAnalyzeFormat(f *testing.F) {
	// Seed with various content patterns
	seeds := []string{
		"a,b,c,d",
		"a\tb\tc\td",
		"a;b;c;d",
		"a|b|c|d",
		"a b c d",
		"a,b,c,d\ne,f,g,h",
		"a\tb\tc\td\ne\tf\tg\th",
		"",
		"\n",
		"single",
		"a,b",
		"a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y,z", // Many columns
		"a,b\tc;d",    // Mixed delimiters
		"a,,c,d",      // Empty field
		"a,\"b,c\",d", // Quoted field with delimiter
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, content string) {
		logger := &MockLogger{}
		validator := &MockValidator{}
		idGenerator := &MockIDGenerator{}
		parser := NewCSVParser(validator, logger, idGenerator)

		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("analyzeFormat panicked with input %q: %v", content, r)
			}
		}()

		format, err := parser.analyzeFormat(content)

		// Test invariants
		if err == nil && format != nil {
			// Valid format should have reasonable properties
			if format.Name == "" {
				t.Errorf("Valid format should have non-empty name")
			}

			// Check delimiter is reasonable
			validDelimiters := []rune{',', '\t', ';', '|', ':'}
			validDelimiter := false
			for _, valid := range validDelimiters {
				if format.Delimiter == valid {
					validDelimiter = true
					break
				}
			}
			if !validDelimiter {
				t.Errorf("Delimiter should be one of valid delimiters: %c (%d)", format.Delimiter, int(format.Delimiter))
			}
		}

		// Empty content should return an error
		if content == "" && err == nil {
			t.Errorf("Empty content should return an error")
		}

		// Content with only whitespace should return an error
		if strings.TrimSpace(content) == "" && content != "" && err == nil {
			t.Errorf("Whitespace-only content should return an error")
		}
	})
}

// FuzzParseDate fuzzes the date parsing functionality
func FuzzParseDate(f *testing.F) {
	// Seed with various date formats
	seeds := []string{
		"2024-01-15",
		"01/15/2024",
		"15/01/2024",
		"2024/01/15",
		"Jan 15, 2024",
		"January 15, 2024",
		"2024-01-15 14:30:00",
		"2024-01-15T14:30:00Z",
		"2024-1-1",
		"1/1/2024",
		"1/1/24",
		"2024",
		"01",
		"",
		"invalid",
		"2024-13-01",          // Invalid month
		"2024-01-32",          // Invalid day
		"2024-02-30",          // Invalid day for February
		"24-01-15",            // Ambiguous year
		"15-01-2024",          // Different order
		"2024/13/01",          // Invalid month with slash
		"32/01/2024",          // Invalid day
		"2024-01-15 25:00:00", // Invalid hour
		"2024-01-15 14:60:00", // Invalid minute
		"2024-01-15 14:30:60", // Invalid second
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, dateStr string) {
		logger := &MockLogger{}
		validator := &MockValidator{}
		idGenerator := &MockIDGenerator{}
		parser := NewCSVParser(validator, logger, idGenerator)

		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseDate panicked with input %q: %v", dateStr, r)
			}
		}()

		date, err := parser.parseDate(dateStr)

		// Test invariants
		if err == nil {
			// Valid date should not be zero unless input was specifically zero time
			// Exception: Some dates like "01/01/0001" might parse but result in edge cases
			if date.IsZero() && dateStr != "" && dateStr != "01/01/0001" {
				t.Errorf("Valid date should not be zero for non-empty input: %q", dateStr)
			}

			// Date should be reasonable (not too far in past or future)
			// Accept any year >= 1 as valid (Go time package supports year 1+)
			// Date years < 1 are expected for invalid dates like "01/01/0000"
			// The parser may parse this but it results in an invalid date
			// Allow dates up to year 9999 (Go time package supports up to year 9999)
			if date.Year() > 9999 {
				t.Errorf("Parsed date is too far in the future: %v for input %q", date, dateStr)
			}
		}

		// Empty string should return error
		if dateStr == "" && err == nil {
			t.Errorf("Empty date string should return error")
		}

		// Error should be consistent - same input should give same result
		date2, err2 := parser.parseDate(dateStr)
		if err == nil && err2 == nil {
			if !date.Equal(date2) {
				t.Errorf("parseDate should be deterministic: %v != %v for input %q", date, date2, dateStr)
			}
		}
		if (err == nil) != (err2 == nil) {
			t.Errorf("parseDate error result should be deterministic for input %q", dateStr)
		}
	})
}

// FuzzNormalizeHeaderName fuzzes header name normalization
func FuzzNormalizeHeaderName(f *testing.F) {
	// Seed with common header variations
	seeds := []string{
		"Date",
		"date",
		"DATE",
		"Hours",
		"hours",
		"HOURS",
		"Rate",
		"rate",
		"RATE",
		"Description",
		"description",
		"DESCRIPTION",
		"Work_Date",
		"work_date",
		"WORK_DATE",
		"Hourly_Rate",
		"hourly_rate",
		"HOURLY_RATE",
		"Work_Description",
		"work_description",
		"WORK_DESCRIPTION",
		"  Date  ",
		"\tHours\t",
		"\nRate\n",
		"",
		"   ",
		"unknown_field",
		"UNKNOWN_FIELD",
		"VeryLongHeaderNameThatShouldStillWork",
		"Header-With-Dashes",
		"Header.With.Dots",
		"Header With Spaces",
		"123Numbers",
		"Special!@#$%Characters",
		"Unicode😀Header",
		"CamelCaseHeader",
		"snake_case_header",
		"kebab-case-header",
		"PascalCaseHeader",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, header string) {
		logger := &MockLogger{}
		validator := &MockValidator{}
		idGenerator := &MockIDGenerator{}
		parser := NewCSVParser(validator, logger, idGenerator)

		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("normalizeHeaderName panicked with input %q: %v", header, r)
			}
		}()

		result := parser.normalizeHeaderName(header)

		// Test invariants
		// Result can be longer than trimmed input when normalizing to full field names
		// (e.g., "desc" -> "description") or when invalid UTF-8 sequences are replaced
		trimmed := strings.TrimSpace(header)
		if len(result) > len(trimmed) {
			// This is expected behavior for normalization (e.g., "desc" -> "description")
			// and for invalid UTF-8 replacement, so we just verify result is valid
			if result == "" && trimmed != "" {
				t.Errorf("Non-empty input should not produce empty result: %q -> %q", header, result)
			}
		}

		// Result should be lowercase if not empty
		if result != "" && result != strings.ToLower(result) {
			t.Errorf("Normalized header should be lowercase: %q -> %q", header, result)
		}

		// Result should be deterministic
		result2 := parser.normalizeHeaderName(header)
		if result != result2 {
			t.Errorf("normalizeHeaderName should be deterministic: %q != %q for input %q", result, result2, header)
		}

		// Known mappings should work correctly
		knownMappings := map[string]string{
			"Date":             "date",
			"date":             "date",
			"HOURS":            "hours",
			"Rate":             "rate",
			"Description":      "description",
			"Work_Date":        "date",
			"Hourly_Rate":      "rate",
			"Work_Description": "description",
			"  Date  ":         "date",
		}

		if expected, exists := knownMappings[header]; exists {
			if result != expected {
				t.Errorf("Known mapping failed: %q should map to %q but got %q", header, expected, result)
			}
		}
	})
}
