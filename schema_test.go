package database

import (
	"strings"
	"testing"
)

func TestPostgresSchemaSupport(t *testing.T) {
	config := Config{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "myapp",
		Username: "user",
		Password: "pass",
		Schema:   "custom_schema",
	}

	dsn := buildPostgresDSN(config)

	if !strings.Contains(dsn, "search_path=custom_schema") {
		t.Errorf("Expected DSN to contain search_path=custom_schema, got: %s", dsn)
	}
}

func TestPostgresDefaultSchema(t *testing.T) {
	config := Config{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "myapp",
		Username: "user",
		Password: "pass",
		// No schema specified - should not add search_path
	}

	dsn := buildPostgresDSN(config)

	if strings.Contains(dsn, "search_path") {
		t.Errorf("Expected DSN to not contain search_path when schema not specified, got: %s", dsn)
	}
}

func TestPostgresSchemaFromConnectionConfig(t *testing.T) {
	config := ConnectionConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "myapp",
		Username: "user",
		Password: "pass",
		Schema:   "tenant_schema",
	}

	dsn := buildPostgresDSNFromConn(config)

	if !strings.Contains(dsn, "search_path=tenant_schema") {
		t.Errorf("Expected DSN to contain search_path=tenant_schema, got: %s", dsn)
	}
}

func TestConfigWithSchema(t *testing.T) {
	config := DefaultConfig().
		WithDriver("postgres").
		WithSchema("my_schema")

	if config.Schema != "my_schema" {
		t.Errorf("Expected schema 'my_schema', got '%s'", config.Schema)
	}
}
