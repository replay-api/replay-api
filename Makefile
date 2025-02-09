.PHONY: docker
.PHONY: docker-down
.PHONY: test-kafka
.PHONY: coverage

build-rest-api:
	@echo "Building API"
	CGO_ENABLED=0 go build -o replay-api-http-service ./cmd/rest-api/main.go

start-rest-api:
	@echo "Running API"
	@export DEV_ENV="true"
	@./replay-api-http-service

build-rest-api-windows:
	@echo "Building API"
	go build -o replay-api-http-service.exe .\cmd\rest-api\main.go

start-rest-api-windows:
	@echo "Running API"
	@set "DEV_ENV=true" && .\replay-api-http-service.exe

run-rest-api-windows:
	@echo "Running API"
	@set "DEV_ENV=true" && go run .\cmd\rest-api\main.go

test-docker:
	@echo "Running tests"
	@docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit

docker:
	@clear
	@printf "$(NEW_BUFFER)"
	@echo $(LOGO)
	@echo "♻️ $(CG)Removing$(CEND) containers and volumes"
	@docker-compose -f docker-compose.dev.yml down -v
	@echo "🔨 $(CC)Building$(CEND) new containers"
	@docker-compose -f docker-compose.dev.yml build
	@echo "🚀 $(CR)⦿ Running$(CEND) containers"
	@docker-compose -f docker-compose.dev.yml up -d

docker-down:
	@clear
	@printf "$(NEW_BUFFER)"
	@echo $(LOGO)
	@echo "♻️ $(CG)Removing$(CEND) containers and volumes"
	@docker-compose -f docker-compose.dev.yml down -v

test-coverage:
	@go test -covermode=atomic -coverprofile=coverage.out ./...
	@mkdir -p ./.coverage  
	@go tool cover -html=coverage.out -o ./.coverage/coverage.html 

test-kafka-produce:
	@go run pkg/infra/events/pub_kafka_poc.go

test-kafka-consume:
	@go run cmd/async-api/main.go

test-rabbit-consume:
	@go run cmd/async-api/main.go

test-rabbit-produce:
	@go run pkg/infra/events/pub_rabbit.go

test-rabbit-consume:
	@go run cmd/async-api/main.go

CG = \033[0;32m
CR = \033[0;31m
CEND = \033[0m
CC = \033[0;36m
B = \033[1m
NEW_BUFFER = \033[H\033[2J
LOGO = "\n\t$(CR)⦿ Replay$(CEND)API\n\n"





# --- Project Configuration ---

PROJECT_NAME     := replay-api
LICENSE_FILE     := LICENSE
IGNORE_DIRS      := vendor test         # Directories to exclude from license checks
ALLOWED_LICENSES := Apache-2.0 MIT      # Specify allowed licenses (comma-separated)

# --- Go Tools ---

GO ?= go
GO_LICENSES ?= go-licenses

# --- Makefile Targets ---

.PHONY: help
help: ## Display this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: licenses-check
licenses-check: ## Check dependencies for allowed licenses
	@$(GO_LICENSES) check $(PROJECT_NAME) --ignore "$(IGNORE_DIRS)" --allow "$(ALLOWED_LICENSES)"

.PHONY: licenses-csv
licenses-csv: ## Generate a CSV report of dependency licenses
	@$(GO_LICENSES) csv $(PROJECT_NAME) --ignore "$(IGNORE_DIRS)" > licenses.csv

.PHONY: licenses-save
licenses-save: ## Save dependency licenses to a directory
	@$(GO_LICENSES) save $(PROJECT_NAME) --ignore "$(IGNORE_DIRS)" --save_path licenses

.PHONY: install-tools
install-tools: ## Install required Go tools (if not already installed)
	@if ! hash $(GO_LICENSES) 2>/dev/null; then \
		$(GO) install github.com/google/go-licenses@latest; \
	fi

# --- Additional Targets (Customize as needed) ---

.PHONY: licenses-report # Example: Generate a custom report 
licenses-report: licenses-csv
	# Process the licenses.csv file (e.g., using a script or another tool) to create a custom report

# --- Default Target ---

.DEFAULT_GOAL := help



#--------- server


# Set your image name and container name (replace with your actual values)
IMAGE_NAME := cs2-server
CONTAINER_NAME := my-cs2-server

# Build the Docker image
build:
	docker build -t $(IMAGE_NAME) .

# Run the Docker container
run:
	docker run -d --name $(CONTAINER_NAME) -p 27015:27015/udp -p 27015:27015 $(IMAGE_NAME)

# Stop the Docker container
stop:
	docker stop $(CONTAINER_NAME)

# Remove the Docker container
rm:
	docker rm $(CONTAINER_NAME)

# Rebuild the Docker image and run the container (cleans up previous container if it exists)
rebuild-run: stop rm build run

# Push the Docker image to a registry (e.g., Docker Hub)
push:
	docker push $(IMAGE_NAME)
