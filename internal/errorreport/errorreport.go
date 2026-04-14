// Package errorreport provides centralized error reporting functionality.
package errorreport

import (
	"log/slog"
	"os"
)

// ReportFatal logs a fatal error and exits the program.
// It funnels unexpected, unrecoverable errors to a central location.
func ReportFatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
