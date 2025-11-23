package main

import (
	"fmt"
	"log"

	database "github.com/donnigundala/dg-database"
	"gorm.io/gorm"
)

// Product model
type Product struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:100"`
	Price float64
}

func main() {
	fmt.Println("=== Read/Write Splitting Example ===")

	// Configure read/write splitting
	config := database.Config{
		Driver:             "sqlite",
		Database:           ":memory:",
		ReadWriteSplitting: true,
		AutoRouting:        true, // Automatic routing enabled!
		SlaveStrategy:      "round-robin",

		// Master configuration
		Master: database.ConnectionConfig{
			Driver:   "sqlite",
			FilePath: ":memory:",
		},

		// Slave configurations
		Slaves: []database.ConnectionConfig{
			{Driver: "sqlite", FilePath: ":memory:", Weight: 2},
			{Driver: "sqlite", FilePath: ":memory:", Weight: 1},
		},

		AutoMigrate: true,
		Models:      []interface{}{&Product{}},
	}

	manager, err := database.NewManager(config, nil)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	fmt.Println("✅ Read/write splitting configured")
	fmt.Printf("   Master: 1 connection\n")
	fmt.Printf("   Slaves: %d connections\n", len(config.Slaves))
	fmt.Printf("   Strategy: %s\n", config.SlaveStrategy)
	fmt.Printf("   Auto-routing: %v\n\n", config.AutoRouting)

	// ========== Automatic Routing ==========
	fmt.Println("=== Automatic Routing (No Code Changes!) ===")

	db := manager.DB()

	// Writes automatically go to master
	fmt.Println("\n1. Creating products (writes → master)...")
	products := []Product{
		{Name: "Laptop", Price: 999.99},
		{Name: "Mouse", Price: 29.99},
		{Name: "Keyboard", Price: 79.99},
	}

	for _, product := range products {
		db.Create(&product)
		fmt.Printf("   ✅ Created: %s\n", product.Name)
	}

	// Reads automatically go to slaves
	fmt.Println("\n2. Reading products (reads → slaves)...")
	var allProducts []Product
	db.Find(&allProducts)
	fmt.Printf("   ✅ Found %d products (from slave)\n", len(allProducts))

	// ========== Manual Control ==========
	fmt.Println("\n=== Manual Control ===")

	// Force read from master
	fmt.Println("\n3. Force read from master...")
	var masterProducts []Product
	manager.Master().Find(&masterProducts)
	fmt.Printf("   ✅ Read %d products from master\n", len(masterProducts))

	// Explicit write
	fmt.Println("\n4. Explicit write to master...")
	newProduct := Product{Name: "Monitor", Price: 299.99}
	manager.Write().Create(&newProduct)
	fmt.Println("   ✅ Written to master")

	// Explicit read from slave
	fmt.Println("\n5. Explicit read from slave...")
	var slaveProducts []Product
	manager.Read().Find(&slaveProducts)
	fmt.Printf("   ✅ Read %d products from slave\n", len(slaveProducts))

	// Read from specific slave
	fmt.Println("\n6. Read from specific slave (index 0)...")
	var slave0Products []Product
	manager.Slave(0).Find(&slave0Products)
	fmt.Printf("   ✅ Read %d products from slave 0\n", len(slave0Products))

	// ========== Health Check ==========
	fmt.Println("\n=== Health Check ===")
	health := manager.HealthCheck()
	for name, status := range health {
		statusStr := "❌ DOWN"
		if status {
			statusStr = "✅ UP"
		}
		fmt.Printf("   %s: %s\n", name, statusStr)
	}

	// ========== Transactions (Always Use Master) ==========
	fmt.Println("\n=== Transactions (Always Master) ===")
	err = manager.WithTx(func(tx *gorm.DB) error {
		// All operations in transaction use master
		tx.Create(&Product{Name: "Headphones", Price: 149.99})
		fmt.Println("   ✅ Created product in transaction (master)")
		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Transaction failed: %v\n", err)
	}

	fmt.Println("\n=== Example Complete ===")
}
