package logger

import (
	"log/slog"
	"os"
	"path/filepath"
)

// InitializeLogger sets up a JSON logger for the given usecase name.
func InitializeLogger(usecaseName string) (*slog.Logger, error) {
	// Ensure the logs directory exists
	logDir := filepath.Join("logs", usecaseName)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// Open or create the log file
	logFilePath := filepath.Join(logDir, usecaseName+".log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Create a JSON handler for structured logging
	handler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return slog.New(handler), nil
}
