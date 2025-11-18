package services

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/mrz/go-invoice/internal/models"
	"github.com/stretchr/testify/suite"
)

type UUIDGeneratorTestSuite struct {
	suite.Suite

	generator *UUIDGenerator
}

func (suite *UUIDGeneratorTestSuite) SetupTest() {
	suite.generator = NewUUIDGenerator()
}

func (suite *UUIDGeneratorTestSuite) TearDownTest() {
	// No teardown needed
}

func TestUUIDGeneratorTestSuite(t *testing.T) {
	suite.Run(t, new(UUIDGeneratorTestSuite))
}

// Test NewUUIDGenerator constructor
func (suite *UUIDGeneratorTestSuite) TestNewUUIDGenerator() {
	generator := NewUUIDGenerator()
	suite.Require().NotNil(generator)
	suite.Require().IsType(&UUIDGenerator{}, generator)
}

// Test GenerateInvoiceID
func (suite *UUIDGeneratorTestSuite) TestGenerateInvoiceID() {
	suite.Run("successful_generation", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := suite.generator.GenerateInvoiceID(ctx)
		suite.Require().NoError(err)
		suite.Require().NotEmpty(id)
		suite.Require().IsType(models.InvoiceID(""), id)

		// Validate UUID format
		suite.assertValidUUID(string(id))
	})

	suite.Run("context_cancellation", func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		id, err := suite.generator.GenerateInvoiceID(cancelCtx)
		suite.Require().Error(err)
		suite.Require().Empty(id)
		suite.Require().Equal(context.Canceled, err)
	})

	suite.Run("context_timeout", func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Sleep to ensure timeout
		time.Sleep(1 * time.Millisecond)

		id, err := suite.generator.GenerateInvoiceID(timeoutCtx)
		suite.Require().Error(err)
		suite.Require().Empty(id)
		suite.Require().Equal(context.DeadlineExceeded, err)
	})

	suite.Run("multiple_generations_are_unique", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		ids := make(map[models.InvoiceID]bool)
		numGenerations := 100

		for i := 0; i < numGenerations; i++ {
			id, err := suite.generator.GenerateInvoiceID(ctx)
			suite.Require().NoError(err)
			suite.Require().NotEmpty(id)

			// Ensure uniqueness
			suite.Require().False(ids[id], "Generated duplicate ID: %s", id)
			ids[id] = true

			// Validate UUID format
			suite.assertValidUUID(string(id))
		}

		suite.Require().Len(ids, numGenerations)
	})
}

// Test GenerateClientID
func (suite *UUIDGeneratorTestSuite) TestGenerateClientID() {
	suite.Run("successful_generation", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := suite.generator.GenerateClientID(ctx)
		suite.Require().NoError(err)
		suite.Require().NotEmpty(id)
		suite.Require().IsType(models.ClientID(""), id)

		// Validate UUID format
		suite.assertValidUUID(string(id))
	})

	suite.Run("context_cancellation", func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		id, err := suite.generator.GenerateClientID(cancelCtx)
		suite.Require().Error(err)
		suite.Require().Empty(id)
		suite.Require().Equal(context.Canceled, err)
	})

	suite.Run("context_timeout", func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Sleep to ensure timeout
		time.Sleep(1 * time.Millisecond)

		id, err := suite.generator.GenerateClientID(timeoutCtx)
		suite.Require().Error(err)
		suite.Require().Empty(id)
		suite.Require().Equal(context.DeadlineExceeded, err)
	})

	suite.Run("multiple_generations_are_unique", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		ids := make(map[models.ClientID]bool)
		numGenerations := 100

		for i := 0; i < numGenerations; i++ {
			id, err := suite.generator.GenerateClientID(ctx)
			suite.Require().NoError(err)
			suite.Require().NotEmpty(id)

			// Ensure uniqueness
			suite.Require().False(ids[id], "Generated duplicate ID: %s", id)
			ids[id] = true

			// Validate UUID format
			suite.assertValidUUID(string(id))
		}

		suite.Require().Len(ids, numGenerations)
	})
}

// Test GenerateWorkItemID
func (suite *UUIDGeneratorTestSuite) TestGenerateWorkItemID() {
	suite.Run("successful_generation", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := suite.generator.GenerateWorkItemID(ctx)
		suite.Require().NoError(err)
		suite.Require().NotEmpty(id)
		suite.Require().IsType("", id)

		// Validate UUID format
		suite.assertValidUUID(id)
	})

	suite.Run("context_cancellation", func() {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		id, err := suite.generator.GenerateWorkItemID(cancelCtx)
		suite.Require().Error(err)
		suite.Require().Empty(id)
		suite.Require().Equal(context.Canceled, err)
	})

	suite.Run("context_timeout", func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Sleep to ensure timeout
		time.Sleep(1 * time.Millisecond)

		id, err := suite.generator.GenerateWorkItemID(timeoutCtx)
		suite.Require().Error(err)
		suite.Require().Empty(id)
		suite.Require().Equal(context.DeadlineExceeded, err)
	})

	suite.Run("multiple_generations_are_unique", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		ids := make(map[string]bool)
		numGenerations := 100

		for i := 0; i < numGenerations; i++ {
			id, err := suite.generator.GenerateWorkItemID(ctx)
			suite.Require().NoError(err)
			suite.Require().NotEmpty(id)

			// Ensure uniqueness
			suite.Require().False(ids[id], "Generated duplicate ID: %s", id)
			ids[id] = true

			// Validate UUID format
			suite.assertValidUUID(id)
		}

		suite.Require().Len(ids, numGenerations)
	})
}

// Test generateUUID internal method (through public methods)
func (suite *UUIDGeneratorTestSuite) TestGenerateUUIDInternal() {
	suite.Run("uuid_format_validation", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test through public methods to ensure internal generateUUID works correctly
		invoiceID, err := suite.generator.GenerateInvoiceID(ctx)
		suite.Require().NoError(err)
		suite.assertValidUUID(string(invoiceID))

		clientID, err := suite.generator.GenerateClientID(ctx)
		suite.Require().NoError(err)
		suite.assertValidUUID(string(clientID))

		workItemID, err := suite.generator.GenerateWorkItemID(ctx)
		suite.Require().NoError(err)
		suite.assertValidUUID(workItemID)
	})

	suite.Run("uuid_version_and_variant_bits", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := suite.generator.GenerateInvoiceID(ctx)
		suite.Require().NoError(err)

		uuidStr := string(id)
		// Remove hyphens to get raw hex string
		rawHex := strings.ReplaceAll(uuidStr, "-", "")

		// Check version bits (4th most significant bit of 7th byte should be 0100)
		versionByte := rawHex[12:14]
		versionInt := suite.parseHexByte(versionByte)
		suite.Require().Equal(uint8(0x40), versionInt&0xF0, "UUID version should be 4")

		// Check variant bits (2 most significant bits of 9th byte should be 10)
		variantByte := rawHex[16:18]
		variantInt := suite.parseHexByte(variantByte)
		suite.Require().Equal(uint8(0x80), variantInt&0xC0, "UUID variant should be RFC 4122")
	})
}

// Test concurrent access
func (suite *UUIDGeneratorTestSuite) TestConcurrentGeneration() {
	suite.Run("concurrent_invoice_id_generation", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		numGoroutines := 50
		numIDsPerGoroutine := 10

		results := make(chan models.InvoiceID, numGoroutines*numIDsPerGoroutine)
		errors := make(chan error, numGoroutines*numIDsPerGoroutine)

		// Start multiple goroutines generating IDs
		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < numIDsPerGoroutine; j++ {
					id, err := suite.generator.GenerateInvoiceID(ctx)
					if err != nil {
						errors <- err
						return
					}
					results <- id
				}
			}()
		}

		// Collect all results
		generatedIDs := make(map[models.InvoiceID]bool)
		for i := 0; i < numGoroutines*numIDsPerGoroutine; i++ {
			select {
			case id := <-results:
				suite.Require().False(generatedIDs[id], "Generated duplicate ID in concurrent test: %s", id)
				generatedIDs[id] = true
				suite.assertValidUUID(string(id))
			case err := <-errors:
				suite.Require().NoError(err)
			case <-time.After(10 * time.Second):
				suite.FailNow("Timeout waiting for concurrent generation results")
			}
		}

		suite.Require().Len(generatedIDs, numGoroutines*numIDsPerGoroutine)
	})
}

// Test different ID types don't collide
func (suite *UUIDGeneratorTestSuite) TestIDTypesSeparation() {
	suite.Run("different_id_types_can_be_same_value", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Generate multiple IDs and ensure types are properly handled
		invoiceIDs := make([]models.InvoiceID, 10)
		clientIDs := make([]models.ClientID, 10)
		workItemIDs := make([]string, 10)

		for i := 0; i < 10; i++ {
			invoiceID, err := suite.generator.GenerateInvoiceID(ctx)
			suite.Require().NoError(err)
			invoiceIDs[i] = invoiceID

			clientID, err := suite.generator.GenerateClientID(ctx)
			suite.Require().NoError(err)
			clientIDs[i] = clientID

			workItemID, err := suite.generator.GenerateWorkItemID(ctx)
			suite.Require().NoError(err)
			workItemIDs[i] = workItemID
		}

		// Ensure all IDs are valid UUIDs
		for _, id := range invoiceIDs {
			suite.assertValidUUID(string(id))
		}
		for _, id := range clientIDs {
			suite.assertValidUUID(string(id))
		}
		for _, id := range workItemIDs {
			suite.assertValidUUID(id)
		}

		// Ensure all IDs are unique across all types
		allIDs := make(map[string]bool)
		for _, id := range invoiceIDs {
			suite.Require().False(allIDs[string(id)], "Duplicate ID found: %s", id)
			allIDs[string(id)] = true
		}
		for _, id := range clientIDs {
			suite.Require().False(allIDs[string(id)], "Duplicate ID found: %s", id)
			allIDs[string(id)] = true
		}
		for _, id := range workItemIDs {
			suite.Require().False(allIDs[id], "Duplicate ID found: %s", id)
			allIDs[id] = true
		}

		suite.Require().Len(allIDs, 30)
	})
}

// Helper method to validate UUID format
func (suite *UUIDGeneratorTestSuite) assertValidUUID(uuid string) {
	// UUID v4 regex pattern
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	regex := regexp.MustCompile(pattern)

	suite.Require().True(regex.MatchString(uuid), "Invalid UUID format: %s", uuid)
	suite.Require().Len(uuid, 36, "UUID should be 36 characters long")
	suite.Require().Equal(4, strings.Count(uuid, "-"), "UUID should have 4 hyphens")
}

// Helper method to parse hex byte
func (suite *UUIDGeneratorTestSuite) parseHexByte(hexStr string) uint8 {
	var result uint8
	for _, char := range hexStr {
		result <<= 4
		if char >= '0' && char <= '9' {
			result |= uint8(char - '0')
		} else if char >= 'a' && char <= 'f' {
			result |= uint8(char - 'a' + 10)
		} else if char >= 'A' && char <= 'F' {
			result |= uint8(char - 'A' + 10)
		}
	}
	return result
}
