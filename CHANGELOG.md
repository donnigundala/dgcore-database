# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-11-24

### Added
- PostgreSQL schema support via `search_path` parameter
- `Schema` field in `Config` and `ConnectionConfig` structs
- `WithSchema(schema string)` fluent configuration method
- `WithMaxConnections(maxOpen, maxIdle int)` helper method
- Schema-based multi-tenancy support
- 4 new tests for schema functionality
- Example 05: Schema support demonstration
- Comprehensive schema documentation (docs/SCHEMA.md)

### Enhanced
- PostgreSQL DSN builders now support schema parameter
- Multi-tenant examples updated with schema usage
- README updated with schema examples and use cases

### Documentation
- Added schema support guide (docs/SCHEMA.md)
- Updated README with PostgreSQL schema examples
- Added schema-based multi-tenancy patterns
- Updated API documentation

### Testing
- Added `TestPostgresSchemaSupport`
- Added `TestPostgresDefaultSchema`
- Added `TestPostgresSchemaFromConnectionConfig`
- Added `TestConfigWithSchema`
- Total tests: 40 (all passing)

## [1.0.0] - 2025-11-23

### Added
- Initial release of dgcore-database plugin
- Support for MySQL, PostgreSQL, and SQLite drivers
- Connection pooling with configurable limits
- Read/write splitting with automatic routing
- Load balancing strategies (round-robin, random, weighted)
- Multi-connection support for managing multiple databases
- Runtime connection management (add/remove connections)
- Comprehensive transaction support with context and savepoints
- Database migration system with up/down/reset functionality
- Health monitoring for all connections
- Fluent configuration API
- Auto-migration support
- Service provider integration with dgcore framework
- Comprehensive test suite (36 tests, 55.9% coverage)
- Complete documentation and examples

### Features
- **Read/Write Splitting**: Automatic routing of reads to slaves, writes to master
- **Multi-Connection**: Manage multiple named database connections
- **Migrations**: Version-controlled database schema changes
- **Transactions**: Full transaction support with automatic commit/rollback
- **Health Checks**: Monitor connection health across all databases
- **Load Balancing**: Distribute read load across multiple slaves

### Documentation
- Comprehensive README with quick start guide
- API reference documentation
- Migration guide with best practices
- 4 complete working examples

### Testing
- 36 unit tests covering all major functionality
- 55.9% code coverage
- Integration tests with SQLite

## [Unreleased]

### Planned
- PostgreSQL-specific features (LISTEN/NOTIFY)
- MySQL-specific features (LOAD DATA INFILE)
- Connection retry logic
- Circuit breaker pattern
- Metrics and monitoring integration
- Query caching layer
- Read replica lag detection
