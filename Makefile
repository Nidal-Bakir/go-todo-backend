include .env
.DEFAULT_GOAL := build

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Create DB container
docker-run:
	@docker compose up
 
# Shutdown DB container
docker-down:
	@docker compose down

# Test the application
test:
	@echo "Testing..."
	@go test ./tests -v

# Live Reload
watch:
	air; \
	echo "Watching...";\

# goose db migrations binary utils
GOOSE_CMD=GOOSE_MIGRATION_DIR=${GOOSE_MIGRATION_DIR} GOOSE_DRIVER=${GOOSE_DRIVER} GOOSE_DBSTRING=${GOOSE_DBSTRING} goose

# Migrate the DB to the most recent version available
goose-up:
	@${GOOSE_CMD} up

# Roll back the version by 1
goose-up-by-one :
	@${GOOSE_CMD} up-by-one

# Migrate the DB to a specific VERSION
goose-up-to :
	@read -p "version: " version; ${GOOSE_CMD} up-to $$version

# Roll back the version by 1
goose-down:
	@${GOOSE_CMD} down

# Roll back to a specific VERSION
goose-down-to :
	@read -p "version: " version; ${GOOSE_CMD} down-to $$version

# Re-run the latest migration
goose-redo :
	@${GOOSE_CMD} redo

# Roll back all migrations
goose-reset :
	@${GOOSE_CMD} reset

# Dump the migration status for the current DB
goose-status:
	@${GOOSE_CMD} status

# Print the current version of the database
goose-version:
	@${GOOSE_CMD} version

# Apply sequential ordering to migrations
goose-fix:
	@${GOOSE_CMD} fix

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main
	@echo "Done"

