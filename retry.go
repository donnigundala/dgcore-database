package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// connectWithRetry attempts to connect to the database with retry logic.
func connectWithRetry(config ConnectionConfig, retryConfig RetryConfig, logger interface{}) (*gorm.DB, error) {
	if !retryConfig.Enabled {
		// Retry disabled, use normal connection
		return connectWithConfig(config, logger)
	}

	var lastErr error
	delay := retryConfig.InitialDelay

	for attempt := 1; attempt <= retryConfig.MaxAttempts; attempt++ {
		// Attempt connection
		db, err := connectWithConfig(config, logger)
		if err == nil {
			// Success!
			if attempt > 1 && logger != nil {
				logInfo(logger, "Database connection successful after retry",
					"attempt", attempt,
					"total_attempts", retryConfig.MaxAttempts)
			}
			return db, nil
		}

		// Connection failed
		lastErr = err

		// Don't sleep after the last attempt
		if attempt < retryConfig.MaxAttempts {
			if logger != nil {
				logWarn(logger, "Database connection failed, retrying",
					"attempt", attempt,
					"max_attempts", retryConfig.MaxAttempts,
					"delay", delay,
					"error", err.Error())
			}

			// Sleep before next retry
			time.Sleep(delay)

			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * retryConfig.BackoffFactor)
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}
		}
	}

	// All attempts failed
	return nil, fmt.Errorf("failed to connect after %d attempts: %w",
		retryConfig.MaxAttempts, lastErr)
}
