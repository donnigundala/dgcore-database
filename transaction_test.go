package database

import (
	"context"
	"os"
	"testing"

	"gorm.io/gorm"
)

type TestOrder struct {
	ID     uint `gorm:"primaryKey"`
	Amount float64
}

func TestWithTransaction_Success(t *testing.T) {
	dbFile := "test_tx_success.db"
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
	if err := manager.AutoMigrate(&TestOrder{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// Successful transaction
	err = WithTransaction(manager.DB(), func(tx *gorm.DB) error {
		order := TestOrder{Amount: 100.50}
		return tx.Create(&order).Error
	})

	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}

	// Verify order was created
	var count int64
	manager.DB().Model(&TestOrder{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 order, got %d", count)
	}
}

func TestWithTransaction_Rollback(t *testing.T) {
	dbFile := "test_tx_rollback.db"
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
	if err := manager.AutoMigrate(&TestOrder{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// Failed transaction
	err = WithTransaction(manager.DB(), func(tx *gorm.DB) error {
		order := TestOrder{Amount: 200.75}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		// Simulate error
		return gorm.ErrInvalidTransaction
	})

	if err == nil {
		t.Error("Expected transaction to fail")
	}

	// Verify rollback - no orders should exist
	var count int64
	manager.DB().Model(&TestOrder{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 orders after rollback, got %d", count)
	}
}

func TestWithTransactionContext(t *testing.T) {
	dbFile := "test_tx_context.db"
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
	if err := manager.AutoMigrate(&TestOrder{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	ctx := context.Background()

	err = WithTransactionContext(ctx, manager.DB(), func(tx *gorm.DB) error {
		order := TestOrder{Amount: 150.25}
		return tx.Create(&order).Error
	})

	if err != nil {
		t.Errorf("Transaction with context failed: %v", err)
	}

	var count int64
	manager.DB().Model(&TestOrder{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 order, got %d", count)
	}
}

func TestTransactionHelper(t *testing.T) {
	dbFile := "test_tx_helper.db"
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
	if err := manager.AutoMigrate(&TestOrder{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	helper := NewTransactionHelper(manager.DB())

	// Test Run
	err = helper.Run(func(tx *gorm.DB) error {
		order := TestOrder{Amount: 99.99}
		return tx.Create(&order).Error
	})

	if err != nil {
		t.Errorf("Helper.Run failed: %v", err)
	}

	var count int64
	manager.DB().Model(&TestOrder{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 order, got %d", count)
	}
}

func TestManager_TX(t *testing.T) {
	dbFile := "test_manager_tx.db"
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
	if err := manager.AutoMigrate(&TestOrder{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	helper := manager.TX()
	if helper == nil {
		t.Error("Expected TX() to return non-nil helper")
	}

	err = helper.Run(func(tx *gorm.DB) error {
		order := TestOrder{Amount: 75.50}
		return tx.Create(&order).Error
	})

	if err != nil {
		t.Errorf("TX helper failed: %v", err)
	}
}

func TestManager_WithTxContext(t *testing.T) {
	dbFile := "test_manager_tx_ctx.db"
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
	if err := manager.AutoMigrate(&TestOrder{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	ctx := context.Background()

	err = manager.WithTxContext(ctx, func(tx *gorm.DB) error {
		order := TestOrder{Amount: 125.00}
		return tx.Create(&order).Error
	})

	if err != nil {
		t.Errorf("WithTxContext failed: %v", err)
	}

	var count int64
	manager.DB().Model(&TestOrder{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 order, got %d", count)
	}
}

func TestBeginCommitRollback(t *testing.T) {
	dbFile := "test_begin_commit.db"
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
	if err := manager.AutoMigrate(&TestOrder{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// Test manual transaction
	tx := BeginTransaction(manager.DB())

	order := TestOrder{Amount: 50.00}
	if err := tx.Create(&order).Error; err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Commit
	if err := Commit(tx); err != nil {
		t.Errorf("Commit failed: %v", err)
	}

	// Verify committed
	var count int64
	manager.DB().Model(&TestOrder{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 order after commit, got %d", count)
	}

	// Test rollback
	tx = BeginTransaction(manager.DB())
	order2 := TestOrder{Amount: 75.00}
	tx.Create(&order2)

	// Rollback
	if err := Rollback(tx); err != nil {
		t.Errorf("Rollback failed: %v", err)
	}

	// Verify rollback - should still have only 1 order
	manager.DB().Model(&TestOrder{}).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 order after rollback, got %d", count)
	}
}
