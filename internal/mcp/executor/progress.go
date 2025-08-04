package executor

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Progress errors
var (
	ErrOperationCanceled      = errors.New("operation canceled")
	ErrOperationAlreadyExists = errors.New("operation already exists")
	ErrOperationNotFound      = errors.New("operation not found")
)

// ProgressTracker manages progress tracking for command execution.
type ProgressTracker interface {
	// StartOperation starts tracking a new operation.
	StartOperation(ctx context.Context, operationID string, description string, totalSteps int) (*Operation, error)

	// GetOperation retrieves an active operation by ID.
	GetOperation(ctx context.Context, operationID string) (*Operation, error)

	// ListOperations lists all active operations.
	ListOperations(ctx context.Context) ([]*Operation, error)

	// Subscribe adds a progress callback for an operation.
	Subscribe(ctx context.Context, operationID string, callback ProgressFunc) error

	// Unsubscribe removes a progress callback.
	Unsubscribe(ctx context.Context, operationID string, callback ProgressFunc) error
}

// Operation represents a tracked operation with progress.
type Operation struct {
	// ID is the unique operation identifier
	ID string

	// Description describes what the operation is doing
	Description string

	// StartTime is when the operation started
	StartTime time.Time

	// EndTime is when the operation completed (zero if still running)
	EndTime time.Time

	// TotalSteps is the total number of steps (0 for indeterminate)
	TotalSteps int

	// CurrentStep is the current step number
	CurrentStep int

	// CurrentStage describes the current operation stage
	CurrentStage string

	// SubOperations contains nested operations
	SubOperations []*Operation

	// Error contains any error that occurred
	Error error

	// Metadata contains additional operation-specific data
	Metadata map[string]interface{}

	// mu protects concurrent access
	mu sync.RWMutex

	// callbacks are progress update callbacks
	callbacks []ProgressFunc

	// canceled indicates if the operation was canceled
	canceled atomic.Bool

	// completed indicates if the operation is complete
	completed atomic.Bool
}

// DefaultProgressTracker implements ProgressTracker with concurrent operation support.
type DefaultProgressTracker struct {
	logger     Logger
	operations sync.Map // map[string]*Operation
}

// NewDefaultProgressTracker creates a new progress tracker.
func NewDefaultProgressTracker(logger Logger) *DefaultProgressTracker {
	if logger == nil {
		panic("logger is required")
	}

	return &DefaultProgressTracker{
		logger: logger,
	}
}

// StartOperation starts tracking a new operation.
func (t *DefaultProgressTracker) StartOperation(ctx context.Context, operationID string, description string, totalSteps int) (*Operation, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Check if operation already exists
	if _, exists := t.operations.Load(operationID); exists {
		return nil, fmt.Errorf("%w: %s", ErrOperationAlreadyExists, operationID)
	}

	// Create new operation
	op := &Operation{
		ID:           operationID,
		Description:  description,
		StartTime:    time.Now(),
		TotalSteps:   totalSteps,
		CurrentStep:  0,
		CurrentStage: "initializing",
		Metadata:     make(map[string]interface{}),
		callbacks:    make([]ProgressFunc, 0),
	}

	// Store operation
	t.operations.Store(operationID, op)

	// Start progress monitoring
	go t.monitorOperation(ctx, op)

	t.logger.Info("operation started",
		"operationID", operationID,
		"description", description,
		"totalSteps", totalSteps,
	)

	return op, nil
}

// GetOperation retrieves an active operation by ID.
func (t *DefaultProgressTracker) GetOperation(ctx context.Context, operationID string) (*Operation, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	value, exists := t.operations.Load(operationID)
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrOperationNotFound, operationID)
	}

	return value.(*Operation), nil
}

// ListOperations lists all active operations.
func (t *DefaultProgressTracker) ListOperations(ctx context.Context) ([]*Operation, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var operations []*Operation
	t.operations.Range(func(_, value interface{}) bool {
		op := value.(*Operation)
		if !op.completed.Load() {
			operations = append(operations, op)
		}
		return true
	})

	return operations, nil
}

// Subscribe adds a progress callback for an operation.
func (t *DefaultProgressTracker) Subscribe(ctx context.Context, operationID string, callback ProgressFunc) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	op, err := t.GetOperation(ctx, operationID)
	if err != nil {
		return err
	}

	op.mu.Lock()
	op.callbacks = append(op.callbacks, callback)
	op.mu.Unlock()

	t.logger.Debug("callback subscribed",
		"operationID", operationID,
	)

	return nil
}

// Unsubscribe removes a progress callback.
func (t *DefaultProgressTracker) Unsubscribe(ctx context.Context, operationID string, callback ProgressFunc) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	op, err := t.GetOperation(ctx, operationID)
	if err != nil {
		return err
	}

	op.mu.Lock()
	defer op.mu.Unlock()

	// Remove callback (comparison by function pointer address)
	filtered := make([]ProgressFunc, 0, len(op.callbacks))
	for _, cb := range op.callbacks {
		if fmt.Sprintf("%p", cb) != fmt.Sprintf("%p", callback) {
			filtered = append(filtered, cb)
		}
	}
	op.callbacks = filtered

	return nil
}

// monitorOperation monitors an operation and sends progress updates.
func (t *DefaultProgressTracker) monitorOperation(ctx context.Context, op *Operation) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// Clean up when done
	defer func() {
		t.operations.Delete(op.ID)
		t.logger.Info("operation completed",
			"operationID", op.ID,
			"duration", time.Since(op.StartTime),
			"error", op.Error,
		)
	}()

	for {
		select {
		case <-ctx.Done():
			op.Cancel()
			return
		case <-ticker.C:
			if op.completed.Load() {
				// Send final update
				op.sendUpdate()
				return
			}
			// Send periodic update
			op.sendUpdate()
		}
	}
}

// Operation methods

// UpdateProgress updates the operation progress.
func (op *Operation) UpdateProgress(step int, stage string, message string) {
	op.mu.Lock()
	defer op.mu.Unlock()

	op.CurrentStep = step
	op.CurrentStage = stage

	// Calculate percentage
	percent := 0
	if op.TotalSteps > 0 {
		percent = int((float64(step) / float64(op.TotalSteps)) * 100)
		if percent > 100 {
			percent = 100
		}
	}

	// Create and send update
	update := &ProgressUpdate{
		Stage:     stage,
		Percent:   percent,
		Message:   message,
		Current:   step,
		Total:     op.TotalSteps,
		Timestamp: time.Now(),
	}

	for _, callback := range op.callbacks {
		go callback(update)
	}
}

// AddSubOperation adds a nested sub-operation.
func (op *Operation) AddSubOperation(subOp *Operation) {
	op.mu.Lock()
	defer op.mu.Unlock()

	op.SubOperations = append(op.SubOperations, subOp)
}

// SetMetadata sets metadata for the operation.
func (op *Operation) SetMetadata(key string, value interface{}) {
	op.mu.Lock()
	defer op.mu.Unlock()

	op.Metadata[key] = value
}

// Complete marks the operation as complete.
func (op *Operation) Complete(err error) {
	op.mu.Lock()
	defer op.mu.Unlock()

	if op.completed.Load() {
		return
	}

	op.EndTime = time.Now()
	op.Error = err
	op.completed.Store(true)

	// Update stage
	if err != nil {
		op.CurrentStage = "failed"
	} else {
		op.CurrentStage = "completed"
		op.CurrentStep = op.TotalSteps
	}
}

// Cancel cancels the operation.
func (op *Operation) Cancel() {
	op.canceled.Store(true)
	op.Complete(ErrOperationCanceled)
}

// IsCanceled checks if the operation was canceled.
func (op *Operation) IsCanceled() bool {
	return op.canceled.Load()
}

// IsComplete checks if the operation is complete.
func (op *Operation) IsComplete() bool {
	return op.completed.Load()
}

// GetProgress returns the current progress percentage.
func (op *Operation) GetProgress() int {
	op.mu.RLock()
	defer op.mu.RUnlock()

	if op.TotalSteps == 0 {
		return -1 // Indeterminate
	}

	percent := int((float64(op.CurrentStep) / float64(op.TotalSteps)) * 100)
	if percent > 100 {
		percent = 100
	}
	return percent
}

// GetDuration returns the operation duration.
func (op *Operation) GetDuration() time.Duration {
	op.mu.RLock()
	defer op.mu.RUnlock()

	if op.EndTime.IsZero() {
		return time.Since(op.StartTime)
	}
	return op.EndTime.Sub(op.StartTime)
}

// sendUpdate sends a progress update to all callbacks.
func (op *Operation) sendUpdate() {
	op.mu.RLock()
	defer op.mu.RUnlock()

	update := &ProgressUpdate{
		Stage:     op.CurrentStage,
		Percent:   op.GetProgress(),
		Message:   fmt.Sprintf("%s: %s", op.Description, op.CurrentStage),
		Current:   op.CurrentStep,
		Total:     op.TotalSteps,
		Timestamp: time.Now(),
	}

	// Add metadata
	if op.Error != nil {
		update.Message = fmt.Sprintf("%s: %v", op.Description, op.Error)
	}

	// Send to all callbacks
	for _, callback := range op.callbacks {
		go callback(update)
	}
}

// LongOperationTracker wraps command execution with progress tracking.
type LongOperationTracker struct {
	executor CommandExecutor
	tracker  ProgressTracker
	logger   Logger
}

// NewLongOperationTracker creates a tracker for long-running operations.
func NewLongOperationTracker(executor CommandExecutor, tracker ProgressTracker, logger Logger) *LongOperationTracker {
	if executor == nil {
		panic("executor is required")
	}
	if tracker == nil {
		panic("tracker is required")
	}
	if logger == nil {
		panic("logger is required")
	}

	return &LongOperationTracker{
		executor: executor,
		tracker:  tracker,
		logger:   logger,
	}
}

// ExecuteWithProgress executes a command with progress tracking.
func (l *LongOperationTracker) ExecuteWithProgress(ctx context.Context, req *ExecutionRequest, operationID string) (*ExecutionResponse, error) {
	// Start operation tracking
	op, err := l.tracker.StartOperation(ctx, operationID, fmt.Sprintf("Executing %s", req.Command), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to start operation tracking: %w", err)
	}

	// Set up progress callback
	req.ProgressCallback = func(update *ProgressUpdate) {
		op.UpdateProgress(update.Current, update.Stage, update.Message)
	}

	// Execute command
	op.UpdateProgress(0, "starting", "Preparing command execution")
	resp, err := l.executor.Execute(ctx, req)

	// Complete operation
	op.Complete(err)

	return resp, err
}

// EstimateOperationTime estimates time for an operation based on historical data.
func EstimateOperationTime(command string, args []string, historicalData map[string]time.Duration) time.Duration {
	// Build operation key
	key := fmt.Sprintf("%s:%s", command, strings.Join(args, ":"))

	// Check historical data
	if duration, exists := historicalData[key]; exists {
		return duration
	}

	// Default estimates based on command type
	cmdBase := filepath.Base(command)
	switch cmdBase {
	case "import":
		return 30 * time.Second
	case "generate":
		return 20 * time.Second
	case "export":
		return 25 * time.Second
	case "summary":
		return 15 * time.Second
	default:
		return 10 * time.Second
	}
}
