.PHONY: all tidy fa fmt lint vet test pulse pulse-migrate pulse-validate

SERVICES = pulse pulse-migrate pulse-validate

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

up:
	@docker-compose -p pulse -f docker-compose.dev.yaml up -d

down:
	@docker-compose -p pulse -f docker-compose.dev.yaml down

$(SERVICES):
	@go build -ldflags "-s -w" -o ./build/$@ ./cmd/$@

build: $(SERVICES)
