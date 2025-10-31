# Performance Optimization Guide

This document describes performance optimization strategies and configurations for the blockchain indexer.

## Table of Contents

- [Cache Infrastructure](#cache-infrastructure)
- [Database Configuration](#database-configuration)
- [Batch Operations](#batch-operations)
- [Concurrency Tuning](#concurrency-tuning)
- [Memory Management](#memory-management)
- [Monitoring Performance](#monitoring-performance)

## Cache Infrastructure

The indexer implements a multi-layer caching strategy to improve performance:

### 1. LRU Cache

Located in `pkg/infrastructure/cache/lru.go`, provides an LRU (Least Recently Used) eviction policy:

```go
cache := cache.NewLRUCache(10000, 5*time.Minute) // 10k items, 5min TTL
value, found := cache.Get("key")
cache.Set("key", value)
```

**Features:**
- Thread-safe operations
- TTL support for automatic expiration
- Background cleanup routine
- Statistics tracking (hits, misses, evictions)

**Use Cases:**
- Block caching (recently accessed blocks)
- Transaction caching
- Chain metadata caching

### 2. Memory Cache

Located in `pkg/infrastructure/cache/cache.go`, provides a simple in-memory cache:

```go
cache := cache.NewMemoryCache(10000, 5*time.Minute)
```

**Features:**
- Simpler than LRU, lower overhead
- Good for smaller datasets
- Automatic expiration

### 3. Statistics Cache

Built into `pkg/infrastructure/storage/pebble/statistics_repository.go`:

```go
type statsCache struct {
    mu          sync.RWMutex
    chainStats  map[string]*models.ChainStatistics
    globalStats *models.GlobalStatistics
    lastUpdate  time.Time
    ttl         time.Duration
}
```

**Configuration:**
- Default TTL: 5 minutes
- Automatic invalidation on updates

## Database Configuration

### PebbleDB Configurations

Three pre-configured profiles are available:

#### 1. Default Configuration

Balanced for general use:

```go
config := pebble.DefaultConfig("./data")
// Cache: 64MB
// Write Buffer: 64MB
// Max Open Files: 1000
// Max Concurrent Memtables: 2
```

#### 2. High Performance Configuration

For systems with ample memory and fast storage:

```go
config := pebble.HighPerformanceConfig("./data")
// Cache: 256MB (4x default)
// Write Buffer: 128MB (2x default)
// Max Open Files: 5000 (5x default)
// Max Concurrent Memtables: 4 (2x default)
// Bytes Per Sync: 1MB (2x default)
```

**Recommended For:**
- Production deployments
- Systems with 8GB+ RAM
- SSD/NVMe storage
- High indexing throughput requirements

**Expected Performance:**
- 2-3x faster writes
- 40-50% faster reads
- Higher memory usage (~400MB total)

#### 3. Low Memory Configuration

For memory-constrained environments:

```go
config := pebble.LowMemoryConfig("./data")
// Cache: 16MB (1/4 default)
// Write Buffer: 16MB (1/4 default)
// Max Open Files: 100 (1/10 default)
// Max Concurrent Memtables: 1 (1/2 default)
```

**Recommended For:**
- Development environments
- Systems with limited RAM (<4GB)
- Testing scenarios
- Embedded systems

### Configuration Parameters

| Parameter | Default | High Perf | Low Memory | Description |
|-----------|---------|-----------|------------|-------------|
| CacheSize | 64MB | 256MB | 16MB | Block cache size |
| WriteBufferSize | 64MB | 128MB | 16MB | Memtable size |
| MaxOpenFiles | 1000 | 5000 | 100 | File descriptor limit |
| MaxConcurrentMem | 2 | 4 | 1 | Concurrent memtables |
| BytesPerSync | 512KB | 1MB | 256KB | Sync frequency |

### Choosing a Configuration

```go
var config *pebble.Config

switch deploymentType {
case "production":
    config = pebble.HighPerformanceConfig(storagePath)
case "development":
    config = pebble.LowMemoryConfig(storagePath)
default:
    config = pebble.DefaultConfig(storagePath)
}

storage, err := pebble.NewStorage(config)
```

## Batch Operations

### Block Processing

Batch operations significantly improve throughput:

```go
// Process multiple blocks in a single transaction
batch := storage.NewBatch()
for _, block := range blocks {
    batch.SaveBlock(ctx, block)
}
err := batch.Commit(ctx)
```

**Performance Impact:**
- Single writes: ~1,000 blocks/sec
- Batch writes: ~10,000 blocks/sec (10x improvement)

### Recommended Batch Sizes

| Operation | Batch Size | Rationale |
|-----------|------------|-----------|
| Block Indexing | 50-100 | Balance memory vs throughput |
| Transaction Indexing | 100-500 | Smaller objects, can batch more |
| Gap Recovery | 10-50 | More complex validation |
| Statistics Updates | 1-10 chains | Updates are lightweight |

## Concurrency Tuning

### Worker Pool Configuration

Configure worker counts based on system resources:

```yaml
chains:
  - chain_id: "eth-mainnet"
    workers: 10        # Concurrent block fetchers
    batch_size: 50     # Blocks per batch
```

**Guidelines:**
- **CPU Cores:** `workers <= 2 * cores`
- **Memory:** Each worker uses ~50-100MB
- **Network:** More workers = more RPC requests

**Example Configurations:**

```yaml
# Small system (4 cores, 8GB RAM)
workers: 4
batch_size: 20

# Medium system (8 cores, 16GB RAM)
workers: 10
batch_size: 50

# Large system (16+ cores, 32GB+ RAM)
workers: 20
batch_size: 100
```

### Event Bus Configuration

```go
eventBusConfig := &event.EventBusConfig{
    WorkerCount: 10,      // Event processing workers
    QueueSize:   10000,   // Event queue buffer
}
```

**Tuning:**
- More workers: Better throughput, higher CPU usage
- Larger queue: Better burst handling, more memory

## Memory Management

### Memory Usage Breakdown

| Component | Typical Usage | Notes |
|-----------|--------------|-------|
| PebbleDB Cache | 64-256MB | Configurable |
| Write Buffers | 64-128MB | Configurable |
| Worker Pools | 50-100MB per worker | Scales with workers |
| Event Bus | 10-50MB | Depends on queue size |
| Statistics Cache | 5-10MB | Fixed overhead |
| Application | 50-100MB | Baseline |

**Total Typical Range:** 500MB - 2GB

### Memory Optimization Tips

1. **Reduce worker counts** if memory-constrained
2. **Use LowMemoryConfig** for PebbleDB
3. **Decrease batch sizes** to reduce buffering
4. **Limit cache sizes** in statistics collector
5. **Monitor memory usage** with metrics

### Memory Monitoring

```bash
# Check memory usage
ps aux | grep indexer

# Get detailed memory stats
pprof -http=:6060 http://localhost:6060/debug/pprof/heap
```

## Monitoring Performance

### Metrics to Monitor

1. **Indexing Rate**
   - Blocks per second
   - Transactions per second

2. **Cache Performance**
   - Hit rate (target: >80%)
   - Eviction rate
   - Size utilization

3. **Database Performance**
   - Read latency (p50, p95, p99)
   - Write throughput
   - Compaction metrics

4. **Resource Usage**
   - Memory consumption
   - CPU utilization
   - Disk I/O

### Prometheus Metrics

The indexer exposes metrics at `http://localhost:9091/metrics`:

```promql
# Indexing rate
rate(indexer_blocks_indexed_total[5m])

# Cache hit rate
indexer_cache_hits_total / (indexer_cache_hits_total + indexer_cache_misses_total)

# Memory usage
go_memstats_alloc_bytes
```

### Performance Benchmarks

Test configurations on your hardware:

```bash
# Run indexer with profiling
./bin/indexer index --config config.yaml --profile

# Analyze CPU profile
go tool pprof cpu.prof

# Analyze memory profile
go tool pprof mem.prof
```

## Best Practices

1. **Start with defaults**, measure, then optimize
2. **Monitor cache hit rates** - low hit rate indicates poor caching strategy
3. **Batch operations** whenever possible
4. **Use high-performance config** in production
5. **Set appropriate worker counts** based on available resources
6. **Regular compaction** of PebbleDB for optimal performance
7. **Monitor memory** to prevent OOM issues

## Troubleshooting

### Slow Indexing

1. Check worker configuration
2. Verify network latency to RPC endpoints
3. Review batch sizes
4. Check database compaction status

### High Memory Usage

1. Reduce worker count
2. Switch to LowMemoryConfig
3. Decrease batch sizes
4. Clear caches periodically

### Poor Cache Hit Rate

1. Increase cache size
2. Increase TTL
3. Review access patterns
4. Consider data locality

## Future Optimizations

Planned improvements:

- [ ] Bloom filters for block existence checks
- [ ] Read-ahead caching for sequential access
- [ ] Write coalescing for small transactions
- [ ] Adaptive batch sizing based on load
- [ ] Compression for cold data
- [ ] Tiered storage (hot/warm/cold)
