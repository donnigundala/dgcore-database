package database

import (
	"strings"
	"testing"
)

func TestBuildMySQLDSN(t *testing.T) {
	config := Config{
		Driver:    "mysql",
		Host:      "localhost",
		Port:      3306,
		Database:  "testdb",
		Username:  "root",
		Password:  "secret",
		Charset:   "utf8mb4",
		ParseTime: true,
	}

	dsn := buildMySQLDSN(config)

	expected := "root:secret@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != expected {
		t.Errorf("Expected DSN:\n%s\nGot:\n%s", expected, dsn)
	}
}

func TestBuildMySQLDSN_Defaults(t *testing.T) {
	config := Config{
		Driver:    "mysql",
		Host:      "localhost",
		Port:      3306,
		Database:  "testdb",
		Username:  "root",
		Password:  "secret",
		ParseTime: true, // Explicitly set to true
	}

	dsn := buildMySQLDSN(config)

	// Should use default charset
	if !strings.Contains(dsn, "charset=utf8mb4") {
		t.Error("Expected default charset utf8mb4")
	}

	// Should use parseTime=True when set
	if !strings.Contains(dsn, "parseTime=True") {
		t.Error("Expected parseTime=True")
	}
}

func TestBuildPostgresDSN(t *testing.T) {
	config := Config{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		Password: "secret",
		SSLMode:  "disable",
		Timezone: "UTC",
	}

	dsn := buildPostgresDSN(config)

	expected := "host=localhost user=postgres password=secret dbname=testdb port=5432 sslmode=disable TimeZone=UTC"
	if dsn != expected {
		t.Errorf("Expected DSN:\n%s\nGot:\n%s", expected, dsn)
	}
}

func TestBuildPostgresDSN_Defaults(t *testing.T) {
	config := Config{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		Password: "secret",
	}

	dsn := buildPostgresDSN(config)

	// Should use default sslmode
	if !strings.Contains(dsn, "sslmode=disable") {
		t.Error("Expected default sslmode=disable")
	}

	// Should use default timezone
	if !strings.Contains(dsn, "TimeZone=UTC") {
		t.Error("Expected default TimeZone=UTC")
	}
}

func TestBuildDSN_SQLite(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		FilePath: "/tmp/test.db",
	}

	dsn := buildDSN(config)

	if dsn != "/tmp/test.db" {
		t.Errorf("Expected DSN '/tmp/test.db', got '%s'", dsn)
	}
}

func TestBuildDSNFromConnectionConfig_MySQL(t *testing.T) {
	config := ConnectionConfig{
		Driver:   "mysql",
		Host:     "db.example.com",
		Port:     3306,
		Database: "mydb",
		Username: "user",
		Password: "pass",
	}

	dsn := buildDSNFromConnectionConfig(config)

	if !strings.Contains(dsn, "user:pass@tcp(db.example.com:3306)/mydb") {
		t.Errorf("Unexpected DSN: %s", dsn)
	}
}

func TestBuildDSNFromConnectionConfig_Postgres(t *testing.T) {
	config := ConnectionConfig{
		Driver:   "postgres",
		Host:     "pg.example.com",
		Port:     5432,
		Database: "pgdb",
		Username: "pguser",
		Password: "pgpass",
	}

	dsn := buildDSNFromConnectionConfig(config)

	if !strings.Contains(dsn, "host=pg.example.com") {
		t.Error("Expected host in DSN")
	}
	if !strings.Contains(dsn, "user=pguser") {
		t.Error("Expected user in DSN")
	}
	if !strings.Contains(dsn, "dbname=pgdb") {
		t.Error("Expected dbname in DSN")
	}
}

func TestConnect_SQLite(t *testing.T) {
	config := Config{
		Driver:   "sqlite",
		FilePath: ":memory:",
	}

	db, err := connect(config, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if db == nil {
		t.Error("Expected db to be non-nil")
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get SQL DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestConnectWithConfig_SQLite(t *testing.T) {
	config := ConnectionConfig{
		Driver:   "sqlite",
		FilePath: ":memory:",
	}

	db, err := connectWithConfig(config, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if db == nil {
		t.Error("Expected db to be non-nil")
	}
}

func TestConnect_UnsupportedDriver(t *testing.T) {
	config := Config{
		Driver: "unsupported",
	}

	_, err := connect(config, nil)
	if err == nil {
		t.Error("Expected error for unsupported driver")
	}

	if !strings.Contains(err.Error(), "unsupported driver") {
		t.Errorf("Expected 'unsupported driver' error, got: %v", err)
	}
}
