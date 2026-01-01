package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

// NoOpLogger is a simple logger that does nothing (for testing)
type NoOpLogger struct{}

func (l *NoOpLogger) Debug(_ string, _ ...interface{}) {}
func (l *NoOpLogger) Info(_ string, _ ...interface{})  {}
func (l *NoOpLogger) Warn(_ string, _ ...interface{})  {}
func (l *NoOpLogger) Error(_ string, _ ...interface{}) {}

// ProgressTrackerTestSuite tests the DefaultProgressTracker implementation
type ProgressTrackerTestSuite struct {
	suite.Suite

	tracker *DefaultProgressTracker
	logger  *NoOpLogger
}

func (suite *ProgressTrackerTestSuite) SetupTest() {
	suite.logger = &NoOpLogger{}
	suite.tracker = NewDefaultProgressTracker(suite.logger)
}

// TestNewDefaultProgressTracker tests the constructor
func (suite *ProgressTrackerTestSuite) TestNewDefaultProgressTracker() {
	tracker := NewDefaultProgressTracker(suite.logger)
	suite.NotNil(tracker)
	suite.Equal(suite.logger, tracker.logger)
}

// TestNewDefaultProgressTrackerPanicsWithNilLogger tests constructor panics without logger
func (suite *ProgressTrackerTestSuite) TestNewDefaultProgressTrackerPanicsWithNilLogger() {
	suite.Panics(func() {
		NewDefaultProgressTracker(nil)
	})
}

// TestStartOperation tests starting a new operation
func (suite *ProgressTrackerTestSuite) TestStartOperation() {
	ctx := context.Background()

	op, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 10)

	suite.Require().NoError(err)
	suite.NotNil(op)
	suite.Equal("test-op-1", op.ID)
	suite.Equal("Test Operation", op.Description)
	suite.Equal(10, op.TotalSteps)
	suite.Equal(0, op.CurrentStep)
	suite.False(op.StartTime.IsZero())
}

// TestStartOperationDuplicate tests starting an operation with existing ID
func (suite *ProgressTrackerTestSuite) TestStartOperationDuplicate() {
	ctx := context.Background()

	// Start first operation
	_, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 10)
	suite.Require().NoError(err)

	// Try to start duplicate
	op, err := suite.tracker.StartOperation(ctx, "test-op-1", "Duplicate Operation", 5)

	suite.Require().ErrorIs(err, ErrOperationAlreadyExists)
	suite.Nil(op)
}

// TestStartOperationContextCancellation tests context cancellation
func (suite *ProgressTrackerTestSuite) TestStartOperationContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	op, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 10)

	suite.Equal(context.Canceled, err)
	suite.Nil(op)
}

// TestGetOperation tests retrieving an operation
func (suite *ProgressTrackerTestSuite) TestGetOperation() {
	ctx := context.Background()

	// Start an operation
	_, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 10)
	suite.Require().NoError(err)

	// Get the operation
	op, err := suite.tracker.GetOperation(ctx, "test-op-1")

	suite.Require().NoError(err)
	suite.NotNil(op)
	suite.Equal("test-op-1", op.ID)
}

// TestGetOperationNotFound tests getting a non-existent operation
func (suite *ProgressTrackerTestSuite) TestGetOperationNotFound() {
	ctx := context.Background()

	op, err := suite.tracker.GetOperation(ctx, "non-existent")

	suite.Require().ErrorIs(err, ErrOperationNotFound)
	suite.Nil(op)
}

// TestGetOperationContextCancellation tests context cancellation
func (suite *ProgressTrackerTestSuite) TestGetOperationContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	op, err := suite.tracker.GetOperation(ctx, "test-op-1")

	suite.Equal(context.Canceled, err)
	suite.Nil(op)
}

// TestListOperations tests listing all operations
func (suite *ProgressTrackerTestSuite) TestListOperations() {
	ctx := context.Background()

	// Start multiple operations
	_, err := suite.tracker.StartOperation(ctx, "op-1", "Operation 1", 10)
	suite.Require().NoError(err)
	_, err = suite.tracker.StartOperation(ctx, "op-2", "Operation 2", 5)
	suite.Require().NoError(err)

	// List operations
	ops, err := suite.tracker.ListOperations(ctx)

	suite.Require().NoError(err)
	suite.Len(ops, 2)
}

// TestListOperationsEmpty tests listing with no operations
func (suite *ProgressTrackerTestSuite) TestListOperationsEmpty() {
	ctx := context.Background()

	ops, err := suite.tracker.ListOperations(ctx)

	suite.Require().NoError(err)
	suite.Empty(ops)
}

// TestOperationUpdateProgress tests updating operation progress
func (suite *ProgressTrackerTestSuite) TestOperationUpdateProgress() {
	ctx := context.Background()

	op, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 10)
	suite.Require().NoError(err)

	// Update progress
	op.UpdateProgress(5, "processing", "Step 5 of 10")

	suite.Equal(5, op.CurrentStep)
	suite.Equal("processing", op.CurrentStage)
}

// TestOperationComplete tests completing an operation
func (suite *ProgressTrackerTestSuite) TestOperationComplete() {
	ctx := context.Background()

	op, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 10)
	suite.Require().NoError(err)

	// Complete the operation
	op.Complete(nil)

	suite.True(op.completed.Load())
	suite.False(op.EndTime.IsZero())
	suite.NoError(op.Error)
}

// TestOperationCompleteWithError tests completing with an error
func (suite *ProgressTrackerTestSuite) TestOperationCompleteWithError() {
	ctx := context.Background()

	op, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 10)
	suite.Require().NoError(err)

	// Complete with error
	testErr := ErrOperationCanceled
	op.Complete(testErr)

	suite.True(op.completed.Load())
	suite.Equal(testErr, op.Error)
}

// TestOperationCancel tests canceling an operation
func (suite *ProgressTrackerTestSuite) TestOperationCancel() {
	ctx := context.Background()

	op, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 10)
	suite.Require().NoError(err)

	// Cancel the operation
	op.Cancel()

	suite.True(op.IsCanceled())
}

// TestOperationGetProgress tests getting progress percentage
func (suite *ProgressTrackerTestSuite) TestOperationGetProgress() {
	ctx := context.Background()

	op, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 10)
	suite.Require().NoError(err)

	// Initial progress
	suite.Equal(0, op.GetProgress())

	// Update to 50%
	op.UpdateProgress(5, "processing", "Half done")
	suite.Equal(50, op.GetProgress())

	// Update to 100%
	op.UpdateProgress(10, "done", "Complete")
	suite.Equal(100, op.GetProgress())
}

// TestOperationGetProgressIndeterminate tests progress with no total steps
func (suite *ProgressTrackerTestSuite) TestOperationGetProgressIndeterminate() {
	ctx := context.Background()

	op, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 0) // Indeterminate
	suite.Require().NoError(err)

	// Should return -1 for indeterminate
	suite.Equal(-1, op.GetProgress())
}

// TestOperationGetDuration tests getting operation duration
func (suite *ProgressTrackerTestSuite) TestOperationGetDuration() {
	ctx := context.Background()

	op, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 10)
	suite.Require().NoError(err)

	// Duration should be positive for running operation
	duration := op.GetDuration()
	suite.GreaterOrEqual(duration.Nanoseconds(), int64(0))

	// Complete the operation
	op.Complete(nil)

	// Duration should still be accessible after completion
	duration2 := op.GetDuration()
	suite.GreaterOrEqual(duration2.Nanoseconds(), int64(0))
}

// TestOperationSetMetadata tests setting operation metadata
func (suite *ProgressTrackerTestSuite) TestOperationSetMetadata() {
	ctx := context.Background()

	op, err := suite.tracker.StartOperation(ctx, "test-op-1", "Test Operation", 10)
	suite.Require().NoError(err)

	// Set metadata
	op.SetMetadata("key1", "value1")
	op.SetMetadata("key2", 42)

	suite.NotNil(op.Metadata)
	suite.Equal("value1", op.Metadata["key1"])
	suite.Equal(42, op.Metadata["key2"])
}

// TestOperationAddSubOperation tests adding sub-operations
func (suite *ProgressTrackerTestSuite) TestOperationAddSubOperation() {
	ctx := context.Background()

	op, err := suite.tracker.StartOperation(ctx, "parent-op", "Parent Operation", 10)
	suite.Require().NoError(err)

	// Add sub-operation
	subOp := &Operation{
		ID:          "sub-op-1",
		Description: "Sub Operation 1",
	}
	op.AddSubOperation(subOp)

	suite.Len(op.SubOperations, 1)
	suite.Equal("sub-op-1", op.SubOperations[0].ID)
}

func TestProgressTrackerTestSuite(t *testing.T) {
	suite.Run(t, new(ProgressTrackerTestSuite))
}

// TestProgressErrorVariables tests that error variables are properly defined
func TestProgressErrorVariables(t *testing.T) {
	tests := []struct {
		err     error
		name    string
		wantMsg string
	}{
		{ErrOperationCanceled, "ErrOperationCanceled", "operation canceled"},
		{ErrOperationAlreadyExists, "ErrOperationAlreadyExists", "operation already exists"},
		{ErrOperationNotFound, "ErrOperationNotFound", "operation not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.wantMsg)
			}
		})
	}
}
