title := "add_needed_care"
include .env

.PHONY: dev prod run test migration migration-up migration-up-force migration-down
dev: ui-build
	docker compose --profile development --env-file .env up --build

dev-down:
	docker compose --profile development --env-file .env down

prod: ui-build
	docker compose --profile integration-tests up --build

prod-down:
	docker compose --profile integration-tests down

run: ui-build
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

build:
	go build -o bin/url-shortener cmd/web/main.go
