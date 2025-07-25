.PHONY: tests
tests:
	go test -v -cover ./...

.PHONY: test-db
test-db:
	go test -v -cover ./internal/database/...

.PHONY: test-service
test-service:
	go test -v -cover ./internal/domain/service/...

.PHONY: test-handler
test-handler:
	go test -v -cover ./internal/handlers/...

.PHONY: mocks
mocks: install-mockgen domain-mocks

.PHONY: install-mockgen
install-mockgen:
	@echo "=====> Installing mockgen"
	@go install go.uber.org/mock/mockgen@latest

.PHONY: domain-mocks
domain-mocks:
	@echo "=====> Generating domain mocks"
	@rm -rf mocks
	@mkdir -p mocks
	
	@for file in internal/domain/contract/*.go; do \
		filename=$$(basename $$file); \
		mockgen -package mocks -source=$$file -destination=mocks/$$filename; \
	done

	@echo "=====> Mocks generated"

.PHONY: build
build:
	go build -o slack-rotation-bot cmd/bot/main.go

.PHONY: run
run:
	go run cmd/bot/main.go