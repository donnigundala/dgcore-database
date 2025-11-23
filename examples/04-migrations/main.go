package main

import (
	"fmt"
	"log"

	database "github.com/donnigundala/dg-database"
	"gorm.io/gorm"
)

// Initial models
type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:100"`
	Email string `gorm:"size:100"`
}

func main() {
	fmt.Println("=== Migration Example ===")

	config := database.Config{
		Driver:   "sqlite",
		FilePath: "test.db",
	}

	manager, err := database.NewManager(config, nil)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Define migrations
	migrations := []database.Migration{
		{
			ID: "001_create_users_table",
			Up: func(db *gorm.DB) error {
				fmt.Println("Running migration: 001_create_users_table")
				return db.AutoMigrate(&User{})
			},
			Down: func(db *gorm.DB) error {
				fmt.Println("Rolling back: 001_create_users_table")
				return db.Migrator().DropTable(&User{})
			},
		},
		{
			ID: "002_add_users_data",
			Up: func(db *gorm.DB) error {
				fmt.Println("Running migration: 002_add_users_data")
				users := []User{
					{Name: "Alice", Email: "alice@example.com"},
					{Name: "Bob", Email: "bob@example.com"},
				}
				return db.Create(&users).Error
			},
			Down: func(db *gorm.DB) error {
				fmt.Println("Rolling back: 002_add_users_data")
				return db.Where("1 = 1").Delete(&User{}).Error
			},
		},
	}

	// Run migrations
	fmt.Println("=== Running Migrations ===")
	if err := manager.Migrate(migrations); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("✅ Migrations completed")

	// Check migration status
	fmt.Println("=== Migration Status ===")
	status, err := manager.MigrationStatus()
	if err != nil {
		log.Printf("Failed to get status: %v", err)
	} else {
		fmt.Println("Completed migrations:")
		for _, id := range status {
			fmt.Printf("  ✅ %s\n", id)
		}
	}

	// Query data
	fmt.Println("\n=== Query Data ===")
	var users []User
	manager.DB().Find(&users)
	fmt.Printf("Found %d users:\n", len(users))
	for _, user := range users {
		fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
	}

	// Rollback example (commented out to preserve data)
	// fmt.Println("\n=== Rolling Back Last Migration ===")
	// if err := manager.Rollback(); err != nil {
	// 	log.Printf("Rollback failed: %v", err)
	// } else {
	// 	fmt.Println("✅ Rollback completed")
	// }

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("Database file: test.db")
}
