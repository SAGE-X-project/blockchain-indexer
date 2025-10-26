package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
	config *Config
}

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, console
	Output     string // stdout, stderr, file
	FilePath   string // log file path (when output=file)
	MaxSize    int    // megabytes
	MaxBackups int    // number of backups
	MaxAge     int    // days
	Compress   bool   // compress old files
}

// New creates a new logger with the given configuration
func New(cfg *Config) (*Logger, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Parse level
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create encoder
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create write syncer
	writeSyncer := getWriteSyncer(cfg)

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		Logger: zapLogger,
		config: cfg,
	}, nil
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}
}

// parseLevel parses log level string
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown level: %s", level)
	}
}

// getWriteSyncer creates write syncer based on output configuration
func getWriteSyncer(cfg *Config) zapcore.WriteSyncer {
	switch cfg.Output {
	case "stdout":
		return zapcore.AddSync(os.Stdout)
	case "stderr":
		return zapcore.AddSync(os.Stderr)
	case "file":
		if cfg.FilePath == "" {
			// Fallback to stdout if no file path specified
			return zapcore.AddSync(os.Stdout)
		}

		// Use lumberjack for log rotation
		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}

		return zapcore.AddSync(lumberJackLogger)
	default:
		return zapcore.AddSync(os.Stdout)
	}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		config: l.config,
	}
}

// WithComponent returns a logger with a component field
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithFields(zap.String("component", component))
}

// WithChain returns a logger with a chain field
func (l *Logger) WithChain(chainID string) *Logger {
	return l.WithFields(zap.String("chain_id", chainID))
}

// WithError returns a logger with an error field
func (l *Logger) WithError(err error) *Logger {
	return l.WithFields(zap.Error(err))
}

// Named returns a new logger with the specified name
func (l *Logger) Named(name string) *Logger {
	return &Logger{
		Logger: l.Logger.Named(name),
		config: l.config,
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Global logger instance
var (
	globalLogger *Logger
)

// init initializes the global logger with default configuration
func init() {
	logger, _ := New(DefaultConfig())
	globalLogger = logger
}

// SetGlobal sets the global logger
func SetGlobal(logger *Logger) {
	globalLogger = logger
}

// Global returns the global logger
func Global() *Logger {
	return globalLogger
}

// Debug logs a debug message using the global logger
func Debug(msg string, fields ...zap.Field) {
	globalLogger.Debug(msg, fields...)
}

// Info logs an info message using the global logger
func Info(msg string, fields ...zap.Field) {
	globalLogger.Info(msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields ...zap.Field) {
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message using the global logger
func Error(msg string, fields ...zap.Field) {
	globalLogger.Error(msg, fields...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(msg string, fields ...zap.Field) {
	globalLogger.Fatal(msg, fields...)
}

// With returns a logger with additional fields
func With(fields ...zap.Field) *Logger {
	return globalLogger.WithFields(fields...)
}

// WithComponent returns a logger with a component field
func WithComponent(component string) *Logger {
	return globalLogger.WithComponent(component)
}

// WithChain returns a logger with a chain field
func WithChain(chainID string) *Logger {
	return globalLogger.WithChain(chainID)
}

// WithError returns a logger with an error field
func WithError(err error) *Logger {
	return globalLogger.WithError(err)
}

// Sync flushes any buffered log entries from the global logger
func Sync() error {
	return globalLogger.Sync()
}
