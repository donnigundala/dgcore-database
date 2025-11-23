package database

import (
	"fmt"
	"math/rand"
	"sync"

	"gorm.io/gorm"
)

// Manager manages database connections with support for read/write splitting and multi-connection.
type Manager struct {
	// Primary connection
	db     *gorm.DB
	config Config
	logger interface{} // Can be *logging.Logger or any logger with Info/Warn/Error methods

	// ========== Read/Write Splitting ==========
	master *gorm.DB
	slaves []*gorm.DB
	plugin *ReadWritePlugin

	slaveIndex int
	slaveMu    sync.Mutex

	// ========== Multi-Connection Support ==========
	connections map[string]*gorm.DB
	connMu      sync.RWMutex
}

// NewManager creates a new database manager.
func NewManager(config Config, logger interface{}) (*Manager, error) {
	manager := &Manager{
		config:      config,
		logger:      logger,
		connections: make(map[string]*gorm.DB),
	}

	// Setup primary/default connection
	if err := manager.setupPrimaryConnection(); err != nil {
		return nil, err
	}

	// Setup read/write splitting if enabled
	if config.ReadWriteSplitting {
		if err := manager.setupReadWriteSplitting(); err != nil {
			return nil, err
		}
	}

	// Setup named connections if configured
	if len(config.Connections) > 0 {
		if err := manager.setupNamedConnections(); err != nil {
			return nil, err
		}
	}

	return manager, nil
}

func (m *Manager) setupPrimaryConnection() error {
	// Create primary connection
	db, err := connect(m.config, m.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(m.config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(m.config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(m.config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(m.config.ConnMaxIdleTime)

	m.db = db
	m.master = db // Master is primary by default

	// Auto migrate if configured
	if m.config.AutoMigrate && len(m.config.Models) > 0 {
		if err := m.AutoMigrate(m.config.Models...); err != nil {
			return fmt.Errorf("auto migration failed: %w", err)
		}
	}

	return nil
}

func (m *Manager) setupReadWriteSplitting() error {
	// Connect to master
	if m.config.Master.Host != "" {
		master, err := connectWithConfig(m.config.Master, m.logger)
		if err != nil {
			return fmt.Errorf("failed to connect to master: %w", err)
		}
		m.master = master
	}

	// Connect to slaves
	for i, slaveConfig := range m.config.Slaves {
		slave, err := connectWithConfig(slaveConfig, m.logger)
		if err != nil {
			m.logWarn("Failed to connect to slave", "index", i, "error", err)
			continue
		}
		m.slaves = append(m.slaves, slave)
	}

	if len(m.slaves) == 0 {
		m.logWarn("No slaves available, using master for reads")
	}

	// Enable automatic routing if configured
	if m.config.AutoRouting {
		m.plugin = NewReadWritePlugin(m)
		m.master.Use(m.plugin)
		m.logInfo("Automatic read/write routing enabled")
	}

	return nil
}

func (m *Manager) setupNamedConnections() error {
	for name, connConfig := range m.config.Connections {
		db, err := connectWithConfig(connConfig, m.logger)
		if err != nil {
			m.logWarn("Failed to connect to named connection", "name", name, "error", err)
			continue
		}

		m.connMu.Lock()
		m.connections[name] = db
		m.connMu.Unlock()

		m.logInfo("Named connection established", "name", name)
	}

	return nil
}

// ========== Primary Connection Methods ==========

// DB returns the primary database connection.
// With auto-routing enabled, reads use slaves, writes use master.
func (m *Manager) DB() *gorm.DB {
	return m.master
}

// Ping tests the database connection.
func (m *Manager) Ping() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close closes all database connections.
func (m *Manager) Close() error {
	// Close primary
	if m.db != nil {
		sqlDB, _ := m.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	// Close slaves
	for _, slave := range m.slaves {
		if sqlDB, err := slave.DB(); err == nil {
			sqlDB.Close()
		}
	}

	// Close named connections
	m.connMu.RLock()
	defer m.connMu.RUnlock()
	for _, conn := range m.connections {
		if sqlDB, err := conn.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return nil
}

// ========== Read/Write Splitting Methods ==========

// Master returns the master connection (forces master for reads).
func (m *Manager) Master() *gorm.DB {
	return m.master.Session(&gorm.Session{})
}

// Read returns a slave connection for read operations.
// Falls back to master if no slaves available.
func (m *Manager) Read() *gorm.DB {
	if !m.config.ReadWriteSplitting || len(m.slaves) == 0 {
		return m.master
	}
	return m.selectSlave()
}

// Write returns the master connection for write operations.
func (m *Manager) Write() *gorm.DB {
	return m.master
}

// Slave returns a specific slave connection by index.
func (m *Manager) Slave(index int) *gorm.DB {
	if index >= 0 && index < len(m.slaves) {
		return m.slaves[index]
	}
	m.logWarn("Invalid slave index, using master", "index", index)
	return m.master
}

func (m *Manager) selectSlave() *gorm.DB {
	m.slaveMu.Lock()
	defer m.slaveMu.Unlock()

	switch m.config.SlaveStrategy {
	case "round-robin":
		slave := m.slaves[m.slaveIndex]
		m.slaveIndex = (m.slaveIndex + 1) % len(m.slaves)
		return slave

	case "random":
		idx := rand.Intn(len(m.slaves))
		return m.slaves[idx]

	case "weighted":
		return m.selectWeightedSlave()

	default:
		return m.slaves[0]
	}
}

func (m *Manager) selectWeightedSlave() *gorm.DB {
	// Calculate total weight
	totalWeight := 0
	for _, slaveConfig := range m.config.Slaves {
		totalWeight += slaveConfig.Weight
	}

	if totalWeight == 0 {
		return m.slaves[0]
	}

	// Select based on weight
	r := rand.Intn(totalWeight)
	cumulative := 0

	for i, slaveConfig := range m.config.Slaves {
		cumulative += slaveConfig.Weight
		if r < cumulative {
			return m.slaves[i]
		}
	}

	return m.slaves[0]
}

// ========== Multi-Connection Methods ==========

// Connection returns a named connection.
func (m *Manager) Connection(name string) *gorm.DB {
	m.connMu.RLock()
	defer m.connMu.RUnlock()

	if conn, exists := m.connections[name]; exists {
		return conn
	}

	m.logWarn("Connection not found, using default", "name", name)
	return m.db
}

// HasConnection checks if a named connection exists.
func (m *Manager) HasConnection(name string) bool {
	m.connMu.RLock()
	defer m.connMu.RUnlock()
	_, exists := m.connections[name]
	return exists
}

// AddConnection adds a new named connection at runtime.
func (m *Manager) AddConnection(name string, config ConnectionConfig) error {
	db, err := connectWithConfig(config, m.logger)
	if err != nil {
		return fmt.Errorf("failed to add connection %s: %w", name, err)
	}

	m.connMu.Lock()
	m.connections[name] = db
	m.connMu.Unlock()

	m.logInfo("Connection added", "name", name)
	return nil
}

// RemoveConnection removes a named connection.
func (m *Manager) RemoveConnection(name string) error {
	m.connMu.Lock()
	defer m.connMu.Unlock()

	if conn, exists := m.connections[name]; exists {
		if sqlDB, err := conn.DB(); err == nil {
			sqlDB.Close()
		}
		delete(m.connections, name)
		m.logInfo("Connection removed", "name", name)
		return nil
	}

	return fmt.Errorf("connection not found: %s", name)
}

// ========== Common Methods ==========

// Transaction runs a function within a transaction.
func (m *Manager) Transaction(fn func(*gorm.DB) error) error {
	return m.master.Transaction(fn)
}

// AutoMigrate runs auto migration for given models.
func (m *Manager) AutoMigrate(models ...interface{}) error {
	return m.master.AutoMigrate(models...)
}

// HealthCheck returns health status of all connections.
func (m *Manager) HealthCheck() map[string]bool {
	health := make(map[string]bool)

	// Check primary
	health["primary"] = m.ping(m.db) == nil

	// Check master
	if m.config.ReadWriteSplitting {
		health["master"] = m.ping(m.master) == nil

		// Check slaves
		for i, slave := range m.slaves {
			key := fmt.Sprintf("slave_%d", i)
			health[key] = m.ping(slave) == nil
		}
	}

	// Check named connections
	m.connMu.RLock()
	for name, conn := range m.connections {
		health[name] = m.ping(conn) == nil
	}
	m.connMu.RUnlock()

	return health
}

func (m *Manager) ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// ========== Logger Helpers ==========

func (m *Manager) logInfo(msg string, args ...interface{}) {
	if m.logger == nil {
		return
	}
	// Try to call Info method if available
	type infoLogger interface {
		Info(string, ...interface{})
	}
	if l, ok := m.logger.(infoLogger); ok {
		l.Info(msg, args...)
	}
}

func (m *Manager) logWarn(msg string, args ...interface{}) {
	if m.logger == nil {
		return
	}
	// Try to call Warn method if available
	type warnLogger interface {
		Warn(string, ...interface{})
	}
	if l, ok := m.logger.(warnLogger); ok {
		l.Warn(msg, args...)
	}
}
