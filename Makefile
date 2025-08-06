.PHONY: test test-setup test-cleanup test-postgres test-mariadb

# Start test databases
test-setup:
	@echo "Starting test databases..."
	docker-compose up -d
	@echo "Waiting for databases to be ready..."
	sleep 10
	@echo "Databases are ready!"

# Stop and remove test databases
test-cleanup:
	@echo "Stopping test databases..."
	docker-compose down -v
	@echo "Test databases stopped and volumes removed."

# Run all sync tests
test: test-setup
	@echo "Running sync tests..."
	go test -v -run "TestPostgresSyncCreateTable|TestMariaDBSyncCreateTable|TestPostgresSyncAddColumn|TestMariaDBSyncAddColumn|TestPostgresSyncWithIndexes|TestMariaDBSyncWithIndexes|TestSyncWithBothDatabases" || true
	@$(MAKE) test-cleanup

# Run only PostgreSQL tests
test-postgres: test-setup
	@echo "Running PostgreSQL sync tests..."
	go test -v -run "TestPostgres" || true
	@$(MAKE) test-cleanup

# Run only MariaDB tests
test-mariadb: test-setup
	@echo "Running MariaDB sync tests..."
	go test -v -run "TestMariaDB" || true
	@$(MAKE) test-cleanup

# Quick test without database setup (assumes databases are already running)
test-quick:
	@echo "Running sync tests (assuming databases are already running)..."
	go test -v -run "TestPostgresSyncCreateTable|TestMariaDBSyncCreateTable|TestPostgresSyncAddColumn|TestMariaDBSyncAddColumn|TestPostgresSyncWithIndexes|TestMariaDBSyncWithIndexes|TestSyncWithBothDatabases"

# Build the project
build:
	go build .

# Check dependencies
deps:
	go mod tidy
	go mod download

# Help
help:
	@echo "Available targets:"
	@echo "  test-setup    - Start test databases using Docker Compose"
	@echo "  test-cleanup  - Stop and remove test databases"
	@echo "  test          - Run all sync tests (includes setup and cleanup)"
	@echo "  test-postgres - Run only PostgreSQL sync tests"
	@echo "  test-mariadb  - Run only MariaDB sync tests"
	@echo "  test-quick    - Run tests without database setup"
	@echo "  build         - Build the project"
	@echo "  deps          - Update dependencies"
	@echo "  help          - Show this help message"
