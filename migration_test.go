package database

import (
	"os"
	"testing"

	"gorm.io/gorm"
)

type TestProduct struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:100"`
	Price float64
}

func TestMigration_UpDown(t *testing.T) {
	dbFile := "test_migration.db"
	defer func() { _ = os.Remove(dbFile) }()

	config := Config{
		Driver:   "sqlite",
		FilePath: dbFile,
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Define migrations
	migrations := []Migration{
		{
			ID: "001_create_products",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&TestProduct{})
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(&TestProduct{})
			},
		},
		{
			ID: "002_add_sample_data",
			Up: func(db *gorm.DB) error {
				products := []TestProduct{
					{Name: "Product 1", Price: 10.99},
					{Name: "Product 2", Price: 20.99},
				}
				return db.Create(&products).Error
			},
			Down: func(db *gorm.DB) error {
				return db.Where("1 = 1").Delete(&TestProduct{}).Error
			},
		},
	}

	// Run migrations
	if err := manager.Migrate(migrations); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify data exists
	var count int64
	manager.DB().Model(&TestProduct{}).Count(&count)
	if count != 2 {
		t.Errorf("Expected 2 products, got %d", count)
	}

	// Check migration status
	status, err := manager.MigrationStatus()
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	if len(status) != 2 {
		t.Errorf("Expected 2 migrations, got %d", len(status))
	}

	// Rollback last migration
	if err := manager.Rollback(migrations); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify data removed
	manager.DB().Model(&TestProduct{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 products after rollback, got %d", count)
	}

	// Check migration status after rollback
	status, err = manager.MigrationStatus()
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	if len(status) != 1 {
		t.Errorf("Expected 1 migration after rollback, got %d", len(status))
	}
}

func TestMigrator_Reset(t *testing.T) {
	dbFile := "test_migrator_reset.db"
	defer func() { _ = os.Remove(dbFile) }()

	config := Config{
		Driver:   "sqlite",
		FilePath: dbFile,
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	migrator := NewMigrator(manager.DB())

	// Add migrations
	migrator.Add(Migration{
		ID: "001_test",
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(&TestProduct{})
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(&TestProduct{})
		},
	})

	// Run migrations
	if err := migrator.Up(); err != nil {
		t.Fatalf("Migration up failed: %v", err)
	}

	// Reset
	if err := migrator.Reset(); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	// Verify all migrations removed
	status, err := migrator.Status()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if len(status) != 0 {
		t.Errorf("Expected 0 migrations after reset, got %d", len(status))
	}
}

func TestMigrator_IdempotentUp(t *testing.T) {
	dbFile := "test_migrator_idempotent.db"
	defer func() { _ = os.Remove(dbFile) }()

	config := Config{
		Driver:   "sqlite",
		FilePath: dbFile,
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	migrations := []Migration{
		{
			ID: "001_test",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&TestProduct{})
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(&TestProduct{})
			},
		},
	}

	// Run migrations first time
	if err := manager.Migrate(migrations); err != nil {
		t.Fatalf("First migration failed: %v", err)
	}

	// Run migrations second time (should be idempotent)
	if err := manager.Migrate(migrations); err != nil {
		t.Fatalf("Second migration failed: %v", err)
	}

	// Should still have only 1 migration record
	status, err := manager.MigrationStatus()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if len(status) != 1 {
		t.Errorf("Expected 1 migration, got %d", len(status))
	}
}
