package database

import (
	"os"
	"testing"
)

func TestManager_SQL(t *testing.T) {
	dbFile := "test_sql_helper.db"
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

	// Get SQL DB
	sqlDB, err := manager.SQL()
	if err != nil {
		t.Fatalf("Failed to get SQL DB: %v", err)
	}

	if sqlDB == nil {
		t.Error("Expected SQL DB to be non-nil")
	}

	// Test ping
	if err := sqlDB.Ping(); err != nil {
		t.Errorf("SQL DB ping failed: %v", err)
	}
}

func TestManager_SQLConnection(t *testing.T) {
	dbFile1 := "test_sql_conn1.db"
	dbFile2 := "test_sql_conn2.db"
	defer func() {
		_ = os.Remove(dbFile1)
		_ = os.Remove(dbFile2)
	}()

	config := Config{
		Driver:   "sqlite",
		FilePath: dbFile1,
		Connections: map[string]ConnectionConfig{
			"secondary": {
				Driver:   "sqlite",
				FilePath: dbFile2,
			},
		},
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Get SQL DB for named connection
	sqlDB, err := manager.SQLConnection("secondary")
	if err != nil {
		t.Fatalf("Failed to get SQL DB for connection: %v", err)
	}

	if sqlDB == nil {
		t.Error("Expected SQL DB to be non-nil")
	}

	// Test ping
	if err := sqlDB.Ping(); err != nil {
		t.Errorf("SQL DB ping failed: %v", err)
	}
}
