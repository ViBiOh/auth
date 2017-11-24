default: go docker

go: deps dev

dev: format lint tst bench build

docker: docker-deps docker-build

deps:
	go get -u github.com/golang/lint/golint
	go get -u github.com/NYTimes/gziphandler
	go get -u github.com/ViBiOh/alcotest/alcotest
	go get -u github.com/ViBiOh/httputils
	go get -u github.com/ViBiOh/httputils/cert
	go get -u github.com/ViBiOh/httputils/cors
	go get -u github.com/ViBiOh/httputils/owasp
	go get -u github.com/ViBiOh/httputils/prometheus
	go get -u github.com/ViBiOh/httputils/rate
	go get -u golang.org/x/crypto/bcrypt
	go get -u golang.org/x/oauth2
	go get -u golang.org/x/oauth2/github
	go get -u golang.org/x/tools/cmd/goimports

format:
	goimports -w **/*.go *.go
	gofmt -s -w **/*.go *.go

lint:
	golint ./...
	go vet ./...

tst:
	script/coverage

bench:
	go test ./... -bench . -benchmem -run Benchmark.*

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/bcrypt bcrypt/bcrypt.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o bin/auth api.go
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo auth/auth.go

start:
	go run api.go -tls=false -basicUsers "1:admin:`go run bcrypt/bcrypt.go admin`"

docker-deps:
	curl -s -o cacert.pem https://curl.haxx.se/ca/cacert.pem

docker-build:
	docker build -t ${DOCKER_USER}/auth .

docker-push:
	docker login -u ${DOCKER_USER} -p ${DOCKER_PASS}
	docker push ${DOCKER_USER}/auth