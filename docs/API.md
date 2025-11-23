# API Reference

## Table of Contents
- [Manager](#manager)
- [Configuration](#configuration)
- [Migrations](#migrations)
- [Transactions](#transactions)

## Manager

The `Manager` is the central component for database operations.

### Constructor

#### `NewManager(config Config, logger interface{}) (*Manager, error)`

Creates a new database manager.

**Parameters:**
- `config` - Database configuration
- `logger` - Optional logger instance (can be nil)

**Returns:**
- `*Manager` - Database manager instance
- `error` - Error if connection fails

**Example:**
```go
config := database.DefaultConfig().WithDriver("mysql")
manager, err := database.NewManager(config, nil)
if err != nil {
    log.Fatal(err)
}
defer manager.Close()
```

### Connection Methods

#### `DB() *gorm.DB`

Returns the primary database connection. With auto-routing enabled, this will automatically route reads to slaves and writes to master.

**Returns:**
- `*gorm.DB` - GORM database instance

**Example:**
```go
db := manager.DB()
db.Find(&users)
```

#### `Connection(name string) *gorm.DB`

Returns a named database connection.

**Parameters:**
- `name` - Connection name

**Returns:**
- `*gorm.DB` - GORM database instance

**Example:**
```go
analyticsDB := manager.Connection("analytics")
analyticsDB.Find(&events)
```

#### `HasConnection(name string) bool`

Checks if a named connection exists.

**Parameters:**
- `name` - Connection name

**Returns:**
- `bool` - True if connection exists

**Example:**
```go
if manager.HasConnection("analytics") {
    // Use analytics connection
}
```

#### `AddConnection(name string, config ConnectionConfig) error`

Adds a new named connection at runtime.

**Parameters:**
- `name` - Connection name
- `config` - Connection configuration

**Returns:**
- `error` - Error if connection fails

**Example:**
```go
config := database.ConnectionConfig{
    Driver:   "postgres",
    Host:     "analytics.db.com",
    Database: "analytics",
}
err := manager.AddConnection("analytics", config)
```

#### `RemoveConnection(name string) error`

Removes a named connection.

**Parameters:**
- `name` - Connection name

**Returns:**
- `error` - Error if connection not found

**Example:**
```go
err := manager.RemoveConnection("analytics")
```

#### `Close() error`

Closes all database connections.

**Returns:**
- `error` - Error if close fails

**Example:**
```go
defer manager.Close()
```

#### `Ping() error`

Tests the primary database connection.

**Returns:**
- `error` - Error if ping fails

**Example:**
```go
if err := manager.Ping(); err != nil {
    log.Fatal("Database connection failed")
}
```

### Read/Write Splitting Methods

#### `Master() *gorm.DB`

Returns the master connection, forcing all operations to use the master database.

**Returns:**
- `*gorm.DB` - Master database instance

**Example:**
```go
// Force read from master
manager.Master().Find(&users)
```

#### `Read() *gorm.DB`

Returns a slave connection for read operations. Falls back to master if no slaves available.

**Returns:**
- `*gorm.DB` - Slave database instance

**Example:**
```go
// Explicit read from slave
manager.Read().Find(&users)
```

#### `Write() *gorm.DB`

Returns the master connection for write operations.

**Returns:**
- `*gorm.DB` - Master database instance

**Example:**
```go
// Explicit write to master
manager.Write().Create(&user)
```

#### `Slave(index int) *gorm.DB`

Returns a specific slave connection by index.

**Parameters:**
- `index` - Slave index (0-based)

**Returns:**
- `*gorm.DB` - Slave database instance

**Example:**
```go
// Use specific slave
manager.Slave(0).Find(&users)
```

### Transaction Methods

#### `WithTx(fn TransactionFunc) error`

Runs a function within a transaction. Automatically commits on success, rolls back on error.

**Parameters:**
- `fn` - Function to run in transaction

**Returns:**
- `error` - Error from transaction

**Example:**
```go
err := manager.WithTx(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    return tx.Create(&profile).Error
})
```

#### `WithTxContext(ctx context.Context, fn TransactionFunc) error`

Runs a function within a transaction with context support.

**Parameters:**
- `ctx` - Context
- `fn` - Function to run in transaction

**Returns:**
- `error` - Error from transaction

**Example:**
```go
ctx := context.Background()
err := manager.WithTxContext(ctx, func(tx *gorm.DB) error {
    return tx.Create(&user).Error
})
```

#### `TX() *TransactionHelper`

Returns a transaction helper for advanced transaction management.

**Returns:**
- `*TransactionHelper` - Transaction helper instance

**Example:**
```go
helper := manager.TX()
err := helper.Run(func(tx *gorm.DB) error {
    return tx.Create(&user).Error
})
```

#### `Transaction(fn func(*gorm.DB) error) error`

Runs a function within a transaction (alias for WithTx).

**Parameters:**
- `fn` - Function to run in transaction

**Returns:**
- `error` - Error from transaction

### Migration Methods

#### `Migrate(migrations []Migration) error`

Runs database migrations.

**Parameters:**
- `migrations` - Array of migration definitions

**Returns:**
- `error` - Error if migration fails

**Example:**
```go
migrations := []database.Migration{
    {
        ID: "001_create_users",
        Up: func(db *gorm.DB) error {
            return db.AutoMigrate(&User{})
        },
        Down: func(db *gorm.DB) error {
            return db.Migrator().DropTable(&User{})
        },
    },
}
err := manager.Migrate(migrations)
```

#### `Rollback(migrations []Migration) error`

Rolls back the last migration.

**Parameters:**
- `migrations` - Array of migration definitions

**Returns:**
- `error` - Error if rollback fails

**Example:**
```go
err := manager.Rollback(migrations)
```

#### `MigrationStatus() ([]string, error)`

Returns the list of applied migrations.

**Returns:**
- `[]string` - Array of migration IDs
- `error` - Error if status check fails

**Example:**
```go
status, err := manager.MigrationStatus()
for _, id := range status {
    fmt.Println("Applied:", id)
}
```

#### `AutoMigrate(models ...interface{}) error`

Automatically migrates database schema for given models.

**Parameters:**
- `models` - Model structs to migrate

**Returns:**
- `error` - Error if migration fails

**Example:**
```go
err := manager.AutoMigrate(&User{}, &Post{}, &Comment{})
```

### Health Check Methods

#### `HealthCheck() map[string]bool`

Returns health status of all database connections.

**Returns:**
- `map[string]bool` - Map of connection names to health status

**Example:**
```go
health := manager.HealthCheck()
for name, status := range health {
    if status {
        fmt.Printf("%s: UP\n", name)
    } else {
        fmt.Printf("%s: DOWN\n", name)
    }
}
```

## Configuration

### Config Struct

```go
type Config struct {
    // Basic connection
    Driver   string
    Host     string
    Port     int
    Database string
    Username string
    Password string
    FilePath string // For SQLite
    
    // Connection pool
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
    
    // Logging
    LogLevel      string
    SlowThreshold time.Duration
    
    // MySQL specific
    Charset   string
    ParseTime bool
    
    // PostgreSQL specific
    SSLMode  string
    Timezone string
    
    // Auto migration
    AutoMigrate bool
    Models      []interface{}
    
    // Read/write splitting
    ReadWriteSplitting bool
    AutoRouting        bool
    SlaveStrategy      string
    Master             ConnectionConfig
    Slaves             []ConnectionConfig
    
    // Multi-connection
    Connections       map[string]ConnectionConfig
    DefaultConnection string
}
```

### Fluent Configuration Methods

#### `DefaultConfig() Config`

Returns a configuration with sensible defaults.

#### `WithDriver(driver string) Config`

Sets the database driver ("mysql", "postgres", "sqlite").

#### `WithHost(host string) Config`

Sets the database host.

#### `WithPort(port int) Config`

Sets the database port.

#### `WithDatabase(database string) Config`

Sets the database name.

#### `WithCredentials(username, password string) Config`

Sets database credentials.

#### `WithMaxConnections(maxOpen, maxIdle int) Config`

Sets connection pool limits.

#### `WithAutoMigrate(models ...interface{}) Config`

Enables auto-migration for specified models.

#### `WithReadWriteSplitting(master ConnectionConfig, slaves ...ConnectionConfig) Config`

Configures read/write splitting.

#### `WithSlaveStrategy(strategy string) Config`

Sets slave selection strategy ("round-robin", "random", "weighted").

#### `WithAutoRouting(enabled bool) Config`

Enables/disables automatic routing.

#### `WithConnection(name string, config ConnectionConfig) Config`

Adds a named connection.

#### `WithDefaultConnection(name string) Config`

Sets the default connection name.

## Migrations

### Migration Struct

```go
type Migration struct {
    ID   string
    Up   func(*gorm.DB) error
    Down func(*gorm.DB) error
}
```

### Migrator

The `Migrator` handles database migrations.

#### `NewMigrator(db *gorm.DB) *Migrator`

Creates a new migrator instance.

#### `Add(migration Migration)`

Adds a migration to the migrator.

#### `Up() error`

Runs all pending migrations.

#### `Down() error`

Rolls back the last migration.

#### `Reset() error`

Rolls back all migrations.

#### `Status() ([]string, error)`

Returns migration status.

## Transactions

### TransactionFunc

```go
type TransactionFunc func(*gorm.DB) error
```

### TransactionHelper

Helper for advanced transaction management.

#### `NewTransactionHelper(db *gorm.DB) *TransactionHelper`

Creates a new transaction helper.

#### `Run(fn TransactionFunc) error`

Runs a function in a transaction.

#### `RunWithContext(ctx context.Context, fn TransactionFunc) error`

Runs a function in a transaction with context.

#### `Begin() *gorm.DB`

Starts a new transaction.

#### `BeginWithOptions(opts *sql.TxOptions) *gorm.DB`

Starts a transaction with options.

### Standalone Functions

#### `WithTransaction(db *gorm.DB, fn TransactionFunc) error`

Runs a function in a transaction.

#### `WithTransactionContext(ctx context.Context, db *gorm.DB, fn TransactionFunc) error`

Runs a function in a transaction with context.

#### `BeginTransaction(db *gorm.DB) *gorm.DB`

Starts a new transaction.

#### `Commit(tx *gorm.DB) error`

Commits a transaction.

#### `Rollback(tx *gorm.DB) error`

Rolls back a transaction.

#### `SavePoint(tx *gorm.DB, name string) error`

Creates a savepoint.

#### `RollbackTo(tx *gorm.DB, name string) error`

Rolls back to a savepoint.
