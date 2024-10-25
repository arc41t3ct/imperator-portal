BINARY_NAME=imperatorapp

build:
	@go mod vendor
	@echo "Building Imperator App"
	@go build -o tmp/${BINARY_NAME} .
	@echo "Imperator App built!"

run: build
	@echo "Starting Imperator App"
	@./tmp/${BINARY_NAME} &
	@echo "Imperator App started!"

clean:
	@echo "Cleaning"
	@go clean
	@rm ./tmp/${BINARY_NAME}
	@echo "Cleaned!"

infra_up:
	@echo "Starting Infrastructure"
	docker compose -f ../imperator-dev-infra/docker-compose.yml up -d 

infra_down:
	@echo "Stopping Infrastructure"
	docker compose -f ../imperator-dev-infra/docker-compose.yml down

test:
	@echo "Testing..."
	@go test ./...
	@echo "Done Testing!"

test_model_integration_test_report:
	@echo "Coverage Report for Models..."
	@go test -coverprofile=coverage.out ./models/ --tags integration
	@go tool cover -html=coverage.out
	@echo "Done creating report!"

test_model_unit_test_report:
	@echo "Coverage Report for Models..."
	@go test -coverprofile=coverage.out ./models/ --tags unit
	@go tool cover -html=coverage.out
	@echo "Done creating report!"

test_model_test_report:
	@echo "Coverage Report for Models..."
	@go test -coverprofile=coverage.out ./models/ --tags unit --tags integration
	@go tool cover -html=coverage.out
	@echo "Done creating report!"

start: run

stop:
	@echo "Stopping Imperator App"
	@-pkill -SIGTERM -f "./tmp/${BINARY_NAME}"
	@echo "Stopped Imperator App"

restart: stop start
