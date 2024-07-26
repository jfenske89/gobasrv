all: install fmt vet lint test

install:
	go mod tidy

fmt:
	goimports -w --local github.com/jfenske89 ./

vet:
	go vet ./...

test:
	go test -v ./...

lint:
	golangci-lint run --verbose

upgrade:
	go get -u ./...
	go mod tidy
