.PHONY: test build integration
build:
	go build -trimpath -o dist/cppreportcard-cli ./cmd/cppreportcard-cli
test:
	go test -race ./...
	go vet ./...
integration:
	./scripts/integration.sh
