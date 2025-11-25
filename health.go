package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// HealthStatus represents the health status of a database connection.
type HealthStatus string

const (
	// HealthStatusHealthy indicates the connection is healthy and responsive.
	HealthStatusHealthy HealthStatus = "healthy"
	// HealthStatusDegraded indicates the connection is working but slow.
	HealthStatusDegraded HealthStatus = "degraded"
	// HealthStatusUnhealthy indicates the connection is not working.
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ConnectionHealth represents detailed health information for a database connection.
type ConnectionHealth struct {
	Status      HealthStatus  // Health status of the connection
	Latency     time.Duration // Ping latency
	Error       error         // Error if unhealthy
	Stats       PoolStats     // Connection pool statistics
	LastChecked time.Time     // When the health check was performed
}

// DetailedHealthCheck returns detailed health status for all database connections.
// It checks connectivity, measures latency, and includes pool statistics.
func (m *Manager) DetailedHealthCheck() map[string]ConnectionHealth {
	result := make(map[string]ConnectionHealth)

	// Check primary connection
	result["primary"] = m.checkConnectionHealth(m.db, "primary")

	// Check master/slave connections if read/write splitting is enabled
	if m.config.ReadWriteSplitting {
		if m.master != nil {
			result["master"] = m.checkConnectionHealth(m.master, "master")
		}

		for i, slave := range m.slaves {
			name := fmt.Sprintf("slave_%d", i)
			result[name] = m.checkConnectionHealth(slave, name)
		}
	}

	// Check named connections
	m.connMu.RLock()
	for name, conn := range m.connections {
		result[name] = m.checkConnectionHealth(conn, name)
	}
	m.connMu.RUnlock()

	return result
}

// checkConnectionHealth performs a health check on a single connection.
func (m *Manager) checkConnectionHealth(db *gorm.DB, name string) ConnectionHealth {
	start := time.Now()

	// Get underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return ConnectionHealth{
			Status:      HealthStatusUnhealthy,
			Error:       fmt.Errorf("failed to get sql.DB: %w", err),
			LastChecked: time.Now(),
		}
	}

	// Ping to check connectivity and measure latency
	err = sqlDB.Ping()
	latency := time.Since(start)

	// Get pool statistics
	stats := sqlDB.Stats()
	poolStats := PoolStats{
		OpenConnections:   stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}

	// Determine health status based on ping result and latency
	status := HealthStatusHealthy
	if err != nil {
		status = HealthStatusUnhealthy
	} else if latency > 100*time.Millisecond {
		// Consider connection degraded if latency > 100ms
		status = HealthStatusDegraded
	}

	return ConnectionHealth{
		Status:      status,
		Latency:     latency,
		Error:       err,
		Stats:       poolStats,
		LastChecked: time.Now(),
	}
}

// IsHealthy returns true if all connections are healthy (not unhealthy).
// Degraded connections are still considered acceptable.
func (m *Manager) IsHealthy() bool {
	health := m.DetailedHealthCheck()
	for _, h := range health {
		if h.Status == HealthStatusUnhealthy {
			return false
		}
	}
	return true
}

// IsFullyHealthy returns true only if all connections are healthy (not degraded or unhealthy).
func (m *Manager) IsFullyHealthy() bool {
	health := m.DetailedHealthCheck()
	for _, h := range health {
		if h.Status != HealthStatusHealthy {
			return false
		}
	}
	return true
}
