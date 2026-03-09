title := "add_needed_care"
include .env

dev:
	docker compose --profile development --env-file .env up --build

prod:
	docker compose --profile integration-tests up --build

run:
	go run cmd/web/main.go

test:
	go test ./...

migration:
	go run cmd/migration/main.go

migration-up:
	go run cmd/migration/main.go -up

migration-up-force:
	go run cmd/migration/main.go -up -force

migration-down:
	go run cmd/migration/main.go -down

ui-build:
	cd pkg/ui && bun run build
