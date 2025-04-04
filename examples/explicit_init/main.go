// Example demonstrating explicit initialization of the logger.
package main

import (
	"fmt"
	"log/slog"
	"os"

	planks_slog "github.com/nakat-t/planks-go/slog"
)

func main() {
	// Explicitly initialize the logger and set as default
	planks_slog.Init()
	slog.Info("Default logger initialized", "mode", "explicit initialization")

	// Create a custom logger without setting as default
	customLogger, err := planks_slog.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build custom logger: %v\n", err)
	} else if customLogger != nil {
		customLogger.Info("Custom logger created",
			"logger", "custom",
			"configured", true)
	}

	// Use the default logger for more logging
	slog.Debug("This message uses the default logger")
	slog.Warn("Warning example", "priority", "medium")

	// Log with attributes
	slog.Info("Operation completed",
		"duration", 157,
		"unit", "ms",
		"success", true)
}
