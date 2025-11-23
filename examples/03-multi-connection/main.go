package main

import (
	"fmt"
	"log"

	database "github.com/donnigundala/dg-database"
)

// Order model
type Order struct {
	ID         uint   `gorm:"primaryKey"`
	CustomerID string `gorm:"size:50"`
	Total      float64
}

// AnalyticsEvent model
type AnalyticsEvent struct {
	ID    uint   `gorm:"primaryKey"`
	Event string `gorm:"size:100"`
	Data  string `gorm:"type:text"`
}

// Log model
type Log struct {
	ID      uint   `gorm:"primaryKey"`
	Level   string `gorm:"size:20"`
	Message string `gorm:"type:text"`
}

func main() {
	fmt.Println("=== Multi-Connection Example ===")

	// Configure multiple named connections
	config := database.Config{
		Driver:   "sqlite",
		Database: ":memory:",

		// Named connections
		Connections: map[string]database.ConnectionConfig{
			"analytics": {
				Driver:   "sqlite",
				FilePath: ":memory:",
			},
			"logs": {
				Driver:   "sqlite",
				FilePath: ":memory:",
			},
		},

		AutoMigrate: true,
		Models:      []interface{}{&Order{}},
	}

	manager, err := database.NewManager(config, nil)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	fmt.Println("✅ Multi-connection configured")
	fmt.Printf("   Primary: orders database\n")
	fmt.Printf("   Analytics: events database\n")
	fmt.Printf("   Logs: logs database\n\n")

	// Migrate other databases
	manager.Connection("analytics").AutoMigrate(&AnalyticsEvent{})
	manager.Connection("logs").AutoMigrate(&Log{})

	// ========== Use Primary Database ==========
	fmt.Println("=== Primary Database (Orders) ===")
	primaryDB := manager.DB()

	orders := []Order{
		{CustomerID: "CUST001", Total: 299.99},
		{CustomerID: "CUST002", Total: 149.99},
	}

	for _, order := range orders {
		primaryDB.Create(&order)
		fmt.Printf("✅ Created order for %s: $%.2f\n", order.CustomerID, order.Total)
	}

	// ========== Use Analytics Database ==========
	fmt.Println("\n=== Analytics Database (Events) ===")
	analyticsDB := manager.Connection("analytics")

	events := []AnalyticsEvent{
		{Event: "page_view", Data: "/products"},
		{Event: "button_click", Data: "checkout"},
	}

	for _, event := range events {
		analyticsDB.Create(&event)
		fmt.Printf("✅ Logged event: %s\n", event.Event)
	}

	// ========== Use Logs Database ==========
	fmt.Println("\n=== Logs Database ===")
	logsDB := manager.Connection("logs")

	logs := []Log{
		{Level: "INFO", Message: "Application started"},
		{Level: "WARN", Message: "High memory usage"},
	}

	for _, logEntry := range logs {
		logsDB.Create(&logEntry)
		fmt.Printf("✅ Logged: [%s] %s\n", logEntry.Level, logEntry.Message)
	}

	// ========== Runtime Connection Management ==========
	fmt.Println("\n=== Runtime Connection Management ===")

	// Add new connection at runtime
	fmt.Println("\n1. Adding 'cache' connection...")
	cacheConfig := database.ConnectionConfig{
		Driver:   "sqlite",
		FilePath: ":memory:",
	}

	if err := manager.AddConnection("cache", cacheConfig); err != nil {
		log.Printf("Failed to add connection: %v", err)
	} else {
		fmt.Println("   ✅ Cache connection added")
	}

	// Check if connection exists
	fmt.Println("\n2. Checking connections...")
	connections := []string{"analytics", "logs", "cache", "nonexistent"}
	for _, name := range connections {
		exists := manager.HasConnection(name)
		status := "❌"
		if exists {
			status = "✅"
		}
		fmt.Printf("   %s Connection '%s': %v\n", status, name, exists)
	}

	// Remove connection
	fmt.Println("\n3. Removing 'cache' connection...")
	if err := manager.RemoveConnection("cache"); err != nil {
		log.Printf("Failed to remove connection: %v", err)
	} else {
		fmt.Println("   ✅ Cache connection removed")
	}

	// ========== Query from Different Databases ==========
	fmt.Println("\n=== Query Summary ===")

	var orderCount, eventCount, logCount int64
	manager.DB().Model(&Order{}).Count(&orderCount)
	manager.Connection("analytics").Model(&AnalyticsEvent{}).Count(&eventCount)
	manager.Connection("logs").Model(&Log{}).Count(&logCount)

	fmt.Printf("Orders: %d\n", orderCount)
	fmt.Printf("Events: %d\n", eventCount)
	fmt.Printf("Logs: %d\n", logCount)

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

	fmt.Println("\n=== Example Complete ===")
}
