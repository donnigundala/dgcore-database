# Migration Guide

This guide covers database migrations in dgcore-database.

## Table of Contents
- [Overview](#overview)
- [Migration Approaches](#migration-approaches)
- [Built-in Migrations](#built-in-migrations)
- [go-migrate Integration](#go-migrate-integration)
- [Choosing an Approach](#choosing-an-approach)
- [Best Practices](#best-practices)

## Overview

The `dgcore-database` plugin supports two migration approaches:

1. **Built-in Migration System** - Simple, Go-based migrations integrated with GORM
2. **go-migrate Integration** - Industry-standard migration tool with CLI support

Both approaches can be used depending on your needs.

## Migration Approaches

### Built-in Migrations

**Best for:**
- Simple applications
- GORM-heavy projects
- Quick prototyping
- Embedded applications
- Type-safe migrations

**Features:**
- ✅ Pure Go migrations
- ✅ GORM integration
- ✅ No external dependencies
- ✅ Simple API
- ❌ No CLI tool
- ❌ No SQL file support

### go-migrate

**Best for:**
- Production applications
- Team environments
- CI/CD pipelines
- SQL-heavy migrations
- Complex migration workflows

**Features:**
- ✅ CLI tool
- ✅ SQL file support
- ✅ Version-based migrations
- ✅ Multiple database drivers
- ✅ Rich feature set
- ❌ External dependency
- ❌ Requires SQL knowledge

## Built-in Migrations

## Creating Migrations

### Basic Migration

```go
migration := database.Migration{
    ID: "001_create_users_table",
    Up: func(db *gorm.DB) error {
        type User struct {
            ID        uint   `gorm:"primaryKey"`
            Name      string `gorm:"size:100"`
            Email     string `gorm:"size:100;uniqueIndex"`
            CreatedAt time.Time
        }
        return db.AutoMigrate(&User{})
    },
    Down: func(db *gorm.DB) error {
        return db.Migrator().DropTable("users")
    },
}
```

### Adding Columns

```go
migration := database.Migration{
    ID: "002_add_user_phone",
    Up: func(db *gorm.DB) error {
        return db.Exec("ALTER TABLE users ADD COLUMN phone VARCHAR(20)").Error
    },
    Down: func(db *gorm.DB) error {
        return db.Exec("ALTER TABLE users DROP COLUMN phone").Error
    },
}
```

### Creating Indexes

```go
migration := database.Migration{
    ID: "003_add_user_email_index",
    Up: func(db *gorm.DB) error {
        return db.Exec("CREATE INDEX idx_users_email ON users(email)").Error
    },
    Down: func(db *gorm.DB) error {
        return db.Exec("DROP INDEX idx_users_email").Error
    },
}
```

### Data Migrations

```go
migration := database.Migration{
    ID: "004_seed_admin_user",
    Up: func(db *gorm.DB) error {
        admin := User{
            Name:  "Admin",
            Email: "admin@example.com",
            Role:  "admin",
        }
        return db.Create(&admin).Error
    },
    Down: func(db *gorm.DB) error {
        return db.Where("email = ?", "admin@example.com").Delete(&User{}).Error
    },
}
```

## Running Migrations

### Using Manager

```go
package main

import (
    "log"
    "github.com/donnigundala/dg-database"
)

func main() {
    // Create manager
    config := database.DefaultConfig().
        WithDriver("mysql").
        WithDatabase("myapp")
    
    manager, err := database.NewManager(config, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer manager.Close()
    
    // Define migrations
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
        {
            ID: "002_create_posts",
            Up: func(db *gorm.DB) error {
                return db.AutoMigrate(&Post{})
            },
            Down: func(db *gorm.DB) error {
                return db.Migrator().DropTable(&Post{})
            },
        },
    }
    
    // Run migrations
    if err := manager.Migrate(migrations); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Migrations completed successfully")
}
```

### Using Migrator Directly

```go
migrator := database.NewMigrator(db)

// Add migrations
migrator.Add(migration1)
migrator.Add(migration2)
migrator.Add(migration3)

// Run all pending migrations
if err := migrator.Up(); err != nil {
    log.Fatal(err)
}
```

## Rolling Back

### Rollback Last Migration

```go
if err := manager.Rollback(migrations); err != nil {
    log.Fatal(err)
}
```

### Reset All Migrations

```go
migrator := database.NewMigrator(db)
for _, m := range migrations {
    migrator.Add(m)
}

if err := migrator.Reset(); err != nil {
    log.Fatal(err)
}
```

## Migration Status

### Check Applied Migrations

```go
status, err := manager.MigrationStatus()
if err != nil {
    log.Fatal(err)
}

fmt.Println("Applied migrations:")
for _, id := range status {
    fmt.Printf("  ✓ %s\n", id)
}
```

### Check Pending Migrations

```go
applied, _ := manager.MigrationStatus()
appliedMap := make(map[string]bool)
for _, id := range applied {
    appliedMap[id] = true
}

fmt.Println("Pending migrations:")
for _, migration := range migrations {
    if !appliedMap[migration.ID] {
        fmt.Printf("  ⏳ %s\n", migration.ID)
    }
}
```

## Best Practices

### 1. Use Descriptive IDs

```go
// Good
ID: "001_create_users_table"
ID: "002_add_user_email_index"
ID: "003_add_posts_table"

// Bad
ID: "migration1"
ID: "update"
```

### 2. Make Migrations Reversible

Always implement both `Up` and `Down` functions:

```go
migration := database.Migration{
    ID: "001_create_users",
    Up: func(db *gorm.DB) error {
        return db.AutoMigrate(&User{})
    },
    Down: func(db *gorm.DB) error {
        return db.Migrator().DropTable(&User{})
    },
}
```

### 3. Test Migrations

```go
func TestMigration(t *testing.T) {
    // Setup test database
    config := database.Config{
        Driver:   "sqlite",
        FilePath: ":memory:",
    }
    manager, _ := database.NewManager(config, nil)
    defer manager.Close()
    
    // Run migration
    err := manager.Migrate(migrations)
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify migration
    var count int64
    manager.DB().Model(&User{}).Count(&count)
    // Add assertions
}
```

### 4. Keep Migrations Small

Break large changes into smaller migrations:

```go
// Instead of one large migration
migrations := []database.Migration{
    {ID: "001_create_users", ...},
    {ID: "002_add_user_indexes", ...},
    {ID: "003_create_posts", ...},
    {ID: "004_add_post_indexes", ...},
}
```

### 5. Use Transactions

```go
migration := database.Migration{
    ID: "001_complex_migration",
    Up: func(db *gorm.DB) error {
        return db.Transaction(func(tx *gorm.DB) error {
            // Multiple operations
            if err := tx.AutoMigrate(&User{}).Error; err != nil {
                return err
            }
            if err := tx.AutoMigrate(&Post{}).Error; err != nil {
                return err
            }
            return nil
        })
    },
    Down: func(db *gorm.DB) error {
        return db.Transaction(func(tx *gorm.DB) error {
            if err := tx.Migrator().DropTable(&Post{}); err != nil {
                return err
            }
            if err := tx.Migrator().DropTable(&User{}); err != nil {
                return err
            }
            return nil
        })
    },
}
```

### 6. Version Control

Store migrations in version control:

```
migrations/
  001_create_users.go
  002_add_user_indexes.go
  003_create_posts.go
```

### 7. Idempotent Migrations

Make migrations safe to run multiple times:

```go
Up: func(db *gorm.DB) error {
    // Check if column exists before adding
    if !db.Migrator().HasColumn(&User{}, "phone") {
        return db.Migrator().AddColumn(&User{}, "phone")
    }
    return nil
},
```

## Common Patterns

### Renaming Columns

```go
migration := database.Migration{
    ID: "005_rename_user_name",
    Up: func(db *gorm.DB) error {
        return db.Migrator().RenameColumn(&User{}, "name", "full_name")
    },
    Down: func(db *gorm.DB) error {
        return db.Migrator().RenameColumn(&User{}, "full_name", "name")
    },
}
```

### Changing Column Types

```go
migration := database.Migration{
    ID: "006_change_user_age_type",
    Up: func(db *gorm.DB) error {
        return db.Exec("ALTER TABLE users MODIFY COLUMN age SMALLINT").Error
    },
    Down: func(db *gorm.DB) error {
        return db.Exec("ALTER TABLE users MODIFY COLUMN age INT").Error
    },
}
```

### Adding Foreign Keys

```go
migration := database.Migration{
    ID: "007_add_post_user_fk",
    Up: func(db *gorm.DB) error {
        return db.Exec(`
            ALTER TABLE posts 
            ADD CONSTRAINT fk_posts_user 
            FOREIGN KEY (user_id) REFERENCES users(id)
        `).Error
    },
    Down: func(db *gorm.DB) error {
        return db.Exec("ALTER TABLE posts DROP FOREIGN KEY fk_posts_user").Error
    },
}
```

## Troubleshooting

### Migration Failed

If a migration fails, it won't be recorded as applied. Fix the issue and run again.

### Stuck Migration

If a migration is partially applied:

1. Manually fix the database state
2. Mark migration as applied:
   ```sql
   INSERT INTO migrations (id) VALUES ('001_problematic_migration');
   ```

### Reset Everything

```go
// Drop all tables and re-run migrations
migrator := database.NewMigrator(db)
for _, m := range migrations {
    migrator.Add(m)
}
migrator.Reset()
migrator.Up()
```
