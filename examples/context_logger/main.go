// Example demonstrating context logger switching functionality.
package main

import (
	"context"
	"log/slog"

	planks_slog "github.com/nakat-t/planks-go/slog"
)

func main() {
	// Initialize the logger with context-aware functionality
	planks_slog.Init()
	slog.Info("Default logger initialized with context-aware functionality")

	// Sample request ID - typically this would come from an HTTP request or similar
	requestID := "req-123456"

	// Create a context for this request
	ctx := context.Background()

	// Process the request with the context
	processRequest(ctx, requestID)
}

func processRequest(ctx context.Context, reqID string) {
	slog.Info("Starting to process request")

	// Create a logger with request ID
	logger := slog.Default().With(slog.String("RequestID", reqID))

	// Add the logger to the context
	ctx = context.WithValue(ctx, planks_slog.ContextLoggerKey{}, logger)

	// Log using the default logger but with context
	// This will automatically include the RequestID field
	slog.InfoContext(ctx, "Processing request with context-aware logger")

	// Regular logging without context doesn't include RequestID
	slog.Info("This log doesn't have RequestID")

	// Pass the context to another function
	doSomething(ctx)
}

func doSomething(ctx context.Context) {
	// When logging with this context, the RequestID will be included automatically
	// even though we're using the default logger
	slog.InfoContext(ctx, "Doing something important")

	// If we create a derived context without a logger, we still get the original context's logger
	childCtx := context.WithValue(ctx, "some-key", "some-value")
	slog.InfoContext(childCtx, "Using child context")

	// Create a new background context - this won't have the RequestID
	emptyCtx := context.Background()
	slog.InfoContext(emptyCtx, "This log doesn't have RequestID either")
}
