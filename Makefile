all: test build

build:
	go build -o bin/ github.com/camaeel/kubectl-ctx/cmd/...
test:
	go test -coverprofile=coverage.out ./... 

vet:
	go vet ./...

lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

format:
	go fmt ./...
clean:
	rm -rf bin/ coverage.out
	