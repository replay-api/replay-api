.PHONY: docker
.PHONY: docker-down
.PHONY: test-kafka
.PHONY: coverage
.PHONY: up
.PHONY: down
.PHONY: seed

#========================================
# Development Environment (make up/down)
#========================================

up: ## Start the complete development environment with seed data
	@echo "$(CC)Starting LeetGaming PRO Development Environment$(CEND)"
	@echo ""
	@echo "$(CC)Step 1/10: Creating Kind cluster...$(CEND)"
	@kind create cluster --config=kind-config.yaml --name=leetgaming-local 2>/dev/null || echo "Cluster already exists"
	@kubectl cluster-info --context kind-leetgaming-local
	@echo ""
	@echo "$(CC)Step 2/10: Installing Strimzi Kafka Operator...$(CEND)"
	@kubectl apply -f 'https://strimzi.io/install/latest?namespace=leetgaming' -n leetgaming 2>/dev/null || echo "Strimzi may already be installed"
	@echo ""
	@echo "$(CC)Step 3/10: Building API image...$(CEND)"
	@docker build -t replay-api:latest -f cmd/rest-api/Dockerfile.rest-api . || { echo "$(CR)Failed to build API image$(CEND)"; exit 1; }
	@kind load docker-image replay-api:latest --name=leetgaming-local
	@echo ""
	@echo "$(CC)Step 4/10: Building Web Frontend image...$(CEND)"
	@if [ -d "../leetgaming-pro-web" ]; then \
		docker build -t leetgaming-web:latest ../leetgaming-pro-web || echo "$(CR)Web frontend build skipped (optional)$(CEND)"; \
		kind load docker-image leetgaming-web:latest --name=leetgaming-local 2>/dev/null || true; \
	else \
		echo "Web frontend directory not found, skipping..."; \
	fi
	@echo ""
	@echo "$(CC)Step 5/10: Applying Kubernetes manifests...$(CEND)"
	@kubectl apply -k k8s/local/
	@echo ""
	@echo "$(CC)Step 6/10: Waiting for Strimzi operator...$(CEND)"
	@kubectl wait --for=condition=ready pod -l strimzi.io/kind=cluster-operator -n leetgaming --timeout=120s 2>/dev/null || echo "$(CR)Strimzi operator not ready, continuing...$(CEND)"
	@echo ""
	@echo "$(CC)Step 7/10: Waiting for Kafka cluster (this may take 60-90 seconds)...$(CEND)"
	@kubectl wait kafka/leetgaming-kafka --for=condition=Ready -n leetgaming --timeout=300s 2>/dev/null || echo "$(CR)Kafka not ready - continuing anyway$(CEND)"
	@echo ""
	@echo "$(CC)Step 8/10: Waiting for other pods to be ready...$(CEND)"
	@kubectl wait --for=condition=ready pod -l app=mongodb -n leetgaming --timeout=300s 2>/dev/null || true
	@kubectl wait --for=condition=ready pod -l app=replay-api -n leetgaming --timeout=300s 2>/dev/null || true
	@kubectl wait --for=condition=ready pod -l app=web-frontend -n leetgaming --timeout=300s 2>/dev/null || true
	@kubectl wait --for=condition=ready pod -l app=prometheus -n leetgaming --timeout=300s 2>/dev/null || true
	@kubectl wait --for=condition=ready pod -l app=grafana -n leetgaming --timeout=300s 2>/dev/null || true
	@kubectl wait --for=condition=ready pod -l app=pyroscope -n leetgaming --timeout=300s 2>/dev/null || true
	@echo ""
	@echo "$(CC)Step 9/10: Seeding database...$(CEND)"
	@sleep 10
	@kubectl delete job database-seed -n leetgaming 2>/dev/null || true
	@kubectl apply -f k8s/local/seed-job.yaml
	@kubectl wait --for=condition=complete job/database-seed -n leetgaming --timeout=120s 2>/dev/null || kubectl logs -l job-name=database-seed -n leetgaming
	@echo ""
	@echo "$(CC)Step 10/10: Starting port forwards...$(CEND)"
	@./scripts/port-forward.sh start
	@echo ""
	@echo "$(CG)Development environment is ready!$(CEND)"
	@echo ""
	@echo "$(CC)Available endpoints:$(CEND)"
	@echo "  - API:           http://localhost:8080"
	@echo "  - Health:        http://localhost:8080/health"
	@echo "  - Web:           http://localhost:3030"
	@echo "  - Prometheus:    http://localhost:9090"
	@echo "  - Grafana:       http://localhost:3031 (admin/leetgaming)"
	@echo "  - Pyroscope:     http://localhost:3041 (Continuous Profiling)"
	@echo "  - Kafka:         localhost:9092"
	@echo ""
	@echo "$(CC)Kafka Topics:$(CEND)"
	@echo "  - matchmaking.queue.events"
	@echo "  - matchmaking.lobby.events"
	@echo "  - matchmaking.matches.created"
	@echo "  - websocket.broadcasts"
	@echo ""
	@echo "$(CC)Useful commands:$(CEND)"
	@echo "  - Stop environment:    make down"
	@echo "  - View API logs:       kubectl logs -f -l app=replay-api -n leetgaming"
	@echo "  - View Kafka logs:     kubectl logs -f -l strimzi.io/name=leetgaming-kafka-kafka -n leetgaming"
	@echo "  - Kafka status:        make kafka-status"
	@echo "  - Run seeds again:     make seed"
	@echo "  - Check status:        kubectl get pods -n leetgaming"

down: ## Stop and clean up the development environment
	@echo "$(CR)Stopping LeetGaming PRO Development Environment$(CEND)"
	@./scripts/port-forward.sh stop 2>/dev/null || true
	@pkill -f "kubectl port-forward" 2>/dev/null || true
	@for port in 8080 9090 3031 3041 9092; do \
		pids=$$(lsof -ti :$$port 2>/dev/null || true); \
		if [ -n "$$pids" ]; then echo "$$pids" | xargs kill -9 2>/dev/null || true; fi; \
	done
	@docker-compose down -v --remove-orphans 2>/dev/null || true
	@kind delete cluster --name=leetgaming-local 2>/dev/null || true
	@echo "$(CG)Environment stopped$(CEND)"

down-clean: down ## Stop environment and clean Docker images (safe for Docker Desktop)
	@echo "$(CR)Cleaning Docker images...$(CEND)"
	@docker rmi -f replay-api:latest 2>/dev/null || true
	@docker rmi -f leetgaming-web:latest 2>/dev/null || true
	@docker image prune -f 2>/dev/null || true
	@docker builder prune -f --keep-storage=2GB 2>/dev/null || true
	@echo "$(CG)Clean complete$(CEND)"

kill-ports: ## Kill processes on API ports (8080, 9090)
	@echo "$(CR)Killing processes on API ports...$(CEND)"
	@for port in 8080 9090 3031 3041 9092 27017; do \
		pids=$$(lsof -ti :$$port 2>/dev/null || true); \
		if [ -n "$$pids" ]; then \
			echo "  Port $$port: killing PIDs $$pids"; \
			echo "$$pids" | xargs kill -9 2>/dev/null || true; \
		fi; \
	done
	@echo "$(CG)Ports cleared$(CEND)"

nuke: kill-ports down-clean ## Full cleanup: kill ports, remove images, delete cluster
	@rm -rf bin/ 2>/dev/null || true
	@echo "$(CG)Nuke complete$(CEND)"

port-forward: ## Start port forwards (use after 'make up' if services become inaccessible)
	@./scripts/port-forward.sh start

port-forward-stop: ## Stop all port forwards
	@./scripts/port-forward.sh stop

port-forward-status: ## Check port forward and service status
	@./scripts/port-forward.sh status

kafka-status: ## Check Kafka cluster status
	@echo "$(CC)Kafka Cluster Status$(CEND)"
	@echo ""
	@kubectl get kafka -n leetgaming 2>/dev/null || echo "Kafka cluster not found"
	@echo ""
	@echo "$(CC)Kafka Topics:$(CEND)"
	@kubectl get kafkatopics -n leetgaming 2>/dev/null || echo "No topics found"
	@echo ""
	@echo "$(CC)Kafka Pods:$(CEND)"
	@kubectl get pods -l strimzi.io/cluster=leetgaming-kafka -n leetgaming 2>/dev/null || echo "No Kafka pods found"

status: ## Check status of the development environment
	@echo "$(CC)LeetGaming PRO Development Environment Status$(CEND)"
	@echo ""
	@kubectl get pods -n leetgaming 2>/dev/null || echo "Cluster not running"
	@echo ""
	@kubectl get svc -n leetgaming 2>/dev/null || true

logs: ## Tail API logs
	@kubectl logs -f -l app=replay-api -n leetgaming --tail=100

build-rest-api:
	@echo "Building API"
	CGO_ENABLED=0 go build -o replay-api-http-service ./cmd/rest-api/main.go

start-rest-api:
	@echo "Running API"
	@export DEV_ENV="true"
	@./replay-api-http-service

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

seed: ## Seed database with realistic esports data
	@echo "🌱 Seeding database with esports teams and players..."
	@go run ./cmd/cli/seed/main.go
	@echo "✅ Seed completed!"

build-seed: ## Build the seed CLI
	@echo "🔨 Building seed CLI..."
	@CGO_ENABLED=0 go build -o seed-cli ./cmd/cli/seed/main.go 

proto:
	@echo "Generating proto files"
	@protoc --proto_path=pkg/app/proto/squad --go_out=pkg/app/proto/squad --go-grpc_out=pkg/app/proto/squad pkg/app/proto/squad/player_profile.proto
	@protoc --proto_path=pkg/app/proto/iam --go_out=pkg/app/proto/iam --go-grpc_out=pkg/app/proto/iam pkg/app/proto/iam/validate_rid.proto
	@protoc --proto_path=pkg/app/proto/billing --go_out=pkg/app/proto/billing --go-grpc_out=pkg/app/proto/billing pkg/app/proto/billing/subscription.proto


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

#========================================
# Kubernetes Deployment (Production-Grade)
#========================================

# Kubernetes Configuration
CLUSTER_NAME := replay-api-cluster
NAMESPACE := replay-api
K8S_IMAGE_NAME := replay-api
K8S_IMAGE_TAG := latest

.PHONY: k8s-cluster-create
k8s-cluster-create: ## Create Kind cluster with production configuration
	@echo "$(CC)creating kind cluster$(CEND)"
	@kind create cluster --config=kind-config.yaml --name=$(CLUSTER_NAME)
	@kubectl cluster-info --context kind-$(CLUSTER_NAME)
	@echo "$(CG)cluster created$(CEND)"

.PHONY: k8s-cluster-delete
k8s-cluster-delete: ## Delete Kind cluster
	@echo "$(CR)deleting kind cluster$(CEND)"
	@kind delete cluster --name=$(CLUSTER_NAME)
	@echo "$(CG)cluster deleted$(CEND)"

.PHONY: k8s-build
k8s-build: ## Build Docker image for Kubernetes
	@echo "$(CC)building docker image$(CEND)"
	@docker build -t $(K8S_IMAGE_NAME):$(K8S_IMAGE_TAG) -f cmd/rest-api/Dockerfile.rest-api .
	@echo "$(CG)image built$(CEND): $(K8S_IMAGE_NAME):$(K8S_IMAGE_TAG)"

.PHONY: k8s-load
k8s-load: k8s-build ## Load Docker image into Kind cluster
	@echo "$(CC)loading image into cluster$(CEND)"
	@kind load docker-image $(K8S_IMAGE_NAME):$(K8S_IMAGE_TAG) --name=$(CLUSTER_NAME)
	@echo "$(CG)image loaded$(CEND)"

.PHONY: k8s-apply
k8s-apply: ## Apply all Kubernetes manifests
	@echo "$(CC)applying manifests$(CEND)"
	@kubectl apply -f k8s/base/
	@echo "$(CG)manifests applied$(CEND)"

.PHONY: k8s-status
k8s-status: ## Check deployment status
	@echo "$(CC)checking deployment status$(CEND)"
	@kubectl get all -n $(NAMESPACE)
	@kubectl get pvc -n $(NAMESPACE)
	@kubectl get hpa -n $(NAMESPACE)

.PHONY: k8s-logs
k8s-logs: ## Tail logs from REST API pods
	@kubectl logs -n $(NAMESPACE) -l app=replay-rest-api --tail=100 -f

.PHONY: k8s-test
k8s-test: ## Run smoke tests
	@echo "$(CC)running smoke tests$(CEND)"
	@./scripts/smoke-test.sh blue
	@echo "$(CG)tests passed$(CEND)"

.PHONY: deploy
deploy: k8s-cluster-create k8s-load k8s-apply ## Single command to deploy everything
	@echo ""
	@echo "$(CC)waiting for pods to be ready$(CEND)"
	@kubectl wait --for=condition=ready pod -l app=mongodb -n $(NAMESPACE) --timeout=300s
	@kubectl wait --for=condition=ready pod -l app=replay-rest-api -n $(NAMESPACE) --timeout=300s
	@echo ""
	@echo "$(CG)deployment complete$(CEND)"
	@echo ""
	@echo "status:"
	@make k8s-status
	@echo ""
	@echo "service endpoints:"
	@kubectl get svc -n $(NAMESPACE) replay-rest-api-service
	@echo ""
	@echo "useful commands:"
	@echo "  - view logs: make k8s-logs"
	@echo "  - run tests: make k8s-test"
	@echo "  - check status: make k8s-status"
	@echo "  - teardown: make k8s-cluster-delete"

.PHONY: redeploy
redeploy: k8s-load ## Redeploy with new image (fast)
	@echo "$(CC)redeploying with new image$(CEND)"
	@kubectl rollout restart deployment/replay-rest-api -n $(NAMESPACE)
	@kubectl rollout status deployment/replay-rest-api -n $(NAMESPACE)
	@echo "$(CG)redeployment complete$(CEND)"

.PHONY: k8s-blue-green-deploy
k8s-blue-green-deploy: ## Deploy using blue-green strategy
	@echo "$(CC)starting blue-green deployment$(CEND)"
	@./scripts/blue-green-deploy.sh $(K8S_IMAGE_TAG)

.PHONY: k8s-rollback
k8s-rollback: ## Rollback to previous deployment
	@echo "$(CR)rolling back$(CEND)"
	@kubectl rollout undo deployment/replay-rest-api -n $(NAMESPACE)
	@kubectl rollout status deployment/replay-rest-api -n $(NAMESPACE)
	@echo "$(CG)rollback complete$(CEND)"

.PHONY: k8s-scale
k8s-scale: ## Scale deployment (usage: make k8s-scale REPLICAS=5)
	@echo "$(CC)scaling to $(REPLICAS) replicas$(CEND)"
	@kubectl scale deployment/replay-rest-api -n $(NAMESPACE) --replicas=$(REPLICAS)
	@echo "$(CG)scaled$(CEND)"

.PHONY: k8s-shell
k8s-shell: ## Open shell in REST API pod
	@kubectl exec -it -n $(NAMESPACE) $$(kubectl get pod -n $(NAMESPACE) -l app=replay-rest-api -o jsonpath='{.items[0].metadata.name}') -- /bin/sh

.PHONY: k8s-port-forward
k8s-port-forward: ## Port forward to REST API (localhost:8080)
	@echo "$(CC)port forwarding to localhost:8080$(CEND)"
	@kubectl port-forward -n $(NAMESPACE) svc/replay-rest-api-service 8080:80

.PHONY: k8s-clean
k8s-clean: ## Clean up Kubernetes resources
	@echo "$(CR)cleaning kubernetes resources$(CEND)"
	@kubectl delete namespace $(NAMESPACE) --ignore-not-found=true
	@echo "$(CG)cleaned$(CEND)"

.PHONY: test
test: k8s-test ## Run all tests

#========================================
# 🎮 Developer Experience (Award-Winning DX)
#========================================

.PHONY: dx-setup
dx-setup: ## One-command developer setup - installs all tools and dependencies
	@echo "$(CG)🎮 LeetGaming PRO - Developer Experience Setup$(CEND)"
	@echo ""
	@echo "$(CC)Step 1: Checking Go installation...$(CEND)"
	@go version || { echo "$(CR)Go is not installed. Please install Go 1.21+$(CEND)"; exit 1; }
	@echo ""
	@echo "$(CC)Step 2: Installing development tools...$(CEND)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/google/go-licenses@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/vektra/mockery/v2@latest
	@echo ""
	@echo "$(CC)Step 3: Verifying dependencies...$(CEND)"
	@go mod download
	@go mod verify
	@echo ""
	@echo "$(CC)Step 4: Building project...$(CEND)"
	@go build ./...
	@echo ""
	@echo "$(CG)✅ Developer environment ready!$(CEND)"
	@echo ""
	@echo "$(CC)Quick start commands:$(CEND)"
	@echo "  make up        - Start full development environment"
	@echo "  make dev       - Start API in watch mode"
	@echo "  make test-unit - Run unit tests"
	@echo "  make docs      - Open API documentation"
	@echo ""

.PHONY: dev
dev: ## Start API in development mode with hot reload
	@echo "$(CC)Starting API in development mode...$(CEND)"
	@echo "$(CC)API will be available at http://localhost:8080$(CEND)"
	@echo "$(CC)Press Ctrl+C to stop$(CEND)"
	@DEV_ENV=true go run ./cmd/rest-api/main.go

.PHONY: dev-watch
dev-watch: ## Start API with file watcher (requires air)
	@command -v air >/dev/null 2>&1 || { echo "Installing air..."; go install github.com/cosmtrek/air@latest; }
	@air -c .air.toml

#========================================
# 📚 API Documentation
#========================================

.PHONY: docs
docs: ## Open API documentation in browser
	@echo "$(CC)Opening API Documentation...$(CEND)"
	@echo "  Swagger UI: http://localhost:8080/api/docs/swagger"
	@echo "  ReDoc:      http://localhost:8080/api/docs/redoc"
	@echo "  OpenAPI:    http://localhost:8080/api/docs/openapi.yaml"
	@open http://localhost:8080/api/docs || xdg-open http://localhost:8080/api/docs 2>/dev/null || echo "Open http://localhost:8080/api/docs in your browser"

.PHONY: docs-generate
docs-generate: ## Generate OpenAPI spec from code annotations
	@echo "$(CC)Generating OpenAPI documentation...$(CEND)"
	@swag init -g cmd/rest-api/main.go -o docs/swagger --parseDependency --parseInternal
	@cp docs/swagger/openapi.yaml cmd/rest-api/docs/openapi.yaml
	@echo "$(CG)Documentation generated!$(CEND)"

.PHONY: docs-validate
docs-validate: ## Validate OpenAPI spec
	@echo "$(CC)Validating OpenAPI specification...$(CEND)"
	@command -v swagger-cli >/dev/null 2>&1 || npm install -g @apidevtools/swagger-cli
	@swagger-cli validate docs/swagger/openapi.yaml
	@echo "$(CG)OpenAPI spec is valid!$(CEND)"

#========================================
# 🧪 Testing & Quality
#========================================

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "$(CC)Running unit tests...$(CEND)"
	@go test -short -race -v ./pkg/... 2>&1 | grep -E "(PASS|FAIL|---)" || go test -short -race ./pkg/...
	@echo "$(CG)Unit tests complete!$(CEND)"

.PHONY: test-integration
test-integration: ## Run integration tests (requires running services)
	@echo "$(CC)Running integration tests...$(CEND)"
	@go test -tags=integration -race -v ./test/integration/...
	@echo "$(CG)Integration tests complete!$(CEND)"

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "$(CC)Running E2E tests...$(CEND)"
	@go test -tags=e2e -race -v ./test/integration/...
	@echo "$(CG)E2E tests complete!$(CEND)"

.PHONY: test-all
test-all: test-unit test-integration test-e2e ## Run all tests
	@echo "$(CG)All tests complete!$(CEND)"

.PHONY: coverage
coverage: ## Generate test coverage report
	@echo "$(CC)Generating coverage report...$(CEND)"
	@go test -coverprofile=coverage.out -covermode=atomic ./pkg/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(CG)Coverage report: coverage.html$(CEND)"
	@go tool cover -func=coverage.out | tail -1

.PHONY: bench
bench: ## Run benchmarks
	@echo "$(CC)Running benchmarks...$(CEND)"
	@go test -bench=. -benchmem ./pkg/...

#========================================
# 🔍 Linting & Static Analysis
#========================================

.PHONY: lint
lint: ## Run linters
	@echo "$(CC)Running linters...$(CEND)"
	@golangci-lint run --timeout 5m ./...
	@echo "$(CG)Linting complete!$(CEND)"

.PHONY: lint-fix
lint-fix: ## Run linters and auto-fix issues
	@echo "$(CC)Fixing lint issues...$(CEND)"
	@golangci-lint run --fix ./...
	@goimports -w .
	@echo "$(CG)Fixes applied!$(CEND)"

.PHONY: fmt
fmt: ## Format code
	@echo "$(CC)Formatting code...$(CEND)"
	@gofmt -s -w .
	@goimports -w .
	@echo "$(CG)Formatting complete!$(CEND)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(CC)Running go vet...$(CEND)"
	@go vet ./...
	@echo "$(CG)Vet complete!$(CEND)"

.PHONY: security
security: ## Run security scanners
	@echo "$(CC)Running security scan...$(CEND)"
	@command -v gosec >/dev/null 2>&1 || go install github.com/securego/gosec/v2/cmd/gosec@latest
	@gosec -quiet ./...
	@echo "$(CG)Security scan complete!$(CEND)"

#========================================
# 🔧 Code Generation
#========================================

.PHONY: generate
generate: ## Run all code generators
	@echo "$(CC)Running code generators...$(CEND)"
	@go generate ./...
	@echo "$(CG)Generation complete!$(CEND)"

.PHONY: mocks
mocks: ## Generate test mocks
	@echo "$(CC)Generating mocks...$(CEND)"
	@mockery --all --keeptree --output=./test/mocks
	@echo "$(CG)Mocks generated!$(CEND)"

#========================================
# 📦 Build & Release
#========================================

.PHONY: build-all
build-all: ## Build all binaries
	@echo "$(CC)Building all binaries...$(CEND)"
	@CGO_ENABLED=0 go build -o bin/rest-api ./cmd/rest-api
	@CGO_ENABLED=0 go build -o bin/event-processor ./cmd/event-processor
	@CGO_ENABLED=0 go build -o bin/seed ./cmd/cli/seed
	@echo "$(CG)Binaries built in ./bin/$(CEND)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(CC)Cleaning...$(CEND)"
	@rm -rf bin/ coverage.out coverage.html
	@go clean -cache -testcache
	@echo "$(CG)Cleaned!$(CEND)"

#========================================
# 🗄️ Database
#========================================

.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "$(CC)Running migrations...$(CEND)"
	@go run ./cmd/cli/migrate/main.go up
	@echo "$(CG)Migrations complete!$(CEND)"

.PHONY: db-seed
db-seed: seed ## Alias for seed

.PHONY: db-reset
db-reset: ## Reset database (WARNING: destroys data)
	@echo "$(CR)WARNING: This will destroy all data!$(CEND)"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	@echo "$(CC)Resetting database...$(CEND)"
	@go run ./cmd/cli/migrate/main.go down
	@go run ./cmd/cli/migrate/main.go up
	@make seed
	@echo "$(CG)Database reset complete!$(CEND)"

#========================================
# 📊 Monitoring & Debugging
#========================================

.PHONY: pprof
pprof: ## Open pprof profiler (requires running API)
	@echo "$(CC)Opening pprof...$(CEND)"
	@go tool pprof http://localhost:8080/debug/pprof/profile

.PHONY: trace
trace: ## Capture and view execution trace
	@echo "$(CC)Capturing trace...$(CEND)"
	@curl -o trace.out http://localhost:8080/debug/pprof/trace?seconds=5
	@go tool trace trace.out

#========================================
# 📋 Project Info
#========================================

.PHONY: version
version: ## Show version information
	@echo "$(CC)LeetGaming PRO API$(CEND)"
	@echo "  Go Version:    $$(go version | cut -d' ' -f3)"
	@echo "  Build Time:    $$(date -u '+%Y-%m-%d %H:%M:%S UTC')"
	@echo "  Git Commit:    $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "  Git Branch:    $$(git branch --show-current 2>/dev/null || echo 'unknown')"

.PHONY: deps
deps: ## Show dependency tree
	@echo "$(CC)Dependencies:$(CEND)"
	@go list -m all

.PHONY: outdated
outdated: ## Check for outdated dependencies
	@echo "$(CC)Checking for outdated dependencies...$(CEND)"
	@go list -u -m all | grep -v '^\s*$$'
