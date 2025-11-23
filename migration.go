package database

import (
	"fmt"

	"gorm.io/gorm"
)

// Migration represents a database migration.
type Migration struct {
	ID   string
	Up   func(*gorm.DB) error
	Down func(*gorm.DB) error
}

// Migrator handles database migrations.
type Migrator struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrator creates a new migrator.
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// Add adds a migration.
func (m *Migrator) Add(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// Up runs all pending migrations.
func (m *Migrator) Up() error {
	// Create migrations table if not exists
	if err := m.ensureMigrationsTable(); err != nil {
		return err
	}

	for _, migration := range m.migrations {
		// Check if already migrated
		if m.isMigrated(migration.ID) {
			continue
		}

		// Run migration
		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.ID, err)
		}

		// Record migration
		if err := m.recordMigration(migration.ID); err != nil {
			return err
		}
	}

	return nil
}

// Down rolls back the last migration.
func (m *Migrator) Down() error {
	// Get last migration
	var lastMigration string
	if err := m.db.Table("migrations").
		Order("id DESC").
		Limit(1).
		Pluck("id", &lastMigration).Error; err != nil {
		return err
	}

	if lastMigration == "" {
		return fmt.Errorf("no migrations to rollback")
	}

	// Find and run down migration
	for _, migration := range m.migrations {
		if migration.ID == lastMigration {
			if err := migration.Down(m.db); err != nil {
				return fmt.Errorf("rollback %s failed: %w", migration.ID, err)
			}

			// Remove migration record
			return m.db.Exec("DELETE FROM migrations WHERE id = ?", migration.ID).Error
		}
	}

	return fmt.Errorf("migration %s not found", lastMigration)
}

// Reset rolls back all migrations.
func (m *Migrator) Reset() error {
	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		if m.isMigrated(migration.ID) {
			if err := migration.Down(m.db); err != nil {
				return fmt.Errorf("rollback %s failed: %w", migration.ID, err)
			}
			m.db.Exec("DELETE FROM migrations WHERE id = ?", migration.ID)
		}
	}
	return nil
}

// Status returns migration status.
func (m *Migrator) Status() ([]string, error) {
	var migrated []string
	if err := m.db.Table("migrations").Pluck("id", &migrated).Error; err != nil {
		return nil, err
	}
	return migrated, nil
}

func (m *Migrator) ensureMigrationsTable() error {
	// Use GORM to create migrations table
	type Migration struct {
		ID         string `gorm:"primaryKey;size:255"`
		MigratedAt int64  `gorm:"autoCreateTime"`
	}
	return m.db.AutoMigrate(&Migration{})
}

func (m *Migrator) isMigrated(id string) bool {
	// Check if migrations table exists
	if !m.db.Migrator().HasTable("migrations") {
		return false
	}

	var count int64
	m.db.Table("migrations").Where("id = ?", id).Count(&count)
	return count > 0
}

func (m *Migrator) recordMigration(id string) error {
	type Migration struct {
		ID string `gorm:"primaryKey;size:255"`
	}
	return m.db.Table("migrations").Create(&Migration{ID: id}).Error
}

// Helper functions for Manager

// Migrate runs migrations using the manager's primary connection.
func (m *Manager) Migrate(migrations []Migration) error {
	migrator := NewMigrator(m.db)
	for _, migration := range migrations {
		migrator.Add(migration)
	}
	return migrator.Up()
}

// Rollback rolls back the last migration.
func (m *Manager) Rollback(migrations []Migration) error {
	migrator := NewMigrator(m.master)
	for _, migration := range migrations {
		migrator.Add(migration)
	}
	return migrator.Down()
}

// MigrationStatus returns the status of migrations.
func (m *Manager) MigrationStatus() ([]string, error) {
	migrator := NewMigrator(m.db)
	return migrator.Status()
}
