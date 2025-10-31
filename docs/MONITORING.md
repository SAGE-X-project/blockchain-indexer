# Monitoring & Observability Guide

This document describes monitoring, health checking, and debugging capabilities of the blockchain indexer.

## Table of Contents

- [Health Checks](#health-checks)
- [Metrics](#metrics)
- [Profiling & Debugging](#profiling--debugging)
- [Logging](#logging)
- [Alerting](#alerting)
- [Troubleshooting](#troubleshooting)

## Health Checks

The indexer provides multiple health check endpoints for monitoring system health.

### Basic Health Check

Simple liveness check:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

**Use Case:** Load balancer health checks, simple liveness probes

### Detailed Health Check

Comprehensive health report:

```bash
curl http://localhost:8080/health/detailed
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "checks": {
    "storage": {
      "name": "storage",
      "status": "healthy",
      "message": "Storage is healthy",
      "timestamp": "2024-01-15T10:30:00Z",
      "duration": 15000000,
      "details": {
        "chain_count": 3,
        "query_duration_ms": 15
      }
    },
    "memory": {
      "name": "memory",
      "status": "healthy",
      "message": "Memory usage: 512MB",
      "timestamp": "2024-01-15T10:30:00Z",
      "duration": 100000,
      "details": {
        "alloc_mb": 512,
        "sys_mb": 768,
        "num_gc": 42
      }
    },
    "goroutines": {
      "name": "goroutines",
      "status": "healthy",
      "message": "Goroutine count: 156",
      "timestamp": "2024-01-15T10:30:00Z",
      "duration": 50000,
      "details": {
        "goroutine_count": 156
      }
    }
  }
}
```

**Status Codes:**
- `200 OK`: System is healthy or degraded
- `503 Service Unavailable`: System is unhealthy

**Health Statuses:**
- `healthy`: All checks passed
- `degraded`: Some issues detected but system is operational
- `unhealthy`: Critical issues detected

**Use Case:** Readiness probes, detailed monitoring, alerting

### Health Check Configuration

Health checks run automatically every 30 seconds. Configure thresholds in `server.go`:

```go
healthChecker.RegisterCheck("storage", health.StorageHealthCheck(chainRepo))
healthChecker.RegisterCheck("memory", health.MemoryHealthCheck(1024)) // 1GB threshold
healthChecker.RegisterCheck("goroutines", health.GoroutineHealthCheck(10000))
```

### Custom Health Checks

Add custom checks for specific chains or components:

```go
healthChecker.RegisterCheck("eth-mainnet", health.ChainConnectivityCheck(
    "eth-mainnet",
    func(ctx context.Context) error {
        // Check if chain is reachable
        return adapter.Ping(ctx)
    },
))
```

## Metrics

The indexer exposes Prometheus metrics at `http://localhost:9091/metrics`.

### Key Metrics

#### Indexing Metrics

```promql
# Blocks indexed per second
rate(indexer_blocks_indexed_total{chain_id="eth-mainnet"}[5m])

# Transactions indexed per second
rate(indexer_transactions_indexed_total{chain_id="eth-mainnet"}[5m])

# Indexing errors
rate(indexer_errors_total{chain_id="eth-mainnet"}[5m])

# Blocks behind
indexer_blocks_behind{chain_id="eth-mainnet"}

# Sync progress percentage
indexer_sync_progress{chain_id="eth-mainnet"}
```

#### Performance Metrics

```promql
# API request duration (p95)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Database query latency
histogram_quantile(0.99, rate(db_query_duration_seconds_bucket[5m]))

# Cache hit rate
indexer_cache_hits_total / (indexer_cache_hits_total + indexer_cache_misses_total)
```

#### Resource Metrics

```promql
# Memory usage
go_memstats_alloc_bytes

# Goroutine count
go_goroutines

# GC duration
rate(go_gc_duration_seconds_sum[5m])

# CPU usage (from container metrics)
rate(process_cpu_seconds_total[5m])
```

#### Gap Recovery Metrics

```promql
# Gaps detected
indexer_gaps_detected_total{chain_id="eth-mainnet"}

# Gaps recovered
indexer_gaps_recovered_total{chain_id="eth-mainnet"}

# Gap recovery errors
rate(indexer_gap_recovery_errors_total[5m])
```

### Grafana Dashboard

Import the provided Grafana dashboard from `deployments/grafana/dashboard.json`:

**Panels Include:**
- Indexing rate per chain
- Sync progress
- Blocks behind
- API latency percentiles
- Error rates
- Memory and CPU usage
- Cache performance
- Gap detection and recovery

### Custom Metrics

Add custom metrics in your code:

```go
import "github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/metrics"

// Counter
metrics.IndexerBlocksIndexed.WithLabelValues(chainID).Inc()

// Gauge
metrics.IndexerBlocksBehind.WithLabelValues(chainID).Set(float64(blocksBehind))

// Histogram
timer := prometheus.NewTimer(metrics.APIRequestDuration)
defer timer.ObserveDuration()
```

## Profiling & Debugging

The indexer provides pprof endpoints for CPU and memory profiling.

### pprof Endpoints

Available at `http://localhost:8080/debug/pprof/`:

- `/debug/pprof/` - Index of available profiles
- `/debug/pprof/profile` - CPU profile (30s by default)
- `/debug/pprof/heap` - Memory allocation profile
- `/debug/pprof/goroutine` - Goroutine stack traces
- `/debug/pprof/threadcreate` - Thread creation profile
- `/debug/pprof/block` - Blocking profile
- `/debug/pprof/mutex` - Mutex contention profile
- `/debug/pprof/allocs` - All memory allocations
- `/debug/pprof/trace` - Execution trace

### CPU Profiling

Capture a 30-second CPU profile:

```bash
# Download profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze with pprof
go tool pprof cpu.prof

# Web UI
go tool pprof -http=:8081 cpu.prof
```

**Common Commands in pprof:**
- `top` - Show top CPU consumers
- `list <function>` - Show function source with annotations
- `web` - Generate SVG call graph
- `pdf` - Generate PDF call graph

### Memory Profiling

Capture heap profile:

```bash
# Download profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# Analyze
go tool pprof -http=:8081 heap.prof
```

**Memory Analysis:**
- `top` - Top memory allocators
- `list <function>` - See allocation sites
- `inuse_space` - Current memory usage
- `alloc_space` - Total allocations

### Goroutine Analysis

Check for goroutine leaks:

```bash
# Get goroutine dump
curl http://localhost:8080/debug/pprof/goroutine?debug=2 > goroutines.txt

# Analyze with pprof
curl http://localhost:8080/debug/pprof/goroutine > goroutine.prof
go tool pprof -http=:8081 goroutine.prof
```

### Runtime Statistics

Get runtime stats:

```bash
curl http://localhost:8080/debug/stats | jq
```

Response:
```json
{
  "memory": {
    "alloc_mb": 512,
    "total_alloc_mb": 1024,
    "sys_mb": 768,
    "num_gc": 42,
    "gc_pause_ms": 0.5
  },
  "runtime": {
    "goroutines": 156,
    "num_cpu": 8,
    "version": "go1.21.0"
  }
}
```

### Execution Tracing

Capture execution trace for detailed analysis:

```bash
# Capture 5-second trace
curl http://localhost:8080/debug/pprof/trace?seconds=5 > trace.out

# Analyze trace
go tool trace trace.out
```

Trace provides:
- Goroutine execution timeline
- Network/syscall blocking
- GC events
- Goroutine creation/destruction

## Logging

### Log Levels

Configure logging in `config.yaml`:

```yaml
logging:
  level: info        # debug, info, warn, error
  format: json       # json, console
  output: stdout     # stdout, stderr, file
  file_path: logs/indexer.log
```

### Structured Logging

Logs are structured with consistent fields:

```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:00.000Z",
  "caller": "indexer/indexer.go:123",
  "msg": "indexed block",
  "chain_id": "eth-mainnet",
  "block_number": 12345678,
  "tx_count": 150,
  "duration_ms": 245
}
```

### Log Aggregation

Forward logs to centralized systems:

**Fluentd:**
```yaml
<source>
  @type forward
  port 24224
</source>

<match indexer.**>
  @type elasticsearch
  host elasticsearch
  port 9200
  index_name indexer-logs
</match>
```

**Loki:**
```yaml
clients:
  - url: http://loki:3100/loki/api/v1/push
    external_labels:
      app: blockchain-indexer
```

### Important Log Patterns

**Errors:**
```bash
# Recent errors
grep -i error logs/indexer.log | tail -20

# Error frequency
grep -i error logs/indexer.log | wc -l
```

**Performance Issues:**
```bash
# Slow operations
jq 'select(.duration_ms > 1000)' logs/indexer.log
```

**Indexing Progress:**
```bash
# Blocks indexed
grep "indexed block" logs/indexer.log | tail -10
```

## Alerting

### Recommended Alerts

#### Critical Alerts

1. **Service Down**
```promql
up{job="blockchain-indexer"} == 0
```

2. **High Error Rate**
```promql
rate(indexer_errors_total[5m]) > 10
```

3. **Indexing Stopped**
```promql
rate(indexer_blocks_indexed_total[5m]) == 0
```

4. **High Memory Usage**
```promql
go_memstats_alloc_bytes > 2e9  # 2GB
```

#### Warning Alerts

1. **Falling Behind**
```promql
indexer_blocks_behind > 1000
```

2. **Slow API Responses**
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
```

3. **Low Cache Hit Rate**
```promql
rate(indexer_cache_hits_total[5m]) / (rate(indexer_cache_hits_total[5m]) + rate(indexer_cache_misses_total[5m])) < 0.5
```

4. **Many Goroutines**
```promql
go_goroutines > 10000
```

### Alert Configuration (Prometheus)

```yaml
groups:
  - name: indexer
    interval: 30s
    rules:
      - alert: IndexerDown
        expr: up{job="blockchain-indexer"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Indexer is down"

      - alert: HighErrorRate
        expr: rate(indexer_errors_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
```

## Troubleshooting

### Common Issues

#### 1. High Memory Usage

**Symptoms:**
- Memory growing continuously
- OOM kills

**Debug:**
```bash
# Check memory profile
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof -top heap.prof

# Check for leaks
go tool pprof -base heap1.prof heap2.prof
```

**Solutions:**
- Reduce worker count
- Use LowMemoryConfig
- Check for goroutine leaks
- Reduce batch sizes

#### 2. Slow Indexing

**Symptoms:**
- Low blocks/sec rate
- Falling behind chain

**Debug:**
```bash
# CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof -top cpu.prof

# Check goroutines
curl http://localhost:8080/debug/pprof/goroutine?debug=2
```

**Solutions:**
- Increase worker count
- Use HighPerformanceConfig
- Check RPC endpoint latency
- Optimize batch sizes

#### 3. High API Latency

**Symptoms:**
- Slow API responses
- Timeouts

**Debug:**
```bash
# Check database performance
curl http://localhost:8080/health/detailed | jq '.checks.storage'

# Profile specific request
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=10
```

**Solutions:**
- Enable caching
- Increase database cache
- Add indices
- Reduce query complexity

#### 4. Goroutine Leak

**Symptoms:**
- Increasing goroutine count
- Memory growth

**Debug:**
```bash
# Compare goroutine profiles
curl http://localhost:8080/debug/pprof/goroutine > g1.prof
# Wait 5 minutes
curl http://localhost:8080/debug/pprof/goroutine > g2.prof

go tool pprof -base g1.prof g2.prof
```

**Solutions:**
- Check for missing context cancellation
- Verify channel cleanup
- Review defer statements

### Performance Baseline

**Expected Performance (Default Config):**
- Indexing rate: 100-500 blocks/sec (depends on chain)
- Memory usage: 500MB-1GB
- Goroutines: 50-500
- API latency (p95): <100ms

**High-Performance Config:**
- Indexing rate: 500-2000 blocks/sec
- Memory usage: 1-2GB
- Goroutines: 100-1000
- API latency (p95): <50ms

### Debug Checklist

When experiencing issues:

1. Check `/health/detailed` endpoint
2. Review recent logs for errors
3. Check Prometheus metrics
4. Capture CPU/memory profile
5. Check goroutine count
6. Verify network connectivity
7. Check disk space
8. Review configuration

## Best Practices

1. **Monitor continuously** - Don't wait for issues
2. **Set up alerts** - Get notified proactively
3. **Regular profiling** - Understand your baseline
4. **Log aggregation** - Centralize logs for analysis
5. **Document incidents** - Learn from issues
6. **Test under load** - Know your limits
7. **Keep metrics** - Historical data is valuable
8. **Review dashboards** - Regular health checks

## Additional Resources

- [Performance Guide](PERFORMANCE.md)
- [Architecture Documentation](ARCHITECTURE.md)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [pprof Guide](https://go.dev/blog/pprof)
- [Go Execution Tracer](https://go.dev/doc/diagnostics#execution-tracer)
