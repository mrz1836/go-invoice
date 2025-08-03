package models

import (
	"context"
	"math"
	"regexp"
	"strings"
	"testing"
	"time"
)

// FuzzValidationBuilder fuzzes the ValidationBuilder with various inputs
func FuzzValidationBuilder(f *testing.F) {
	// Seed with various string inputs
	stringSeeds := []string{
		"valid string",
		"",
		"   ",
		"a",
		strings.Repeat("a", 1000),
		"test@example.com",
		"invalid.email",
		"special!@#$%chars",
		"unicode🚀string",
		"\n\t\r",
		"<script>alert('xss')</script>",
		"../../etc/passwd",
		"DROP TABLE users;",
	}

	for _, str := range stringSeeds {
		f.Add(str, str, str, 0, 0.0, 0.0) // Use same string for different fields, add default numeric values
	}

	// Add numeric seeds
	numericSeeds := []float64{
		0.0, 1.0, -1.0, 100.0, -100.0,
		math.MaxFloat64, -math.MaxFloat64,
		math.SmallestNonzeroFloat64,
		math.NaN(), math.Inf(1), math.Inf(-1),
		3.14159, -3.14159,
		0.001, -0.001,
		1000000.0, -1000000.0,
	}

	for _, num := range numericSeeds {
		f.Add("test", "test@example.com", "description", int(num), num, num)
	}

	f.Fuzz(func(t *testing.T, name, email, description string, intVal int, floatVal1, floatVal2 float64) {
		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ValidationBuilder panicked: %v", r)
			}
		}()

		// Test ValidationBuilder methods
		builder := NewValidationBuilder()

		// Test all validation methods
		builder.
			AddRequired("name", name).
			AddRequired("email", email).
			AddRequired("description", description).
			AddMaxLength("name", name, 100).
			AddMaxLength("email", email, 255).
			AddMaxLength("description", description, 1000).
			AddMinLength("name", name, 1).
			AddMinLength("description", description, 3).
			AddLengthRange("name", name, 1, 100).
			AddEmail("email", email).
			AddNonNegative("floatVal1", floatVal1).
			AddNonNegative("floatVal2", floatVal2).
			AddNonNegativeInt("intVal", intVal).
			AddValidFloat("floatVal1", floatVal1).
			AddValidFloat("floatVal2", floatVal2).
			AddPositive("floatVal1", floatVal1).
			AddMaxValue("floatVal1", floatVal1, 1000.0, "1000 units")

		// Test HasErrors method
		hasErrors := builder.HasErrors()
		if hasErrors {
			// If there are errors, Build should return an error
			err := builder.Build(ErrValidationFailed)
			if err == nil {
				t.Errorf("Build should return error when HasErrors is true")
			}
		} else {
			// If no errors, Build should return nil
			err := builder.Build(ErrValidationFailed)
			if err != nil {
				t.Errorf("Build should return nil when HasErrors is false, got: %v", err)
			}
		}

		// Test BuildWithMessage
		errWithMessage := builder.BuildWithMessage("custom message")
		if hasErrors && errWithMessage == nil {
			t.Errorf("BuildWithMessage should return error when HasErrors is true")
		}
		if !hasErrors && errWithMessage != nil {
			t.Errorf("BuildWithMessage should return nil when HasErrors is false")
		}

		// Test multiple builds are consistent
		err1 := builder.Build(ErrValidationFailed)
		err2 := builder.Build(ErrValidationFailed)
		if (err1 == nil) != (err2 == nil) {
			t.Errorf("Multiple Build calls should be consistent")
		}

		// Test specific validation logic
		if name == "" {
			// Empty name should cause error
			nameBuilder := NewValidationBuilder().AddRequired("name", name)
			if !nameBuilder.HasErrors() {
				t.Errorf("Empty name should cause validation error")
			}
		}

		if email != "" && !strings.Contains(email, "@") {
			// Invalid email should cause error
			emailBuilder := NewValidationBuilder().AddEmail("email", email)
			if !emailBuilder.HasErrors() {
				t.Errorf("Invalid email should cause validation error: %s", email)
			}
		}

		if floatVal1 < 0 {
			// Negative value should cause error for non-negative validation
			negBuilder := NewValidationBuilder().AddNonNegative("floatVal1", floatVal1)
			if !negBuilder.HasErrors() {
				t.Errorf("Negative value should cause non-negative validation error: %f", floatVal1)
			}
		}

		if intVal < 0 {
			// Negative int should cause error
			negIntBuilder := NewValidationBuilder().AddNonNegativeInt("intVal", intVal)
			if !negIntBuilder.HasErrors() {
				t.Errorf("Negative int should cause non-negative validation error: %d", intVal)
			}
		}

		if math.IsNaN(floatVal1) || math.IsInf(floatVal1, 0) {
			// NaN/Inf should cause error
			floatBuilder := NewValidationBuilder().AddValidFloat("floatVal1", floatVal1)
			if !floatBuilder.HasErrors() {
				t.Errorf("NaN/Inf should cause validation error: %f", floatVal1)
			}
		}

		if floatVal1 <= 0 && !math.IsNaN(floatVal1) {
			// Non-positive should cause error for positive validation
			posBuilder := NewValidationBuilder().AddPositive("floatVal1", floatVal1)
			if !posBuilder.HasErrors() {
				t.Errorf("Non-positive value should cause positive validation error: %f", floatVal1)
			}
		}

		if len(name) > 100 {
			// Too long name should cause error
			lengthBuilder := NewValidationBuilder().AddMaxLength("name", name, 100)
			if !lengthBuilder.HasErrors() {
				t.Errorf("Too long name should cause max length validation error: %d", len(name))
			}
		}
	})
}

// FuzzEmailValidation fuzzes email validation specifically
func FuzzEmailValidation(f *testing.F) {
	// Seed with various email examples
	emails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"user+tag@domain.com",
		"user_name@domain-name.com",
		"123@domain.com",
		"user@sub.domain.com",
		"invalid.email",
		"@domain.com",
		"user@",
		"user@domain",
		"user.domain.com",
		"user@@domain.com",
		"user@domain..com",
		"user @domain.com",  // Space
		"user@domain .com",  // Space
		"user@domain.com.",  // Trailing dot
		".user@domain.com",  // Leading dot
		"user.@domain.com",  // Trailing dot in local
		"us..er@domain.com", // Double dot
		"user@",
		"",
		"very.long.email.address.that.might.cause.issues@very.long.domain.name.that.might.also.cause.issues.com",
		"unicode@доменное.имя",
		"user@domain.toolongextension",
		"user@domain.co",
		"user@domain.c",
		"user@localhost",
		"user@127.0.0.1",
		"user@[192.168.1.1]",
		"user@domain-",
		"user@-domain.com",
		"user@domain..com",
	}

	for _, email := range emails {
		f.Add(email)
	}

	f.Fuzz(func(t *testing.T, email string) {
		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Email validation panicked with input %q: %v", email, r)
			}
		}()

		builder := NewValidationBuilder()
		builder.AddEmail("email", email)

		hasErrors := builder.HasErrors()

		// Test some basic email validation rules
		if email == "" {
			// Empty email should not cause error (it's optional unless required)
			if hasErrors {
				t.Errorf("Empty email should not cause email validation error")
			}
		} else {
			// Non-empty email validation
			containsAt := strings.Count(email, "@")

			if containsAt != 1 {
				// Email without exactly one @ should fail
				if !hasErrors {
					t.Errorf("Email without exactly one @ should fail validation: %q", email)
				}
			} else if strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
				// Email starting or ending with @ should fail
				if !hasErrors {
					t.Errorf("Email starting/ending with @ should fail validation: %q", email)
				}
			} else if strings.Contains(email, "..") {
				// Email with consecutive dots should fail
				if !hasErrors {
					t.Errorf("Email with consecutive dots should fail validation: %q", email)
				}
			} else if strings.HasPrefix(email, ".") || strings.HasSuffix(email, ".") {
				// Email starting or ending with dot should fail
				if !hasErrors {
					t.Errorf("Email starting/ending with dot should fail validation: %q", email)
				}
			}
		}

		// Test validation is consistent
		builder2 := NewValidationBuilder()
		builder2.AddEmail("email", email)
		hasErrors2 := builder2.HasErrors()

		if hasErrors != hasErrors2 {
			t.Errorf("Email validation should be consistent: %v != %v for %q", hasErrors, hasErrors2, email)
		}
	})
}

// FuzzPatternValidation fuzzes pattern-based validation
func FuzzPatternValidation(f *testing.F) {
	// Create some test patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^[A-Z0-9-]+$`),        // Invoice ID pattern
		regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`), // Date pattern
		regexp.MustCompile(`^\d+(\.\d{2})?$`),     // Money pattern
		regexp.MustCompile(`^[a-zA-Z\s]+$`),       // Name pattern
		regexp.MustCompile(`^.{3,}$`),             // Minimum length pattern
		regexp.MustCompile(`^.{0,100}$`),          // Maximum length pattern
	}

	// Seed with various strings
	strings := []string{
		"VALID-123",
		"2024-01-15",
		"100.50",
		"John Doe",
		"abc",
		"toolongstringthatexceedslimits" + strings.Repeat("x", 100),
		"",
		"invalid@chars",
		"123",
		"UPPER",
		"lower",
		"Mixed-Case_123",
		"special!@#$%chars",
		"unicode🚀string",
		"2024/01/15", // Wrong date format
		"100.5",      // Wrong money format
		"123-ABC",    // Mixed case in ID
	}

	for i := range patterns {
		for _, str := range strings {
			f.Add(i, str)
		}
	}

	f.Fuzz(func(t *testing.T, patternIdx int, value string) {
		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Pattern validation panicked: %v", r)
			}
		}()

		// Clamp pattern index
		if patternIdx < 0 || patternIdx >= len(patterns) {
			patternIdx = patternIdx % len(patterns)
			if patternIdx < 0 {
				patternIdx = -patternIdx
			}
		}

		selectedPattern := patterns[patternIdx]

		builder := NewValidationBuilder()
		builder.AddPattern("field", value, selectedPattern, "must match pattern")

		hasErrors := builder.HasErrors()

		// Test pattern matching consistency
		expectedMatch := selectedPattern.MatchString(value)
		if value == "" {
			// Empty values should not cause pattern errors (pattern is only checked if value is non-empty)
			if hasErrors {
				t.Errorf("Empty value should not cause pattern validation error")
			}
		} else {
			// Non-empty values should match expected behavior
			if expectedMatch && hasErrors {
				t.Errorf("Value %q should match pattern but validation failed", value)
			}
			if !expectedMatch && !hasErrors {
				t.Errorf("Value %q should not match pattern but validation passed", value)
			}
		}

		// Test consistency
		builder2 := NewValidationBuilder()
		builder2.AddPattern("field", value, selectedPattern, "must match pattern")
		hasErrors2 := builder2.HasErrors()

		if hasErrors != hasErrors2 {
			t.Errorf("Pattern validation should be consistent")
		}
	})
}

// FuzzDateValidation fuzzes date-related validation
func FuzzDateValidation(f *testing.F) {
	// Seed with various date scenarios
	now := time.Now()
	dates := []time.Time{
		now,
		now.AddDate(0, 0, 1),  // Tomorrow
		now.AddDate(0, 0, -1), // Yesterday
		now.AddDate(1, 0, 0),  // Next year
		now.AddDate(-1, 0, 0), // Last year
		now.AddDate(0, 0, 7),  // One week ahead
		now.AddDate(0, 0, -7), // One week ago
		{},                    // Zero time
		time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2050, 12, 31, 23, 59, 59, 0, time.UTC),
		time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	for i, date1 := range dates {
		for _, date2 := range dates {
			f.Add(date1.Unix(), date2.Unix(), i*24) // i*24 as hours offset
		}
	}

	f.Fuzz(func(t *testing.T, unix1, unix2 int64, allowedFutureHours int) {
		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Date validation panicked: %v", r)
			}
		}()

		// Convert unix timestamps to dates
		date1 := time.Unix(unix1, 0)
		date2 := time.Unix(unix2, 0)

		// Clamp allowedFutureHours to reasonable range
		if allowedFutureHours < 0 {
			allowedFutureHours = 0
		}
		if allowedFutureHours > 8760 { // Max 1 year
			allowedFutureHours = 8760
		}

		builder := NewValidationBuilder()

		// Test various date validations
		builder.
			AddTimeRequired("date1", date1).
			AddTimeRequired("date2", date2).
			AddTimeOrder("order", date1, date2, "date1", "date2").
			AddDateRange("range", date1, date2, "from", "to").
			AddDateNotFuture("date1", date1, allowedFutureHours).
			AddDateNotFuture("date2", date2, allowedFutureHours)

		hasErrors := builder.HasErrors()

		// Test specific date validation logic
		if date1.IsZero() {
			// Zero date should cause required validation error
			reqBuilder := NewValidationBuilder().AddTimeRequired("date1", date1)
			if !reqBuilder.HasErrors() {
				t.Errorf("Zero date should cause required validation error")
			}
		}

		if !date1.IsZero() && !date2.IsZero() && date2.Before(date1) {
			// Invalid order should cause error
			orderBuilder := NewValidationBuilder().AddTimeOrder("order", date1, date2, "date1", "date2")
			if !orderBuilder.HasErrors() {
				t.Errorf("Invalid date order should cause validation error: %v before %v", date2, date1)
			}
		}

		now := time.Now()
		futureLimit := now.Add(time.Duration(allowedFutureHours) * time.Hour)
		if date1.After(futureLimit) {
			// Date too far in future should cause error
			futureBuilder := NewValidationBuilder().AddDateNotFuture("date1", date1, allowedFutureHours)
			if !futureBuilder.HasErrors() {
				t.Errorf("Date too far in future should cause validation error: %v", date1)
			}
		}

		// Test consistency
		builder2 := NewValidationBuilder()
		builder2.
			AddTimeRequired("date1", date1).
			AddTimeRequired("date2", date2).
			AddTimeOrder("order", date1, date2, "date1", "date2").
			AddDateRange("range", date1, date2, "from", "to").
			AddDateNotFuture("date1", date1, allowedFutureHours).
			AddDateNotFuture("date2", date2, allowedFutureHours)

		hasErrors2 := builder2.HasErrors()
		if hasErrors != hasErrors2 {
			t.Errorf("Date validation should be consistent")
		}
	})
}

// FuzzValidationBuilderChaining fuzzes method chaining
func FuzzValidationBuilderChaining(f *testing.F) {
	// Seed with various combinations
	for i := 0; i < 10; i++ {
		f.Add(i, "test", "test@example.com", 100.0, -50.0, 1000)
	}

	f.Fuzz(func(t *testing.T, scenario int, str1, str2 string, val1, val2 float64, intVal int) {
		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validation builder chaining panicked: %v", r)
			}
		}()

		// Test various chaining scenarios
		builder := NewValidationBuilder()

		switch scenario % 5 {
		case 0:
			// Test basic chaining
			builder.
				AddRequired("str1", str1).
				AddMaxLength("str1", str1, 50).
				AddNonNegative("val1", val1).
				AddPositive("val2", val2)

		case 1:
			// Test email and pattern chaining
			pattern := regexp.MustCompile(`^[a-zA-Z]+$`)
			builder.
				AddEmail("email", str2).
				AddPattern("pattern", str1, pattern, "letters only").
				AddValidFloat("val1", val1).
				AddValidFloat("val2", val2)

		case 2:
			// Test length validations
			builder.
				AddMinLength("str1", str1, 3).
				AddMaxLength("str1", str1, 100).
				AddLengthRange("str2", str2, 5, 50).
				AddNonNegativeInt("intVal", intVal)

		case 3:
			// Test conditional validations
			builder.
				AddIf(len(str1) > 10, "str1", "too long", str1).
				AddIf(val1 < 0, "val1", "negative", val1).
				AddIf(val2 > 1000, "val2", "too high", val2).
				AddCustom("custom", "always fails", "test")

		case 4:
			// Test numeric validations
			builder.
				AddValidFloat("val1", val1).
				AddValidFloat("val2", val2).
				AddNonNegative("val1", val1).
				AddPositive("val2", val2).
				AddMaxValue("val1", val1, 100.0, "100 units").
				AddNonNegativeInt("intVal", intVal)
		}

		// Builder should still work after chaining
		hasErrors := builder.HasErrors()
		err := builder.Build(ErrValidationFailed)

		if hasErrors && err == nil {
			t.Errorf("Build should return error when HasErrors is true")
		}
		if !hasErrors && err != nil {
			t.Errorf("Build should return nil when HasErrors is false")
		}

		// Test that we can still add more validations
		builder.AddRequired("extra", "extra")
		hasErrors2 := builder.HasErrors()

		// Should be able to build again
		err2 := builder.Build(ErrValidationFailed)
		if hasErrors2 && err2 == nil {
			t.Errorf("Build should return error after adding more validations")
		}
	})
}

// FuzzValidationWithContext fuzzes validation with context
func FuzzValidationWithContext(f *testing.F) {
	// Seed with basic values
	f.Add("test", 100.0)

	f.Fuzz(func(t *testing.T, str string, val float64) {
		// Should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validation with context panicked: %v", r)
			}
		}()

		// Test deprecated WithContext method (should still work)
		ctx := context.Background()
		builder := NewValidationBuilder().WithContext(ctx)

		builder.
			AddRequired("str", str).
			AddNonNegative("val", val)

		// Should work normally despite deprecated context usage
		hasErrors := builder.HasErrors()
		err := builder.Build(ErrValidationFailed)

		if hasErrors && err == nil {
			t.Errorf("Build should return error when HasErrors is true")
		}
		if !hasErrors && err != nil {
			t.Errorf("Build should return nil when HasErrors is false")
		}

		// Test that context doesn't affect validation logic
		builder2 := NewValidationBuilder()
		builder2.
			AddRequired("str", str).
			AddNonNegative("val", val)

		hasErrors2 := builder2.HasErrors()
		if hasErrors != hasErrors2 {
			t.Errorf("Context should not affect validation logic")
		}
	})
}
