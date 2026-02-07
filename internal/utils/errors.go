package utils

import (
	"log/slog"
	"os"
)

// ReportError logs an error using slog.
// This is the centralized error reporting function.
func ReportError(err error, msg string, args ...any) {
	allArgs := append([]any{"err", err}, args...)
	slog.Error(msg, allArgs...)
}

// ReportFatal logs an error and exits the application with status 1.
func ReportFatal(err error, msg string, args ...any) {
	ReportError(err, msg, args...)
	os.Exit(1)
}
