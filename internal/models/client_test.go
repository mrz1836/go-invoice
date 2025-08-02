package models

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite

	ctx        context.Context //nolint:containedctx // Test suite context is acceptable
	cancelFunc context.CancelFunc
}

func (suite *ClientTestSuite) SetupTest() {
	suite.ctx, suite.cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)
}

func (suite *ClientTestSuite) TearDownTest() {
	suite.cancelFunc()
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (suite *ClientTestSuite) TestNewClient() {
	t := suite.T()

	tests := []struct {
		name        string
		id          ClientID
		clientName  string
		email       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "ValidClient",
			id:          "CLIENT-001",
			clientName:  "Test Client",
			email:       "test@example.com",
			expectError: false,
		},
		{
			name:        "EmptyID",
			id:          "",
			clientName:  "Test Client",
			email:       "test@example.com",
			expectError: true,
			errorMsg:    "validation failed for field 'id': is required",
		},
		{
			name:        "EmptyName",
			id:          "CLIENT-002",
			clientName:  "",
			email:       "test@example.com",
			expectError: true,
			errorMsg:    "validation failed for field 'name': is required",
		},
		{
			name:        "EmptyEmail",
			id:          "CLIENT-003",
			clientName:  "Test Client",
			email:       "",
			expectError: true,
			errorMsg:    "validation failed for field 'email': is required",
		},
		{
			name:        "InvalidEmail",
			id:          "CLIENT-004",
			clientName:  "Test Client",
			email:       "invalid-email",
			expectError: true,
			errorMsg:    "validation failed for field 'email': must be a valid email address",
		},
		{
			name:        "LongName",
			id:          "CLIENT-005",
			clientName:  strings.Repeat("a", 201),
			email:       "test@example.com",
			expectError: true,
			errorMsg:    "validation failed for field 'name': cannot exceed 200 characters",
		},
		{
			name:        "WhitespaceOnlyName",
			id:          "CLIENT-006",
			clientName:  "   ",
			email:       "test@example.com",
			expectError: true,
			errorMsg:    "validation failed for field 'name': is required",
		},
		{
			name:        "ValidComplexEmail",
			id:          "CLIENT-007",
			clientName:  "Test Client",
			email:       "test.user+tag@sub.example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			client, err := NewClient(suite.ctx, tt.id, tt.clientName, tt.email)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				assert.Equal(t, tt.id, client.ID)
				assert.Equal(t, tt.clientName, client.Name)
				assert.Equal(t, tt.email, client.Email)
				assert.True(t, client.Active)
				assert.False(t, client.CreatedAt.IsZero())
				assert.False(t, client.UpdatedAt.IsZero())
				assert.Equal(t, client.CreatedAt, client.UpdatedAt)
			}
		})
	}
}

func (suite *ClientTestSuite) TestNewClientWithContext() {
	t := suite.T()

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client, err := NewClient(ctx, "CLIENT-001", "Test Client", "test@example.com")
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, client)
}

func (suite *ClientTestSuite) TestClientValidate() {
	t := suite.T()

	tests := []struct {
		name        string
		client      Client
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidClient",
			client: Client{
				ID:        "CLIENT-001",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "ValidClientWithOptionalFields",
			client: Client{
				ID:        "CLIENT-002",
				Name:      "Test Client",
				Email:     "test@example.com",
				Phone:     "+1-555-123-4567",
				Address:   "123 Main St, City, State 12345",
				TaxID:     "12-3456789",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "EmptyID",
			client: Client{
				ID:        "",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'id': is required",
		},
		{
			name: "WhitespaceID",
			client: Client{
				ID:        "   ",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'id': is required",
		},
		{
			name: "InvalidEmailFormat",
			client: Client{
				ID:        "CLIENT-003",
				Name:      "Test Client",
				Email:     "not-an-email",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'email': must be a valid email address",
		},
		{
			name: "ShortPhone",
			client: Client{
				ID:        "CLIENT-004",
				Name:      "Test Client",
				Email:     "test@example.com",
				Phone:     "123456789", // 9 chars, min is 10
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'phone': must be between 10 and 20 characters",
		},
		{
			name: "LongPhone",
			client: Client{
				ID:        "CLIENT-005",
				Name:      "Test Client",
				Email:     "test@example.com",
				Phone:     strings.Repeat("1", 21), // 21 chars, max is 20
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'phone': must be between 10 and 20 characters",
		},
		{
			name: "LongAddress",
			client: Client{
				ID:        "CLIENT-006",
				Name:      "Test Client",
				Email:     "test@example.com",
				Address:   strings.Repeat("a", 501), // 501 chars, max is 500
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'address': cannot exceed 500 characters",
		},
		{
			name: "LongTaxID",
			client: Client{
				ID:        "CLIENT-007",
				Name:      "Test Client",
				Email:     "test@example.com",
				TaxID:     strings.Repeat("1", 51), // 51 chars, max is 50
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'tax_id': cannot exceed 50 characters",
		},
		{
			name: "MissingCreatedAt",
			client: Client{
				ID:        "CLIENT-008",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Time{},
				UpdatedAt: time.Now(),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'created_at': is required",
		},
		{
			name: "MissingUpdatedAt",
			client: Client{
				ID:        "CLIENT-009",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Time{},
			},
			expectError: true,
			errorMsg:    "validation failed for field 'updated_at': is required",
		},
		{
			name: "UpdatedBeforeCreated",
			client: Client{
				ID:        "CLIENT-010",
				Name:      "Test Client",
				Email:     "test@example.com",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'updated_at': must be on or after created_at",
		},
		{
			name: "EmptyOptionalFields",
			client: Client{
				ID:        "CLIENT-011",
				Name:      "Test Client",
				Email:     "test@example.com",
				Phone:     "",
				Address:   "",
				TaxID:     "",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.client.Validate(suite.ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (suite *ClientTestSuite) TestUpdateName() {
	t := suite.T()

	client := &Client{
		ID:        "CLIENT-001",
		Name:      "Original Name",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	originalUpdatedAt := client.UpdatedAt

	tests := []struct {
		name         string
		newName      string
		expectError  bool
		errorMsg     string
		expectedName string
	}{
		{
			name:         "ValidUpdate",
			newName:      "Updated Name",
			expectError:  false,
			expectedName: "Updated Name",
		},
		{
			name:         "TrimmedName",
			newName:      "  Trimmed Name  ",
			expectError:  false,
			expectedName: "Trimmed Name",
		},
		{
			name:        "EmptyName",
			newName:     "",
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name:        "WhitespaceOnlyName",
			newName:     "   ",
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name:        "TooLongName",
			newName:     strings.Repeat("a", 201),
			expectError: true,
			errorMsg:    "name cannot exceed 200 characters",
		},
		{
			name:         "MaxLengthName",
			newName:      strings.Repeat("a", 200),
			expectError:  false,
			expectedName: strings.Repeat("a", 200),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset client for each test
			client.Name = "Original Name"
			client.UpdatedAt = originalUpdatedAt

			err := client.UpdateName(suite.ctx, tt.newName)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Equal(t, "Original Name", client.Name)
				assert.Equal(t, originalUpdatedAt, client.UpdatedAt)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, client.Name)
				assert.True(t, client.UpdatedAt.After(originalUpdatedAt))
			}
		})
	}
}

func (suite *ClientTestSuite) TestUpdateEmail() {
	t := suite.T()

	client := &Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "original@example.com",
		Active:    true,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	originalUpdatedAt := client.UpdatedAt

	tests := []struct {
		name          string
		newEmail      string
		expectError   bool
		errorMsg      string
		expectedEmail string
	}{
		{
			name:          "ValidUpdate",
			newEmail:      "updated@example.com",
			expectError:   false,
			expectedEmail: "updated@example.com",
		},
		{
			name:          "TrimmedEmail",
			newEmail:      "  trimmed@example.com  ",
			expectError:   false,
			expectedEmail: "trimmed@example.com",
		},
		{
			name:        "EmptyEmail",
			newEmail:    "",
			expectError: true,
			errorMsg:    "email cannot be empty",
		},
		{
			name:        "WhitespaceOnlyEmail",
			newEmail:    "   ",
			expectError: true,
			errorMsg:    "email cannot be empty",
		},
		{
			name:        "InvalidEmailFormat",
			newEmail:    "not-an-email",
			expectError: true,
			errorMsg:    "email must be a valid email address",
		},
		{
			name:          "ComplexValidEmail",
			newEmail:      "test.user+tag@sub.example.co.uk",
			expectError:   false,
			expectedEmail: "test.user+tag@sub.example.co.uk",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset client for each test
			client.Email = "original@example.com"
			client.UpdatedAt = originalUpdatedAt

			err := client.UpdateEmail(suite.ctx, tt.newEmail)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Equal(t, "original@example.com", client.Email)
				assert.Equal(t, originalUpdatedAt, client.UpdatedAt)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedEmail, client.Email)
				assert.True(t, client.UpdatedAt.After(originalUpdatedAt))
			}
		})
	}
}

func (suite *ClientTestSuite) TestUpdatePhone() {
	t := suite.T()

	client := &Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		Phone:     "1234567890",
		Active:    true,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	originalUpdatedAt := client.UpdatedAt

	tests := []struct {
		name          string
		newPhone      string
		expectError   bool
		errorMsg      string
		expectedPhone string
	}{
		{
			name:          "ValidUpdate",
			newPhone:      "9876543210",
			expectError:   false,
			expectedPhone: "9876543210",
		},
		{
			name:          "TrimmedPhone",
			newPhone:      "  9876543210  ",
			expectError:   false,
			expectedPhone: "9876543210",
		},
		{
			name:          "EmptyPhone",
			newPhone:      "",
			expectError:   false,
			expectedPhone: "",
		},
		{
			name:          "WhitespaceOnlyPhone",
			newPhone:      "   ",
			expectError:   false,
			expectedPhone: "",
		},
		{
			name:        "TooShortPhone",
			newPhone:    "123456789", // 9 chars
			expectError: true,
			errorMsg:    "phone must be between 10 and 20 characters",
		},
		{
			name:        "TooLongPhone",
			newPhone:    strings.Repeat("1", 21), // 21 chars
			expectError: true,
			errorMsg:    "phone must be between 10 and 20 characters",
		},
		{
			name:          "MinLengthPhone",
			newPhone:      "1234567890", // 10 chars
			expectError:   false,
			expectedPhone: "1234567890",
		},
		{
			name:          "MaxLengthPhone",
			newPhone:      strings.Repeat("1", 20), // 20 chars
			expectError:   false,
			expectedPhone: strings.Repeat("1", 20),
		},
		{
			name:          "FormattedPhone",
			newPhone:      "+1-555-123-4567",
			expectError:   false,
			expectedPhone: "+1-555-123-4567",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset client for each test
			client.Phone = "1234567890"
			client.UpdatedAt = originalUpdatedAt

			err := client.UpdatePhone(suite.ctx, tt.newPhone)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Equal(t, "1234567890", client.Phone)
				assert.Equal(t, originalUpdatedAt, client.UpdatedAt)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPhone, client.Phone)
				assert.True(t, client.UpdatedAt.After(originalUpdatedAt))
			}
		})
	}
}

func (suite *ClientTestSuite) TestUpdateAddress() {
	t := suite.T()

	client := &Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		Address:   "123 Main St",
		Active:    true,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	originalUpdatedAt := client.UpdatedAt

	tests := []struct {
		name            string
		newAddress      string
		expectError     bool
		errorMsg        string
		expectedAddress string
	}{
		{
			name:            "ValidUpdate",
			newAddress:      "456 New Avenue, City, State 12345",
			expectError:     false,
			expectedAddress: "456 New Avenue, City, State 12345",
		},
		{
			name:            "TrimmedAddress",
			newAddress:      "  789 Trimmed Blvd  ",
			expectError:     false,
			expectedAddress: "789 Trimmed Blvd",
		},
		{
			name:            "EmptyAddress",
			newAddress:      "",
			expectError:     false,
			expectedAddress: "",
		},
		{
			name:            "WhitespaceOnlyAddress",
			newAddress:      "   ",
			expectError:     false,
			expectedAddress: "",
		},
		{
			name:        "TooLongAddress",
			newAddress:  strings.Repeat("a", 501),
			expectError: true,
			errorMsg:    "address cannot exceed 500 characters",
		},
		{
			name:            "MaxLengthAddress",
			newAddress:      strings.Repeat("a", 500),
			expectError:     false,
			expectedAddress: strings.Repeat("a", 500),
		},
		{
			name:            "MultilineAddress",
			newAddress:      "123 Main St\nSuite 100\nCity, State 12345",
			expectError:     false,
			expectedAddress: "123 Main St\nSuite 100\nCity, State 12345",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset client for each test
			client.Address = "123 Main St"
			client.UpdatedAt = originalUpdatedAt

			err := client.UpdateAddress(suite.ctx, tt.newAddress)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Equal(t, "123 Main St", client.Address)
				assert.Equal(t, originalUpdatedAt, client.UpdatedAt)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedAddress, client.Address)
				assert.True(t, client.UpdatedAt.After(originalUpdatedAt))
			}
		})
	}
}

func (suite *ClientTestSuite) TestUpdateTaxID() {
	t := suite.T()

	client := &Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		TaxID:     "12-3456789",
		Active:    true,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	originalUpdatedAt := client.UpdatedAt

	tests := []struct {
		name          string
		newTaxID      string
		expectError   bool
		errorMsg      string
		expectedTaxID string
	}{
		{
			name:          "ValidUpdate",
			newTaxID:      "98-7654321",
			expectError:   false,
			expectedTaxID: "98-7654321",
		},
		{
			name:          "TrimmedTaxID",
			newTaxID:      "  87-6543210  ",
			expectError:   false,
			expectedTaxID: "87-6543210",
		},
		{
			name:          "EmptyTaxID",
			newTaxID:      "",
			expectError:   false,
			expectedTaxID: "",
		},
		{
			name:          "WhitespaceOnlyTaxID",
			newTaxID:      "   ",
			expectError:   false,
			expectedTaxID: "",
		},
		{
			name:        "TooLongTaxID",
			newTaxID:    strings.Repeat("1", 51),
			expectError: true,
			errorMsg:    "tax ID cannot exceed 50 characters",
		},
		{
			name:          "MaxLengthTaxID",
			newTaxID:      strings.Repeat("1", 50),
			expectError:   false,
			expectedTaxID: strings.Repeat("1", 50),
		},
		{
			name:          "InternationalTaxID",
			newTaxID:      "GB123456789",
			expectError:   false,
			expectedTaxID: "GB123456789",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Reset client for each test
			client.TaxID = "12-3456789"
			client.UpdatedAt = originalUpdatedAt

			err := client.UpdateTaxID(suite.ctx, tt.newTaxID)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Equal(t, "12-3456789", client.TaxID)
				assert.Equal(t, originalUpdatedAt, client.UpdatedAt)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTaxID, client.TaxID)
				assert.True(t, client.UpdatedAt.After(originalUpdatedAt))
			}
		})
	}
}

func (suite *ClientTestSuite) TestActivateDeactivate() {
	t := suite.T()

	client := &Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	originalUpdatedAt := client.UpdatedAt

	// Test deactivation
	err := client.Deactivate(suite.ctx)
	require.NoError(t, err)
	assert.False(t, client.Active)
	assert.True(t, client.UpdatedAt.After(originalUpdatedAt))

	// Update the reference time
	deactivatedUpdatedAt := client.UpdatedAt

	// Test activation
	err = client.Activate(suite.ctx)
	require.NoError(t, err)
	assert.True(t, client.Active)
	assert.True(t, client.UpdatedAt.After(deactivatedUpdatedAt))
}

func (suite *ClientTestSuite) TestContextCancellation() {
	t := suite.T()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &Client{
		ID:        "CLIENT-001",
		Name:      "Test Client",
		Email:     "test@example.com",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test all update methods with canceled context
	err := client.UpdateName(ctx, "New Name")
	assert.Equal(t, context.Canceled, err)

	err = client.UpdateEmail(ctx, "new@example.com")
	assert.Equal(t, context.Canceled, err)

	err = client.UpdatePhone(ctx, "1234567890")
	assert.Equal(t, context.Canceled, err)

	err = client.UpdateAddress(ctx, "New Address")
	assert.Equal(t, context.Canceled, err)

	err = client.UpdateTaxID(ctx, "NEW-TAX-ID")
	assert.Equal(t, context.Canceled, err)

	err = client.Activate(ctx)
	assert.Equal(t, context.Canceled, err)

	err = client.Deactivate(ctx)
	assert.Equal(t, context.Canceled, err)

	// Validate with canceled context
	err = client.Validate(ctx)
	assert.Equal(t, context.Canceled, err)
}

func (suite *ClientTestSuite) TestGetDisplayName() {
	t := suite.T()

	tests := []struct {
		name         string
		client       Client
		expectedName string
	}{
		{
			name: "WithName",
			client: Client{
				ID:   "CLIENT-001",
				Name: "Test Client",
			},
			expectedName: "Test Client",
		},
		{
			name: "WithoutName",
			client: Client{
				ID:   "CLIENT-002",
				Name: "",
			},
			expectedName: "CLIENT-002",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			assert.Equal(t, tt.expectedName, tt.client.GetDisplayName())
		})
	}
}

func (suite *ClientTestSuite) TestGetContactInfo() {
	t := suite.T()

	tests := []struct {
		name         string
		client       Client
		expectedInfo string
	}{
		{
			name: "EmailOnly",
			client: Client{
				Email: "test@example.com",
			},
			expectedInfo: "test@example.com",
		},
		{
			name: "PhoneOnly",
			client: Client{
				Phone: "+1-555-123-4567",
			},
			expectedInfo: "+1-555-123-4567",
		},
		{
			name: "EmailAndPhone",
			client: Client{
				Email: "test@example.com",
				Phone: "+1-555-123-4567",
			},
			expectedInfo: "test@example.com | +1-555-123-4567",
		},
		{
			name: "NoContactInfo",
			client: Client{
				Email: "",
				Phone: "",
			},
			expectedInfo: "",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			assert.Equal(t, tt.expectedInfo, tt.client.GetContactInfo())
		})
	}
}

func (suite *ClientTestSuite) TestHasCompleteInfo() {
	t := suite.T()

	tests := []struct {
		name     string
		client   Client
		expected bool
	}{
		{
			name: "CompleteInfo",
			client: Client{
				Name:    "Test Client",
				Email:   "test@example.com",
				Address: "123 Main St",
			},
			expected: true,
		},
		{
			name: "MissingName",
			client: Client{
				Name:    "",
				Email:   "test@example.com",
				Address: "123 Main St",
			},
			expected: false,
		},
		{
			name: "MissingEmail",
			client: Client{
				Name:    "Test Client",
				Email:   "",
				Address: "123 Main St",
			},
			expected: false,
		},
		{
			name: "MissingAddress",
			client: Client{
				Name:    "Test Client",
				Email:   "test@example.com",
				Address: "",
			},
			expected: false,
		},
		{
			name: "AllMissing",
			client: Client{
				Name:    "",
				Email:   "",
				Address: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			assert.Equal(t, tt.expected, tt.client.HasCompleteInfo())
		})
	}
}

func (suite *ClientTestSuite) TestCreateClientRequestValidate() {
	t := suite.T()

	tests := []struct {
		name        string
		request     CreateClientRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidRequest",
			request: CreateClientRequest{
				Name:  "Test Client",
				Email: "test@example.com",
			},
			expectError: false,
		},
		{
			name: "ValidRequestWithOptionalFields",
			request: CreateClientRequest{
				Name:    "Test Client",
				Email:   "test@example.com",
				Phone:   "+1-555-123-4567",
				Address: "123 Main St, City, State 12345",
				TaxID:   "12-3456789",
			},
			expectError: false,
		},
		{
			name: "EmptyName",
			request: CreateClientRequest{
				Name:  "",
				Email: "test@example.com",
			},
			expectError: true,
			errorMsg:    "validation failed for field 'name': is required",
		},
		{
			name: "EmptyEmail",
			request: CreateClientRequest{
				Name:  "Test Client",
				Email: "",
			},
			expectError: true,
			errorMsg:    "validation failed for field 'email': is required",
		},
		{
			name: "InvalidEmail",
			request: CreateClientRequest{
				Name:  "Test Client",
				Email: "invalid-email",
			},
			expectError: true,
			errorMsg:    "validation failed for field 'email': must be a valid email address",
		},
		{
			name: "InvalidPhone",
			request: CreateClientRequest{
				Name:  "Test Client",
				Email: "test@example.com",
				Phone: "123", // Too short
			},
			expectError: true,
			errorMsg:    "validation failed for field 'phone': must be between 10 and 20 characters",
		},
		{
			name: "LongName",
			request: CreateClientRequest{
				Name:  strings.Repeat("a", 201),
				Email: "test@example.com",
			},
			expectError: true,
			errorMsg:    "validation failed for field 'name': cannot exceed 200 characters",
		},
		{
			name: "LongAddress",
			request: CreateClientRequest{
				Name:    "Test Client",
				Email:   "test@example.com",
				Address: strings.Repeat("a", 501),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'address': cannot exceed 500 characters",
		},
		{
			name: "LongTaxID",
			request: CreateClientRequest{
				Name:  "Test Client",
				Email: "test@example.com",
				TaxID: strings.Repeat("1", 51),
			},
			expectError: true,
			errorMsg:    "validation failed for field 'tax_id': cannot exceed 50 characters",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.request.Validate(suite.ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
