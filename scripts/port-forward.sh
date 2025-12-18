#!/bin/bash
# Port forwarding script for LeetGaming PRO development environment
# This script ensures all services are accessible via localhost

set -e

NAMESPACE="leetgaming"
PID_FILE="/tmp/leetgaming-port-forwards.pids"
LOG_DIR="/tmp/leetgaming-logs"

# Colors for output
GREEN='\033[0;32m'
CYAN='\033[0;36m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Service definitions: service_name:local_port:target_port
SERVICES=(
    "replay-api-service:8080:8080"
    "web-frontend-service:3030:3030"
    "grafana-service:3031:3000"
    "prometheus-service:9090:9090"
    "pyroscope-service:3041:4040"
    "leetgaming-kafka-kafka-bootstrap:9092:9092"
)

usage() {
    echo "Usage: $0 {start|stop|status|restart}"
    echo ""
    echo "Commands:"
    echo "  start   - Start all port forwards"
    echo "  stop    - Stop all port forwards"
    echo "  status  - Check status of port forwards and services"
    echo "  restart - Stop and start all port forwards"
    exit 1
}

log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

check_cluster() {
    if ! kubectl cluster-info --context kind-leetgaming-local &>/dev/null; then
        log_error "Kind cluster 'leetgaming-local' not found or not accessible"
        echo "Run 'make up' to create the development environment first"
        exit 1
    fi
}

check_namespace() {
    if ! kubectl get namespace "$NAMESPACE" &>/dev/null; then
        log_error "Namespace '$NAMESPACE' not found"
        echo "Run 'make up' to create the development environment first"
        exit 1
    fi
}

stop_port_forwards() {
    log_info "Stopping existing port forwards..."

    # Kill by PID file
    if [ -f "$PID_FILE" ]; then
        while read -r pid; do
            if kill -0 "$pid" 2>/dev/null; then
                kill "$pid" 2>/dev/null || true
            fi
        done < "$PID_FILE"
        rm -f "$PID_FILE"
    fi

    # Also kill any stray port-forward processes for our services
    pkill -f "kubectl port-forward.*-n $NAMESPACE" 2>/dev/null || true

    sleep 1
    log_success "Port forwards stopped"
}

start_port_forward() {
    local service=$1
    local local_port=$2
    local target_port=$3

    # Check if service exists
    if ! kubectl get svc "$service" -n "$NAMESPACE" &>/dev/null; then
        log_warn "Service $service not found, skipping..."
        return 1
    fi

    # Check if port is already in use
    if lsof -i ":$local_port" &>/dev/null; then
        log_warn "Port $local_port already in use (existing port forward may still be active)"
        return 0  # Return success since the port forward is already active
    fi

    # Start port forward with nohup to persist
    mkdir -p "$LOG_DIR"
    nohup kubectl port-forward "svc/$service" "$local_port:$target_port" -n "$NAMESPACE" \
        > "$LOG_DIR/${service}.log" 2>&1 &

    local pid=$!
    echo "$pid" >> "$PID_FILE"

    # Wait a bit and check if it's still running
    sleep 1
    if kill -0 "$pid" 2>/dev/null; then
        log_success "$service -> localhost:$local_port"
        return 0
    else
        log_error "Failed to start port forward for $service"
        return 1
    fi
}

start_port_forwards() {
    log_info "Starting port forwards..."

    check_cluster
    check_namespace

    # Clean up any existing port forwards first
    stop_port_forwards

    mkdir -p "$LOG_DIR"
    > "$PID_FILE"  # Clear PID file

    local success_count=0
    local total_count=${#SERVICES[@]}

    for service_def in "${SERVICES[@]}"; do
        IFS=':' read -r service local_port target_port <<< "$service_def"
        if start_port_forward "$service" "$local_port" "$target_port"; then
            ((success_count++))
        fi
    done

    echo ""
    log_info "Started $success_count/$total_count port forwards"
    echo ""

    # Wait for services to be accessible
    sleep 2
    verify_services
}

verify_services() {
    log_info "Verifying service accessibility..."
    echo ""

    local all_ok=true

    # Test API
    local api_status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health 2>/dev/null || echo "000")
    if [ "$api_status" = "200" ]; then
        log_success "API:        http://localhost:8080 (HTTP $api_status)"
    else
        log_error "API:        http://localhost:8080 (HTTP $api_status)"
        all_ok=false
    fi

    # Test Web Frontend
    local web_status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3030 2>/dev/null || echo "000")
    if [ "$web_status" = "200" ]; then
        log_success "Web:        http://localhost:3030 (HTTP $web_status)"
    else
        log_error "Web:        http://localhost:3030 (HTTP $web_status)"
        all_ok=false
    fi

    # Test Grafana
    local grafana_status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3031/api/health 2>/dev/null || echo "000")
    if [ "$grafana_status" = "200" ]; then
        log_success "Grafana:    http://localhost:3031 (HTTP $grafana_status)"
    else
        log_warn "Grafana:    http://localhost:3031 (HTTP $grafana_status)"
    fi

    # Test Prometheus
    local prom_status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:9090/-/healthy 2>/dev/null || echo "000")
    if [ "$prom_status" = "200" ]; then
        log_success "Prometheus: http://localhost:9090 (HTTP $prom_status)"
    else
        log_warn "Prometheus: http://localhost:9090 (HTTP $prom_status)"
    fi

    # Test Pyroscope
    local pyro_status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3041/ready 2>/dev/null || echo "000")
    if [ "$pyro_status" = "200" ]; then
        log_success "Pyroscope:  http://localhost:3041 (HTTP $pyro_status)"
    else
        log_warn "Pyroscope:  http://localhost:3041 (HTTP $pyro_status)"
    fi

    echo ""

    if [ "$all_ok" = true ]; then
        log_success "All critical services are accessible!"
    else
        log_warn "Some services may not be fully accessible yet"
        echo "Wait a few seconds and try again: $0 status"
    fi
}

show_status() {
    log_info "Port Forward Status"
    echo ""

    check_cluster
    check_namespace

    # Show running port forwards
    local running=0
    if [ -f "$PID_FILE" ]; then
        while read -r pid; do
            if kill -0 "$pid" 2>/dev/null; then
                ((running++))
            fi
        done < "$PID_FILE"
    fi

    echo "Running port forwards: $running"
    echo ""

    # Verify services
    verify_services

    echo ""
    log_info "Pod Status"
    kubectl get pods -n "$NAMESPACE" --no-headers | while read -r line; do
        name=$(echo "$line" | awk '{print $1}')
        ready=$(echo "$line" | awk '{print $2}')
        status=$(echo "$line" | awk '{print $3}')

        if [ "$status" = "Running" ]; then
            echo -e "  ${GREEN}●${NC} $name ($ready ready)"
        elif [ "$status" = "CrashLoopBackOff" ] || [ "$status" = "Error" ]; then
            echo -e "  ${RED}●${NC} $name ($status)"
        else
            echo -e "  ${YELLOW}●${NC} $name ($status)"
        fi
    done
}

# Main script logic
case "${1:-}" in
    start)
        start_port_forwards
        ;;
    stop)
        stop_port_forwards
        ;;
    restart)
        stop_port_forwards
        start_port_forwards
        ;;
    status)
        show_status
        ;;
    *)
        usage
        ;;
esac
