package database

import "time"

// Config holds the database configuration
type Config struct {
	// Driver: mysql, postgres, sqlite, sqlserver
	Driver string

	// Connection details
	Host     string
	Port     int
	Database string
	Username string
	Password string

	// SQLite specific
	FilePath string

	// Connection pool
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Options
	Charset   string
	Timezone  string
	ParseTime bool
	SSLMode   string
	Schema    string // PostgreSQL schema (default: public), MySQL: not used

	// Logging
	LogLevel      string // silent, error, warn, info
	SlowThreshold time.Duration

	// Slow query logging
	SlowQuery SlowQueryConfig

	// Auto migration
	AutoMigrate bool
	Models      []interface{}

	// ========== Read/Write Splitting ==========
	ReadWriteSplitting bool
	AutoRouting        bool // Automatic routing (default: true)

	// Master (write) connection
	Master ConnectionConfig

	// Slaves (read) connections
	Slaves []ConnectionConfig

	// Load balancing strategy
	SlaveStrategy string // round-robin, random, weighted

	// ========== Multi-Connection Support ==========
	// Named connections for multiple databases
	Connections map[string]ConnectionConfig

	// Default connection name
	DefaultConnection string
}

// SlowQueryConfig holds configuration for slow query logging.
type SlowQueryConfig struct {
	Enabled   bool          // Enable slow query logging
	Threshold time.Duration // Queries slower than this are logged
	LogStack  bool          // Include stack trace in logs
}

// ConnectionConfig holds configuration for a single database connection
type ConnectionConfig struct {
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string

	// SQLite
	FilePath string

	// Connection pool (optional, inherits from main config if not set)
	MaxOpenConns    *int
	MaxIdleConns    *int
	ConnMaxLifetime *time.Duration

	// For weighted load balancing
	Weight int

	// Options
	Charset   string
	Timezone  string
	ParseTime bool
	SSLMode   string
	Schema    string // PostgreSQL schema (default: public)
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Driver:          "mysql",
		Host:            "localhost",
		Port:            3306,
		Charset:         "utf8mb4",
		ParseTime:       true,
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		LogLevel:        "warn",
		SlowThreshold:   200 * time.Millisecond,

		// Read/Write splitting defaults
		ReadWriteSplitting: false,
		AutoRouting:        true,
		SlaveStrategy:      "round-robin",

		// Multi-connection defaults
		DefaultConnection: "default",
		Connections:       make(map[string]ConnectionConfig),
	}
}

// Fluent setters

// WithDriver sets the database driver
func (c Config) WithDriver(driver string) Config {
	c.Driver = driver
	return c
}

// WithHost sets the database host
func (c Config) WithHost(host string) Config {
	c.Host = host
	return c
}

// WithPort sets the database port
func (c Config) WithPort(port int) Config {
	c.Port = port
	return c
}

// WithDatabase sets the database name
func (c Config) WithDatabase(database string) Config {
	c.Database = database
	return c
}

// WithCredentials sets the database credentials
func (c Config) WithCredentials(username, password string) Config {
	c.Username = username
	c.Password = password
	return c
}

// WithAutoMigrate enables auto-migration for the given models
func (c Config) WithAutoMigrate(models ...interface{}) Config {
	c.AutoMigrate = true
	c.Models = models
	return c
}

// ========== Read/Write Splitting Setters ==========

// WithReadWriteSplitting enables read/write splitting
func (c Config) WithReadWriteSplitting(master ConnectionConfig, slaves ...ConnectionConfig) Config {
	c.ReadWriteSplitting = true
	c.Master = master
	c.Slaves = slaves
	return c
}

// WithAutoRouting enables or disables automatic routing
func (c Config) WithAutoRouting(enabled bool) Config {
	c.AutoRouting = enabled
	return c
}

// WithSlaveStrategy sets the slave selection strategy
func (c Config) WithSlaveStrategy(strategy string) Config {
	c.SlaveStrategy = strategy
	return c
}

// ========== Multi-Connection Setters ==========

// WithConnection adds a named connection
func (c Config) WithConnection(name string, config ConnectionConfig) Config {
	if c.Connections == nil {
		c.Connections = make(map[string]ConnectionConfig)
	}
	c.Connections[name] = config
	return c
}

// WithDefaultConnection sets the default connection name
func (c Config) WithDefaultConnection(name string) Config {
	c.DefaultConnection = name
	return c
}

// WithSchema sets the database schema (PostgreSQL)
func (c Config) WithSchema(schema string) Config {
	c.Schema = schema
	return c
}

// WithSlowQueryLogging enables slow query logging with the specified threshold.
func (c Config) WithSlowQueryLogging(threshold time.Duration) Config {
	c.SlowQuery = SlowQueryConfig{
		Enabled:   true,
		Threshold: threshold,
		LogStack:  false,
	}
	return c
}

// WithSlowQueryLoggingAndStack enables slow query logging with stack traces.
func (c Config) WithSlowQueryLoggingAndStack(threshold time.Duration) Config {
	c.SlowQuery = SlowQueryConfig{
		Enabled:   true,
		Threshold: threshold,
		LogStack:  true,
	}
	return c
}

// WithMaxConnections sets the connection pool limits
func (c Config) WithMaxConnections(maxOpen, maxIdle int) Config {
	c.MaxOpenConns = maxOpen
	c.MaxIdleConns = maxIdle
	return c
}
