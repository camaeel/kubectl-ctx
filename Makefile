all: test build

build:
	go build -o bin/ github.com/camaeel/kubectl-ctx/cmd/...
test:
	go test -coverprofile=coverage.out ./... 

vet:
	go vet ./...
lint:
	golangci-lint run
format:
	go fmt ./...
clean:
	rm -rf bin/ coverage.out
	