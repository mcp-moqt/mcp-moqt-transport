.PHONY: help build install server stdio client doctor test docker-up docker-down fmt vet vuln coverage

MODULE := github.com/mcp-moqt/mcp-moqt-transport
BIN    := mcp-moqt
ADDR   ?= 127.0.0.1:8080

help:
	@echo "Friendly targets:"
	@echo "  make install   - install CLI to GOBIN/GOPATH/bin"
	@echo "  make build     - build ./cmd/mcp-moqt"
	@echo "  make server    - run QUIC MCP server"
	@echo "  make stdio     - run stdio MCP server"
	@echo "  make client    - run demo client"
	@echo "  make doctor    - local environment + capabilities check"
	@echo "  make test      - run unit/integration tests"
	@echo "  make coverage  - coverage report for pkg + unit tests"
	@echo "  make vuln      - govulncheck"
	@echo "  make docker-up - docker compose up --build"

build:
	go build -o bin/$(BIN)$(shell go env GOEXE) ./cmd/mcp-moqt

install:
	go install ./cmd/mcp-moqt

server: build
	./bin/$(BIN)$(shell go env GOEXE) server -addr $(ADDR)

stdio: build
	./bin/$(BIN)$(shell go env GOEXE) server -stdio

client: build
	./bin/$(BIN)$(shell go env GOEXE) client -addr $(ADDR)

doctor: build
	./bin/$(BIN)$(shell go env GOEXE) doctor

multi-client: build
	./bin/$(BIN)$(shell go env GOEXE) server -addr $(ADDR) -multi &
	sleep 2
	go run ./examples/multi_client -addr $(ADDR) -clients 3

test:
	go test ./test/...

coverage:
	go test -coverprofile=coverage.txt -covermode=atomic ./pkg/... ./test/unit/...
	go tool cover -func=coverage.txt

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down
