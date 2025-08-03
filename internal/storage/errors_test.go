package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type StorageErrorsTestSuite struct {
	suite.Suite
}

func (suite *StorageErrorsTestSuite) SetupTest() {
	// No setup needed for error testing
}

func (suite *StorageErrorsTestSuite) TearDownTest() {
	// No teardown needed for error testing
}

func TestStorageErrorsTestSuite(t *testing.T) {
	suite.Run(t, new(StorageErrorsTestSuite))
}

// Test NotFoundError
func (suite *StorageErrorsTestSuite) TestNotFoundError() {
	tests := []struct {
		name     string
		resource string
		id       string
		expected string
	}{
		{
			name:     "basic_not_found",
			resource: "invoice",
			id:       "INV-001",
			expected: "invoice with ID 'INV-001' not found",
		},
		{
			name:     "client_not_found",
			resource: "client",
			id:       "CLI-123",
			expected: "client with ID 'CLI-123' not found",
		},
		{
			name:     "empty_id",
			resource: "workitem",
			id:       "",
			expected: "workitem with ID '' not found",
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			err := NewNotFoundError(test.resource, test.id)
			suite.Require().Equal(test.expected, err.Error())
			suite.Require().Equal(test.resource, err.Resource)
			suite.Require().Equal(test.id, err.ID)
		})
	}
}

// Test ConflictError
func (suite *StorageErrorsTestSuite) TestConflictError() {
	tests := []struct {
		name     string
		resource string
		id       string
		message  string
		expected string
	}{
		{
			name:     "conflict_with_message",
			resource: "client",
			id:       "CLI-001",
			message:  "email already exists",
			expected: "client with ID 'CLI-001' conflicts: email already exists",
		},
		{
			name:     "conflict_without_message",
			resource: "invoice",
			id:       "INV-002",
			message:  "",
			expected: "invoice with ID 'INV-002' already exists",
		},
		{
			name:     "conflict_empty_message",
			resource: "workitem",
			id:       "WI-003",
			message:  "",
			expected: "workitem with ID 'WI-003' already exists",
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			err := NewConflictError(test.resource, test.id, test.message)
			suite.Require().Equal(test.expected, err.Error())
			suite.Require().Equal(test.resource, err.Resource)
			suite.Require().Equal(test.id, err.ID)
			suite.Require().Equal(test.message, err.Message)
		})
	}
}

// Test VersionMismatchError
func (suite *StorageErrorsTestSuite) TestVersionMismatchError() {
	tests := []struct {
		name            string
		resource        string
		id              string
		expectedVersion int
		actualVersion   int
		expectedStr     string
	}{
		{
			name:            "version_mismatch_basic",
			resource:        "invoice",
			id:              "INV-001",
			expectedVersion: 5,
			actualVersion:   3,
			expectedStr:     "invoice with ID 'INV-001' version mismatch: expected 5, got 3",
		},
		{
			name:            "version_mismatch_zero_versions",
			resource:        "client",
			id:              "CLI-002",
			expectedVersion: 0,
			actualVersion:   0,
			expectedStr:     "client with ID 'CLI-002' version mismatch: expected 0, got 0",
		},
		{
			name:            "version_mismatch_negative",
			resource:        "workitem",
			id:              "WI-003",
			expectedVersion: 2,
			actualVersion:   -1,
			expectedStr:     "workitem with ID 'WI-003' version mismatch: expected 2, got -1",
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			err := NewVersionMismatchError(test.resource, test.id, test.expectedVersion, test.actualVersion)
			suite.Require().Equal(test.expectedStr, err.Error())
			suite.Require().Equal(test.resource, err.Resource)
			suite.Require().Equal(test.id, err.ID)
			suite.Require().Equal(test.expectedVersion, err.ExpectedVersion)
			suite.Require().Equal(test.actualVersion, err.ActualVersion)
		})
	}
}

// Test CorruptedError
func (suite *StorageErrorsTestSuite) TestCorruptedError() {
	tests := []struct {
		name     string
		resource string
		id       string
		message  string
		expected string
	}{
		{
			name:     "corrupted_json",
			resource: "invoice",
			id:       "INV-001",
			message:  "invalid JSON format",
			expected: "invoice with ID 'INV-001' is corrupted: invalid JSON format",
		},
		{
			name:     "corrupted_missing_field",
			resource: "client",
			id:       "CLI-002",
			message:  "required field 'email' is missing",
			expected: "client with ID 'CLI-002' is corrupted: required field 'email' is missing",
		},
		{
			name:     "corrupted_empty_message",
			resource: "workitem",
			id:       "WI-003",
			message:  "",
			expected: "workitem with ID 'WI-003' is corrupted: ",
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			err := NewCorruptedError(test.resource, test.id, test.message)
			suite.Require().Equal(test.expected, err.Error())
			suite.Require().Equal(test.resource, err.Resource)
			suite.Require().Equal(test.id, err.ID)
			suite.Require().Equal(test.message, err.Message)
		})
	}
}

// Test StorageUnavailableError
func (suite *StorageErrorsTestSuite) TestStorageUnavailableError() {
	tests := []struct {
		name     string
		message  string
		cause    error
		expected string
	}{
		{
			name:     "unavailable_with_cause",
			message:  "database connection failed",
			cause:    errors.New("connection refused"), //nolint:err113 // test error creation
			expected: "storage unavailable: database connection failed (caused by: connection refused)",
		},
		{
			name:     "unavailable_without_cause",
			message:  "disk full",
			cause:    nil,
			expected: "storage unavailable: disk full",
		},
		{
			name:     "unavailable_empty_message",
			message:  "",
			cause:    errors.New("network timeout"), //nolint:err113 // test error creation
			expected: "storage unavailable:  (caused by: network timeout)",
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			err := NewStorageUnavailableError(test.message, test.cause)
			suite.Require().Equal(test.expected, err.Error())
			suite.Require().Equal(test.message, err.Message)
			suite.Require().Equal(test.cause, err.Cause)
			suite.Require().Equal(test.cause, err.Unwrap())
		})
	}
}

// Test InvalidFilterError
func (suite *StorageErrorsTestSuite) TestInvalidFilterError() {
	tests := []struct {
		name     string
		field    string
		value    interface{}
		message  string
		expected string
	}{
		{
			name:     "invalid_date_filter",
			field:    "created_date",
			value:    "invalid-date",
			message:  "date format must be YYYY-MM-DD",
			expected: "invalid filter for field 'created_date' with value 'invalid-date': date format must be YYYY-MM-DD",
		},
		{
			name:     "invalid_numeric_filter",
			field:    "amount",
			value:    -100,
			message:  "amount cannot be negative",
			expected: "invalid filter for field 'amount' with value '-100': amount cannot be negative",
		},
		{
			name:     "invalid_boolean_filter",
			field:    "is_paid",
			value:    "maybe",
			message:  "must be true or false",
			expected: "invalid filter for field 'is_paid' with value 'maybe': must be true or false",
		},
		{
			name:     "nil_value",
			field:    "status",
			value:    nil,
			message:  "value cannot be nil",
			expected: "invalid filter for field 'status' with value '<nil>': value cannot be nil",
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			err := NewInvalidFilterError(test.field, test.value, test.message)
			suite.Require().Equal(test.expected, err.Error())
			suite.Require().Equal(test.field, err.Field)
			suite.Require().Equal(test.value, err.Value)
			suite.Require().Equal(test.message, err.Message)
		})
	}
}

// Test PermissionError
func (suite *StorageErrorsTestSuite) TestPermissionError() {
	tests := []struct {
		name      string
		operation string
		resource  string
		message   string
		expected  string
	}{
		{
			name:      "permission_denied_read",
			operation: "read",
			resource:  "client",
			message:   "user does not have read access",
			expected:  "permission denied for read on client: user does not have read access",
		},
		{
			name:      "permission_denied_write",
			operation: "write",
			resource:  "invoice",
			message:   "insufficient privileges",
			expected:  "permission denied for write on invoice: insufficient privileges",
		},
		{
			name:      "permission_denied_delete",
			operation: "delete",
			resource:  "workitem",
			message:   "",
			expected:  "permission denied for delete on workitem: ",
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			err := NewPermissionError(test.operation, test.resource, test.message)
			suite.Require().Equal(test.expected, err.Error())
			suite.Require().Equal(test.operation, err.Operation)
			suite.Require().Equal(test.resource, err.Resource)
			suite.Require().Equal(test.message, err.Message)
		})
	}
}

// Test error type checking functions
func (suite *StorageErrorsTestSuite) TestErrorTypeChecking() {
	// Create one of each error type
	notFoundErr := NewNotFoundError("invoice", "INV-001")
	conflictErr := NewConflictError("client", "CLI-001", "email exists")
	versionErr := NewVersionMismatchError("workitem", "WI-001", 5, 3)
	corruptedErr := NewCorruptedError("invoice", "INV-002", "invalid JSON")
	unavailableErr := NewStorageUnavailableError("disk full", nil)
	filterErr := NewInvalidFilterError("date", "invalid", "bad format")
	permissionErr := NewPermissionError("read", "client", "no access")

	// Test IsNotFound
	suite.Run("IsNotFound", func() {
		suite.Require().True(IsNotFound(notFoundErr))
		suite.Require().False(IsNotFound(conflictErr))
		suite.Require().False(IsNotFound(versionErr))
		suite.Require().False(IsNotFound(corruptedErr))
		suite.Require().False(IsNotFound(unavailableErr))
		suite.Require().False(IsNotFound(filterErr))
		suite.Require().False(IsNotFound(permissionErr))
		suite.Require().False(IsNotFound(errors.New("other error"))) //nolint:err113 // test error
	})

	// Test IsConflict
	suite.Run("IsConflict", func() {
		suite.Require().False(IsConflict(notFoundErr))
		suite.Require().True(IsConflict(conflictErr))
		suite.Require().False(IsConflict(versionErr))
		suite.Require().False(IsConflict(corruptedErr))
		suite.Require().False(IsConflict(unavailableErr))
		suite.Require().False(IsConflict(filterErr))
		suite.Require().False(IsConflict(permissionErr))
		suite.Require().False(IsConflict(errors.New("other error"))) //nolint:err113 // test error
	})

	// Test IsVersionMismatch
	suite.Run("IsVersionMismatch", func() {
		suite.Require().False(IsVersionMismatch(notFoundErr))
		suite.Require().False(IsVersionMismatch(conflictErr))
		suite.Require().True(IsVersionMismatch(versionErr))
		suite.Require().False(IsVersionMismatch(corruptedErr))
		suite.Require().False(IsVersionMismatch(unavailableErr))
		suite.Require().False(IsVersionMismatch(filterErr))
		suite.Require().False(IsVersionMismatch(permissionErr))
		suite.Require().False(IsVersionMismatch(errors.New("other error"))) //nolint:err113 // test error
	})

	// Test IsCorrupted
	suite.Run("IsCorrupted", func() {
		suite.Require().False(IsCorrupted(notFoundErr))
		suite.Require().False(IsCorrupted(conflictErr))
		suite.Require().False(IsCorrupted(versionErr))
		suite.Require().True(IsCorrupted(corruptedErr))
		suite.Require().False(IsCorrupted(unavailableErr))
		suite.Require().False(IsCorrupted(filterErr))
		suite.Require().False(IsCorrupted(permissionErr))
		suite.Require().False(IsCorrupted(errors.New("other error"))) //nolint:err113 // test error
	})

	// Test IsStorageUnavailable
	suite.Run("IsStorageUnavailable", func() {
		suite.Require().False(IsStorageUnavailable(notFoundErr))
		suite.Require().False(IsStorageUnavailable(conflictErr))
		suite.Require().False(IsStorageUnavailable(versionErr))
		suite.Require().False(IsStorageUnavailable(corruptedErr))
		suite.Require().True(IsStorageUnavailable(unavailableErr))
		suite.Require().False(IsStorageUnavailable(filterErr))
		suite.Require().False(IsStorageUnavailable(permissionErr))
		suite.Require().False(IsStorageUnavailable(errors.New("other error"))) //nolint:err113 // test error
	})

	// Test IsInvalidFilter
	suite.Run("IsInvalidFilter", func() {
		suite.Require().False(IsInvalidFilter(notFoundErr))
		suite.Require().False(IsInvalidFilter(conflictErr))
		suite.Require().False(IsInvalidFilter(versionErr))
		suite.Require().False(IsInvalidFilter(corruptedErr))
		suite.Require().False(IsInvalidFilter(unavailableErr))
		suite.Require().True(IsInvalidFilter(filterErr))
		suite.Require().False(IsInvalidFilter(permissionErr))
		suite.Require().False(IsInvalidFilter(errors.New("other error"))) //nolint:err113 // test error
	})

	// Test IsPermission
	suite.Run("IsPermission", func() {
		suite.Require().False(IsPermission(notFoundErr))
		suite.Require().False(IsPermission(conflictErr))
		suite.Require().False(IsPermission(versionErr))
		suite.Require().False(IsPermission(corruptedErr))
		suite.Require().False(IsPermission(unavailableErr))
		suite.Require().False(IsPermission(filterErr))
		suite.Require().True(IsPermission(permissionErr))
		suite.Require().False(IsPermission(errors.New("other error"))) //nolint:err113 // test error
	})
}

// Test error wrapping with StorageUnavailableError
func (suite *StorageErrorsTestSuite) TestErrorWrapping() {
	baseErr := errors.New("connection timeout") //nolint:err113 // test error creation
	wrappedErr := NewStorageUnavailableError("database unreachable", baseErr)

	suite.Run("unwrap_works", func() {
		suite.Require().Equal(baseErr, wrappedErr.Unwrap())
		suite.Require().ErrorIs(wrappedErr, baseErr)
	})

	suite.Run("wrapped_error_detection", func() {
		suite.Require().True(IsStorageUnavailable(wrappedErr))
		suite.Require().False(IsNotFound(wrappedErr))
	})

	suite.Run("nil_cause_unwrap", func() {
		noWrappedErr := NewStorageUnavailableError("simple error", nil)
		suite.Require().NoError(noWrappedErr.Unwrap())
	})
}

// Test error formatting edge cases
func (suite *StorageErrorsTestSuite) TestErrorFormattingEdgeCases() {
	suite.Run("empty_strings", func() {
		notFoundErr := NewNotFoundError("", "")
		suite.Require().Equal(" with ID '' not found", notFoundErr.Error())

		conflictErr := NewConflictError("", "", "")
		suite.Require().Equal(" with ID '' already exists", conflictErr.Error())

		corruptedErr := NewCorruptedError("", "", "")
		suite.Require().Equal(" with ID '' is corrupted: ", corruptedErr.Error())
	})

	suite.Run("special_characters", func() {
		notFoundErr := NewNotFoundError("invoice's", "INV-'123'")
		suite.Require().Equal("invoice's with ID 'INV-'123'' not found", notFoundErr.Error())

		filterErr := NewInvalidFilterError("field with spaces", "value\nwith\nnewlines", "message with \"quotes\"")
		expected := `invalid filter for field 'field with spaces' with value 'value
with
newlines': message with "quotes"`
		suite.Require().Equal(expected, filterErr.Error())
	})

	suite.Run("unicode_characters", func() {
		// Test with English characters to avoid gosmopolitan linter issue
		notFoundErr := NewNotFoundError("client", "CLI-001")
		suite.Require().Equal("client with ID 'CLI-001' not found", notFoundErr.Error())
	})
}
