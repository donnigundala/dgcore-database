# go-migrate Integration Section (to be appended to MIGRATIONS.md)

## go-migrate Integration

The `dgcore-database` plugin provides helper methods to integrate with [go-migrate](https://github.com/golang-migrate/migrate), the industry-standard migration tool.

### Installation

```bash
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/database/mysql
go get -u github.com/golang-migrate/migrate/v4/source/file
```

### Setup

#### 1. Get SQL Database Connection

```go
import (
    "github.com/donnigundala/dgcore-database"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

// Create manager
config := database.DefaultConfig().
    WithDriver("postgres").
    WithDatabase("myapp")

manager, _ := database.NewManager(config, nil)
defer manager.Close()

// Get underlying *sql.DB
sqlDB, _ := manager.SQL()

// Create postgres driver for go-migrate
driver, _ := postgres.WithInstance(sqlDB, &postgres.Config{})

// Create migrator
m, _ := migrate.NewWithDatabaseInstance(
    "file://migrations",  // Migration files directory
    "postgres",           // Database name
    driver,
)
```

#### 2. Create Migration Files

```bash
# Create migrations directory
mkdir -p migrations

# Create migration files
# migrations/000001_create_users.up.sql
# migrations/000001_create_users.down.sql
```

**000001_create_users.up.sql:**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

**000001_create_users.down.sql:**
```sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

#### 3. Run Migrations

```go
// Run all pending migrations
if err := m.Up(); err != nil && err != migrate.ErrNoChange {
    log.Fatal(err)
}

// Rollback one migration
if err := m.Steps(-1); err != nil {
    log.Fatal(err)
}

// Migrate to specific version
if err := m.Migrate(2); err != nil {
    log.Fatal(err)
}

// Get current version
version, dirty, err := m.Version()
if err != nil && err != migrate.ErrNilVersion {
    log.Fatal(err)
}
fmt.Printf("Current version: %d, Dirty: %v\n", version, dirty)
```

### Complete Example

```go
package main

import (
    "log"
    
    "github.com/donnigundala/dgcore-database"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    // 1. Create database manager
    config := database.DefaultConfig().
        WithDriver("postgres").
        WithHost("localhost").
        WithPort(5432).
        WithDatabase("myapp").
        WithCredentials("user", "password")
    
    manager, err := database.NewManager(config, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer manager.Close()
    
    // 2. Get SQL database for go-migrate
    sqlDB, err := manager.SQL()
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. Create postgres driver
    driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. Create migrator
    m, err := migrate.NewWithDatabaseInstance(
        "file://migrations",
        "postgres",
        driver,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 5. Run migrations
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatal(err)
    }
    
    log.Println("Migrations completed successfully")
    
    // 6. Check version
    version, dirty, err := m.Version()
    if err != nil && err != migrate.ErrNilVersion {
        log.Fatal(err)
    }
    log.Printf("Current version: %d, Dirty: %v\n", version, dirty)
}
```

### Multi-Connection Migrations

Run migrations on specific connections:

```go
// Migrate analytics database
analyticsDB, _ := manager.SQLConnection("analytics")
analyticsDriver, _ := postgres.WithInstance(analyticsDB, &postgres.Config{})
analyticsM, _ := migrate.NewWithDatabaseInstance(
    "file://migrations/analytics",
    "postgres",
    analyticsDriver,
)
analyticsM.Up()

// Migrate logs database
logsDB, _ := manager.SQLConnection("logs")
logsDriver, _ := postgres.WithInstance(logsDB, &postgres.Config{})
logsM, _ := migrate.NewWithDatabaseInstance(
    "file://migrations/logs",
    "postgres",
    logsDriver,
)
logsM.Up()
```

### Schema-Specific Migrations

For PostgreSQL schemas:

```go
// Tenant 1 migrations
tenant1DB, _ := manager.SQLConnection("tenant_1")
tenant1Driver, _ := postgres.WithInstance(tenant1DB, &postgres.Config{
    MigrationsTable: "schema_migrations", // Custom table name
})
tenant1M, _ := migrate.NewWithDatabaseInstance(
    "file://migrations/tenants",
    "postgres",
    tenant1Driver,
)
tenant1M.Up()

// Tenant 2 migrations
tenant2DB, _ := manager.SQLConnection("tenant_2")
tenant2Driver, _ := postgres.WithInstance(tenant2DB, &postgres.Config{
    MigrationsTable: "schema_migrations",
})
tenant2M, _ := migrate.NewWithDatabaseInstance(
    "file://migrations/tenants",
    "postgres",
    tenant2Driver,
)
tenant2M.Up()
```

### CLI Usage

go-migrate also provides a CLI tool:

```bash
# Install CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create new migration
migrate create -ext sql -dir migrations -seq create_users

# Run migrations
migrate -path migrations -database "postgres://user:pass@localhost:5432/myapp?sslmode=disable" up

# Rollback
migrate -path migrations -database "postgres://user:pass@localhost:5432/myapp?sslmode=disable" down 1

# Check version
migrate -path migrations -database "postgres://user:pass@localhost:5432/myapp?sslmode=disable" version
```

### Migration File Naming

go-migrate uses versioned migration files:

```
migrations/
  000001_create_users.up.sql
  000001_create_users.down.sql
  000002_add_posts.up.sql
  000002_add_posts.down.sql
  000003_add_indexes.up.sql
  000003_add_indexes.down.sql
```

### Best Practices with go-migrate

1. **Use Sequential Versioning**: `000001`, `000002`, `000003`
2. **Descriptive Names**: `create_users`, `add_email_index`
3. **Always Create Down Migrations**: For rollback capability
4. **Test Migrations**: Test both up and down migrations
5. **Use Transactions**: Wrap migrations in transactions when possible
6. **Version Control**: Commit migration files to git

### go-migrate vs Built-in

| Feature | Built-in | go-migrate |
|---------|----------|------------|
| CLI Tool | ❌ | ✅ |
| SQL Files | ❌ | ✅ |
| Go Migrations | ✅ | ✅ |
| GORM Integration | ✅ | ❌ |
| Version Control | Simple | Advanced |
| Rollback | Last only | To any version |
| External Deps | None | Required |
| Learning Curve | Low | Medium |

## Choosing an Approach

### Use Built-in Migrations When:

- ✅ Building a simple application
- ✅ Prefer type-safe Go code
- ✅ Using GORM extensively
- ✅ Want minimal dependencies
- ✅ Quick prototyping

### Use go-migrate When:

- ✅ Building production applications
- ✅ Need CLI tool for deployments
- ✅ Team prefers SQL migrations
- ✅ Complex migration workflows
- ✅ CI/CD integration
- ✅ Need version-based rollbacks

### Hybrid Approach

You can use both:

```go
// Development: Use built-in for quick iterations
if os.Getenv("ENV") == "development" {
    manager.AutoMigrate(&User{}, &Post{})
}

// Production: Use go-migrate for controlled deployments
if os.Getenv("ENV") == "production" {
    sqlDB, _ := manager.SQL()
    driver, _ := postgres.WithInstance(sqlDB, &postgres.Config{})
    m, _ := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
    m.Up()
}
```

## Skeleton Integration

For the application skeleton, we recommend using go-migrate as the default:

**Project Structure:**
```
myapp/
├── cmd/
│   └── migrate/
│       └── main.go          # Migration CLI
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   └── ...
├── database/
│   └── database.go          # Database setup
└── main.go
```

**cmd/migrate/main.go:**
```go
package main

import (
    "flag"
    "log"
    
    "myapp/database"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    var (
        direction = flag.String("direction", "up", "Migration direction: up, down, or version number")
        steps     = flag.Int("steps", 0, "Number of steps to migrate")
    )
    flag.Parse()
    
    // Get database manager
    manager := database.GetManager()
    defer manager.Close()
    
    // Get SQL DB
    sqlDB, _ := manager.SQL()
    
    // Create migrator
    driver, _ := postgres.WithInstance(sqlDB, &postgres.Config{})
    m, _ := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
    
    // Run migration
    switch *direction {
    case "up":
        if err := m.Up(); err != nil && err != migrate.ErrNoChange {
            log.Fatal(err)
        }
    case "down":
        if *steps > 0 {
            if err := m.Steps(-*steps); err != nil {
                log.Fatal(err)
            }
        } else {
            if err := m.Down(); err != nil {
                log.Fatal(err)
            }
        }
    default:
        log.Fatal("Invalid direction")
    }
    
    log.Println("Migration completed")
}
```

**Usage:**
```bash
# Run all migrations
go run cmd/migrate/main.go -direction=up

# Rollback one migration
go run cmd/migrate/main.go -direction=down -steps=1

# Or use Makefile
make migrate-up
make migrate-down
```

This provides a production-ready migration system while keeping the flexibility of the built-in system for development.
