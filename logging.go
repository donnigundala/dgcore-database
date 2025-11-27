package database

// logInfo logs an info message using the logger.
func logInfo(logger Logger, msg string, args ...interface{}) {
	if logger != nil {
		logger.Info(msg, args...)
	}
}

// logWarn logs a warning message using the logger.
func logWarn(logger Logger, msg string, args ...interface{}) {
	if logger != nil {
		logger.Warn(msg, args...)
	}
}
