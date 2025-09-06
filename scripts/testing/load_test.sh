#!/bin/bash

# CloudBridge Relay - Load Testing Script
# Этот скрипт выполняет нагрузочное тестирование системы

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="${NAMESPACE:-cloudbridge}"
TEST_DURATION="${TEST_DURATION:-300}"  # 5 minutes
CONCURRENT_USERS="${CONCURRENT_USERS:-100}"
RAMP_UP_TIME="${RAMP_UP_TIME:-60}"     # 1 minute
TARGET_URL="${TARGET_URL:-http://cloudbridge-relay-service:9090}"
RESULTS_DIR="${RESULTS_DIR:-./test-results}"

# Test scenarios
SCENARIOS=(
    "health_check"
    "api_endpoints"
    "quic_connections"
    "wireguard_tunnels"
    "mixed_workload"
    "stress_test"
)

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_section() {
    echo -e "${PURPLE}[SECTION]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    for tool in kubectl curl jq; do
        if ! command -v "$tool" &> /dev/null; then
            missing_deps+=("$tool")
        fi
    done
    
    # Check if k6 is available
    if ! command -v k6 &> /dev/null; then
        log_warning "k6 not found, will install it"
        install_k6
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        exit 1
    fi
    
    log_success "Dependencies check passed"
}

# Install k6
install_k6() {
    log_info "Installing k6..."
    
    # Detect OS
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    
    case $os in
        linux)
            sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
            echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
            sudo apt-get update
            sudo apt-get install k6
            ;;
        darwin)
            brew install k6
            ;;
        *)
            log_error "Unsupported OS: $os"
            exit 1
            ;;
    esac
    
    log_success "k6 installed successfully"
}

# Create results directory
create_results_dir() {
    log_info "Creating results directory..."
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local test_dir="$RESULTS_DIR/load-test-$timestamp"
    
    mkdir -p "$test_dir"
    
    echo "$test_dir"
}

# Check system health before testing
check_system_health() {
    log_section "Checking system health before testing..."
    
    # Check if services are running
    local deployments=$(kubectl get deployments -n "$NAMESPACE" --no-headers | wc -l)
    local ready_deployments=$(kubectl get deployments -n "$NAMESPACE" --no-headers | grep "1/1" | wc -l)
    
    log_info "Deployments: $ready_deployments/$deployments ready"
    
    if [ "$ready_deployments" -ne "$deployments" ]; then
        log_error "Not all deployments are ready"
        kubectl get deployments -n "$NAMESPACE"
        exit 1
    fi
    
    # Check pod status
    local pods=$(kubectl get pods -n "$NAMESPACE" --no-headers | wc -l)
    local running_pods=$(kubectl get pods -n "$NAMESPACE" --no-headers | grep "Running" | wc -l)
    
    log_info "Pods: $running_pods/$pods running"
    
    if [ "$running_pods" -ne "$pods" ]; then
        log_error "Not all pods are running"
        kubectl get pods -n "$NAMESPACE"
        exit 1
    fi
    
    # Test basic connectivity
    log_info "Testing basic connectivity..."
    if ! kubectl run connectivity-test --image=curlimages/curl --rm -i --restart=Never -- \
        curl -f "$TARGET_URL/health" >/dev/null 2>&1; then
        log_error "Cannot reach target URL: $TARGET_URL"
        exit 1
    fi
    
    log_success "System health check passed"
}

# Create k6 test script for health checks
create_health_check_script() {
    local script_file="$1"
    
    cat > "$script_file" <<'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export let options = {
  stages: [
    { duration: '1m', target: 10 },
    { duration: '2m', target: 50 },
    { duration: '2m', target: 100 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'],
    http_req_failed: ['rate<0.1'],
    errors: ['rate<0.1'],
  },
};

export default function() {
  let response = http.get(`${__ENV.TARGET_URL}/health`);
  
  let result = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
    'response has body': (r) => r.body.length > 0,
  });
  
  errorRate.add(!result);
  sleep(1);
}
EOF
}

# Create k6 test script for API endpoints
create_api_test_script() {
    local script_file="$1"
    
    cat > "$script_file" <<'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export let options = {
  stages: [
    { duration: '1m', target: 20 },
    { duration: '3m', target: 100 },
    { duration: '2m', target: 200 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.05'],
  },
};

const endpoints = [
  '/health',
  '/metrics',
  '/api/v1/status',
  '/api/v1/tenants',
];

export default function() {
  let endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  let response = http.get(`${__ENV.TARGET_URL}${endpoint}`);
  
  let result = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
    'response has body': (r) => r.body.length > 0,
  });
  
  errorRate.add(!result);
  sleep(0.5);
}
EOF
}

# Create k6 test script for mixed workload
create_mixed_workload_script() {
    local script_file="$1"
    
    cat > "$script_file" <<'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export let options = {
  stages: [
    { duration: '2m', target: 50 },
    { duration: '5m', target: 150 },
    { duration: '3m', target: 300 },
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed: ['rate<0.1'],
    errors: ['rate<0.1'],
  },
};

const scenarios = [
  { name: 'health_check', weight: 30, endpoint: '/health' },
  { name: 'metrics', weight: 20, endpoint: '/metrics' },
  { name: 'api_status', weight: 25, endpoint: '/api/v1/status' },
  { name: 'api_tenants', weight: 25, endpoint: '/api/v1/tenants' },
];

export default function() {
  let scenario = scenarios[Math.floor(Math.random() * scenarios.length)];
  let response = http.get(`${__ENV.TARGET_URL}${scenario.endpoint}`);
  
  let result = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 1000ms': (r) => r.timings.duration < 1000,
    'response has body': (r) => r.body.length > 0,
  });
  
  errorRate.add(!result);
  sleep(Math.random() * 2);
}
EOF
}

# Create k6 test script for stress test
create_stress_test_script() {
    local script_file="$1"
    
    cat > "$script_file" <<'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export let options = {
  stages: [
    { duration: '1m', target: 100 },
    { duration: '2m', target: 500 },
    { duration: '3m', target: 1000 },
    { duration: '2m', target: 1500 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.2'],
    errors: ['rate<0.2'],
  },
};

export default function() {
  let response = http.get(`${__ENV.TARGET_URL}/health`);
  
  let result = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 2000ms': (r) => r.timings.duration < 2000,
  });
  
  errorRate.add(!result);
  sleep(0.1);
}
EOF
}

# Run k6 test
run_k6_test() {
    local test_name="$1"
    local script_file="$2"
    local results_dir="$3"
    
    log_section "Running $test_name test..."
    
    local output_file="$results_dir/${test_name}-results.json"
    local summary_file="$results_dir/${test_name}-summary.txt"
    
    # Run k6 test
    k6 run \
        --env TARGET_URL="$TARGET_URL" \
        --out json="$output_file" \
        --summary-export="$summary_file" \
        "$script_file"
    
    # Parse results
    if [ -f "$summary_file" ]; then
        log_info "Test results for $test_name:"
        cat "$summary_file"
    fi
    
    log_success "$test_name test completed"
}

# Monitor system during test
monitor_system() {
    local test_name="$1"
    local results_dir="$2"
    local duration="$3"
    
    log_info "Starting system monitoring for $test_name..."
    
    local monitor_file="$results_dir/${test_name}-monitoring.log"
    
    # Start monitoring in background
    (
        echo "Timestamp,CPU_Usage,Memory_Usage,Pods_Count,Deployments_Ready" > "$monitor_file"
        
        local end_time=$(($(date +%s) + duration))
        while [ $(date +%s) -lt $end_time ]; do
            local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
            local cpu_usage=$(kubectl top nodes --no-headers 2>/dev/null | awk '{sum+=$2} END {print sum/NR}' || echo "N/A")
            local memory_usage=$(kubectl top nodes --no-headers 2>/dev/null | awk '{sum+=$4} END {print sum/NR}' || echo "N/A")
            local pods_count=$(kubectl get pods -n "$NAMESPACE" --no-headers | wc -l)
            local deployments_ready=$(kubectl get deployments -n "$NAMESPACE" --no-headers | grep "1/1" | wc -l)
            
            echo "$timestamp,$cpu_usage,$memory_usage,$pods_count,$deployments_ready" >> "$monitor_file"
            sleep 10
        done
    ) &
    
    local monitor_pid=$!
    echo $monitor_pid
}

# Generate test report
generate_test_report() {
    local results_dir="$1"
    local test_duration="$2"
    
    log_section "Generating test report..."
    
    local report_file="$results_dir/load-test-report.txt"
    
    cat > "$report_file" <<EOF
CloudBridge Relay Load Test Report
==================================

Test Date: $(date)
Test Duration: $test_duration seconds
Target URL: $TARGET_URL
Namespace: $NAMESPACE

Test Scenarios Executed:
EOF
    
    # Add results for each test
    for scenario in "${SCENARIOS[@]}"; do
        local summary_file="$results_dir/${scenario}-summary.txt"
        if [ -f "$summary_file" ]; then
            echo "" >> "$report_file"
            echo "=== $scenario ===" >> "$report_file"
            cat "$summary_file" >> "$report_file"
        fi
    done
    
    # Add system monitoring summary
    echo "" >> "$report_file"
    echo "=== System Monitoring ===" >> "$report_file"
    for scenario in "${SCENARIOS[@]}"; do
        local monitor_file="$results_dir/${scenario}-monitoring.log"
        if [ -f "$monitor_file" ]; then
            echo "" >> "$report_file"
            echo "--- $scenario monitoring ---" >> "$report_file"
            tail -5 "$monitor_file" >> "$report_file"
        fi
    done
    
    # Add recommendations
    cat >> "$report_file" <<EOF

Recommendations:
- Monitor system performance under sustained load
- Consider horizontal scaling if response times exceed thresholds
- Review resource limits and requests based on actual usage
- Implement circuit breakers for better fault tolerance
- Set up alerting for performance degradation

EOF
    
    log_success "Test report generated: $report_file"
    cat "$report_file"
}

# Main function
main() {
    log_info "Starting CloudBridge Relay load testing..."
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --scenario)
                SCENARIOS=("$2")
                shift 2
                ;;
            --duration)
                TEST_DURATION="$2"
                shift 2
                ;;
            --users)
                CONCURRENT_USERS="$2"
                shift 2
                ;;
            --url)
                TARGET_URL="$2"
                shift 2
                ;;
            --namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Execute load testing
    check_dependencies
    local results_dir=$(create_results_dir)
    check_system_health
    
    local monitor_pids=()
    
    # Run each test scenario
    for scenario in "${SCENARIOS[@]}"; do
        local script_file="$results_dir/${scenario}.js"
        
        # Create test script
        case $scenario in
            health_check)
                create_health_check_script "$script_file"
                ;;
            api_endpoints)
                create_api_test_script "$script_file"
                ;;
            mixed_workload)
                create_mixed_workload_script "$script_file"
                ;;
            stress_test)
                create_stress_test_script "$script_file"
                ;;
            *)
                log_warning "Unknown scenario: $scenario, skipping"
                continue
                ;;
        esac
        
        # Start monitoring
        local monitor_pid=$(monitor_system "$scenario" "$results_dir" "$TEST_DURATION")
        monitor_pids+=($monitor_pid)
        
        # Run test
        run_k6_test "$scenario" "$script_file" "$results_dir"
        
        # Stop monitoring
        kill $monitor_pid 2>/dev/null || true
    done
    
    # Wait for all monitoring to complete
    for pid in "${monitor_pids[@]}"; do
        wait $pid 2>/dev/null || true
    done
    
    # Generate report
    generate_test_report "$results_dir" "$TEST_DURATION"
    
    log_success "Load testing completed successfully!"
    log_info ""
    log_info "Results directory: $results_dir"
    log_info "Test scenarios: ${SCENARIOS[*]}"
    log_info "Total duration: $TEST_DURATION seconds"
}

# Run main function
main "$@"

