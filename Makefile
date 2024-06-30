DATABASE_URL=$(shell echo $$DATABASE_URL)
DATABASE_URL_DEV=postgres://postgres:postgres@postgres-dev:5432/test?sslmode=disable
DOCKER_COMPOSE_TOOLS= docker compose up -d && docker compose exec tools

ifndef DATABASE_URL 
DATABASE_URL=postgres://postgres:postgres@postgres:5432/test?sslmode=disable
endif

up:
	docker compose up --build -d 
down: 
	docker compose down
cleanup:
	docker compose down -v --rmi all --remove-orphans
atlas-run:
	$(DOCKER_COMPOSE_TOOLS) atlas $(args)
atlas-diff:
	$(DOCKER_COMPOSE_TOOLS) atlas migrate diff $(args) \
  --dir "file://migrations?format=golang-migrate" \
  --dev-url "$(DATABASE_URL_DEV)" \
  --to "file://migrations/schema/schema.sql" \
   --format '{{ sql . "  " }}'
atlas-hash:
	$(DOCKER_COMPOSE_TOOLS) atlas migrate hash $(args) \
  --dir "file://migrations/?format=golang-migrate" 
atlas-migrate:
	atlas migrate apply \
  --dir "file://migrations?format=golang-migrate" \
  --url "$(DATABASE_URL)" 
atlas-migrate-tools:
	$(DOCKER_COMPOSE_TOOLS)  atlas migrate apply \
  --dir "file://migrations?format=golang-migrate" \
  --url "$(DATABASE_URL)" 
atlas-status:
  	atlas migrate status \
  --dir "file://migrations?format=golang-migrate" \
  --url "$(DATABASE_URL_DEV)" 
setup-hooks:
	cp .github/pre-commit/pre-commit .git/hooks/pre-commit

