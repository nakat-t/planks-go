// Package slog extends Go's standard log/slog package to provide
// environment-variable-based configuration for loggers.
//
// It also provides a context-aware logging functionality where logs can
// automatically use a logger stored in the context instead of the default logger.
package slog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Default values for the logger configuration.
const (
	DefaultHandlerType = "text"
	DefaultWriterType  = "stderr"
	DefaultFilePerm    = 0644
)

var (
	// ErrNoEnvVarSet is returned when no relevant environment variables are set.
	ErrNoEnvVarSet = errors.New("no relevant environment variables set")
	// ErrInvalidLevel is returned when an invalid log level is specified.
	ErrInvalidLevel = errors.New("invalid log level")
	// ErrInvalidHandlerType is returned when an invalid handler type is specified.
	ErrInvalidHandlerType = errors.New("invalid handler type")
	// ErrInvalidWriterType is returned when an invalid writer type is specified.
	ErrInvalidWriterType = errors.New("invalid writer type")
	// ErrMissingFilePath is returned when file writer is specified but no file path is provided.
	ErrMissingFilePath = errors.New("file path is required when writer type is 'file'")
	// ErrInvalidFilePermission is returned when an invalid file permission is specified.
	ErrInvalidFilePermission = errors.New("invalid file permission")
)

// Environment variable names used for configuration.
const (
	EnvLoggerLevel          = "LOGGER_LEVEL"
	EnvLoggerAddSource      = "LOGGER_ADD_SOURCE"
	EnvLoggerHandler        = "LOGGER_HANDLER"
	EnvLoggerWriter         = "LOGGER_WRITER"
	EnvLoggerWriterFilePath = "LOGGER_WRITER_FILE_PATH"
	EnvLoggerWriterNoAppend = "LOGGER_WRITER_FILE_NO_APPEND"
	EnvLoggerWriterFilePerm = "LOGGER_WRITER_FILE_PERM"
	EnvPlanksNoPanicOnError = "PLANKS_NO_PANIC_ON_ERROR"
	EnvPlanksEnvPrefix      = "PLANKS_ENV_PREFIX"
)

// ContextLoggerKey is a key for context.Context values. It is used to store
// a logger in context so that context-aware logs can use this logger instead
// of the default logger.
type ContextLoggerKey struct{}

// contextAwareHandler is a wrapper handler that checks for a logger in the context
// and delegates logging to that logger's handler if found. Otherwise, it delegates
// to its internal handler.
type contextAwareHandler struct {
	internal slog.Handler
}

// Enabled implements slog.Handler.Enabled.
func (h *contextAwareHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if ctx != nil {
		if loggerValue := ctx.Value(ContextLoggerKey{}); loggerValue != nil {
			if logger, ok := loggerValue.(*slog.Logger); ok && logger != nil {
				// Ensure handler is not itself to prevent recursive loops
				contextHandler := logger.Handler()
				if contextHandler != h {
					return contextHandler.Enabled(ctx, level)
				}
			}
		}
	}
	return h.internal.Enabled(ctx, level)
}

// Handle implements slog.Handler.Handle.
func (h *contextAwareHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if loggerValue := ctx.Value(ContextLoggerKey{}); loggerValue != nil {
			if logger, ok := loggerValue.(*slog.Logger); ok && logger != nil {
				// Ensure handler is not itself to prevent recursive loops
				contextHandler := logger.Handler()
				if contextHandler != h {
					return contextHandler.Handle(ctx, r)
				}
			}
		}
	}
	return h.internal.Handle(ctx, r)
}

// WithAttrs implements slog.Handler.WithAttrs.
func (h *contextAwareHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextAwareHandler{
		internal: h.internal.WithAttrs(attrs),
	}
}

// WithGroup implements slog.Handler.WithGroup.
func (h *contextAwareHandler) WithGroup(name string) slog.Handler {
	return &contextAwareHandler{
		internal: h.internal.WithGroup(name),
	}
}

// newContextAwareHandler creates a new handler that wraps the given handler
// with context-aware functionality.
func newContextAwareHandler(handler slog.Handler) slog.Handler {
	return &contextAwareHandler{
		internal: handler,
	}
}

func FromContext(ctx context.Context) *slog.Logger {
	if loggerValue := ctx.Value(ContextLoggerKey{}); loggerValue != nil {
		if logger, ok := loggerValue.(*slog.Logger); ok && logger != nil {
			return logger
		}
	}
	return slog.Default()
}

// Config represents the logger configuration derived from environment variables.
type Config struct {
	// Level is the minimum level to log.
	Level slog.Level
	// AddSource determines whether to add source information to logs.
	AddSource bool
	// HandlerType is the type of handler to use.
	HandlerType string
	// WriterType is the type of writer to use.
	WriterType string
	// WriterFilePath is the path to the log file.
	WriterFilePath string
	// WriterFileNoAppend determines whether to append to the log file.
	WriterFileNoAppend bool
	// WriterFilePerm is the permission for the log file.
	WriterFilePerm os.FileMode
	// NoPanicOnError determines whether to panic on configuration errors.
	NoPanicOnError bool
}

// ReadConfig reads the logger configuration from environment variables.
func ReadConfig() (*Config, error) {
	prefix := os.Getenv(EnvPlanksEnvPrefix)
	noPanicOnError := os.Getenv(EnvPlanksNoPanicOnError) != ""

	// Only proceed with configuration if at least one logger-related env var is set
	if !isAnyLoggerEnvVarSet(prefix) {
		return nil, nil
	}

	config := &Config{
		HandlerType:    DefaultHandlerType,
		WriterType:     DefaultWriterType,
		WriterFilePerm: DefaultFilePerm,
		NoPanicOnError: noPanicOnError,
	}

	// Parse level
	levelStr := getEnv(prefix, EnvLoggerLevel)
	if levelStr != "" {
		var level slog.Level
		if err := level.UnmarshalText([]byte(levelStr)); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidLevel, err)
		}
		config.Level = level
	}

	// Parse add source
	config.AddSource = getEnv(prefix, EnvLoggerAddSource) != ""

	// Parse handler type
	if handlerType := getEnv(prefix, EnvLoggerHandler); handlerType != "" {
		handlerType = strings.ToLower(handlerType)
		if !isValidHandlerType(handlerType) {
			return nil, fmt.Errorf("%w: %v", ErrInvalidHandlerType, handlerType)
		}
		config.HandlerType = handlerType
	}

	// Parse writer type
	if writerType := getEnv(prefix, EnvLoggerWriter); writerType != "" {
		writerType = strings.ToLower(writerType)
		if !isValidWriterType(writerType) {
			return nil, fmt.Errorf("%w: %v", ErrInvalidWriterType, writerType)
		}
		config.WriterType = writerType
	}

	// Parse file-related settings if writer type is 'file'
	if config.WriterType == "file" {
		filePath := getEnv(prefix, EnvLoggerWriterFilePath)
		if filePath == "" {
			return nil, ErrMissingFilePath
		}
		config.WriterFilePath = filePath
		config.WriterFileNoAppend = getEnv(prefix, EnvLoggerWriterNoAppend) != ""

		if permStr := getEnv(prefix, EnvLoggerWriterFilePerm); permStr != "" {
			perm, err := strconv.ParseUint(permStr, 8, 32)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrInvalidFilePermission, err)
			}
			config.WriterFilePerm = os.FileMode(perm)
		}
	}

	return config, nil
}

// isAnyLoggerEnvVarSet checks if any of the logger-related environment variables are set.
func isAnyLoggerEnvVarSet(prefix string) bool {
	envVars := []string{
		EnvLoggerLevel,
		EnvLoggerAddSource,
		EnvLoggerHandler,
		EnvLoggerWriter,
		EnvLoggerWriterFilePath,
		EnvLoggerWriterNoAppend,
		EnvLoggerWriterFilePerm,
	}

	for _, envVar := range envVars {
		if getEnv(prefix, envVar) != "" {
			return true
		}
	}

	return false
}

// isValidHandlerType checks if the given handler type is valid.
func isValidHandlerType(handlerType string) bool {
	validTypes := map[string]bool{
		"json":    true,
		"text":    true,
		"discard": true,
	}
	return validTypes[handlerType]
}

// isValidWriterType checks if the given writer type is valid.
func isValidWriterType(writerType string) bool {
	validTypes := map[string]bool{
		"stdout": true,
		"stderr": true,
		"file":   true,
	}
	return validTypes[writerType]
}

// getEnv gets an environment variable with the given prefix.
func getEnv(prefix, key string) string {
	if prefix != "" {
		return os.Getenv(prefix + "_" + key)
	}
	return os.Getenv(key)
}

// createHandler creates a handler based on the given config.
func createHandler(config *Config, w io.Writer) slog.Handler {
	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
	}

	switch config.HandlerType {
	case "json":
		return newContextAwareHandler(slog.NewJSONHandler(w, opts))
	case "text":
		return newContextAwareHandler(slog.NewTextHandler(w, opts))
	case "discard":
		return slog.DiscardHandler // Discard handler does not log anything, so no need for context awareness
	default:
		// This should never happen due to validation in ReadConfig
		return newContextAwareHandler(slog.NewTextHandler(w, opts))
	}
}

// createWriter creates a writer based on the given config.
func createWriter(config *Config) (io.Writer, error) {
	switch config.WriterType {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	case "file":
		flag := os.O_CREATE | os.O_WRONLY
		if !config.WriterFileNoAppend {
			flag |= os.O_APPEND
		} else {
			flag |= os.O_TRUNC
		}
		return os.OpenFile(config.WriterFilePath, flag, config.WriterFilePerm)
	default:
		// This should never happen due to validation in ReadConfig
		return os.Stderr, nil
	}
}

// Build creates a logger based on environment variables.
// If no relevant environment variables are set, it returns (nil, ErrNoEnvVarSet).
// If an error occurs during configuration, it returns (nil, error).
func Build() (*slog.Logger, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}

	// If no configuration is provided, return nil
	if config == nil {
		return nil, ErrNoEnvVarSet
	}

	writer, err := createWriter(config)
	if err != nil {
		return nil, err
	}

	handler := createHandler(config, writer)

	return slog.New(handler), nil
}

// Init creates a logger based on environment variables and sets it as the default logger.
// If no relevant environment variables are set, it does nothing.
// If an error occurs during configuration, it will either panic (by default) or log the error
// and continue without changing the default logger (if PLANKS_NO_PANIC_ON_ERROR is set).
func Init() {
	logger, err := Build()
	if err != nil && !errors.Is(err, ErrNoEnvVarSet) {
		if os.Getenv(EnvPlanksNoPanicOnError) != "" {
			return
		}
		panic(err)
	}

	if logger != nil {
		slog.SetDefault(logger)
	}
}
