.PHONY: help build install server client doctor test docker-up docker-down fmt vet

MODULE := github.com/mcp-moqt/mcp-moqt-transport
BIN    := mcp-moqt
ADDR   ?= 127.0.0.1:8080

help:
	@echo "Friendly targets:"
	@echo "  make install   - install CLI to GOPATH/bin"
	@echo "  make build     - build ./cmd/mcp-moqt"
	@echo "  make server    - run MCP server (tools/prompts/resources)"
	@echo "  make client    - run demo client"
	@echo "  make doctor    - local environment check"
	@echo "  make test      - run unit/integration tests"
	@echo "  make docker-up - docker compose up --build"

build:
	go build -o bin/$(BIN) ./cmd/mcp-moqt

install:
	go install $(MODULE)/cmd/mcp-moqt@latest

server: build
	./bin/$(BIN) server -addr $(ADDR)

client: build
	./bin/$(BIN) client -addr $(ADDR)

doctor: build
	./bin/$(BIN) doctor

test:
	go test ./test/...

fmt:
	go fmt ./...

vet:
	go vet ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down
