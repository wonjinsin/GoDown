package context

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Manager manages contexts for different operations
type Manager struct {
	contexts map[string]context.CancelFunc
	mu       sync.RWMutex
	logger   zerolog.Logger
}

// NewManager creates a new context manager
func NewManager(logger zerolog.Logger) *Manager {
	return &Manager{
		contexts: make(map[string]context.CancelFunc),
		logger:   logger,
	}
}

// CreateContext creates a new context with timeout and cancellation
func (cm *Manager) CreateContext(parent context.Context, id string, timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)

	cm.mu.Lock()
	cm.contexts[id] = cancel
	cm.mu.Unlock()

	cm.logger.Debug().
		Str("context_id", id).
		Dur("timeout", timeout).
		Msg("Context created")

	// Return a wrapped cancel function that cleans up
	wrappedCancel := func() {
		cm.mu.Lock()
		delete(cm.contexts, id)
		cm.mu.Unlock()

		cancel()

		cm.logger.Debug().
			Str("context_id", id).
			Msg("Context cancelled and cleaned up")
	}

	return ctx, wrappedCancel
}

// CancelContext cancels a specific context by ID
func (cm *Manager) CancelContext(id string) bool {
	cm.mu.RLock()
	cancel, exists := cm.contexts[id]
	cm.mu.RUnlock()

	if exists {
		cancel()
		cm.logger.Info().
			Str("context_id", id).
			Msg("Context cancelled by ID")
		return true
	}

	cm.logger.Warn().
		Str("context_id", id).
		Msg("Context not found for cancellation")
	return false
}

// CancelAll cancels all managed contexts
func (cm *Manager) CancelAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	count := len(cm.contexts)
	for id, cancel := range cm.contexts {
		cancel()
		cm.logger.Debug().
			Str("context_id", id).
			Msg("Context cancelled during shutdown")
	}

	cm.contexts = make(map[string]context.CancelFunc)

	cm.logger.Info().
		Int("cancelled_count", count).
		Msg("All contexts cancelled")
}

// GetActiveCount returns the number of active contexts
func (cm *Manager) GetActiveCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.contexts)
}

// WithDeadline creates a context with a specific deadline
func (cm *Manager) WithDeadline(parent context.Context, id string, deadline time.Time) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(parent, deadline)

	cm.mu.Lock()
	cm.contexts[id] = cancel
	cm.mu.Unlock()

	cm.logger.Debug().
		Str("context_id", id).
		Time("deadline", deadline).
		Msg("Context with deadline created")

	wrappedCancel := func() {
		cm.mu.Lock()
		delete(cm.contexts, id)
		cm.mu.Unlock()

		cancel()

		cm.logger.Debug().
			Str("context_id", id).
			Msg("Deadline context cancelled and cleaned up")
	}

	return ctx, wrappedCancel
}

// WithValue creates a context with a value
func WithValue(parent context.Context, key, value interface{}) context.Context {
	return context.WithValue(parent, key, value)
}

// GetValue retrieves a value from context
func GetValue(ctx context.Context, key interface{}) interface{} {
	return ctx.Value(key)
}

// Common context keys
type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
	SessionIDKey ContextKey = "session_id"
	UserIDKey    ContextKey = "user_id"
	TraceIDKey   ContextKey = "trace_id"
	OperationKey ContextKey = "operation"
)

// Helper functions for common context operations
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return WithValue(ctx, RequestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return WithValue(ctx, SessionIDKey, sessionID)
}

func GetSessionID(ctx context.Context) string {
	if id, ok := ctx.Value(SessionIDKey).(string); ok {
		return id
	}
	return ""
}

func WithOperation(ctx context.Context, operation string) context.Context {
	return WithValue(ctx, OperationKey, operation)
}

func GetOperation(ctx context.Context) string {
	if op, ok := ctx.Value(OperationKey).(string); ok {
		return op
	}
	return ""
}
