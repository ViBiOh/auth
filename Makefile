default: deps fmt lint tst build

deps:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/lint/golint
	go get -u github.com/ViBiOh/httputils
	go get -u github.com/ViBiOh/alcotest
	go get -u golang.org/x/crypto/bcrypt
	go get -u golang.org/x/oauth2

fmt:
	goimports -w **/*.go *.go
	gofmt -s -w **/*.go *.go

lint:
	golint ./...
	go vet ./...

tst:
	script/coverage

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/bcrypt bcrypt/bcrypt.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/auth auth.go
