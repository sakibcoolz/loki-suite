package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global logger instance
var Logger *zap.Logger

// Initialize initializes the logger with the given configuration
func Initialize(level, encoding, outputPath string) error {
	config := zap.NewProductionConfig()

	// Set log level
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Set encoding
	if encoding == "console" {
		config.Encoding = "console"
		config.EncoderConfig = zap.NewDevelopmentEncoderConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config.Encoding = "json"
		config.EncoderConfig = zap.NewProductionEncoderConfig()
	}

	// Set output path
	if outputPath != "stdout" {
		config.OutputPaths = []string{outputPath}
	}

	// Add timestamp to logs
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Build logger
	var err error
	Logger, err = config.Build()
	if err != nil {
		return err
	}

	// Replace global logger
	zap.ReplaceGlobals(Logger)

	return nil
}

// Sync syncs the logger
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
	os.Exit(1)
}

// With creates a child logger with additional fields
func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}

// Field helpers
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

func Duration(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}
