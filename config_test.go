package database

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Driver != "mysql" {
		t.Errorf("Expected driver 'mysql', got '%s'", config.Driver)
	}

	if config.Port != 3306 {
		t.Errorf("Expected port 3306, got %d", config.Port)
	}

	if config.MaxOpenConns != 100 {
		t.Errorf("Expected MaxOpenConns 100, got %d", config.MaxOpenConns)
	}

	if config.AutoRouting != true {
		t.Error("Expected AutoRouting to be true")
	}
}

func TestConfigFluentSetters(t *testing.T) {
	config := DefaultConfig().
		WithDriver("postgres").
		WithHost("localhost").
		WithPort(5432).
		WithDatabase("testdb").
		WithCredentials("user", "pass")

	if config.Driver != "postgres" {
		t.Errorf("Expected driver 'postgres', got '%s'", config.Driver)
	}

	if config.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", config.Host)
	}

	if config.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", config.Port)
	}

	if config.Database != "testdb" {
		t.Errorf("Expected database 'testdb', got '%s'", config.Database)
	}

	if config.Username != "user" {
		t.Errorf("Expected username 'user', got '%s'", config.Username)
	}

	if config.Password != "pass" {
		t.Errorf("Expected password 'pass', got '%s'", config.Password)
	}
}

func TestConfigReadWriteSplitting(t *testing.T) {
	masterConfig := ConnectionConfig{
		Driver:   "mysql",
		Host:     "master.db.com",
		Port:     3306,
		Database: "mydb",
	}

	slaveConfig := ConnectionConfig{
		Driver:   "mysql",
		Host:     "slave.db.com",
		Port:     3306,
		Database: "mydb",
	}

	config := DefaultConfig().
		WithReadWriteSplitting(masterConfig, slaveConfig)

	if !config.ReadWriteSplitting {
		t.Error("Expected ReadWriteSplitting to be true")
	}

	if config.Master.Host != "master.db.com" {
		t.Errorf("Expected master host 'master.db.com', got '%s'", config.Master.Host)
	}

	if len(config.Slaves) != 1 {
		t.Errorf("Expected 1 slave, got %d", len(config.Slaves))
	}

	if config.Slaves[0].Host != "slave.db.com" {
		t.Errorf("Expected slave host 'slave.db.com', got '%s'", config.Slaves[0].Host)
	}
}

func TestConfigMultiConnection(t *testing.T) {
	analyticsConfig := ConnectionConfig{
		Driver:   "postgres",
		Host:     "analytics.db.com",
		Database: "analytics",
	}

	config := DefaultConfig().
		WithConnection("analytics", analyticsConfig).
		WithDefaultConnection("primary")

	if len(config.Connections) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(config.Connections))
	}

	conn, exists := config.Connections["analytics"]
	if !exists {
		t.Error("Expected 'analytics' connection to exist")
	}

	if conn.Host != "analytics.db.com" {
		t.Errorf("Expected analytics host 'analytics.db.com', got '%s'", conn.Host)
	}

	if config.DefaultConnection != "primary" {
		t.Errorf("Expected default connection 'primary', got '%s'", config.DefaultConnection)
	}
}

func TestConfigAutoMigrate(t *testing.T) {
	type User struct {
		ID   uint
		Name string
	}

	config := DefaultConfig().
		WithAutoMigrate(&User{})

	if !config.AutoMigrate {
		t.Error("Expected AutoMigrate to be true")
	}

	if len(config.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(config.Models))
	}
}

func TestConfigSlaveStrategy(t *testing.T) {
	config := DefaultConfig().
		WithSlaveStrategy("weighted")

	if config.SlaveStrategy != "weighted" {
		t.Errorf("Expected strategy 'weighted', got '%s'", config.SlaveStrategy)
	}
}

func TestConfigAutoRouting(t *testing.T) {
	config := DefaultConfig().
		WithAutoRouting(false)

	if config.AutoRouting {
		t.Error("Expected AutoRouting to be false")
	}
}

func TestConnectionConfigPoolSettings(t *testing.T) {
	maxOpen := 50
	maxIdle := 10
	lifetime := time.Hour

	config := ConnectionConfig{
		Driver:          "mysql",
		Host:            "localhost",
		MaxOpenConns:    &maxOpen,
		MaxIdleConns:    &maxIdle,
		ConnMaxLifetime: &lifetime,
	}

	if *config.MaxOpenConns != 50 {
		t.Errorf("Expected MaxOpenConns 50, got %d", *config.MaxOpenConns)
	}

	if *config.MaxIdleConns != 10 {
		t.Errorf("Expected MaxIdleConns 10, got %d", *config.MaxIdleConns)
	}

	if *config.ConnMaxLifetime != time.Hour {
		t.Errorf("Expected ConnMaxLifetime 1h, got %v", *config.ConnMaxLifetime)
	}
}
