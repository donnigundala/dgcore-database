package database

import (
	"gorm.io/gorm"
)

// ReadWritePlugin is a GORM plugin that automatically routes queries to master/slave.
type ReadWritePlugin struct {
	manager *Manager
}

// NewReadWritePlugin creates a new read/write routing plugin.
func NewReadWritePlugin(manager *Manager) *ReadWritePlugin {
	return &ReadWritePlugin{
		manager: manager,
	}
}

// Name returns the plugin name.
func (p *ReadWritePlugin) Name() string {
	return "dgcore:read_write_plugin"
}

// Initialize initializes the plugin.
func (p *ReadWritePlugin) Initialize(db *gorm.DB) error {
	// Register callbacks for automatic routing

	// Query callback - use slave for reads
	db.Callback().Query().Before("gorm:query").Register("dgcore:route_read", p.routeRead)

	// Create callback - use master for writes
	db.Callback().Create().Before("gorm:create").Register("dgcore:route_write", p.routeWrite)

	// Update callback - use master for writes
	db.Callback().Update().Before("gorm:update").Register("dgcore:route_write", p.routeWrite)

	// Delete callback - use master for writes
	db.Callback().Delete().Before("gorm:delete").Register("dgcore:route_write", p.routeWrite)

	// Raw callback - use master for raw SQL
	db.Callback().Raw().Before("gorm:raw").Register("dgcore:route_write", p.routeWrite)

	return nil
}

// routeRead routes read queries to slave connections.
func (p *ReadWritePlugin) routeRead(db *gorm.DB) {
	// Skip if already in a transaction (transactions always use master)
	if db.Statement.DB.Statement.ConnPool != nil {
		return
	}

	// Check if this is explicitly a write operation
	if isWriteOperation(db) {
		return
	}

	// Use slave for reads
	if len(p.manager.slaves) > 0 {
		slave := p.manager.selectSlave()
		db.Statement.ConnPool = slave.Statement.ConnPool
	}
}

// routeWrite routes write queries to master connection.
func (p *ReadWritePlugin) routeWrite(db *gorm.DB) {
	// Writes always go to master
	// This is a no-op since master is the default connection
	// But we keep it for clarity and future enhancements
}

// isWriteOperation checks if the operation is a write.
func isWriteOperation(db *gorm.DB) bool {
	// Check for write operations
	sql := db.Statement.SQL.String()
	if sql == "" {
		return false
	}

	// Simple check for write operations
	// In production, you might want a more sophisticated check
	writeKeywords := []string{"INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP"}
	for _, keyword := range writeKeywords {
		if len(sql) >= len(keyword) && sql[:len(keyword)] == keyword {
			return true
		}
	}

	return false
}
