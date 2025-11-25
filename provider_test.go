package database

import (
	"testing"

	"github.com/donnigundala/dg-core/foundation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServiceProvider_Metadata tests the provider metadata methods
func TestServiceProvider_Metadata(t *testing.T) {
	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	provider := NewServiceProvider(config)

	// Test Name
	assert.Equal(t, "database", provider.Name(), "Provider name should be 'database'")

	// Test Version
	version := provider.Version()
	assert.NotEmpty(t, version, "Provider version should not be empty")
	assert.Equal(t, "1.0.0", version, "Provider version should be 1.0.0")

	// Test Dependencies
	deps := provider.Dependencies()
	assert.NotNil(t, deps, "Dependencies should not be nil")
	assert.Empty(t, deps, "Database provider should have no dependencies")
}

// TestServiceProvider_Register tests the provider registration
func TestServiceProvider_Register(t *testing.T) {
	// Create application
	app := foundation.New(".")

	// Create provider with SQLite (no external dependencies needed)
	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	provider := NewServiceProvider(config)

	// Register provider
	err := provider.Register(app)
	require.NoError(t, err, "Provider registration should not fail")

	// Test that "db" binding exists
	t.Run("db binding registered", func(t *testing.T) {
		db, err := app.Make("db")
		require.NoError(t, err, "Should be able to resolve 'db' binding")
		assert.NotNil(t, db, "Database manager should not be nil")

		// Verify it's the correct type
		manager, ok := db.(*Manager)
		assert.True(t, ok, "Should be able to cast to *Manager")
		assert.NotNil(t, manager, "Manager should not be nil")

		// Verify we can get GORM instance directly
		gormDB := manager.DB()
		assert.NotNil(t, gormDB, "GORM instance should be accessible via manager.DB()")
	})

	// Test singleton behavior
	t.Run("singleton behavior", func(t *testing.T) {
		db1, _ := app.Make("db")
		db2, _ := app.Make("db")

		// Should return the same instance
		assert.Same(t, db1, db2, "Should return the same database manager instance")
	})
}

// TestServiceProvider_Register_WithLogger tests registration with logger available
func TestServiceProvider_Register_WithLogger(t *testing.T) {
	app := foundation.New(".")

	// Register a mock logger
	app.Instance("logger", &mockLogger{})

	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	provider := NewServiceProvider(config)

	err := provider.Register(app)
	require.NoError(t, err, "Provider registration should not fail with logger")

	// Verify db can be resolved
	db, err := app.Make("db")
	require.NoError(t, err)
	assert.NotNil(t, db)
}

// TestServiceProvider_Boot tests the provider boot phase
func TestServiceProvider_Boot(t *testing.T) {
	app := foundation.New(".")

	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	provider := NewServiceProvider(config)

	// Register first
	err := provider.Register(app)
	require.NoError(t, err)

	// Boot should succeed
	err = provider.Boot(app)
	assert.NoError(t, err, "Boot should succeed with valid SQLite config")

	// Verify connection is working
	dbInstance, _ := app.Make("db")
	manager := dbInstance.(*Manager)

	err = manager.Ping()
	assert.NoError(t, err, "Database should be pingable after boot")
}

// TestServiceProvider_Boot_ConnectionFailure tests boot with connection failure
func TestServiceProvider_Boot_ConnectionFailure(t *testing.T) {
	app := foundation.New(".")

	// Use invalid configuration
	config := DefaultConfig().
		WithDriver("mysql"). // MySQL without server running
		WithHost("invalid-host-that-does-not-exist").
		WithPort(3306).
		WithDatabase("test")

	provider := NewServiceProvider(config)

	// Register
	err := provider.Register(app)
	require.NoError(t, err)

	// Boot should panic when trying to resolve db with invalid connection
	assert.Panics(t, func() {
		_ = provider.Boot(app)
	}, "Boot should panic with invalid connection")
}

// TestServiceProvider_Boot_WithLogger tests boot with logger
func TestServiceProvider_Boot_WithLogger(t *testing.T) {
	app := foundation.New(".")

	// Register mock logger
	logger := &mockLogger{}
	app.Instance("logger", logger)

	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	provider := NewServiceProvider(config)

	err := provider.Register(app)
	require.NoError(t, err)

	err = provider.Boot(app)
	assert.NoError(t, err)

	// Logger should have been used (we can't easily verify this without
	// modifying the Manager, but we can at least verify boot succeeded)
}

// TestServiceProvider_Boot_WithReadWriteSplitting tests boot with read/write splitting
func TestServiceProvider_Boot_WithReadWriteSplitting(t *testing.T) {
	app := foundation.New(".")

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

	provider := NewServiceProvider(config)

	err := provider.Register(app)
	require.NoError(t, err)

	err = provider.Boot(app)
	assert.NoError(t, err, "Boot should succeed with read/write splitting")
}

// TestServiceProvider_Boot_WithMultipleConnections tests boot with named connections
func TestServiceProvider_Boot_WithMultipleConnections(t *testing.T) {
	app := foundation.New(".")

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

	provider := NewServiceProvider(config)

	err := provider.Register(app)
	require.NoError(t, err)

	err = provider.Boot(app)
	assert.NoError(t, err, "Boot should succeed with multiple connections")

	// Verify named connections are accessible
	dbInstance, _ := app.Make("db")
	manager := dbInstance.(*Manager)

	assert.True(t, manager.HasConnection("analytics"), "Should have analytics connection")
	assert.True(t, manager.HasConnection("logs"), "Should have logs connection")
}

// TestServiceProvider_IntegrationWithDgCore tests full integration
func TestServiceProvider_IntegrationWithDgCore(t *testing.T) {
	// Create application
	app := foundation.New(".")

	// Create and register provider
	config := DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:")

	provider := NewServiceProvider(config)

	// Register
	err := provider.Register(app)
	require.NoError(t, err, "Registration should succeed")

	// Boot
	err = provider.Boot(app)
	require.NoError(t, err, "Boot should succeed")

	// Use the database
	dbInstance, err := app.Make("db")
	require.NoError(t, err)

	manager := dbInstance.(*Manager)

	// Perform a simple operation
	type TestModel struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	// Auto-migrate
	err = manager.DB().AutoMigrate(&TestModel{})
	require.NoError(t, err, "Auto-migrate should succeed")

	// Create a record
	testRecord := TestModel{Name: "test"}
	result := manager.DB().Create(&testRecord)
	assert.NoError(t, result.Error, "Create should succeed")
	assert.NotZero(t, testRecord.ID, "ID should be set")

	// Read the record
	var retrieved TestModel
	result = manager.DB().First(&retrieved, testRecord.ID)
	assert.NoError(t, result.Error, "Read should succeed")
	assert.Equal(t, "test", retrieved.Name, "Name should match")
}

// mockLogger is a simple mock logger for testing
type mockLogger struct {
	logs []string
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) Debug(msg string, args ...interface{}) {
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) Warn(msg string, args ...interface{}) {
	m.logs = append(m.logs, msg)
}
