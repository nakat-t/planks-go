// Example demonstrating automatic initialization of the logger using the init package.
package main

import (
	"log/slog"

	_ "github.com/nakat-t/planks-go/slog/init" // Import for side effects
)

func main() {
	// The logger has already been configured by the init package
	// based on environment variables.

	slog.Info("Application started", "mode", "automatic initialization")
	slog.Debug("This is a debug message")
	slog.Warn("This is a warning", "code", 123)
	slog.Error("An error occurred", "err", "example error")

	// Additional structured logging
	slog.Info("User logged in",
		"userId", "user123",
		"role", "admin",
		"loginAttempts", 1)
}
