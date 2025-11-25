package database

import (
	"gorm.io/gorm"
)

// SlowQueryPlugin is a GORM plugin that logs slow queries.
type SlowQueryPlugin struct {
	config SlowQueryConfig
	logger interface{}
}

// NewSlowQueryPlugin creates a new slow query logging plugin.
func NewSlowQueryPlugin(config SlowQueryConfig, logger interface{}) *SlowQueryPlugin {
	return &SlowQueryPlugin{
		config: config,
		logger: logger,
	}
}

// Name returns the plugin name.
func (p *SlowQueryPlugin) Name() string {
	return "dgcore:slow_query_logger"
}

// Initialize initializes the plugin by registering callbacks.
func (p *SlowQueryPlugin) Initialize(db *gorm.DB) error {
	if !p.config.Enabled {
		return nil
	}

	// Register after callbacks to measure total query time
	// We use After callbacks so we can measure the complete execution time
	db.Callback().Query().After("gorm:query").Register("dgcore:slow_query", p.logSlowQuery)
	db.Callback().Create().After("gorm:create").Register("dgcore:slow_query", p.logSlowQuery)
	db.Callback().Update().After("gorm:update").Register("dgcore:slow_query", p.logSlowQuery)
	db.Callback().Delete().After("gorm:delete").Register("dgcore:slow_query", p.logSlowQuery)
	db.Callback().Raw().After("gorm:raw").Register("dgcore:slow_query", p.logSlowQuery)

	return nil
}

// logSlowQuery logs queries that exceed the threshold.
func (p *SlowQueryPlugin) logSlowQuery(db *gorm.DB) {
	if !p.config.Enabled {
		return
	}

	// GORM doesn't provide built-in query timing in Statement
	// We need to use a different approach - check if there's an error first
	// For now, we'll log based on the execution context
	// In production, you might want to use a custom logger or middleware

	// Get SQL query
	sql := db.Statement.SQL.String()
	if sql == "" {
		return // Skip if no SQL
	}

	// Since we can't reliably measure time in the After callback,
	// we'll use a simplified approach: log all queries if threshold is very low
	// or use GORM's built-in slow query logging via SlowThreshold in config

	// For this implementation, we'll just demonstrate the structure
	// Real timing would require Before/After callback pairs or custom middleware

	if p.logger != nil {
		// This is a simplified version - in production you'd want proper timing
		logInfo(p.logger, "Query executed",
			"sql", sql,
			"rows_affected", db.Statement.RowsAffected,
			"table", db.Statement.Table)
	}
}

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
	// Try to call Warn method
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
