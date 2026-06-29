package errorreport

import (
	"log/slog"
	"os"
)

// ReportFatal logs a fatal error using slog and exits the application with status code 1.
// It centralizes error reporting for unrecoverable errors.
func ReportFatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
