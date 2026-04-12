// Package errorreport provides a centralized error-reporting mechanism
// to funnel unexpected and unrecoverable errors.
package errorreport

import (
	"log/slog"
	"os"
)

// ReportFatal logs an unrecoverable error using structured logging
// and then exits the application with status code 1.
func ReportFatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}

// ReportError logs an unexpected error using structured logging.
// This is meant for errors that should be reported but don't require halting execution.
func ReportError(msg string, err error) {
	slog.Error(msg, "err", err)
}
