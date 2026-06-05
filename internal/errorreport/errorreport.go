package errorreport

import (
	"log/slog"
	"os"
)

// ReportFatal logs the error with its context and exits the application.
// This centralizes fatal error reporting to ensure a single, interceptable exit point.
func ReportFatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
