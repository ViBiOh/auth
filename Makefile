default: api docker-api

api: deps go

go: format lint tst bench build

deps:
	go get github.com/golang/dep/cmd/dep
	go get github.com/golang/lint/golint
	go get github.com/kisielk/errcheck
	go get golang.org/x/tools/cmd/goimports
	dep ensure

format:
	goimports -w */*/*.go
	gofmt -s -w */*/*.go

lint:
	golint `go list ./... | grep -v vendor`
	errcheck -ignoretests `go list ./... | grep -v vendor`
	go vet ./...

tst:
	script/coverage

bench:
	go test ./... -bench . -benchmem -run Benchmark.*

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/bcrypt cmd/bcrypt/bcrypt.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/auth cmd/auth/auth.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo pkg/auth/auth.go

start-api:
	go run -race cmd/auth/auth.go \
		-tls=false \
		-basicUsers "1:admin:`go run -race cmd/bcrypt/bcrypt.go admin`"

docker-deps:
	curl -s -o cacert.pem https://curl.haxx.se/ca/cacert.pem

docker-login:
	echo $(DOCKER_PASS) | docker login -u $(DOCKER_USER) --password-stdin

docker-api: docker-build-api docker-push-api

docker-build-api: docker-deps
	docker build -t $(DOCKER_USER)/auth .

docker-push-api: docker-login
	docker push $(DOCKER_USER)/auth

.PHONY: api go deps format lint tst bench build start docker-deps docker-login docker-api docker-build-api docker-push-api
