package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManager_DetailedHealthCheck tests the DetailedHealthCheck method
func TestManager_DetailedHealthCheck(t *testing.T) {
	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// Get detailed health check
	health := manager.DetailedHealthCheck()

	// Should have primary connection
	assert.Contains(t, health, "primary", "Should have primary connection health")

	// Check primary health
	primaryHealth := health["primary"]
	assert.Equal(t, HealthStatusHealthy, primaryHealth.Status, "Primary should be healthy")
	assert.Nil(t, primaryHealth.Error, "Should have no error")
	assert.Greater(t, primaryHealth.Latency, time.Duration(0), "Should have measured latency")
	assert.NotZero(t, primaryHealth.LastChecked, "Should have LastChecked timestamp")

	// Should have pool stats
	assert.GreaterOrEqual(t, primaryHealth.Stats.OpenConnections, 0, "Should have stats")
}

// TestManager_DetailedHealthCheck_WithReadWriteSplitting tests health check with master/slaves
func TestManager_DetailedHealthCheck_WithReadWriteSplitting(t *testing.T) {
	config := Config{
		Driver:             "sqlite",
		Database:           ":memory:",
		ReadWriteSplitting: true,
		Master: ConnectionConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		Slaves: []ConnectionConfig{
			{
				Driver:   "sqlite",
				Database: ":memory:",
			},
			{
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// Get detailed health check
	health := manager.DetailedHealthCheck()

	// Should have primary, master, and slaves
	assert.Contains(t, health, "primary")
	assert.Contains(t, health, "master")
	assert.Contains(t, health, "slave_0")
	assert.Contains(t, health, "slave_1")

	// All should be healthy
	for name, h := range health {
		assert.Equal(t, HealthStatusHealthy, h.Status, "%s should be healthy", name)
		assert.Nil(t, h.Error, "%s should have no error", name)
	}
}

// TestManager_DetailedHealthCheck_WithNamedConnections tests health check with named connections
func TestManager_DetailedHealthCheck_WithNamedConnections(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		Database: ":memory:",
		Connections: map[string]ConnectionConfig{
			"analytics": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
			"logs": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// Get detailed health check
	health := manager.DetailedHealthCheck()

	// Should have primary and named connections
	assert.Contains(t, health, "primary")
	assert.Contains(t, health, "analytics")
	assert.Contains(t, health, "logs")

	// All should be healthy
	assert.Equal(t, HealthStatusHealthy, health["primary"].Status)
	assert.Equal(t, HealthStatusHealthy, health["analytics"].Status)
	assert.Equal(t, HealthStatusHealthy, health["logs"].Status)
}

// TestManager_DetailedHealthCheck_UnhealthyConnection tests unhealthy connection detection
func TestManager_DetailedHealthCheck_UnhealthyConnection(t *testing.T) {
	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	manager, err := NewManager(config, nil)
	require.NoError(t, err)

	// Close the connection to make it unhealthy
	manager.Close()

	// Get detailed health check
	health := manager.DetailedHealthCheck()

	// Primary should be unhealthy
	primaryHealth := health["primary"]
	assert.Equal(t, HealthStatusUnhealthy, primaryHealth.Status, "Closed connection should be unhealthy")
	assert.NotNil(t, primaryHealth.Error, "Should have error")
}

// TestManager_IsHealthy tests the IsHealthy method
func TestManager_IsHealthy(t *testing.T) {
	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// Should be healthy
	assert.True(t, manager.IsHealthy(), "Manager should be healthy")

	// Close and check again
	manager.Close()
	assert.False(t, manager.IsHealthy(), "Closed manager should not be healthy")
}

// TestManager_IsFullyHealthy tests the IsFullyHealthy method
func TestManager_IsFullyHealthy(t *testing.T) {
	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// Should be fully healthy (SQLite is fast)
	assert.True(t, manager.IsFullyHealthy(), "Manager should be fully healthy")
}

// TestHealthStatus_Constants tests that health status constants are defined
func TestHealthStatus_Constants(t *testing.T) {
	assert.Equal(t, HealthStatus("healthy"), HealthStatusHealthy)
	assert.Equal(t, HealthStatus("degraded"), HealthStatusDegraded)
	assert.Equal(t, HealthStatus("unhealthy"), HealthStatusUnhealthy)
}

// TestConnectionHealth_Structure tests the ConnectionHealth structure
func TestConnectionHealth_Structure(t *testing.T) {
	health := ConnectionHealth{
		Status:      HealthStatusHealthy,
		Latency:     10 * time.Millisecond,
		Error:       nil,
		Stats:       PoolStats{OpenConnections: 1},
		LastChecked: time.Now(),
	}

	assert.Equal(t, HealthStatusHealthy, health.Status)
	assert.Equal(t, 10*time.Millisecond, health.Latency)
	assert.Nil(t, health.Error)
	assert.Equal(t, 1, health.Stats.OpenConnections)
	assert.NotZero(t, health.LastChecked)
}
