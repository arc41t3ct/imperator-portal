## test: runs all tests
test:
	@go test -v ./...

## cover: opens coverage in browser
cover:
	@go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out && rm coverage.out

## coverage: displays test coverage
coverage:
	@go test -cover ./...

## build_cli: builds the command line tool imperator and copies it to myapp
build_cli:
	@go build -o ../../../bin/imperator ./cmd/cli
	@go build -o ../imperator_app/imperator ./cmd/cli
