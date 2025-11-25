package database

import (
	"fmt"

	"github.com/donnigundala/dg-core/contracts/foundation"
)

// ServiceProvider is the database service provider
type ServiceProvider struct {
	config Config
}

// NewServiceProvider creates a new database service provider
func NewServiceProvider(config Config) *ServiceProvider {
	return &ServiceProvider{config: config}
}

// Name returns the provider name
func (p *ServiceProvider) Name() string {
	return "database"
}

// Version returns the provider version
func (p *ServiceProvider) Version() string {
	return "1.0.0"
}

// Dependencies returns the provider dependencies
func (p *ServiceProvider) Dependencies() []string {
	return []string{}
}

// Register registers the database services
func (p *ServiceProvider) Register(app foundation.Application) error {
	// Register database manager
	app.Singleton("db", func() interface{} {
		// Note: Logger should be configured via Config.Logger field
		// Calling app.Make("logger") here causes deadlock
		manager, err := NewManager(p.config, nil)
		if err != nil {
			panic(fmt.Sprintf("failed to create database manager: %v", err))
		}
		return manager
	})

	// Note: We don't register "gorm" here because calling Make("db") inside
	// a resolver causes deadlock. Users can get GORM via manager.DB() directly.

	return nil
}

// Boot boots the database services
func (p *ServiceProvider) Boot(app foundation.Application) error {
	// Get manager
	managerInstance, err := app.Make("db")
	if err != nil {
		return fmt.Errorf("failed to get database manager: %w", err)
	}
	manager := managerInstance.(*Manager)

	// Test connection
	if err := manager.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// Log success if logger available
	if manager.logger != nil {
		manager.logInfo("Database connected successfully",
			"driver", p.config.Driver,
			"database", p.config.Database)

		if p.config.ReadWriteSplitting {
			manager.logInfo("Read/write splitting enabled",
				"slaves", len(p.config.Slaves),
				"strategy", p.config.SlaveStrategy)
		}

		if len(p.config.Connections) > 0 {
			manager.logInfo("Named connections established",
				"count", len(p.config.Connections))
		}
	}

	return nil
}
