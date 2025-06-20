#!/bin/bash

# Integration Test Runner for Temporal Flow Demo
# This script runs comprehensive integration tests against the complete system

set -e

echo "🚀 Starting Integration Tests for Temporal Flow Demo"
echo "=================================================="

# Configuration
API_GATEWAY_URL=${API_GATEWAY_URL:-"http://api-gateway-test:4000"}
SVC_BALANCE_URL=${SVC_BALANCE_URL:-"http://svc-balance-test:4000"}
SVC_TRANSACTION_URL=${SVC_TRANSACTION_URL:-"http://svc-transaction-test:4001"}
RESULTS_DIR="/app/results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create results directory
mkdir -p $RESULTS_DIR

# Test result tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a $RESULTS_DIR/test_run_$TIMESTAMP.log
}

# Test function
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log "🧪 Running test: $test_name"
    
    if eval "$test_command"; then
        log "✅ PASSED: $test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        log "❌ FAILED: $test_name"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Wait for services to be ready
wait_for_service() {
    local service_url="$1"
    local service_name="$2"
    local max_attempts=30
    local attempt=1
    
    log "⏳ Waiting for $service_name to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -sf "$service_url/health" > /dev/null 2>&1; then
            log "✅ $service_name is ready"
            return 0
        fi
        
        log "🔄 Attempt $attempt/$max_attempts - $service_name not ready yet..."
        sleep 5
        attempt=$((attempt + 1))
    done
    
    log "❌ $service_name failed to become ready after $max_attempts attempts"
    return 1
}

# Health check tests
test_health_checks() {
    log "🏥 Testing service health checks..."
    
    run_test "API Gateway Health Check" \
        "curl -sf $API_GATEWAY_URL/health | jq -e '.status == \"healthy\"'"
    
    run_test "Balance Service Health Check" \
        "curl -sf $SVC_BALANCE_URL/health | jq -e '.status == \"healthy\"'"
    
    run_test "Transaction Service Health Check" \
        "curl -sf $SVC_TRANSACTION_URL/health | jq -e '.status == \"healthy\"'"
}

# Basic transfer tests
test_basic_transfers() {
    log "💸 Testing basic money transfers..."
    
    # Test successful transfer
    run_test "Successful Transfer" \
        "curl -sf -X POST $API_GATEWAY_URL/transfer \
         -H 'Content-Type: application/json' \
         -d '{
           \"from_account\": \"1000000001\",
           \"to_account\": \"1000000002\",
           \"amount\": 100.50,
           \"currency\": \"USD\",
           \"description\": \"Integration test transfer\"
         }' | jq -e '.status == \"TRANSFER_STATUS_PENDING\"'"
    
    # Test insufficient funds
    run_test "Insufficient Funds Transfer" \
        "curl -sf -X POST $API_GATEWAY_URL/transfer \
         -H 'Content-Type: application/json' \
         -d '{
           \"from_account\": \"1000000003\",
           \"to_account\": \"1000000002\",
           \"amount\": 1000.00,
           \"currency\": \"USD\",
           \"description\": \"Should fail - insufficient funds\"
         }' | jq -e '.error // false'"
    
    # Test frozen account
    run_test "Frozen Account Transfer" \
        "curl -sf -X POST $API_GATEWAY_URL/transfer \
         -H 'Content-Type: application/json' \
         -d '{
           \"from_account\": \"1000000004\",
           \"to_account\": \"1000000002\",
           \"amount\": 100.00,
           \"currency\": \"USD\",
           \"description\": \"Should fail - frozen account\"
         }' | jq -e '.error // false'"
}

# Failure simulation tests
test_failure_simulation() {
    log "🔥 Testing failure simulation scenarios..."
    
    # Test failure simulation status
    run_test "Balance Service Failure Simulation Status" \
        "curl -sf $SVC_BALANCE_URL/failure-simulation/stats | jq -e '.learning_mode == true'"
    
    run_test "Transaction Service Failure Simulation Status" \
        "curl -sf $SVC_TRANSACTION_URL/failure-simulation/stats | jq -e '.learning_mode == true'"
    
    # Test failure simulation scenarios
    run_test "Balance Service Learning Scenarios" \
        "curl -sf $SVC_BALANCE_URL/failure-simulation/scenarios | jq -e 'length > 0'"
    
    run_test "Transaction Service Learning Scenarios" \
        "curl -sf $SVC_TRANSACTION_URL/failure-simulation/scenarios | jq -e 'length > 0'"
}

# Compensation tests
test_compensation_scenarios() {
    log "🔄 Testing compensation scenarios..."
    
    # Test compensation audit trail
    run_test "Compensation Audit Trail" \
        "curl -sf $SVC_TRANSACTION_URL/compensation-audit/stats | jq -e 'type == \"object\"'"
    
    # Test enhanced compensation scenarios
    run_test "Enhanced Compensation Scenarios" \
        "curl -sf $SVC_TRANSACTION_URL/compensation-audit/scenarios | jq -e 'length > 0'"
}

# Currency tests
test_currency_support() {
    log "💱 Testing multi-currency support..."
    
    # Test EUR transfer
    run_test "EUR Currency Transfer" \
        "curl -sf -X POST $API_GATEWAY_URL/transfer \
         -H 'Content-Type: application/json' \
         -d '{
           \"from_account\": \"1000000010\",
           \"to_account\": \"1000000011\",
           \"amount\": 50.00,
           \"currency\": \"EUR\",
           \"description\": \"EUR test transfer\"
         }' | jq -e '.status == \"TRANSFER_STATUS_PENDING\"'"
    
    # Test invalid currency
    run_test "Invalid Currency Transfer" \
        "curl -sf -X POST $API_GATEWAY_URL/transfer \
         -H 'Content-Type: application/json' \
         -d '{
           \"from_account\": \"1000000001\",
           \"to_account\": \"1000000002\",
           \"amount\": 100.00,
           \"currency\": \"INVALID\",
           \"description\": \"Should fail - invalid currency\"
         }' | jq -e '.error // false'"
}

# Performance tests
test_performance() {
    log "⚡ Testing system performance..."
    
    # Test concurrent transfers
    run_test "Concurrent Transfers" \
        "for i in {1..5}; do
           curl -sf -X POST $API_GATEWAY_URL/transfer \
           -H 'Content-Type: application/json' \
           -d '{
             \"from_account\": \"1000000012\",
             \"to_account\": \"1000000013\",
             \"amount\": 10.00,
             \"currency\": \"USD\",
             \"description\": \"Concurrent test transfer \$i\"
           }' &
         done
         wait
         echo 'All concurrent transfers initiated'"
}

# Main test execution
main() {
    log "🔧 Waiting for services to be ready..."
    
    # Wait for all services
    wait_for_service "$API_GATEWAY_URL" "API Gateway" || exit 1
    wait_for_service "$SVC_BALANCE_URL" "Balance Service" || exit 1
    wait_for_service "$SVC_TRANSACTION_URL" "Transaction Service" || exit 1
    
    log "🎯 All services are ready. Starting integration tests..."
    
    # Run test suites
    test_health_checks
    test_basic_transfers
    test_failure_simulation
    test_compensation_scenarios
    test_currency_support
    test_performance
    
    # Generate test report
    log "📊 Test Results Summary"
    log "======================"
    log "Total Tests: $TOTAL_TESTS"
    log "Passed: $PASSED_TESTS"
    log "Failed: $FAILED_TESTS"
    log "Success Rate: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
    
    # Create JSON report
    cat > $RESULTS_DIR/test_report_$TIMESTAMP.json << EOF
{
  "timestamp": "$TIMESTAMP",
  "total_tests": $TOTAL_TESTS,
  "passed_tests": $PASSED_TESTS,
  "failed_tests": $FAILED_TESTS,
  "success_rate": $(( PASSED_TESTS * 100 / TOTAL_TESTS )),
  "test_environment": {
    "api_gateway_url": "$API_GATEWAY_URL",
    "svc_balance_url": "$SVC_BALANCE_URL",
    "svc_transaction_url": "$SVC_TRANSACTION_URL"
  }
}
EOF
    
    log "📄 Test report saved to: $RESULTS_DIR/test_report_$TIMESTAMP.json"
    
    # Exit with appropriate code
    if [ $FAILED_TESTS -eq 0 ]; then
        log "🎉 All tests passed!"
        exit 0
    else
        log "💥 Some tests failed!"
        exit 1
    fi
}

# Run main function
main "$@" 