# dgcore-database

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A powerful, production-ready database plugin for the DG Framework with support for read/write splitting, multi-connection management, and comprehensive database operations.

## Features

### 🚀 Core Features
- **Multiple Database Drivers**: MySQL, PostgreSQL, SQLite
- **Connection Pooling**: Configurable connection pool settings
- **Auto-Migration**: Automatic database schema migration
- **Health Monitoring**: Built-in health checks for all connections
- **Transaction Support**: Full transaction support with context and savepoints

### ⚡ Advanced Features
- **Read/Write Splitting**: Automatic routing of reads to slaves and writes to master
- **Load Balancing**: Round-robin, random, and weighted strategies
- **Multi-Connection**: Manage multiple named database connections
- **Runtime Management**: Add/remove connections at runtime
- **Fluent Configuration**: Easy-to-use configuration API

## Installation

```bash
go get github.com/donnigundala/dgcore-database
```

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/donnigundala/dgcore-database"
    "github.com/donnigundala/dgcore/foundation"
)

func main() {
    // Create application
    app := foundation.NewApplication()
    
    // Configure database
    config := database.DefaultConfig().
        WithDriver("mysql").
        WithHost("localhost").
        WithPort(3306).
        WithDatabase("myapp").
        WithCredentials("user", "password")
    
    // Register provider
    provider := database.NewServiceProvider(config)
    app.Register(provider)
    app.Boot()
    
    // Get database manager
    db := app.Make("db").(*database.Manager)
    
    // Use database
    var users []User
    db.DB().Find(&users)
}
```

### Read/Write Splitting

```go
config := database.Config{
    Driver:             "mysql",
    ReadWriteSplitting: true,
    AutoRouting:        true,
    SlaveStrategy:      "round-robin",
    
    Master: database.ConnectionConfig{
        Host:     "master.db.com",
        Port:     3306,
        Database: "myapp",
        Username: "user",
        Password: "pass",
    },
    
    Slaves: []database.ConnectionConfig{
        {Host: "slave1.db.com", Port: 3306, Database: "myapp", Username: "user", Password: "pass", Weight: 2},
        {Host: "slave2.db.com", Port: 3306, Database: "myapp", Username: "user", Password: "pass", Weight: 1},
    },
}

manager, _ := database.NewManager(config, nil)

// Automatic routing
manager.DB().Find(&users)        // → Routes to slave
manager.DB().Create(&newUser)    // → Routes to master

// Manual control
manager.Master().Find(&users)    // → Force master
manager.Read().Find(&users)      // → Force slave
manager.Slave(0).Find(&users)    // → Specific slave
```

### Multi-Connection

```go
config := database.Config{
    Driver:   "mysql",
    Database: "primary",
    
    Connections: map[string]database.ConnectionConfig{
        "analytics": {
            Driver:   "postgres",
            Host:     "analytics.db.com",
            Database: "analytics",
        },
        "logs": {
            Driver:   "postgres",
            Host:     "logs.db.com",
            Database: "logs",
        },
    },
}

manager, _ := database.NewManager(config, nil)

// Use different connections
manager.DB().Find(&users)                      // Primary
manager.Connection("analytics").Find(&events)  // Analytics
manager.Connection("logs").Find(&logs)         // Logs

// Runtime management
manager.AddConnection("cache", cacheConfig)
manager.RemoveConnection("cache")
```

## Configuration

### Basic Configuration

```go
config := database.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "myapp",
    Username: "user",
    Password: "password",
    
    // Connection Pool
    MaxOpenConns:    100,
    MaxIdleConns:    10,
    ConnMaxLifetime: time.Hour,
    ConnMaxIdleTime: 10 * time.Minute,
    
    // Logging
    LogLevel:      "warn",
    SlowThreshold: 200 * time.Millisecond,
    
    // Auto Migration
    AutoMigrate: true,
    Models:      []interface{}{&User{}, &Post{}},
}
```

### Fluent API

```go
config := database.DefaultConfig().
    WithDriver("postgres").
    WithHost("localhost").
    WithPort(5432).
    WithDatabase("myapp").
    WithCredentials("user", "pass").
    WithMaxConnections(100, 10).
    WithAutoMigrate(&User{}, &Post{}).
    WithSlaveStrategy("weighted").
    WithAutoRouting(true)
```

## Features Guide

### Transactions

```go
// Automatic transaction
err := manager.WithTx(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err // Automatic rollback
    }
    if err := tx.Create(&profile).Error; err != nil {
        return err // Automatic rollback
    }
    return nil // Automatic commit
})

// With context
ctx := context.Background()
err = manager.WithTxContext(ctx, func(tx *gorm.DB) error {
    // Transaction operations
    return nil
})

// Manual transaction
tx := manager.TX().Begin()
tx.Create(&user)
tx.Commit() // or tx.Rollback()
```

### Migrations

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

// Run migrations
manager.Migrate(migrations)

// Rollback
manager.Rollback(migrations)

// Check status
status, _ := manager.MigrationStatus()
```

### Health Checks

```go
health := manager.HealthCheck()

for name, status := range health {
    if status {
        fmt.Printf("%s: ✅ UP\n", name)
    } else {
        fmt.Printf("%s: ❌ DOWN\n", name)
    }
}

// Output:
// primary: ✅ UP
// master: ✅ UP
// slave_0: ✅ UP
// slave_1: ✅ UP
// analytics: ✅ UP
```

## Load Balancing Strategies

### Round-Robin
```go
config.WithSlaveStrategy("round-robin")
// Distributes queries evenly: slave1 → slave2 → slave1 → ...
```

### Random
```go
config.WithSlaveStrategy("random")
// Randomly selects a slave for each query
```

### Weighted
```go
config.Slaves = []database.ConnectionConfig{
    {Host: "slave1", Weight: 3}, // 75% of traffic
    {Host: "slave2", Weight: 1}, // 25% of traffic
}
config.WithSlaveStrategy("weighted")
```

## Use Cases

### Multi-Tenancy
```go
// Add tenant database at runtime
tenantConfig := database.ConnectionConfig{
    Driver:   "mysql",
    Host:     "tenant1.db.com",
    Database: "tenant1",
}
manager.AddConnection("tenant_1", tenantConfig)

// Use tenant database
manager.Connection("tenant_1").Find(&users)
```

### Microservices
```go
config := database.Config{
    Connections: map[string]database.ConnectionConfig{
        "users":    {Database: "users_service"},
        "orders":   {Database: "orders_service"},
        "products": {Database: "products_service"},
    },
}
```

### Analytics Separation
```go
// Operational data
manager.DB().Create(&order)

// Analytics data
manager.Connection("analytics").Create(&event)
```

## API Reference

### Manager Methods

#### Connection Management
- `DB() *gorm.DB` - Get primary database connection
- `Connection(name string) *gorm.DB` - Get named connection
- `HasConnection(name string) bool` - Check if connection exists
- `AddConnection(name string, config ConnectionConfig) error` - Add connection at runtime
- `RemoveConnection(name string) error` - Remove connection
- `Close() error` - Close all connections
- `Ping() error` - Test primary connection

#### Read/Write Splitting
- `Master() *gorm.DB` - Force master connection
- `Read() *gorm.DB` - Get slave for reads
- `Write() *gorm.DB` - Get master for writes
- `Slave(index int) *gorm.DB` - Get specific slave

#### Transactions
- `WithTx(fn TransactionFunc) error` - Run transaction
- `WithTxContext(ctx context.Context, fn TransactionFunc) error` - Transaction with context
- `TX() *TransactionHelper` - Get transaction helper
- `Transaction(fn func(*gorm.DB) error) error` - Run transaction

#### Migrations
- `Migrate(migrations []Migration) error` - Run migrations
- `Rollback(migrations []Migration) error` - Rollback last migration
- `MigrationStatus() ([]string, error)` - Get migration status
- `AutoMigrate(models ...interface{}) error` - Auto-migrate models

#### Health
- `HealthCheck() map[string]bool` - Check all connections

## Examples

See the [examples](./examples) directory for complete working examples:

- [01-basic](./examples/01-basic) - Basic CRUD operations
- [02-read-write-splitting](./examples/02-read-write-splitting) - Master/slave setup
- [03-multi-connection](./examples/03-multi-connection) - Multiple databases
- [04-migrations](./examples/04-migrations) - Database migrations

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific test
go test -run TestManager_CRUD
```

## Performance Tips

1. **Connection Pooling**: Tune `MaxOpenConns` and `MaxIdleConns` based on your workload
2. **Read/Write Splitting**: Use slaves for read-heavy workloads
3. **Weighted Load Balancing**: Assign higher weights to more powerful slaves
4. **Health Checks**: Monitor connection health to detect issues early
5. **Transaction Scope**: Keep transactions small and focused

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [Full documentation](https://github.com/donnigundala/dgcore-database)
- **Issues**: [GitHub Issues](https://github.com/donnigundala/dgcore-database/issues)
- **Discussions**: [GitHub Discussions](https://github.com/donnigundala/dgcore-database/discussions)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release history.
