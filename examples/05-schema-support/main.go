package main

import (
	"fmt"

	database "github.com/donnigundala/dg-database"
)

// User model
type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:100"`
	Email string `gorm:"size:100"`
}

func main() {
	fmt.Println("=== PostgreSQL Schema Support Example ===")

	// Example 1: Using custom schema
	config := database.DefaultConfig().
		WithDriver("postgres").
		WithHost("localhost").
		WithPort(5432).
		WithDatabase("myapp").
		WithCredentials("user", "password").
		WithSchema("tenant_1") // Use tenant_1 schema

	fmt.Println("\n1. Configuration with custom schema:")
	fmt.Printf("   Database: %s\n", config.Database)
	fmt.Printf("   Schema: %s\n", config.Schema)

	// Example 2: Multi-tenant with different schemas
	multiTenantConfig := database.Config{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "myapp",
		Username: "user",
		Password: "password",
		Schema:   "public", // Default schema

		// Each tenant gets their own schema
		Connections: map[string]database.ConnectionConfig{
			"tenant_1": {
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "myapp",
				Username: "user",
				Password: "password",
				Schema:   "tenant_1",
			},
			"tenant_2": {
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "myapp",
				Username: "user",
				Password: "password",
				Schema:   "tenant_2",
			},
		},
	}

	fmt.Println("\n2. Multi-tenant schema configuration:")
	fmt.Printf("   Default schema: %s\n", multiTenantConfig.Schema)
	for name, conn := range multiTenantConfig.Connections {
		fmt.Printf("   %s schema: %s\n", name, conn.Schema)
	}

	// Example 3: Usage with manager (commented out - requires actual PostgreSQL)
	/*
		manager, err := database.NewManager(config, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer manager.Close()

		// Auto-migrate in custom schema
		manager.AutoMigrate(&User{})

		// All operations will use the specified schema
		manager.DB().Create(&User{Name: "Alice", Email: "alice@example.com"})

		// For multi-tenant
		manager.Connection("tenant_1").Create(&User{Name: "Tenant 1 User"})
		manager.Connection("tenant_2").Create(&User{Name: "Tenant 2 User"})
	*/

	// Example 4: Schema-specific migrations
	fmt.Println("\n3. Schema-specific migrations:")
	fmt.Println("   Each tenant schema can have independent migrations")
	fmt.Println("   - tenant_1 schema: users, posts, comments")
	fmt.Println("   - tenant_2 schema: users, posts, comments")
	fmt.Println("   - public schema: shared configuration tables")

	// Example 5: Read/Write splitting with schemas
	rwConfig := database.Config{
		Driver:             "postgres",
		ReadWriteSplitting: true,
		AutoRouting:        true,

		Master: database.ConnectionConfig{
			Driver:   "postgres",
			Host:     "master.db.com",
			Port:     5432,
			Database: "myapp",
			Username: "user",
			Password: "password",
			Schema:   "tenant_1", // Master uses tenant_1 schema
		},

		Slaves: []database.ConnectionConfig{
			{
				Driver:   "postgres",
				Host:     "slave1.db.com",
				Port:     5432,
				Database: "myapp",
				Username: "user",
				Password: "password",
				Schema:   "tenant_1", // Slave also uses tenant_1 schema
			},
		},
	}

	fmt.Println("\n4. Read/Write splitting with schema:")
	fmt.Printf("   Master schema: %s\n", rwConfig.Master.Schema)
	fmt.Printf("   Slave schema: %s\n", rwConfig.Slaves[0].Schema)

	fmt.Println("\n=== Use Cases ===")
	fmt.Println("\n1. Multi-Tenancy:")
	fmt.Println("   - Each tenant gets their own schema")
	fmt.Println("   - Data isolation at database level")
	fmt.Println("   - Shared database, separate schemas")

	fmt.Println("\n2. Environment Separation:")
	fmt.Println("   - dev schema for development")
	fmt.Println("   - staging schema for staging")
	fmt.Println("   - prod schema for production")

	fmt.Println("\n3. Feature Isolation:")
	fmt.Println("   - core schema for core features")
	fmt.Println("   - analytics schema for analytics")
	fmt.Println("   - reporting schema for reports")

	fmt.Println("\n=== Example Complete ===")
}
