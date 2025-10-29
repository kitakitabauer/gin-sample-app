.PHONY: run build test lint dev vuln docker-build docker-run docker-clean migrate-up migrate-down migrate-steps migrate-lint openapi-lint

APP_NAME := gin-sample-app
DOCKER_IMAGE ?= $(APP_NAME):latest
MIGRATE := go run ./cmd/migrate

run:
	go run main.go

build:
	go build -o app main.go

test:
	go test -v ./...

lint:
	go vet ./...

vuln:
	if ! command -v govulncheck >/dev/null 2>&1; then \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	govulncheck ./...

dev:
	APP_ENV=dev go run main.go

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run --rm -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

docker-clean:
	docker rmi $(DOCKER_IMAGE) || true

migrate-up:
	$(MIGRATE) -cmd up

migrate-down:
	$(MIGRATE) -cmd down

migrate-steps:
	@if [ -z "$(STEPS)" ]; then \
		echo "Usage: make migrate-steps STEPS=1"; \
		exit 1; \
	fi
	$(MIGRATE) -cmd steps -steps $(STEPS)

migrate-lint:
	go test ./internal/database -run TestMigrationFilesHavePairs -count=1

openapi-lint:
	npx --yes @stoplight/spectral-cli lint docs/openapi.yaml
