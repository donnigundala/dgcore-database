package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManager_Stats tests the Stats() method
func TestManager_Stats(t *testing.T) {
	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// Get stats
	stats := manager.Stats()

	// Verify stats structure
	assert.GreaterOrEqual(t, stats.OpenConnections, 0, "OpenConnections should be non-negative")
	assert.GreaterOrEqual(t, stats.InUse, 0, "InUse should be non-negative")
	assert.GreaterOrEqual(t, stats.Idle, 0, "Idle should be non-negative")
	assert.GreaterOrEqual(t, stats.WaitCount, int64(0), "WaitCount should be non-negative")

	// For SQLite :memory: with MaxOpenConns=1, we should have exactly 1 connection
	assert.Equal(t, 1, stats.OpenConnections, "Should have 1 open connection for :memory:")
}

// TestManager_ConnectionStats tests the ConnectionStats() method
func TestManager_ConnectionStats(t *testing.T) {
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

	// Get stats for named connection
	analyticsStats := manager.ConnectionStats("analytics")
	assert.GreaterOrEqual(t, analyticsStats.OpenConnections, 0, "Should have valid stats")

	logsStats := manager.ConnectionStats("logs")
	assert.GreaterOrEqual(t, logsStats.OpenConnections, 0, "Should have valid stats")

	// Non-existent connection should return empty stats
	nonExistentStats := manager.ConnectionStats("nonexistent")
	assert.Equal(t, 0, nonExistentStats.OpenConnections, "Non-existent connection should return empty stats")
}

// TestManager_AllStats tests the AllStats() method
func TestManager_AllStats(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		Database: ":memory:",
		Connections: map[string]ConnectionConfig{
			"analytics": {
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// Get all stats
	allStats := manager.AllStats()

	// Should have primary and analytics
	assert.Contains(t, allStats, "primary", "Should have primary connection stats")
	assert.Contains(t, allStats, "analytics", "Should have analytics connection stats")

	// Verify stats are valid
	assert.GreaterOrEqual(t, allStats["primary"].OpenConnections, 0)
	assert.GreaterOrEqual(t, allStats["analytics"].OpenConnections, 0)
}

// TestManager_AllStats_WithReadWriteSplitting tests AllStats with read/write splitting
func TestManager_AllStats_WithReadWriteSplitting(t *testing.T) {
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

	// Get all stats
	allStats := manager.AllStats()

	// Should have primary, master, and slaves
	assert.Contains(t, allStats, "primary", "Should have primary stats")
	assert.Contains(t, allStats, "master", "Should have master stats")
	assert.Contains(t, allStats, "slave_0", "Should have slave_0 stats")
	assert.Contains(t, allStats, "slave_1", "Should have slave_1 stats")

	// Verify all stats are valid
	for name, stats := range allStats {
		assert.GreaterOrEqual(t, stats.OpenConnections, 0, "Stats for %s should be valid", name)
	}
}

// TestPoolStats_AfterQuery tests that stats change after queries
func TestPoolStats_AfterQuery(t *testing.T) {
	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// Create a test table
	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	err = manager.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	// Get stats before query
	statsBefore := manager.Stats()

	// Execute a query
	var count int64
	manager.DB().Model(&TestModel{}).Count(&count)

	// Get stats after query
	statsAfter := manager.Stats()

	// Stats should still be valid (exact values may vary)
	assert.GreaterOrEqual(t, statsAfter.OpenConnections, statsBefore.OpenConnections)
}
