package logger

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	t.Run("create logger with default config", func(t *testing.T) {
		logger, err := New(nil)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		if logger == nil {
			t.Fatal("New() returned nil logger")
		}

		if logger.config == nil {
			t.Error("logger.config should not be nil")
		}
	})

	t.Run("create logger with custom config", func(t *testing.T) {
		cfg := &Config{
			Level:  "debug",
			Format: "console",
			Output: "stdout",
		}

		logger, err := New(cfg)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		if logger.config.Level != "debug" {
			t.Errorf("logger.config.Level = %v, want debug", logger.config.Level)
		}
	})

	t.Run("create logger with invalid level", func(t *testing.T) {
		cfg := &Config{
			Level:  "invalid",
			Format: "json",
			Output: "stdout",
		}

		_, err := New(cfg)
		if err == nil {
			t.Error("New() should return error for invalid level")
		}
	})

	t.Run("create logger with file output", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")

		cfg := &Config{
			Level:    "info",
			Format:   "json",
			Output:   "file",
			FilePath: logFile,
			MaxSize:  10,
		}

		logger, err := New(cfg)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		// Write a log
		logger.Info("test message")
		logger.Sync()

		// Verify file was created
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Error("log file was not created")
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Level != "info" {
		t.Errorf("Level = %v, want info", cfg.Level)
	}

	if cfg.Format != "json" {
		t.Errorf("Format = %v, want json", cfg.Format)
	}

	if cfg.Output != "stdout" {
		t.Errorf("Output = %v, want stdout", cfg.Output)
	}
}

func TestLogger_WithFields(t *testing.T) {
	logger, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Add fields
	loggerWithFields := logger.WithFields(
		zap.String("key1", "value1"),
		zap.Int("key2", 42),
	)

	if loggerWithFields == nil {
		t.Error("WithFields() returned nil")
	}

	// Original logger should be unchanged
	if logger.Logger == loggerWithFields.Logger {
		t.Error("WithFields() should return a new logger instance")
	}
}

func TestLogger_WithComponent(t *testing.T) {
	logger, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	componentLogger := logger.WithComponent("test-component")

	if componentLogger == nil {
		t.Error("WithComponent() returned nil")
	}
}

func TestLogger_WithChain(t *testing.T) {
	logger, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	chainLogger := logger.WithChain("ethereum")

	if chainLogger == nil {
		t.Error("WithChain() returned nil")
	}
}

func TestLogger_WithError(t *testing.T) {
	logger, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	testErr := errors.New("test error")
	errorLogger := logger.WithError(testErr)

	if errorLogger == nil {
		t.Error("WithError() returned nil")
	}
}

func TestLogger_Named(t *testing.T) {
	logger, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	namedLogger := logger.Named("test-logger")

	if namedLogger == nil {
		t.Error("Named() returned nil")
	}
}

func TestLogger_Sync(t *testing.T) {
	logger, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = logger.Sync()
	// Sync may return error on some platforms (e.g., /dev/stdout on Linux)
	// so we just check that it doesn't panic
	if err != nil {
		t.Logf("Sync() returned error (this may be expected): %v", err)
	}
}

func TestGlobalLogger(t *testing.T) {
	t.Run("get global logger", func(t *testing.T) {
		logger := Global()
		if logger == nil {
			t.Fatal("Global() returned nil")
		}
	})

	t.Run("set global logger", func(t *testing.T) {
		newLogger, err := New(&Config{
			Level:  "debug",
			Format: "console",
			Output: "stdout",
		})
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		SetGlobal(newLogger)

		if Global() != newLogger {
			t.Error("SetGlobal() did not update global logger")
		}
	})
}

func TestGlobalLogFunctions(t *testing.T) {
	// These tests just verify the functions don't panic
	// We can't easily verify the output without capturing stdout

	t.Run("debug", func(t *testing.T) {
		Debug("debug message", zap.String("key", "value"))
	})

	t.Run("info", func(t *testing.T) {
		Info("info message", zap.String("key", "value"))
	})

	t.Run("warn", func(t *testing.T) {
		Warn("warn message", zap.String("key", "value"))
	})

	t.Run("error", func(t *testing.T) {
		Error("error message", zap.String("key", "value"))
	})

	t.Run("with component", func(t *testing.T) {
		logger := WithComponent("test")
		logger.Info("test message")
	})

	t.Run("with chain", func(t *testing.T) {
		logger := WithChain("ethereum")
		logger.Info("test message")
	})

	t.Run("with error", func(t *testing.T) {
		logger := WithError(errors.New("test error"))
		logger.Info("test message")
	})

	t.Run("with fields", func(t *testing.T) {
		logger := With(zap.String("key", "value"))
		logger.Info("test message")
	})
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{"debug", "debug", false},
		{"info", "info", false},
		{"warn", "warn", false},
		{"warning", "warning", false},
		{"error", "error", false},
		{"fatal", "fatal", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetWriteSyncer(t *testing.T) {
	t.Run("stdout syncer", func(t *testing.T) {
		cfg := &Config{Output: "stdout"}
		syncer := getWriteSyncer(cfg)
		if syncer == nil {
			t.Error("getWriteSyncer() returned nil")
		}
	})

	t.Run("stderr syncer", func(t *testing.T) {
		cfg := &Config{Output: "stderr"}
		syncer := getWriteSyncer(cfg)
		if syncer == nil {
			t.Error("getWriteSyncer() returned nil")
		}
	})

	t.Run("file syncer", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &Config{
			Output:   "file",
			FilePath: filepath.Join(tmpDir, "test.log"),
		}
		syncer := getWriteSyncer(cfg)
		if syncer == nil {
			t.Error("getWriteSyncer() returned nil")
		}
	})

	t.Run("file syncer without path falls back to stdout", func(t *testing.T) {
		cfg := &Config{Output: "file"}
		syncer := getWriteSyncer(cfg)
		if syncer == nil {
			t.Error("getWriteSyncer() returned nil")
		}
	})

	t.Run("unknown output defaults to stdout", func(t *testing.T) {
		cfg := &Config{Output: "unknown"}
		syncer := getWriteSyncer(cfg)
		if syncer == nil {
			t.Error("getWriteSyncer() returned nil")
		}
	})
}

func BenchmarkLogger_Info(b *testing.B) {
	logger, _ := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test message", zap.Int("iteration", i))
	}
}

func BenchmarkLogger_WithFields(b *testing.B) {
	logger, _ := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(
			zap.String("key1", "value1"),
			zap.Int("key2", i),
		).Info("test message")
	}
}
