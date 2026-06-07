package errorreport

import (
	"log/slog"
	"os"
)

// ReportFatal logs the given message and error using slog, then exits the application.
// This function centralizes fatal error reporting to ensure consistency across the application.
func ReportFatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
