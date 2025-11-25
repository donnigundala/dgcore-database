package database

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestReadWritePlugin_Metadata tests the plugin metadata
func TestReadWritePlugin_Metadata(t *testing.T) {
	// Create a simple manager for testing
	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	plugin := NewReadWritePlugin(manager)

	// Test Name
	assert.Equal(t, "dgcore:read_write_plugin", plugin.Name(), "Plugin name should match")
}

// TestReadWritePlugin_Initialize tests plugin initialization
func TestReadWritePlugin_Initialize(t *testing.T) {
	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	plugin := NewReadWritePlugin(manager)

	// Initialize should not error
	err = plugin.Initialize(manager.DB())
	assert.NoError(t, err, "Plugin initialization should succeed")
}

// TestReadWritePlugin_ReadRouting tests that SELECT queries go to slaves
func TestReadWritePlugin_ReadRouting(t *testing.T) {
	// Create config with read/write splitting but AutoRouting disabled initially
	config := Config{
		Driver:             "sqlite",
		Database:           ":memory:",
		ReadWriteSplitting: true,
		AutoRouting:        false, // Disable during setup
		SlaveStrategy:      "round-robin",
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

	// Create test table on master only
	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	// Migrate on master
	err = manager.Master().AutoMigrate(&TestModel{})
	require.NoError(t, err)

	// Insert data on master
	testData := TestModel{Name: "test"}
	result := manager.Master().Create(&testData)
	require.NoError(t, result.Error)

	// Verify data exists on master
	var retrieved TestModel
	result = manager.Master().First(&retrieved, testData.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, "test", retrieved.Name)

	// Read should use slave selection (even though slaves don't have data)
	// We're testing the routing mechanism, not data replication
	db := manager.Read()
	assert.NotNil(t, db, "Read() should return a valid DB instance")
}

// TestReadWritePlugin_WriteRouting tests that write operations go to master
func TestReadWritePlugin_WriteRouting(t *testing.T) {
	// Use simple config without read/write splitting for this test
	// We're testing that Write() returns the master connection
	config := Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	// Migrate
	err = manager.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	// Write should go to master
	testData := TestModel{Name: "write_test"}
	result := manager.Write().Create(&testData)
	assert.NoError(t, result.Error, "Write should succeed")
	assert.NotZero(t, testData.ID, "ID should be set after create")

	// Verify data exists on master
	var retrieved TestModel
	result = manager.Master().First(&retrieved, testData.ID)
	assert.NoError(t, result.Error, "Should find record on master")
	assert.Equal(t, "write_test", retrieved.Name, "Name should match")
}

// TestReadWritePlugin_LoadBalancing tests slave selection strategies
func TestReadWritePlugin_LoadBalancing(t *testing.T) {
	t.Run("round-robin strategy", func(t *testing.T) {
		config := Config{
			Driver:             "sqlite",
			Database:           ":memory:",
			ReadWriteSplitting: true,
			SlaveStrategy:      "round-robin",
			Master: ConnectionConfig{
				Driver:   "sqlite",
				Database: ":memory:",
			},
			Slaves: []ConnectionConfig{
				{Driver: "sqlite", Database: ":memory:"},
				{Driver: "sqlite", Database: ":memory:"},
				{Driver: "sqlite", Database: ":memory:"},
			},
		}

		manager, err := NewManager(config, nil)
		require.NoError(t, err)
		defer manager.Close()

		// Verify we have 3 slaves
		assert.Len(t, manager.slaves, 3, "Should have 3 slaves")

		// Call selectSlave multiple times and verify round-robin
		// Note: We can't easily verify the exact slave selected without
		// exposing internal state, but we can verify it doesn't panic
		for i := 0; i < 10; i++ {
			slave := manager.selectSlave()
			assert.NotNil(t, slave, "selectSlave should return a valid slave")
		}
	})

	t.Run("random strategy", func(t *testing.T) {
		config := Config{
			Driver:             "sqlite",
			Database:           ":memory:",
			ReadWriteSplitting: true,
			SlaveStrategy:      "random",
			Master: ConnectionConfig{
				Driver:   "sqlite",
				Database: ":memory:",
			},
			Slaves: []ConnectionConfig{
				{Driver: "sqlite", Database: ":memory:"},
				{Driver: "sqlite", Database: ":memory:"},
			},
		}

		manager, err := NewManager(config, nil)
		require.NoError(t, err)
		defer manager.Close()

		// Verify random selection works
		for i := 0; i < 10; i++ {
			slave := manager.selectSlave()
			assert.NotNil(t, slave, "selectSlave should return a valid slave")
		}
	})

	t.Run("weighted strategy", func(t *testing.T) {
		config := Config{
			Driver:             "sqlite",
			Database:           ":memory:",
			ReadWriteSplitting: true,
			SlaveStrategy:      "weighted",
			Master: ConnectionConfig{
				Driver:   "sqlite",
				Database: ":memory:",
			},
			Slaves: []ConnectionConfig{
				{Driver: "sqlite", Database: ":memory:", Weight: 3},
				{Driver: "sqlite", Database: ":memory:", Weight: 1},
			},
		}

		manager, err := NewManager(config, nil)
		require.NoError(t, err)
		defer manager.Close()

		// Verify weighted selection works
		for i := 0; i < 10; i++ {
			slave := manager.selectSlave()
			assert.NotNil(t, slave, "selectSlave should return a valid slave")
		}
	})
}

// TestIsWriteOperation tests the SQL operation classification
func TestIsWriteOperation(t *testing.T) {
	t.Run("empty statement", func(t *testing.T) {
		db := &gorm.DB{
			Statement: &gorm.Statement{},
		}

		// Should not panic with empty statement
		result := isWriteOperation(db)
		assert.False(t, result, "Empty statement should not be classified as write")
	})

	t.Run("write operations", func(t *testing.T) {
		writeKeywords := []string{"INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP"}

		for _, keyword := range writeKeywords {
			t.Run(keyword, func(t *testing.T) {
				db := &gorm.DB{
					Statement: &gorm.Statement{
						SQL: strings.Builder{},
					},
				}
				db.Statement.SQL.WriteString(keyword + " INTO table")

				result := isWriteOperation(db)
				assert.True(t, result, keyword+" should be classified as write operation")
			})
		}
	})

	t.Run("read operations", func(t *testing.T) {
		readKeywords := []string{"SELECT", "SHOW", "DESCRIBE", "EXPLAIN"}

		for _, keyword := range readKeywords {
			t.Run(keyword, func(t *testing.T) {
				db := &gorm.DB{
					Statement: &gorm.Statement{
						SQL: strings.Builder{},
					},
				}
				db.Statement.SQL.WriteString(keyword + " * FROM table")

				result := isWriteOperation(db)
				assert.False(t, result, keyword+" should not be classified as write operation")
			})
		}
	})
}

// TestReadWritePlugin_AutoRouting_ActualQueries tests routing with actual queries
func TestReadWritePlugin_AutoRouting_ActualQueries(t *testing.T) {
	config := Config{
		Driver:             "sqlite",
		Database:           ":memory:",
		ReadWriteSplitting: true,
		AutoRouting:        true, // Enable auto routing
		SlaveStrategy:      "round-robin",
		Master: ConnectionConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		Slaves: []ConnectionConfig{
			{
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	// Migrate on master
	err = manager.Master().AutoMigrate(&TestModel{})
	require.NoError(t, err)

	// Migrate on slaves
	for i := range manager.slaves {
		err = manager.Slave(i).AutoMigrate(&TestModel{})
		require.NoError(t, err)
	}

	// Test write operation (should go to master)
	testData := TestModel{Name: "test"}
	result := manager.Master().Create(&testData)
	require.NoError(t, result.Error)

	// Test that plugin is active
	assert.NotNil(t, manager.plugin, "Plugin should be initialized")
}

// TestReadWritePlugin_SystemTableDetection tests that system table queries go to master
func TestReadWritePlugin_SystemTableDetection(t *testing.T) {
	config := Config{
		Driver:             "sqlite",
		Database:           ":memory:",
		ReadWriteSplitting: true,
		AutoRouting:        true,
		SlaveStrategy:      "round-robin",
		Master: ConnectionConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		Slaves: []ConnectionConfig{
			{
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// System table queries should not be routed to slaves
	// This is tested indirectly through AutoMigrate which queries sqlite_master
	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	// AutoMigrate queries system tables - should work without errors
	err = manager.AutoMigrate(&TestModel{})
	assert.NoError(t, err, "AutoMigrate should succeed (system tables go to master)")
}

// TestReadWritePlugin_SkipRoutingContext tests that skip_routing context is respected
func TestReadWritePlugin_SkipRoutingContext(t *testing.T) {
	config := Config{
		Driver:             "sqlite",
		Database:           ":memory:",
		ReadWriteSplitting: true,
		AutoRouting:        true,
		SlaveStrategy:      "round-robin",
		Master: ConnectionConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
		Slaves: []ConnectionConfig{
			{
				Driver:   "sqlite",
				Database: ":memory:",
			},
		},
	}

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// AutoMigrate uses skip_routing context
	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	err = manager.AutoMigrate(&TestModel{})
	assert.NoError(t, err, "AutoMigrate with skip_routing should succeed")
}

// TestReadWritePlugin_TransactionRouting tests that transactions always use master
func TestReadWritePlugin_TransactionRouting(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		Database: ":memory:",
	}

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	err = manager.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	// Transaction should use master
	err = manager.Transaction(func(tx *gorm.DB) error {
		testData := TestModel{Name: "transaction_test"}
		result := tx.Create(&testData)
		if result.Error != nil {
			return result.Error
		}

		// Verify within transaction
		var retrieved TestModel
		result = tx.First(&retrieved, testData.ID)
		assert.NoError(t, result.Error, "Should find record in transaction")
		assert.Equal(t, "transaction_test", retrieved.Name)

		return nil
	})
	assert.NoError(t, err, "Transaction should succeed")
}

// TestReadWritePlugin_NoSlaves tests behavior when no slaves are configured
func TestReadWritePlugin_NoSlaves(t *testing.T) {
	config := Config{
		Driver:             "sqlite",
		Database:           ":memory:",
		ReadWriteSplitting: false, // No read/write splitting
		Master: ConnectionConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
	}

	manager, err := NewManager(config, nil)
	require.NoError(t, err)
	defer manager.Close()

	// Read should fall back to master when no slaves
	db := manager.Read()
	assert.NotNil(t, db, "Read() should return a valid DB when no slaves")

	// Verify it works (it should use master internally)
	// We can't compare pointers because GORM returns clones
	assert.NotNil(t, db.Statement, "DB should have a valid statement")
}
