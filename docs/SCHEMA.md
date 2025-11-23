# PostgreSQL Schema Support Guide

This guide covers PostgreSQL schema support in dgcore-database.

## Overview

PostgreSQL schemas provide a way to organize database objects into logical groups. The `dgcore-database` plugin supports schemas through the `search_path` parameter, allowing you to:

- Use custom schemas instead of the default `public` schema
- Implement multi-tenancy with schema isolation
- Separate environments (dev, staging, prod) in the same database
- Organize features into separate schemas

## Configuration

### Basic Schema Usage

```go
config := database.DefaultConfig().
    WithDriver("postgres").
    WithDatabase("myapp").
    WithSchema("my_schema")

manager, _ := database.NewManager(config, nil)

// All operations use my_schema
manager.DB().Find(&users) // SELECT * FROM my_schema.users
```

### Multi-Tenant Schemas

```go
config := database.Config{
    Driver:   "postgres",
    Database: "myapp",
    Schema:   "public", // Default/shared schema
    
    Connections: map[string]database.ConnectionConfig{
        "tenant_1": {
            Driver:   "postgres",
            Database: "myapp",
            Schema:   "tenant_1",
        },
        "tenant_2": {
            Driver:   "postgres",
            Database: "myapp",
            Schema:   "tenant_2",
        },
        "tenant_3": {
            Driver:   "postgres",
            Database: "myapp",
            Schema:   "tenant_3",
        },
    },
}

manager, _ := database.NewManager(config, nil)

// Each tenant isolated by schema
manager.Connection("tenant_1").Find(&users) // tenant_1.users
manager.Connection("tenant_2").Find(&orders) // tenant_2.orders
```

### Read/Write Splitting with Schemas

```go
config := database.Config{
    Driver:             "postgres",
    ReadWriteSplitting: true,
    AutoRouting:        true,
    
    Master: database.ConnectionConfig{
        Driver:   "postgres",
        Host:     "master.db.com",
        Database: "myapp",
        Schema:   "production",
    },
    
    Slaves: []database.ConnectionConfig{
        {
            Driver:   "postgres",
            Host:     "slave1.db.com",
            Database: "myapp",
            Schema:   "production",
        },
        {
            Driver:   "postgres",
            Host:     "slave2.db.com",
            Database: "myapp",
            Schema:   "production",
        },
    },
}
```

## Use Cases

### 1. Multi-Tenancy

Each tenant gets their own schema in the same database:

```go
// Setup
config := database.Config{
    Driver:   "postgres",
    Database: "saas_app",
    
    Connections: map[string]database.ConnectionConfig{
        "acme_corp":   {Schema: "tenant_acme"},
        "widgets_inc": {Schema: "tenant_widgets"},
        "tech_co":     {Schema: "tenant_tech"},
    },
}

manager, _ := database.NewManager(config, nil)

// Usage
func handleRequest(tenantID string) {
    db := manager.Connection(tenantID)
    
    var users []User
    db.Find(&users) // Automatically uses correct schema
}
```

**Benefits:**
- Data isolation at database level
- Shared database infrastructure
- Easy backup/restore per tenant
- Cost-effective scaling

### 2. Environment Separation

Separate dev, staging, and production in the same database:

```go
config := database.Config{
    Driver:   "postgres",
    Database: "myapp",
    
    Connections: map[string]database.ConnectionConfig{
        "dev":     {Schema: "dev"},
        "staging": {Schema: "staging"},
        "prod":    {Schema: "production"},
    },
}

// Use based on environment
env := os.Getenv("APP_ENV")
db := manager.Connection(env)
```

### 3. Feature Isolation

Organize features into separate schemas:

```go
config := database.Config{
    Driver:   "postgres",
    Database: "myapp",
    Schema:   "core", // Default schema
    
    Connections: map[string]database.ConnectionConfig{
        "analytics": {Schema: "analytics"},
        "reporting": {Schema: "reporting"},
        "audit":     {Schema: "audit"},
    },
}

// Core operations
manager.DB().Find(&users)

// Analytics operations
manager.Connection("analytics").Find(&events)

// Reporting operations
manager.Connection("reporting").Find(&reports)
```

### 4. Microservices

Each microservice uses its own schema:

```go
config := database.Config{
    Driver:   "postgres",
    Database: "microservices",
    
    Connections: map[string]database.ConnectionConfig{
        "users":    {Schema: "users_service"},
        "orders":   {Schema: "orders_service"},
        "products": {Schema: "products_service"},
        "payments": {Schema: "payments_service"},
    },
}
```

## Migrations with Schemas

### Schema-Specific Migrations

```go
// Create schema first
migration1 := database.Migration{
    ID: "001_create_tenant_schema",
    Up: func(db *gorm.DB) error {
        return db.Exec("CREATE SCHEMA IF NOT EXISTS tenant_1").Error
    },
    Down: func(db *gorm.DB) error {
        return db.Exec("DROP SCHEMA IF EXISTS tenant_1 CASCADE").Error
    },
}

// Then migrate tables
migration2 := database.Migration{
    ID: "002_create_users_table",
    Up: func(db *gorm.DB) error {
        return db.AutoMigrate(&User{})
    },
    Down: func(db *gorm.DB) error {
        return db.Migrator().DropTable(&User{})
    },
}

// Run migrations for specific tenant
tenantDB := manager.Connection("tenant_1")
migrator := database.NewMigrator(tenantDB)
migrator.Add(migration1)
migrator.Add(migration2)
migrator.Up()
```

### Shared Schema Migrations

```go
// Migrations that apply to all tenants
sharedMigrations := []database.Migration{
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

// Apply to all tenants
tenants := []string{"tenant_1", "tenant_2", "tenant_3"}
for _, tenant := range tenants {
    db := manager.Connection(tenant)
    migrator := database.NewMigrator(db)
    for _, m := range sharedMigrations {
        migrator.Add(m)
    }
    migrator.Up()
}
```

## Best Practices

### 1. Schema Naming Convention

```go
// Good
"tenant_acme"
"tenant_widgets"
"env_production"
"service_users"

// Bad
"acme"
"widgets"
"prod"
"users"
```

### 2. Schema Creation

Always create schemas before using them:

```sql
CREATE SCHEMA IF NOT EXISTS tenant_1;
CREATE SCHEMA IF NOT EXISTS tenant_2;
```

Or in Go:

```go
db.Exec("CREATE SCHEMA IF NOT EXISTS tenant_1")
```

### 3. Default Schema

Set a sensible default schema:

```go
config := database.Config{
    Schema: "public", // Fallback to public
    Connections: map[string]database.ConnectionConfig{
        // Tenant-specific schemas
    },
}
```

### 4. Schema Permissions

Ensure proper permissions:

```sql
GRANT USAGE ON SCHEMA tenant_1 TO app_user;
GRANT ALL ON ALL TABLES IN SCHEMA tenant_1 TO app_user;
```

### 5. Connection Pooling

Configure appropriate pool sizes for multi-tenant scenarios:

```go
config := database.Config{
    MaxOpenConns: 100, // Total connections
    MaxIdleConns: 10,  // Per connection
    
    Connections: map[string]database.ConnectionConfig{
        "tenant_1": {
            MaxOpenConns: intPtr(20), // Limit per tenant
            MaxIdleConns: intPtr(2),
        },
    },
}
```

## Limitations

### MySQL

MySQL doesn't have true schemas like PostgreSQL. In MySQL:
- Database = Schema
- Use separate databases for multi-tenancy
- Schema field is ignored for MySQL connections

### SQLite

SQLite doesn't support schemas:
- Single schema per database file
- Use separate database files for multi-tenancy
- Schema field is ignored for SQLite connections

## Performance Considerations

1. **Search Path**: Setting `search_path` is lightweight and doesn't impact performance
2. **Connection Pooling**: Each schema uses the same connection pool
3. **Indexes**: Create indexes in each schema independently
4. **Statistics**: PostgreSQL maintains statistics per schema

## Troubleshooting

### Schema Not Found

```
ERROR: schema "tenant_1" does not exist
```

**Solution**: Create the schema first:
```sql
CREATE SCHEMA IF NOT EXISTS tenant_1;
```

### Permission Denied

```
ERROR: permission denied for schema tenant_1
```

**Solution**: Grant proper permissions:
```sql
GRANT USAGE ON SCHEMA tenant_1 TO app_user;
GRANT ALL ON ALL TABLES IN SCHEMA tenant_1 TO app_user;
```

### Wrong Schema

If operations are using the wrong schema, verify:

```go
// Check DSN
dsn := buildPostgresDSN(config)
fmt.Println(dsn) // Should contain search_path=your_schema

// Verify connection
db := manager.Connection("tenant_1")
var result string
db.Raw("SELECT current_schema()").Scan(&result)
fmt.Println("Current schema:", result)
```

## Example: Complete Multi-Tenant Setup

```go
package main

import (
    "fmt"
    "github.com/donnigundala/dgcore-database"
)

func main() {
    // 1. Configure multi-tenant database
    config := database.Config{
        Driver:   "postgres",
        Host:     "localhost",
        Port:     5432,
        Database: "saas_app",
        Username: "app_user",
        Password: "password",
        Schema:   "public", // Shared/default schema
        
        Connections: map[string]database.ConnectionConfig{
            "tenant_acme":    {Schema: "tenant_acme"},
            "tenant_widgets": {Schema: "tenant_widgets"},
        },
    }
    
    manager, _ := database.NewManager(config, nil)
    defer manager.Close()
    
    // 2. Create schemas (one-time setup)
    manager.DB().Exec("CREATE SCHEMA IF NOT EXISTS tenant_acme")
    manager.DB().Exec("CREATE SCHEMA IF NOT EXISTS tenant_widgets")
    
    // 3. Run migrations for each tenant
    tenants := []string{"tenant_acme", "tenant_widgets"}
    for _, tenant := range tenants {
        db := manager.Connection(tenant)
        db.AutoMigrate(&User{}, &Order{})
    }
    
    // 4. Use tenant-specific connections
    acmeDB := manager.Connection("tenant_acme")
    acmeDB.Create(&User{Name: "Acme User"})
    
    widgetsDB := manager.Connection("tenant_widgets")
    widgetsDB.Create(&User{Name: "Widgets User"})
    
    // 5. Verify isolation
    var acmeUsers, widgetUsers []User
    acmeDB.Find(&acmeUsers)
    widgetsDB.Find(&widgetUsers)
    
    fmt.Printf("Acme users: %d\n", len(acmeUsers))       // 1
    fmt.Printf("Widgets users: %d\n", len(widgetUsers))  // 1
}
```
