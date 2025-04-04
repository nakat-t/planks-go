package slog

import (
	"errors"
	"log/slog"
	"os"
	"testing"
)

func TestReadConfig(t *testing.T) {
	// Save original environment variables
	origEnvs := saveEnvVars()
	defer restoreEnvVars(origEnvs)

	tests := []struct {
		name          string
		envVars       map[string]string
		expected      *Config
		expectErr     bool
		expectedError error
	}{
		{
			name:     "No Environment Variables",
			envVars:  map[string]string{},
			expected: nil,
		},
		{
			name: "Basic Configuration",
			envVars: map[string]string{
				EnvLoggerLevel:     "info",
				EnvLoggerAddSource: "true",
				EnvLoggerHandler:   "json",
				EnvLoggerWriter:    "stdout",
			},
			expected: &Config{
				Level:          slog.LevelInfo,
				AddSource:      true,
				HandlerType:    "json",
				WriterType:     "stdout",
				WriterFilePerm: DefaultFilePerm,
				NoPanicOnError: false,
			},
		},
		{
			name: "Invalid Log Level",
			envVars: map[string]string{
				EnvLoggerLevel: "invalid",
			},
			expectErr: true,
		},
		{
			name: "Invalid Handler Type",
			envVars: map[string]string{
				EnvLoggerHandler: "invalid",
			},
			expectErr:     true,
			expectedError: ErrInvalidHandlerType,
		},
		{
			name: "Invalid Writer Type",
			envVars: map[string]string{
				EnvLoggerWriter: "invalid",
			},
			expectErr:     true,
			expectedError: ErrInvalidWriterType,
		},
		{
			name: "File Writer Without Path",
			envVars: map[string]string{
				EnvLoggerWriter: "file",
			},
			expectErr:     true,
			expectedError: ErrMissingFilePath,
		},
		{
			name: "File Writer With Path",
			envVars: map[string]string{
				EnvLoggerWriter:         "file",
				EnvLoggerWriterFilePath: "/tmp/test.log",
			},
			expected: &Config{
				Level:              0,
				AddSource:          false,
				HandlerType:        DefaultHandlerType,
				WriterType:         "file",
				WriterFilePath:     "/tmp/test.log",
				WriterFileNoAppend: false,
				WriterFilePerm:     DefaultFilePerm,
				NoPanicOnError:     false,
			},
		},
		{
			name: "File Writer With Invalid Permission",
			envVars: map[string]string{
				EnvLoggerWriter:         "file",
				EnvLoggerWriterFilePath: "/tmp/test.log",
				EnvLoggerWriterFilePerm: "invalid",
			},
			expectErr:     true,
			expectedError: ErrInvalidFilePermission,
		},
		{
			name: "With Panic Prevention",
			envVars: map[string]string{
				EnvLoggerLevel:          "debug",
				EnvPlanksNoPanicOnError: "true",
			},
			expected: &Config{
				Level:          slog.LevelDebug,
				AddSource:      false,
				HandlerType:    DefaultHandlerType,
				WriterType:     DefaultWriterType,
				WriterFilePerm: DefaultFilePerm,
				NoPanicOnError: true,
			},
		},
		{
			name: "With Prefix",
			envVars: map[string]string{
				EnvPlanksEnvPrefix:       "TEST",
				"TEST_" + EnvLoggerLevel: "warn",
			},
			expected: &Config{
				Level:          slog.LevelWarn,
				AddSource:      false,
				HandlerType:    DefaultHandlerType,
				WriterType:     DefaultWriterType,
				WriterFilePerm: DefaultFilePerm,
				NoPanicOnError: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnvVars()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Call ReadConfig
			config, err := ReadConfig()

			// Check for expected error
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tt.expectedError != nil && !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error to be '%v' but got '%v'", tt.expectedError, err)
				}
				return
			}

			// Check for unexpected error
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check if config matches expected
			if tt.expected == nil {
				if config != nil {
					t.Errorf("expected nil config but got %+v", config)
				}
				return
			}

			if config == nil {
				t.Errorf("expected config %+v but got nil", tt.expected)
				return
			}

			// Compare fields
			if config.Level != tt.expected.Level {
				t.Errorf("Level: expected %v, got %v", tt.expected.Level, config.Level)
			}
			if config.AddSource != tt.expected.AddSource {
				t.Errorf("AddSource: expected %v, got %v", tt.expected.AddSource, config.AddSource)
			}
			if config.HandlerType != tt.expected.HandlerType {
				t.Errorf("HandlerType: expected %v, got %v", tt.expected.HandlerType, config.HandlerType)
			}
			if config.WriterType != tt.expected.WriterType {
				t.Errorf("WriterType: expected %v, got %v", tt.expected.WriterType, config.WriterType)
			}
			if config.WriterFilePath != tt.expected.WriterFilePath {
				t.Errorf("WriterFilePath: expected %v, got %v", tt.expected.WriterFilePath, config.WriterFilePath)
			}
			if config.WriterFileNoAppend != tt.expected.WriterFileNoAppend {
				t.Errorf("WriterFileNoAppend: expected %v, got %v", tt.expected.WriterFileNoAppend, config.WriterFileNoAppend)
			}
			if config.WriterFilePerm != tt.expected.WriterFilePerm {
				t.Errorf("WriterFilePerm: expected %v, got %v", tt.expected.WriterFilePerm, config.WriterFilePerm)
			}
			if config.NoPanicOnError != tt.expected.NoPanicOnError {
				t.Errorf("NoPanicOnError: expected %v, got %v", tt.expected.NoPanicOnError, config.NoPanicOnError)
			}
		})
	}
}

func TestCreateHandler(t *testing.T) {
	config := &Config{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	// Test JSON handler
	config.HandlerType = "json"
	jsonHandler := createHandler(config, os.Stderr)
	if jsonHandler == nil {
		t.Errorf("createHandler returned nil for JSON handler")
	}

	// Test Text handler
	config.HandlerType = "text"
	textHandler := createHandler(config, os.Stderr)
	if textHandler == nil {
		t.Errorf("createHandler returned nil for Text handler")
	}

	// Test Discard handler
	config.HandlerType = "discard"
	discardHandler := createHandler(config, os.Stderr)
	if discardHandler == nil {
		t.Errorf("createHandler returned nil for Discard handler")
	}
}

func TestBuild(t *testing.T) {
	// Save original environment variables
	origEnvs := saveEnvVars()
	defer restoreEnvVars(origEnvs)

	// Test with no environment variables
	clearEnvVars()
	logger, err := Build()
	if !errors.Is(err, ErrNoEnvVarSet) {
		t.Errorf("unexpected error: %v", err)
	}
	if logger != nil {
		t.Errorf("expected nil logger when no env vars set")
	}

	// Test with valid configuration
	os.Setenv(EnvLoggerLevel, "info")
	os.Setenv(EnvLoggerHandler, "json")
	logger, err = Build()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if logger == nil {
		t.Errorf("expected non-nil logger with valid config")
	}

	// Test with invalid configuration
	os.Setenv(EnvLoggerHandler, "invalid")
	logger, err = Build()
	if err == nil {
		t.Errorf("expected error for invalid handler type")
	}
	if logger != nil {
		t.Errorf("expected nil logger with invalid config")
	}
}

func TestCreateWriter(t *testing.T) {
	// Test stdout writer
	config := &Config{
		WriterType: "stdout",
	}
	writer, err := createWriter(config)
	if err != nil {
		t.Errorf("unexpected error creating stdout writer: %v", err)
	}
	if writer != os.Stdout {
		t.Errorf("expected os.Stdout but got different writer")
	}

	// Test stderr writer
	config.WriterType = "stderr"
	writer, err = createWriter(config)
	if err != nil {
		t.Errorf("unexpected error creating stderr writer: %v", err)
	}
	if writer != os.Stderr {
		t.Errorf("expected os.Stderr but got different writer")
	}

	// We can't easily test file writer without creating actual files,
	// but we can verify the code path works without errors
	tempFile := os.TempDir() + "/planks-test.log"
	defer os.Remove(tempFile)

	config.WriterType = "file"
	config.WriterFilePath = tempFile
	config.WriterFilePerm = 0644

	// Test append mode
	config.WriterFileNoAppend = false
	writer, err = createWriter(config)
	if err != nil {
		t.Errorf("unexpected error creating file writer (append): %v", err)
	}
	file, ok := writer.(*os.File)
	if !ok {
		t.Errorf("expected *os.File but got %T", writer)
	} else {
		file.Close()
	}

	// Test truncate mode
	config.WriterFileNoAppend = true
	writer, err = createWriter(config)
	if err != nil {
		t.Errorf("unexpected error creating file writer (truncate): %v", err)
	}
	file, ok = writer.(*os.File)
	if !ok {
		t.Errorf("expected *os.File but got %T", writer)
	} else {
		file.Close()
	}
}

// Helper functions for managing environment variables in tests
func saveEnvVars() map[string]string {
	envVars := []string{
		EnvLoggerLevel,
		EnvLoggerAddSource,
		EnvLoggerHandler,
		EnvLoggerWriter,
		EnvLoggerWriterFilePath,
		EnvLoggerWriterNoAppend,
		EnvLoggerWriterFilePerm,
		EnvPlanksNoPanicOnError,
		EnvPlanksEnvPrefix,
	}

	saved := make(map[string]string)
	for _, env := range envVars {
		saved[env] = os.Getenv(env)
	}
	return saved
}

func clearEnvVars() {
	envVars := []string{
		EnvLoggerLevel,
		EnvLoggerAddSource,
		EnvLoggerHandler,
		EnvLoggerWriter,
		EnvLoggerWriterFilePath,
		EnvLoggerWriterNoAppend,
		EnvLoggerWriterFilePerm,
		EnvPlanksNoPanicOnError,
		EnvPlanksEnvPrefix,
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}
}

func restoreEnvVars(saved map[string]string) {
	for k, v := range saved {
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}
}
