# planks-go

`planks-go` is a library that extends Go's standard package `log/slog` to easily configure customized loggers based on environment variables.

## Overview

- Configure loggers using environment variables
- Automatically set up the default logger for `log/slog`
- Support for different log levels, handler types, and output destinations
- Context-aware logging functionality that automatically switches loggers based on context values

## Usage

### Automatic Initialization (Simplest)

```go
import "log/slog"
import _ "github.com/nakat-t/planks-go/slog/init"

func main() {
    // Default logger is automatically configured based on environment variables
    slog.Info("Application started")
}
```

### Explicit Initialization

```go
import "log/slog"
import planks_slog "github.com/nakat-t/planks-go/slog"

func main() {
    // Explicitly initialize and set as default logger
    planks_slog.Init()
    slog.Info("Application started")
    
    // Or, create your own logger instance
    logger := planks_slog.Build()
    logger.Info("Message from custom logger")
}
```

## Configuration via Environment Variables

### Basic Logger Settings

| Environment Variable | Description | Possible Values | Default |
|----------------------|-------------|-----------------|---------|
| `LOGGER_LEVEL` | Set log level | debug, info, warn, error, etc. | info |
| `LOGGER_ADD_SOURCE` | Include source code position | Any value (enabled if set) | Not set (disabled) |
| `LOGGER_HANDLER` | Log output format | json, text, discard | text |
| `LOGGER_WRITER` | Log destination | stdout, stderr, file | stderr |

### File Output Settings

| Environment Variable | Description | Possible Values | Default |
|----------------------|-------------|-----------------|---------|
| `LOGGER_WRITER_FILE_PATH` | Log file path | Any file path | Required (when `file` is specified) |
| `LOGGER_WRITER_FILE_NO_APPEND` | Use overwrite mode | Any value (enabled if set) | Not set (append mode) |
| `LOGGER_WRITER_FILE_PERM` | File permissions | e.g., 0644 | 0644 |

### Other Settings

| Environment Variable | Description | Possible Values | Default |
|----------------------|-------------|-----------------|---------|
| `PLANKS_NO_PANIC_ON_ERROR` | Prevent panics on errors | Any value (enabled if set) | Not set (will panic) |
| `PLANKS_ENV_PREFIX` | Change environment variable prefix | Any string | Not set |

## Examples

### Output JSON logs to stdout

```
LOGGER_LEVEL=debug LOGGER_HANDLER=json LOGGER_WRITER=stdout go run examples/auto_init/main.go
```

### Output logs to a file

```
LOGGER_LEVEL=info LOGGER_WRITER=file LOGGER_WRITER_FILE_PATH=./app.log go run examples/auto_init/main.go
```

### Using a prefix

```
PLANKS_ENV_PREFIX=APP APP_LOGGER_LEVEL=debug APP_LOGGER_HANDLER=json go run examples/auto_init/main.go
```

## Context-Aware Logging

The library provides a context-aware logging functionality that automatically uses the logger stored in a `context.Context` object when logging with `slog.InfoContext` and similar functions, even when calling through the default logger.

### Example

```go
import (
    "context"
    "log/slog"
    
    planks_slog "github.com/nakat-t/planks-go/slog"
)

func processRequest(ctx context.Context, requestID string) {
 	// If a logger has already been set for context, use that, otherwise get slog.Default()
	logger := planks_slog.FromContext(ctx)

   // Create a logger with specific attributes
    logger = logger.With(slog.String("RequestID", requestID))
    
    // Store the logger in the context
    ctx = context.WithValue(ctx, planks_slog.ContextLoggerKey{}, logger)
    
    // Log with context - automatically uses the logger from context
    // Output includes RequestID
    slog.InfoContext(ctx, "Processing request with context")

    // Log with other context - no RequestID
    slog.InfoContext(context.Background(), "Processing request with other context")
    
    // Regular logging without context - no RequestID
    slog.Info("Regular log without context")
    
    // Pass the context to other functions
    doSomething(ctx)
}

func doSomething(ctx context.Context) {
    // This will also include RequestID automatically
    slog.InfoContext(ctx, "Doing something important")
}
```

## License

See [License Information](./LICENSE)
