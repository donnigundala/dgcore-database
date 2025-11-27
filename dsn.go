package database

import "fmt"

// DSNBuilder defines the interface for building Data Source Names (DSNs).
// This allows reusing DSN building logic for both Config and ConnectionConfig.
type DSNBuilder interface {
	GetDriver() string
	GetHost() string
	GetPort() int
	GetDatabase() string
	GetUsername() string
	GetPassword() string
	GetCharset() string
	GetParseTime() bool
	GetSSLMode() string
	GetTimezone() string
	GetSchema() string
	GetFilePath() string
}

// Config implementation of DSNBuilder
func (c Config) GetDriver() string   { return c.Driver }
func (c Config) GetHost() string     { return c.Host }
func (c Config) GetPort() int        { return c.Port }
func (c Config) GetDatabase() string { return c.Database }
func (c Config) GetUsername() string { return c.Username }
func (c Config) GetPassword() string { return c.Password }
func (c Config) GetCharset() string  { return c.Charset }
func (c Config) GetParseTime() bool  { return c.ParseTime }
func (c Config) GetSSLMode() string  { return c.SSLMode }
func (c Config) GetTimezone() string { return c.Timezone }
func (c Config) GetSchema() string   { return c.Schema }
func (c Config) GetFilePath() string { return c.FilePath }

// ConnectionConfig implementation of DSNBuilder
func (c ConnectionConfig) GetDriver() string   { return c.Driver }
func (c ConnectionConfig) GetHost() string     { return c.Host }
func (c ConnectionConfig) GetPort() int        { return c.Port }
func (c ConnectionConfig) GetDatabase() string { return c.Database }
func (c ConnectionConfig) GetUsername() string { return c.Username }
func (c ConnectionConfig) GetPassword() string { return c.Password }
func (c ConnectionConfig) GetCharset() string  { return c.Charset }
func (c ConnectionConfig) GetParseTime() bool  { return c.ParseTime }
func (c ConnectionConfig) GetSSLMode() string  { return c.SSLMode }
func (c ConnectionConfig) GetTimezone() string { return c.Timezone }
func (c ConnectionConfig) GetSchema() string   { return c.Schema }
func (c ConnectionConfig) GetFilePath() string { return c.FilePath }

// buildDSN builds a DSN string from any configuration implementing DSNBuilder.
func buildDSN(config DSNBuilder) string {
	switch config.GetDriver() {
	case "mysql":
		return buildMySQLDSN(config)
	case "postgres":
		return buildPostgresDSN(config)
	case "sqlite":
		return config.GetFilePath()
	default:
		return ""
	}
}

// buildMySQLDSN builds MySQL DSN.
func buildMySQLDSN(config DSNBuilder) string {
	charset := config.GetCharset()
	if charset == "" {
		charset = "utf8mb4"
	}

	parseTime := "True"
	if !config.GetParseTime() {
		parseTime = "False"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%s&loc=Local",
		config.GetUsername(),
		config.GetPassword(),
		config.GetHost(),
		config.GetPort(),
		config.GetDatabase(),
		charset,
		parseTime,
	)
}

// buildPostgresDSN builds PostgreSQL DSN.
func buildPostgresDSN(config DSNBuilder) string {
	sslMode := config.GetSSLMode()
	if sslMode == "" {
		sslMode = "disable"
	}

	timezone := config.GetTimezone()
	if timezone == "" {
		timezone = "UTC"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.GetHost(),
		config.GetUsername(),
		config.GetPassword(),
		config.GetDatabase(),
		config.GetPort(),
		sslMode,
		timezone,
	)

	// Add schema (search_path) if specified
	if config.GetSchema() != "" {
		dsn += fmt.Sprintf(" search_path=%s", config.GetSchema())
	}

	return dsn
}
