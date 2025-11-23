package main

import (
	"fmt"
	"log"

	database "github.com/donnigundala/dg-database"
	"gorm.io/gorm"
)

// User model
type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:100"`
	Email string `gorm:"size:100;uniqueIndex"`
}

func main() {
	fmt.Println("=== Basic Single Connection Example ===")

	// Create configuration
	config := database.DefaultConfig().
		WithDriver("sqlite").
		WithDatabase(":memory:").
		WithAutoMigrate(&User{})

	// Alternative: explicit configuration
	// config := database.Config{
	// 	Driver:   "sqlite",
	// 	Database: ":memory:",
	// 	AutoMigrate: true,
	// 	Models: []interface{}{&User{}},
	// }

	// Create manager
	manager, err := database.NewManager(config, nil)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Get database connection
	db := manager.DB()

	// Create users
	fmt.Println("Creating users...")
	users := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
		{Name: "Charlie", Email: "charlie@example.com"},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user: %v", err)
		} else {
			fmt.Printf("✅ Created user: %s (ID: %d)\n", user.Name, user.ID)
		}
	}

	// Query users
	fmt.Println("\nQuerying users...")
	var allUsers []User
	if err := db.Find(&allUsers).Error; err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}

	fmt.Printf("Found %d users:\n", len(allUsers))
	for _, user := range allUsers {
		fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
	}

	// Update user
	fmt.Println("\nUpdating user...")
	if err := db.Model(&User{}).Where("name = ?", "Alice").Update("email", "alice.new@example.com").Error; err != nil {
		log.Printf("Failed to update: %v", err)
	} else {
		fmt.Println("✅ Updated Alice's email")
	}

	// Delete user
	fmt.Println("\nDeleting user...")
	if err := db.Where("name = ?", "Bob").Delete(&User{}).Error; err != nil {
		log.Printf("Failed to delete: %v", err)
	} else {
		fmt.Println("✅ Deleted Bob")
	}

	// Count remaining users
	var count int64
	db.Model(&User{}).Count(&count)
	fmt.Printf("\nRemaining users: %d\n", count)

	// Transaction example
	fmt.Println("\n=== Transaction Example ===")
	err = manager.WithTx(func(tx *gorm.DB) error {
		// Create user in transaction
		newUser := User{Name: "David", Email: "david@example.com"}
		if err := tx.Create(&newUser).Error; err != nil {
			return err
		}
		fmt.Println("✅ Created David in transaction")

		// Simulate error to test rollback
		// return fmt.Errorf("simulated error")

		return nil
	})

	if err != nil {
		fmt.Printf("❌ Transaction failed: %v\n", err)
	} else {
		fmt.Println("✅ Transaction committed successfully")
	}

	// Final count
	db.Model(&User{}).Count(&count)
	fmt.Printf("\nFinal user count: %d\n", count)

	fmt.Println("\n=== Example Complete ===")
}
