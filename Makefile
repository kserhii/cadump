.PHONY: update build test test-cov clean

update:
	@echo "Updating dependencies"
	@go mod tidy
	@go mod verify

build:
	@echo "Create build for Linux"
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o build/cadump cmd/cadump/main.go

test:
	@echo "Run tests"
	@cd cadump && go test

test-cov:
	@echo "Run tests with coverage"
	@cd cadump && go test -coverprofile=/tmp/cover.out && go tool cover -html=/tmp/cover.out

clean:
	@echo "Cleanup"
	@rm -fr /tmp/cover*
