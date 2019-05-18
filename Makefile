SHELL = /bin/sh

APP_NAME = auth
PACKAGES ?= ./...

GOBIN=bin
BINARY_PATH=$(GOBIN)/$(APP_NAME)

SERVER_SOURCE = cmd/auth/auth.go
SERVER_RUNNER = go run $(SERVER_SOURCE)
ifeq ($(DEBUG), true)
	SERVER_RUNNER = dlv debug $(SERVER_SOURCE) --
endif

.DEFAULT_GOAL := app

## help: Display list of commands
.PHONY: help
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sed -e 's|^| |'

## $(APP_NAME): Build app with dependencies download
$(APP_NAME): deps go

## go: Build app
.PHONY: go
go: format lint test bench build

## name: Output name
.PHONY: name
name:
	@echo -n $(APP_NAME)

## dist: Output build output path
.PHONY: dist
dist:
	@echo -n $(BINARY_PATH)

## version: Output sha1 of last commit
.PHONY: version
version:
	@echo -n $(shell git rev-parse --short HEAD)

## author: Output author's name of last commit
.PHONY: author
author:
	@python -c 'import sys; import urllib; sys.stdout.write(urllib.quote_plus(sys.argv[1]))' "$(shell git log --pretty=format:'%an' -n 1)"

## deps: Download dependencies
.PHONY: deps
deps:
	go get github.com/kisielk/errcheck
	go get golang.org/x/lint/golint
	go get golang.org/x/tools/cmd/goimports

## format: Format code
.PHONY: format
format:
	goimports -w */*/*.go
	gofmt -s -w */*/*.go

## lint: Lint code
.PHONY: lint
lint:
	golint $(PACKAGES)
	errcheck -ignoretests $(PACKAGES)
	go vet $(PACKAGES)

## test: Test code with coverage
.PHONY: test
test:
	script/coverage

## bench: Benchmark code
.PHONY: bench
bench:
	go test $(PACKAGES) -bench . -benchmem -run Benchmark.*

## build: Build binary
.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo $(PACKAGES)
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(BINARY_PATH) $(SERVER_SOURCE)
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(GOBIN)/bcrypt cmd/bcrypt/bcrypt.go

## start: Start app
.PHONY: start
start:
	$(SERVER_RUNNER) \
		-basicUsers "1:admin:$(go run cmd/bcrypt/bcrypt.go admin)"
