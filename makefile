.DEFAULT_GOAL := all

all: test format lint

build:
	@go build ./...

test: build
	@gotestsum ./... -cover -race -shuffle=on

format:
	@go mod tidy
	@gofumpt -l -w .

lint:
	@go vet ./...
	@govulncheck ./...
	@gosec ./...
	@golangci-lint run

deps:
	@go install gotest.tools/gotestsum@latest
	@go install mvdan.cc/gofumpt@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
