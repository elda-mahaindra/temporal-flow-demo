# TEMP-010: Observability Enhancement

This directory contains the observability configuration for the Temporal Flow Demo banking system, implementing comprehensive monitoring and dashboards for workflow orchestration and system health.

## 🎯 Observability Objectives

**Aligns with PRD Objective 4: Simplify Development**
- Showcase Temporal's built-in observability features
- Enable workflow monitoring and debugging
- Provide real-time system health visibility
- Simplify troubleshooting and performance analysis

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │───▶│   Prometheus    │───▶│    Grafana      │
│   Services      │    │  (Metrics       │    │  (Dashboards &  │
│                 │    │   Collection)   │    │   Visualization)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  Temporal UI    │
                       │ (Workflow       │
                       │  Monitoring)    │
                       └─────────────────┘
```

## 🚀 Quick Start

### 1. Start the Complete Stack
```bash
cd /path/to/temporal-flow-demo
docker-compose up -d
```

### 2. Access Observability Dashboards

| Service         | URL                   | Credentials | Purpose                       |
| --------------- | --------------------- | ----------- | ----------------------------- |
| **Temporal UI** | http://localhost:8080 | None        | Workflow execution monitoring |
| **Grafana**     | http://localhost:3001 | admin/admin | System metrics & dashboards   |
| **Prometheus**  | http://localhost:9090 | None        | Metrics collection & queries  |

### 3. Monitor Banking Workflows

1. **Temporal UI (http://localhost:8080)**:
   - View real-time workflow executions
   - Inspect transfer workflow histories
   - Debug failed workflows with full stack traces
   - Monitor activity execution details

2. **Grafana Dashboard (http://localhost:3001)**:
   - Banking system overview
   - Transfer workflow success/failure rates
   - Service health status
   - API response times
   - Banking activity performance metrics

## 📊 Available Dashboards

### 1. Temporal Flow Demo - Banking System Overview
**Dashboard ID**: `temporal-flow-demo-overview`

**Panels Include**:
- **Service Health Status**: Real-time health of all services
- **Transfer Workflow Rates**: Completed vs failed workflow rates
- **API Response Times**: 95th and 50th percentile response times
- **Banking Activity Rates**: CheckBalance, DebitAccount, CreditAccount operations
- **System Resource Usage**: CPU and memory utilization

**Key Metrics**:
```promql
# Workflow completion rate
rate(temporal_workflow_completed_total[5m])

# Workflow failure rate
rate(temporal_workflow_failed_total[5m])

# Activity execution rates
rate(temporal_activity_completed_total{activity_type="CheckBalance"}[5m])
rate(temporal_activity_completed_total{activity_type="DebitAccount"}[5m])
rate(temporal_activity_completed_total{activity_type="CreditAccount"}[5m])

# API response times
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

## 🔧 Configuration Details

### Prometheus Configuration
**File**: `prometheus/prometheus.yml`

**Scrape Targets**:
- `prometheus:9090` - Prometheus self-monitoring
- `temporal-server:7233` - Temporal server metrics
- `api-gateway:8080` - API Gateway metrics (port 8084 externally)
- `flowngine:8080` - FlowEngine metrics (port 8083 externally)
- `svc-transaction:8080` - Transaction service metrics (port 8081 externally)
- `svc-balance:8080` - Balance service metrics (port 8082 externally)

**Scrape Intervals**:
- Temporal Server: 30s (longer interval for stability)
- Application Services: 15s (frequent for real-time monitoring)

### Grafana Configuration
**Datasource**: Prometheus at `http://prometheus:9090`
**Dashboard Provisioning**: Auto-loaded from `/var/lib/grafana/dashboards`
**Refresh Rate**: 30 seconds for real-time monitoring

## 🎯 Banking-Specific Monitoring

### Workflow Observability
- **Transfer Workflow Execution**: Monitor money transfer workflows from start to completion
- **Compensation Tracking**: Observe compensation workflows when transfers fail
- **Activity Execution**: Track individual banking activities (CheckBalance, DebitAccount, CreditAccount)

### Business Metrics
- **Transfer Success Rate**: Percentage of successful money transfers
- **Average Transfer Time**: End-to-end transfer completion time
- **Failure Patterns**: Analysis of common failure scenarios
- **Compensation Rate**: Frequency of compensation workflows

### Performance Monitoring
- **Activity Timeouts**: Monitor activity execution times vs configured timeouts
- **Worker Utilization**: Track Temporal worker capacity and performance
- **Database Performance**: Monitor PostgreSQL connection pool and query performance
- **API Gateway Load**: Track REST API request rates and response times

## 🚨 Alerting (Future Enhancement)

**Recommended Alerts**:
- Workflow failure rate > 5%
- API response time > 2 seconds (95th percentile)
- Service health check failures
- Database connection pool exhaustion
- Memory usage > 80%

## 🛠️ Troubleshooting

### Common Issues

1. **No Metrics in Grafana**:
   - Check Prometheus targets: http://localhost:9090/targets
   - Verify services are exposing metrics on port 8080
   - Check network connectivity between containers

2. **Temporal UI Connection Issues**:
   - Verify Temporal server is running: `docker-compose ps temporal-server`
   - Check Temporal server logs: `docker-compose logs temporal-server`

3. **Dashboard Not Loading**:
   - Check Grafana logs: `docker-compose logs grafana`
   - Verify dashboard provisioning: `/var/lib/grafana/dashboards/`

### Debug Commands
```bash
# Check service health
curl http://localhost:4000/health

# Check Prometheus metrics
curl http://localhost:8084/metrics  # API Gateway
curl http://localhost:8083/metrics  # FlowEngine
curl http://localhost:8081/metrics  # Transaction Service
curl http://localhost:8082/metrics  # Balance Service

# View container logs
docker-compose logs -f prometheus
docker-compose logs -f grafana
docker-compose logs -f temporal-ui
```

## 📈 Performance Optimization Integration

**TEMP-010 enhances TEMP-009 optimizations**:
- Monitor the effectiveness of worker option tuning
- Validate activity timeout configurations
- Observe retry policy performance
- Track heartbeat and session worker efficiency

## 🎓 Learning Objectives

This observability setup demonstrates:
1. **Temporal Observability**: Native workflow and activity monitoring
2. **Metrics Collection**: Prometheus-based metrics gathering
3. **Dashboard Creation**: Grafana visualization for banking workflows
4. **System Health**: Comprehensive service health monitoring
5. **Performance Analysis**: Real-time performance insights

## 🔗 Related Documentation

- [Temporal UI Documentation](https://docs.temporal.io/web-ui)
- [Prometheus Configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/)
- [Grafana Dashboard Guide](https://grafana.com/docs/grafana/latest/dashboards/)
- [Banking System Architecture](../docs/temporal_flow_demo_tech_spec.md)

---

**Implementation Status**: ✅ COMPLETED (TEMP-010)
**Priority**: Medium - Required for **Objective 4: Simplify Development** 