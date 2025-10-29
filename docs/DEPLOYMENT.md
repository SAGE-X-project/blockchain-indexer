# Deployment Guide

This guide covers various deployment options for the Blockchain Indexer.

## Table of Contents

- [Docker Deployment](#docker-deployment)
- [Docker Compose Deployment](#docker-compose-deployment)
- [Systemd Service](#systemd-service)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Configuration](#configuration)
- [Monitoring](#monitoring)

---

## Docker Deployment

### Building the Docker Image

```bash
# Build using script
./scripts/docker-build.sh

# Or build manually
docker build -t blockchain-indexer:latest .
```

### Running the Container

```bash
# Create config file
cp config/config.example.yaml config/config.yaml
# Edit config.yaml with your settings

# Run the container
docker run -d \
  --name blockchain-indexer \
  -p 8080:8080 \
  -p 50051:50051 \
  -p 9091:9091 \
  -v $(pwd)/config/config.yaml:/app/config/config.yaml:ro \
  -v indexer-data:/app/data \
  blockchain-indexer:latest
```

### Checking Logs

```bash
# View logs
docker logs -f blockchain-indexer

# Check health
docker ps
curl http://localhost:8080/health
```

---

## Docker Compose Deployment

Docker Compose provides a complete monitoring stack with Prometheus and Grafana.

### Services Included

- **Blockchain Indexer**: Main application
- **Prometheus**: Metrics collection
- **Grafana**: Metrics visualization

### Starting Services

```bash
# Create config file
cp config/config.example.yaml config/config.yaml
# Edit config.yaml with your settings

# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

### Accessing Services

- **Indexer REST API**: http://localhost:8080/api
- **Indexer GraphQL**: http://localhost:8080/graphql
- **Indexer gRPC**: localhost:50051
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Metrics**: http://localhost:9091/metrics

### Stopping Services

```bash
# Stop services
docker-compose stop

# Stop and remove
docker-compose down

# Stop and remove with volumes
docker-compose down -v
```

---

## Systemd Service

For production deployments on Linux servers.

### Installation

```bash
# Create user
sudo useradd -r -s /bin/false indexer

# Create directories
sudo mkdir -p /opt/blockchain-indexer/bin
sudo mkdir -p /opt/blockchain-indexer/data
sudo mkdir -p /etc/blockchain-indexer

# Copy binary
sudo cp bin/indexer /opt/blockchain-indexer/bin/

# Copy configuration
sudo cp config/config.example.yaml /etc/blockchain-indexer/config.yaml
# Edit /etc/blockchain-indexer/config.yaml with your settings

# Set permissions
sudo chown -R indexer:indexer /opt/blockchain-indexer
sudo chown -R indexer:indexer /etc/blockchain-indexer
sudo chmod 600 /etc/blockchain-indexer/config.yaml

# Install service
sudo cp deployments/systemd/blockchain-indexer.service /etc/systemd/system/
sudo systemctl daemon-reload
```

### Managing the Service

```bash
# Start service
sudo systemctl start blockchain-indexer

# Enable on boot
sudo systemctl enable blockchain-indexer

# Check status
sudo systemctl status blockchain-indexer

# View logs
sudo journalctl -u blockchain-indexer -f

# Restart service
sudo systemctl restart blockchain-indexer

# Stop service
sudo systemctl stop blockchain-indexer
```

---

## Kubernetes Deployment

Coming soon. See [deployments/kubernetes/](../deployments/kubernetes/) for examples.

---

## Configuration

### Environment-Specific Configs

```bash
# Development (EVM chains)
cp config/config.example.yaml config/config.yaml

# Development (Solana)
cp config/config-solana.example.yaml config/config.yaml

# Production
# Create custom config with:
# - Production RPC endpoints
# - Increased workers and batch sizes
# - TLS enabled for APIs
# - Proper logging configuration
```

### Required Configuration

At minimum, configure:

1. **Chain RPC endpoints**: Update with your RPC provider URLs
2. **Storage path**: Ensure writable data directory
3. **Server ports**: Adjust if needed (8080, 50051, 9091)

### Security Best Practices

- **Never commit** `config.yaml` to git (it's in .gitignore)
- Use **read-only** mounts for config in Docker
- Set **proper file permissions** (600) for config files
- Use **TLS** for production APIs
- Limit **API access** with firewall rules

---

## Monitoring

### Prometheus Metrics

The indexer exposes Prometheus metrics at `/metrics` (default port 9091).

**Key metrics:**
- `indexer_blocks_indexed_total`: Total blocks indexed
- `indexer_transactions_indexed_total`: Total transactions indexed
- `indexer_rpc_requests_total`: RPC request count
- `indexer_rpc_errors_total`: RPC error count
- `indexer_block_processing_duration_seconds`: Block processing time

### Grafana Dashboards

Access Grafana at http://localhost:3000 (when using docker-compose).

Default credentials:
- Username: `admin`
- Password: `admin`

Dashboards are automatically provisioned from `deployments/grafana/dashboards/`.

### Health Checks

```bash
# Check application health
curl http://localhost:8080/health

# Response
{
  "status": "ok",
  "version": "0.1.0-beta"
}
```

### Logging

The indexer logs to stdout/stderr by default.

**Log levels:**
- `debug`: Development debugging
- `info`: General information
- `warn`: Warning messages
- `error`: Error messages

Configure in `config.yaml`:

```yaml
logging:
  level: info
  format: json  # or "console"
  output: stdout
```

---

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker logs blockchain-indexer

# Common issues:
# - Config file not found: Mount config correctly
# - Permission denied: Check volume permissions
# - Port already in use: Change port mapping
```

### High Memory Usage

```bash
# Reduce batch size and workers in config.yaml
chains:
  - batch_size: 10  # Lower from 50
    workers: 2      # Lower from 5

storage:
  pebble:
    cache_size: 33554432  # Lower from 64MB
```

### Connection Issues

```bash
# Test RPC endpoints
curl <your-rpc-endpoint>

# Check network connectivity
docker exec -it blockchain-indexer wget --spider <rpc-endpoint>

# Verify DNS resolution
docker exec -it blockchain-indexer nslookup <rpc-host>
```

### Performance Tuning

For high-performance indexing:

```yaml
chains:
  - batch_size: 100      # Increase batch size
    workers: 20          # Increase workers
    confirmation_blocks: 0  # Lower for faster indexing

storage:
  pebble:
    cache_size: 134217728  # 128MB
    write_buffer_size: 67108864  # 64MB
```

---

## Support

For issues and questions:
- GitHub Issues: https://github.com/sage-x-project/blockchain-indexer/issues
- Documentation: https://github.com/sage-x-project/blockchain-indexer/docs
