package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// connect creates a database connection from the main config.
func connect(config Config, log interface{}) (*gorm.DB, error) {
	// Use retry logic if enabled
	if config.Retry.Enabled {
		// Convert Config to ConnectionConfig for retry
		connConfig := ConnectionConfig{
			Driver:    config.Driver,
			Host:      config.Host,
			Port:      config.Port,
			Database:  config.Database,
			Username:  config.Username,
			Password:  config.Password,
			FilePath:  config.FilePath,
			Charset:   config.Charset,
			Timezone:  config.Timezone,
			ParseTime: config.ParseTime,
			SSLMode:   config.SSLMode,
			Schema:    config.Schema,
		}
		return connectWithRetry(connConfig, config.Retry, log)
	}

	// No retry - use direct connection
	// Build DSN
	dsn := buildDSN(config)

	// Select driver
	var dialector gorm.Dialector
	switch config.Driver {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(config.FilePath)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", config.Driver)
	}

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: getLogger(config.LogLevel, config.SlowThreshold),
	}

	// Open connection
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// connectWithConfig creates a connection from ConnectionConfig.
func connectWithConfig(config ConnectionConfig, log interface{}) (*gorm.DB, error) {
	// Build DSN
	dsn := buildDSNFromConnectionConfig(config)

	// Select driver
	var dialector gorm.Dialector
	switch config.Driver {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(config.FilePath)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", config.Driver)
	}

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// Open connection
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, err
	}

	// Configure connection pool if specified
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if config.MaxOpenConns != nil {
		sqlDB.SetMaxOpenConns(*config.MaxOpenConns)
	}
	if config.MaxIdleConns != nil {
		sqlDB.SetMaxIdleConns(*config.MaxIdleConns)
	}
	if config.ConnMaxLifetime != nil {
		sqlDB.SetConnMaxLifetime(*config.ConnMaxLifetime)
	}

	return db, nil
}

// buildDSN builds a DSN string from Config.
func buildDSN(config Config) string {
	switch config.Driver {
	case "mysql":
		return buildMySQLDSN(config)
	case "postgres":
		return buildPostgresDSN(config)
	case "sqlite":
		return config.FilePath
	default:
		return ""
	}
}

// buildDSNFromConnectionConfig builds a DSN from ConnectionConfig.
func buildDSNFromConnectionConfig(config ConnectionConfig) string {
	switch config.Driver {
	case "mysql":
		return buildMySQLDSNFromConn(config)
	case "postgres":
		return buildPostgresDSNFromConn(config)
	case "sqlite":
		return config.FilePath
	default:
		return ""
	}
}

// buildMySQLDSN builds MySQL DSN.
func buildMySQLDSN(config Config) string {
	charset := config.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	parseTime := "True"
	if !config.ParseTime {
		parseTime = "False"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%s&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		charset,
		parseTime,
	)
}

// buildMySQLDSNFromConn builds MySQL DSN from ConnectionConfig.
func buildMySQLDSNFromConn(config ConnectionConfig) string {
	charset := config.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	parseTime := "True"
	if !config.ParseTime {
		parseTime = "False"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%s&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		charset,
		parseTime,
	)
}

// buildPostgresDSN builds PostgreSQL DSN.
func buildPostgresDSN(config Config) string {
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	timezone := config.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.Host,
		config.Username,
		config.Password,
		config.Database,
		config.Port,
		sslMode,
		timezone,
	)

	// Add schema (search_path) if specified
	if config.Schema != "" {
		dsn += fmt.Sprintf(" search_path=%s", config.Schema)
	}

	return dsn
}

// buildPostgresDSNFromConn builds PostgreSQL DSN from ConnectionConfig.
func buildPostgresDSNFromConn(config ConnectionConfig) string {
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	timezone := config.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.Host,
		config.Username,
		config.Password,
		config.Database,
		config.Port,
		sslMode,
		timezone,
	)

	// Add schema (search_path) if specified
	if config.Schema != "" {
		dsn += fmt.Sprintf(" search_path=%s", config.Schema)
	}

	return dsn
}

// getLogger returns a GORM logger based on log level.
func getLogger(logLevel string, slowThreshold time.Duration) logger.Interface {
	var level logger.LogLevel

	switch logLevel {
	case "silent":
		level = logger.Silent
	case "error":
		level = logger.Error
	case "warn":
		level = logger.Warn
	case "info":
		level = logger.Info
	default:
		level = logger.Warn
	}

	return logger.Default.LogMode(level)
}
