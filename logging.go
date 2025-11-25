package database

// logInfo logs an info message using the logger.
func logInfo(logger interface{}, msg string, args ...interface{}) {
	type infoer interface {
		Info(string, ...interface{})
	}

	if i, ok := logger.(infoer); ok {
		i.Info(msg, args...)
	}
}

// logWarn logs a warning message using the logger.
func logWarn(logger interface{}, msg string, args ...interface{}) {
	type warner interface {
		Warn(string, ...interface{})
	}

	if w, ok := logger.(warner); ok {
		w.Warn(msg, args...)
		return
	}

	// Fallback to Info if Warn not available
	logInfo(logger, msg, args...)
}
