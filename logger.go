package database

// Logger defines the interface for database logging.
// Any logger implementation must provide Info and Warn methods to be used with the database manager.
//
// This interface is satisfied by dg-core's logging.Logger and any other logger
// that implements these two methods.
//
// Example:
//
//	import "github.com/donnigundala/dg-core/logging"
//
//	logger := logging.Default()
//	manager, err := database.NewManager(config, logger)
type Logger interface {
	// Info logs an informational message with optional key-value pairs
	Info(msg string, args ...interface{})

	// Warn logs a warning message with optional key-value pairs
	Warn(msg string, args ...interface{})
}
