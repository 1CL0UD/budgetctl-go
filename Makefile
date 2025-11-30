-include .env
export

DEV_URL=docker://postgres/17/dev
SCHEMA_FILE=file://internal/database/schema.hcl
MIGRATION_PATH=internal/database/migrations
MIGRATION_URL=file://$(MIGRATION_PATH)
DB_CONTAINER=budgetctl-db

.PHONY: dev diff apply sqlc gen check-env

# --- COMMANDS ---

dev: check-env swagger
	@echo "🐳 Ensuring Database is up..."
	@if [ -z "$$(docker ps -q -f name=$(DB_CONTAINER))" ]; then \
		echo "🐳 Database '$(DB_CONTAINER)' not running. Starting it..."; \
	else \
		echo "✅ Database '$(DB_CONTAINER)' is already running."; \
	fi
	@echo "🚀 Starting Hot Reload..."
	@air

check-env:
ifndef DATABASE_URL
	$(error DATABASE_URL is undefined. Ensure .env file exists in root and contains DATABASE_URL)
endif
	@echo "Env loaded! DATABASE_URL is set."

# 1. Create Migration File
diff: check-env
	@echo "1. Creating migration directory..."
	@mkdir -p $(MIGRATION_PATH)
	@echo "2. Generating migration diff..."
	atlas migrate diff $(name) \
	  --to "$(SCHEMA_FILE)" \
	  --dev-url "$(DEV_URL)" \
	  --dir "$(MIGRATION_URL)" \
	  --format "{{ sql . \"  \" }}"

# 2. Apply to Database
apply: check-env
	@echo "3. Applying migrations to $(DATABASE_URL)..."
	atlas migrate apply \
	  --url "$(DATABASE_URL)" \
	  --dir "$(MIGRATION_URL)"

# 3. Generate Go Code
sqlc:
	@echo "4. Generating Go code..."
	sqlc generate

swagger:
	swag init -g cmd/api/main.go --output docs

gen:
	@$(MAKE) diff name=$(name)
	@$(MAKE) apply
	@$(MAKE) sqlc
	@$(MAKE) swagger
