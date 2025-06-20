# Integration Test Environment

This directory contains the integration test environment for the Temporal Flow Demo banking system. It provides an isolated testing environment with dedicated test data and configurations.

## 🎯 Purpose

The integration test environment validates:
- **End-to-end system functionality** across all services
- **Service communication** and data flow
- **Error handling and compensation** workflows
- **Performance under load** scenarios
- **Failure simulation** capabilities

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  postgres-test  │───▶│temporal-server- │───▶│ Integration     │
│  (Test DB)      │    │test             │    │ Test Runner     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  svc-balance-   │    │  svc-transaction│    │  flowngine-test │
│  test           │    │  -test          │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌─────────────────┐
                    │ api-gateway-test│
                    └─────────────────┘
```

## 🚀 Quick Start

### 1. Run Integration Tests
```bash
# Start test environment and run tests
docker-compose -f docker-compose.yml -f integration-tests/docker-compose.test.yml up --build

# Or run tests against existing development environment
cd integration-tests
./run-integration-tests.sh
```

### 2. View Test Results
```bash
# Check test logs
docker logs integration-tests

# View detailed results
ls -la integration-tests/results/
```

## 📁 Directory Structure

```
integration-tests/
├── configs/                    # Test-specific service configurations
│   ├── api-gateway-test.json   # API Gateway test config
│   ├── flowngine-test.json     # FlowEngine test config
│   ├── svc-balance-test.json   # Balance service test config
│   └── svc-transaction-test.json # Transaction service test config
├── test-data/                  # Test-specific database data
│   └── test-accounts.sql       # Test accounts and transactions
├── results/                    # Test execution results (created at runtime)
├── docker-compose.test.yml     # Test environment orchestration
├── Dockerfile                  # Integration test runner image
├── run-integration-tests.sh    # Main test execution script
└── README.md                   # This file
```

## 🧪 Test Categories

### 1. Health Check Tests
- Service availability and health endpoints
- Database connectivity validation
- Temporal server communication

### 2. Basic Transfer Tests
- Successful money transfers
- Insufficient funds scenarios
- Account validation failures (frozen, closed accounts)

### 3. Failure Simulation Tests
- Balance service failure scenarios
- Transaction service failure scenarios
- Learning scenario validation

### 4. Compensation Tests
- Compensation workflow execution
- Audit trail validation
- Enhanced compensation scenarios

### 5. Currency Support Tests
- Multi-currency transfers (USD, EUR, GBP)
- Invalid currency handling

### 6. Performance Tests
- Concurrent transfer processing
- System throughput validation

## 🔧 Test Configuration

### Test Database
- **Database**: `temporal_flow_demo_test_db`
- **User**: `test_user`
- **Password**: `test_password`
- **Port**: `5433` (to avoid conflicts with development DB)

### Test Services Ports
- **API Gateway**: `4012` (external) → `4000` (internal)
- **Balance Service**: `4010` (external) → `4000` (internal)
- **Transaction Service**: `4011` (external) → `4001` (internal)
- **FlowEngine**: `50052` (external) → `50051` (internal)
- **Temporal Server**: `7234` (external) → `7233` (internal)
- **PostgreSQL**: `5433` (external) → `5432` (internal)

### Test Accounts
The test environment includes predefined accounts for various scenarios:

| Account Number | Purpose                  | Balance | Status |
| -------------- | ------------------------ | ------- | ------ |
| 1000000001     | Successful transfers     | $1,000  | Active |
| 1000000002     | Transfer recipient       | $500    | Active |
| 1000000003     | Insufficient funds tests | $10     | Active |
| 1000000004     | Frozen account tests     | $1,000  | Frozen |
| 1000000005     | Closed account tests     | $0      | Closed |
| 111111111111   | Failure simulation       | $1,000  | Active |
| 222222222222   | Compensation tests       | $1,000  | Active |
| 123456789012   | Timeout tests            | $1,000  | Active |
| 999999999999   | Panic tests              | $1,000  | Active |

## 🎯 Test Scenarios

### Successful Transfer Flow
```bash
POST /transfer
{
  "from_account": "1000000001",
  "to_account": "1000000002", 
  "amount": 100.50,
  "currency": "USD"
}
```

### Insufficient Funds Test
```bash
POST /transfer
{
  "from_account": "1000000003",  # Only has $10
  "to_account": "1000000002",
  "amount": 1000.00,             # Requesting $1000
  "currency": "USD"
}
```

### Failure Simulation Test
```bash
POST /transfer
{
  "from_account": "111111111111",  # Triggers failure simulation
  "to_account": "1000000002",
  "amount": 100.00,
  "currency": "USD"
}
```

## 📊 Test Results

Test results are saved in the `results/` directory:

- **test_run_YYYYMMDD_HHMMSS.log**: Detailed execution log
- **test_report_YYYYMMDD_HHMMSS.json**: JSON test report

### Sample Test Report
```json
{
  "timestamp": "20250116_143022",
  "total_tests": 15,
  "passed_tests": 14,
  "failed_tests": 1,
  "success_rate": 93,
  "test_environment": {
    "api_gateway_url": "http://api-gateway-test:4000",
    "svc_balance_url": "http://svc-balance-test:4000",
    "svc_transaction_url": "http://svc-transaction-test:4001"
  }
}
```

## 🐛 Troubleshooting

### Common Issues

1. **Services Not Ready**
   ```bash
   # Check service health individually
   curl http://localhost:4010/health  # Balance Service
   curl http://localhost:4011/health  # Transaction Service
   curl http://localhost:4012/health  # API Gateway
   ```

2. **Database Connection Issues**
   ```bash
   # Check test database
   docker exec postgres-test psql -U test_user -d temporal_flow_demo_test_db -c "\dt"
   ```

3. **Temporal Server Issues**
   ```bash
   # Check Temporal server
   docker logs temporal-server-test
   ```

### Debug Commands
```bash
# View all test containers
docker-compose -f docker-compose.yml -f integration-tests/docker-compose.test.yml ps

# Check test network
docker network inspect temporal-flow-demo-test

# View integration test logs
docker logs integration-tests -f

# Access test database
docker exec -it postgres-test psql -U test_user -d temporal_flow_demo_test_db
```

## 🔗 Related Documentation

- [Main Project README](../README.md)
- [API Specification](../docs/api_specification.md)
- [Development Backlog](../docs/development_backlog.md)
- [Testing Strategy](../docs/testing_strategy.md)

---

**Implementation Status**: ✅ COMPLETED (INT-001)  
**Epic**: Epic 5 - End-to-End Integration  
**Priority**: High - Required for system validation 