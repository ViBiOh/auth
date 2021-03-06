SHELL = /bin/bash

ifneq ("$(wildcard .env)","")
	include .env
	export
endif

APP_NAME = bcrypt
PACKAGES ?= ./...

MAIN_SOURCE = cmd/bcrypt/bcrypt.go
MAIN_RUNNER = go run $(MAIN_SOURCE)
ifeq ($(DEBUG), true)
	MAIN_RUNNER = gdlv -d $(shell dirname $(MAIN_SOURCE)) debug --
endif

MEMORY_SOURCE = cmd/example_memory/memory.go
MEMORY_RUNNER = go run $(MEMORY_SOURCE)
ifeq ($(DEBUG), true)
	MEMORY_RUNNER = gdlv -d $(shell dirname $(MEMORY_SOURCE)) debug --
endif

DB_SOURCE = cmd/example_db/db.go
DB_RUNNER = go run $(DB_SOURCE)
ifeq ($(DEBUG), true)
	DB_RUNNER = gdlv -d $(shell dirname $(DB_SOURCE)) debug --
endif

SCORE_SOURCE = cmd/bcrypt_score/bcrypt_score.go
SCORE_RUNNER = go run $(SCORE_SOURCE)
ifeq ($(DEBUG), true)
	SCORE_RUNNER = gdlv -d $(shell dirname $(SCORE_SOURCE)) debug --
endif

.DEFAULT_GOAL := app

## help: Display list of commands
.PHONY: help
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sort

## name: Output app name
.PHONY: name
name:
	@printf "$(APP_NAME)"

## version: Output last commit sha1
.PHONY: version
version:
	@printf "$(shell git rev-parse --short HEAD)"

## dev: Build app
.PHONY: dev
dev: format style test build

## app: Build whole app
.PHONY: app
app: init dev

## init: Bootstrap your application. e.g. fetch some data files, make some API calls, request user input etc...
.PHONY: init
init:
	@curl --disable --silent --show-error --location --max-time 30 "https://raw.githubusercontent.com/ViBiOh/scripts/main/bootstrap" | bash -s -- "-c" "git_hooks" "coverage" "release"
	go install github.com/kisielk/errcheck@latest
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go mod tidy

## format: Format code. e.g Prettier (js), format (golang)
.PHONY: format
format:
	goimports -w $(shell find . -name "*.go")
	gofmt -s -w $(shell find . -name "*.go")

## style: Check lint, code styling rules. e.g. pylint, phpcs, eslint, style (java) etc ...
.PHONY: style
style:
	golint $(PACKAGES)
	errcheck -ignoretests $(PACKAGES)
	go vet $(PACKAGES)

## test: Shortcut to launch all the test tasks (unit, functional and integration).
.PHONY: test
test:
	scripts/coverage
	$(MAKE) bench

## bench: Shortcut to launch benchmark tests.
.PHONY: bench
bench:
	go test $(PACKAGES) -bench . -benchmem -run Benchmark.*

## build: Build the application.
.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/$(APP_NAME) $(MAIN_SOURCE)

## run: Locally run the application, e.g. node index.js, python -m myapp, go run myapp etc ...
.PHONY: run
run:
	$(MAIN_RUNNER) "password" 12

## run-memory: Run memory app
.PHONY: run-memory
run-memory:
	$(MEMORY_RUNNER) -users "1:`htpasswd -nBb admin password`"

## run-db: Run db app
.PHONY: run-db
run-db:
	$(DB_RUNNER)

## run: Locally compute the best bcrypt score
.PHONY: run-score
run-score:
	$(SCORE_RUNNER)
