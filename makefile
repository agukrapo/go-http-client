.DEFAULT_GOAL := all

NAME := $(shell basename $(CURDIR))

all: build test format lint

build:
	@echo "Building ${NAME}..."
	@go build ./...

test: build
	@echo "Testing ${NAME}..."
	@go test ./... -cover -race -shuffle=on

format:
	@echo "Formatting ${NAME}..."
	@go mod tidy
	@gofumpt -l -w . #go install mvdan.cc/gofumpt@latest

lint:
	@echo "Linting ${NAME}..."
	@go vet ./...
	@golangci-lint run #https://golangci-lint.run/usage/install/
