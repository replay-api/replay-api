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
	@echo "$(CC)Step 1/7: Creating Kind cluster...$(CEND)"
	@kind create cluster --config=kind-config.yaml --name=leetgaming-local 2>/dev/null || echo "Cluster already exists"
	@kubectl cluster-info --context kind-leetgaming-local
	@echo ""
	@echo "$(CC)Step 2/7: Building API image...$(CEND)"
	@docker build -t replay-api:latest -f cmd/rest-api/Dockerfile.rest-api . || { echo "$(CR)Failed to build API image$(CEND)"; exit 1; }
	@kind load docker-image replay-api:latest --name=leetgaming-local
	@echo ""
	@echo "$(CC)Step 3/7: Building Web Frontend image...$(CEND)"
	@if [ -d "../leetgaming-pro-web" ]; then \
		docker build -t leetgaming-web:latest ../leetgaming-pro-web || echo "$(CR)Web frontend build skipped (optional)$(CEND)"; \
		kind load docker-image leetgaming-web:latest --name=leetgaming-local 2>/dev/null || true; \
	else \
		echo "Web frontend directory not found, skipping..."; \
	fi
	@echo ""
	@echo "$(CC)Step 4/7: Applying Kubernetes manifests...$(CEND)"
	@kubectl apply -k k8s/local/
	@echo ""
	@echo "$(CC)Step 5/7: Waiting for pods to be ready...$(CEND)"
	@kubectl wait --for=condition=ready pod -l app=mongodb -n leetgaming --timeout=300s 2>/dev/null || true
	@kubectl wait --for=condition=ready pod -l app=replay-api -n leetgaming --timeout=300s 2>/dev/null || true
	@kubectl wait --for=condition=ready pod -l app=web-frontend -n leetgaming --timeout=300s 2>/dev/null || true
	@kubectl wait --for=condition=ready pod -l app=prometheus -n leetgaming --timeout=300s 2>/dev/null || true
	@kubectl wait --for=condition=ready pod -l app=grafana -n leetgaming --timeout=300s 2>/dev/null || true
	@echo ""
	@echo "$(CC)Step 6/7: Seeding database...$(CEND)"
	@sleep 10
	@kubectl delete job database-seed -n leetgaming 2>/dev/null || true
	@kubectl apply -f k8s/local/seed-job.yaml
	@kubectl wait --for=condition=complete job/database-seed -n leetgaming --timeout=120s 2>/dev/null || kubectl logs -l job-name=database-seed -n leetgaming
	@echo ""
	@echo "$(CC)Step 7/7: Starting port forwards...$(CEND)"
	@pkill -f "port-forward" 2>/dev/null || true
	@sleep 1
	@kubectl port-forward svc/replay-api-service 8080:8080 -n leetgaming &
	@kubectl port-forward svc/web-frontend-service 3030:3030 -n leetgaming 2>/dev/null &
	@kubectl port-forward svc/prometheus-service 9090:9090 -n leetgaming 2>/dev/null &
	@kubectl port-forward svc/grafana-service 3000:3000 -n leetgaming 2>/dev/null &
	@echo ""
	@echo "$(CG)✅ Development environment is ready!$(CEND)"
	@echo ""
	@echo "$(CC)Available endpoints:$(CEND)"
	@echo "  - API:         http://localhost:8080"
	@echo "  - Health:      http://localhost:8080/health"
	@echo "  - Web:         http://localhost:3030"
	@echo "  - Prometheus:  http://localhost:9090"
	@echo "  - Grafana:     http://localhost:3000 (admin/admin)"
	@echo ""
	@echo "$(CC)Useful commands:$(CEND)"
	@echo "  - Stop environment:    make down"
	@echo "  - View API logs:       kubectl logs -f -l app=replay-api -n leetgaming"
	@echo "  - View Web logs:       kubectl logs -f -l app=web-frontend -n leetgaming"
	@echo "  - Run seeds again:     make seed"
	@echo "  - Check status:        kubectl get pods -n leetgaming"

down: ## Stop and clean up the development environment
	@echo "$(CR)Stopping LeetGaming PRO Development Environment$(CEND)"
	@pkill -f "port-forward" 2>/dev/null || true
	@kind delete cluster --name=leetgaming-local 2>/dev/null || true
	@echo "$(CG)✅ Environment stopped$(CEND)"

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
