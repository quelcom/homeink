.DEFAULT_GOAL := help

.PHONY: help
# From: http://disq.us/p/16327nq
help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: buildzero
buildzero: ## Build the Go project.
	# swag init -g main.go
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.buildCommit=$(shell git rev-parse --short HEAD)" -trimpath -v -o homeinkzero

.PHONY: build
build: ## Build the Go project.
	swag init -g main.go
	go build -ldflags="-s -w -X main.buildCommit=$(shell git rev-parse --short HEAD)" -trimpath -v -o homeink

.PHONY: dietcopy
dietcopy: ## Copy the binary to the DietPi
	cat homeinkzero | ssh dietpi.lan "cat > homeink"

.PHONY: fmt
fmt: ## Format the project with gofmt
	gofmt -l -w -s .

.PHONY: lint
lint: ## Lint code with golangci-lint
	golangci-lint run

.PHONY: test
test: ## Run the tests
	go test -v -cover ./...

.PHONY: check-vuln
check-vuln: ## Check for vulnerabilities
	govulncheck -show verbose ./...
