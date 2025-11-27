package database

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"gorm.io/gorm"
)

// Test models
type TestUser struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:100"`
	Email string `gorm:"size:100"`
}

func TestNewManager_SQLite(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		FilePath: ":memory:",
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	if manager.db == nil {
		t.Error("Expected db to be initialized")
	}

	// Test ping
	if err := manager.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestManager_CRUD(t *testing.T) {
	dbFile := "test_crud.db"
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

	// Explicitly migrate
	if err := manager.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	db := manager.DB()

	// Create
	user := TestUser{Name: "Alice", Email: "alice@example.com"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set")
	}

	// Read
	var foundUser TestUser
	if err := db.First(&foundUser, user.ID).Error; err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}

	if foundUser.Name != "Alice" {
		t.Errorf("Expected name 'Alice', got '%s'", foundUser.Name)
	}

	// Update
	if err := db.Model(&foundUser).Update("email", "alice.new@example.com").Error; err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify update
	db.First(&foundUser, user.ID)
	if foundUser.Email != "alice.new@example.com" {
		t.Errorf("Expected email 'alice.new@example.com', got '%s'", foundUser.Email)
	}

	// Delete
	if err := db.Delete(&foundUser).Error; err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify delete
	var count int64
	db.Model(&TestUser{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 users, got %d", count)
	}
}

func TestManager_Transaction(t *testing.T) {
	dbFile := "test_transaction.db"
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

	// Explicitly migrate
	if err := manager.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// Successful transaction
	err = manager.WithTx(func(tx *gorm.DB) error {
		user := TestUser{Name: "Bob", Email: "bob@example.com"}
		return tx.Create(&user).Error
	})

	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}

	// Verify user was created
	var count int64
	manager.DB().Model(&TestUser{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 user, got %d", count)
	}

	// Failed transaction (should rollback)
	err = manager.WithTx(func(tx *gorm.DB) error {
		user := TestUser{Name: "Charlie", Email: "charlie@example.com"}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		// Simulate error
		return gorm.ErrInvalidTransaction
	})

	if err == nil {
		t.Error("Expected transaction to fail")
	}

	// Verify rollback - should still have only 1 user
	manager.DB().Model(&TestUser{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 user after rollback, got %d", count)
	}
}

func TestManager_HealthCheck(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		FilePath: ":memory:",
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	health := manager.HealthCheck()

	if !health["primary"] {
		t.Error("Expected primary connection to be healthy")
	}
}

func TestManager_AutoMigrate(t *testing.T) {
	dbFile := "test_automigrate.db"
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

	// Auto migrate
	if err := manager.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// Verify table exists by creating a record
	user := TestUser{Name: "Test", Email: "test@example.com"}
	if err := manager.DB().Create(&user).Error; err != nil {
		t.Errorf("Failed to create user after migration: %v", err)
	}
}

func TestManager_MultiConnection(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		FilePath: ":memory:",
		Connections: map[string]ConnectionConfig{
			"analytics": {
				Driver:   "sqlite",
				FilePath: ":memory:",
			},
		},
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Check connection exists
	if !manager.HasConnection("analytics") {
		t.Error("Expected 'analytics' connection to exist")
	}

	// Get connection
	analyticsDB := manager.Connection("analytics")
	if analyticsDB == nil {
		t.Error("Expected analytics connection to be non-nil")
	}

	// Test connection
	if err := manager.ping(analyticsDB); err != nil {
		t.Errorf("Analytics connection ping failed: %v", err)
	}
}

func TestManager_AddRemoveConnection(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		FilePath: ":memory:",
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Add connection
	cacheConfig := ConnectionConfig{
		Driver:   "sqlite",
		FilePath: ":memory:",
	}

	if err := manager.AddConnection("cache", cacheConfig); err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	// Verify connection exists
	if !manager.HasConnection("cache") {
		t.Error("Expected 'cache' connection to exist after adding")
	}

	// Remove connection
	if err := manager.RemoveConnection("cache"); err != nil {
		t.Fatalf("Failed to remove connection: %v", err)
	}

	// Verify connection removed
	if manager.HasConnection("cache") {
		t.Error("Expected 'cache' connection to not exist after removal")
	}
}

func TestManager_Close(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		FilePath: ":memory:",
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Close should not error
	if err := manager.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		FilePath: ":memory:",
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	var wg sync.WaitGroup
	workers := 20

	// Concurrent GetConnection
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Should not panic or race
			_ = manager.Connection("default")
			_ = manager.HasConnection("default")
		}()
	}

	// Concurrent Add/Remove Connection
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("conn_%d", id)
			connConfig := ConnectionConfig{
				Driver:   "sqlite",
				FilePath: ":memory:",
			}

			// Add
			if err := manager.AddConnection(name, connConfig); err != nil {
				t.Errorf("Failed to add connection %s: %v", name, err)
				return
			}

			// Check
			if !manager.HasConnection(name) {
				t.Errorf("Connection %s should exist", name)
			}

			// Remove
			if err := manager.RemoveConnection(name); err != nil {
				t.Errorf("Failed to remove connection %s: %v", name, err)
			}
		}(i)
	}

	wg.Wait()
}

func TestManager_ConcurrentTransactions(t *testing.T) {
	dbFile := "test_concurrent_tx.db"
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

	// Migrate
	if err := manager.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	var wg sync.WaitGroup
	workers := 10

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := manager.WithTx(func(tx *gorm.DB) error {
				user := TestUser{
					Name:  fmt.Sprintf("User_%d", id),
					Email: fmt.Sprintf("user%d@example.com", id),
				}
				return tx.Create(&user).Error
			})
			if err != nil {
				t.Errorf("Transaction failed for worker %d: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify count
	var count int64
	manager.DB().Model(&TestUser{}).Count(&count)
	if count != int64(workers) {
		t.Errorf("Expected %d users, got %d", workers, count)
	}
}
