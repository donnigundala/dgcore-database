# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
