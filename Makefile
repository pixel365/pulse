.PHONY: all tidy fa fmt lint vet test

all: tidy fa fmt lint

tidy:
	go mod tidy

fa:
	@fieldalignment -fix ./...

fmt:
	@goimports -w -local github.com/pixel365/pulse .
	@gofmt -w .
	@golines -w .

lint:
	@golangci-lint run

vet:
	go vet ./...

test:
	@go test ./...
