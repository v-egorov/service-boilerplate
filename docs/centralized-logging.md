# Centralized Logging with Loki, Grafana, and Promtail

## Overview

The service boilerplate implements a comprehensive centralized logging stack using Grafana Loki for log aggregation, Promtail for log shipping, and Grafana for visualization and alerting. This setup provides production-ready log management with structured JSON logging, real-time monitoring, and powerful querying capabilities.

## Architecture

### Components

- **Grafana Loki**: Log aggregation system that indexes and stores logs
- **Promtail**: Log shipping agent that tails log files and forwards them to Loki
- **Grafana**: Visualization and dashboard platform with Loki as a data source
- **Service Logs**: Structured JSON logs written by each microservice

### Data Flow

```
Services → Log Files → Promtail → Loki → Grafana
    ↓         ↓         ↓        ↓       ↓
  JSON     Rotation   Parsing  Indexing  Dashboards
```

## Configuration

### Loki Configuration

Loki is configured via `docker/loki-config.yml` with the following key settings:

```yaml
auth_enabled: false  # Disabled for development
server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1

schema_config:
  configs:
    - from: 2020-10-24
      store: tsdb
      object_store: filesystem
      schema: v13
      index:
        prefix: index_
        period: 24h

limits_config:
  allow_structured_metadata: true
```

### Promtail Configuration

Promtail is configured via `docker/promtail-config.yml` and includes job definitions for each service:

```yaml
server:
  http_listen_port: 9080

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: api-gateway
    static_configs:
      - targets: [localhost]
        labels:
          job: api-gateway
          service: api-gateway
          __path__: /var/log/api-gateway/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: timestamp
            service: service
            request_id: request_id
            method: method
            path: path
            status: status
            duration_ms: duration_ms
      - labels:
          level:
          service:
          request_id:
          method:
          path:
          status:
      - timestamp:
          source: timestamp
          format: RFC3339
```

### Grafana Configuration

Grafana is automatically configured with Loki as a data source via provisioning:

```yaml
# docker/grafana/provisioning/datasources/loki.yml
apiVersion: 1
datasources:
  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    isDefault: true
```

## Service Configuration

### Log Output Requirements

For services to integrate with the centralized logging stack, they must:

1. **Write structured JSON logs** to files in `/app/logs/`
2. **Include consistent fields** for proper parsing and querying
3. **Use appropriate log levels** (debug, info, warn, error)
4. **Include request correlation** via `request_id`

### Required Log Fields

Each log entry should include these fields for optimal querying:

```json
{
  "time": "2025-09-27T10:30:00Z",
  "level": "info",
  "msg": "Request completed",
  "service": "api-gateway",
  "request_id": "abc-123-def",
  "method": "GET",
  "path": "/api/v1/users",
  "status": 200,
  "duration_ms": 150,
  "user_id": "user-456"
}
```

### Environment Variables

Services are configured with these logging environment variables:

```bash
# Core logging settings
LOGGING_LEVEL=info                    # debug, info, warn, error
LOGGING_FORMAT=json                   # json for structured logs
LOGGING_OUTPUT=file                   # Write to files for Promtail
LOGGING_DUAL_OUTPUT=true              # Also output to stdout for Docker logs
LOGGING_STRIP_ANSI_FROM_FILES=true   # Clean JSON for Loki parsing
```

## Docker Integration

### Volume Mounts

Each service has dedicated log volumes defined in `docker-compose.yml`:

```yaml
volumes:
  api_gateway_logs:
    name: ${API_GATEWAY_LOGS_VOLUME}
    driver: local
    driver_opts:
      type: none
      o: bind
      device: ${PWD}/docker/volumes/api-gateway/logs
```

### Service Configuration

Services mount their log volumes and configure logging:

```yaml
api-gateway:
  environment:
    - LOGGING_OUTPUT=file
    - LOGGING_DUAL_OUTPUT=true
  volumes:
    - api_gateway_logs:/app/logs
```

## Adding New Services

### Automatic Configuration

When creating a new service with `./scripts/create-service.sh`, the system automatically:

1. **Creates log volumes** in `docker-compose.yml`
2. **Configures environment variables** for structured logging
3. **Adds Promtail job configuration** for log shipping
4. **Creates volume directories** for log storage

### Manual Promtail Configuration

If you need to manually add a service to Promtail, add a new job to `docker/promtail-config.yml`:

```yaml
- job_name: new-service
  static_configs:
    - targets: [localhost]
      labels:
        job: new-service
        service: new-service
        __path__: /var/log/new-service/*.log
  pipeline_stages:
    - json:
        expressions:
          level: level
          timestamp: timestamp
          service: service
          request_id: request_id
          # Add service-specific fields
    - labels:
        level:
        service:
        request_id:
    - timestamp:
        source: timestamp
        format: RFC3339
```

### Volume Creation

Ensure log volume directories exist:

```bash
mkdir -p docker/volumes/new-service/logs
```

## Grafana Dashboards

### Pre-configured Dashboard

The system includes a default dashboard at `docker/grafana/dashboards/service-logs.json` with:

- **Log Volume by Service**: Bar chart showing log volume per service
- **Error Logs**: Dedicated panel for error-level logs
- **All Service Logs**: Comprehensive log viewer with filtering

### Custom Dashboards

Create custom dashboards by:

1. **Access Grafana** at http://localhost:3001 (admin/admin)
2. **Create new dashboard** from the Loki data source
3. **Use LogQL queries** for advanced filtering

### Useful LogQL Queries

```logql
# All error logs
{level="error"}

# Errors from specific service
{level="error", service="api-gateway"}

# Slow requests (>1 second)
{duration_ms > 1000}

# Requests by status code
{status="404"}

# Logs with specific request ID
{request_id="abc-123-def"}
```

## Operations Guide

### Starting the Stack

Start all logging components:

```bash
# Start complete stack
make dev

# Or start individual components
docker-compose up loki promtail grafana
```

### Accessing Interfaces

- **Grafana**: http://localhost:3001 (admin/admin)
- **Loki**: http://localhost:3100 (API only)
- **Promtail**: http://localhost:9080 (metrics only)

### Monitoring Health

Check component health:

```bash
# Loki health
curl http://localhost:3100/ready

# Promtail metrics
curl http://localhost:9080/metrics

# Grafana health
curl http://localhost:3001/api/health
```

### Log Inspection

View logs directly from files:

```bash
# Service logs
tail -f docker/volumes/api-gateway/logs/api-gateway.log

# Docker logs (includes stdout)
docker logs service-boilerplate-api-gateway
```

### Troubleshooting

#### Logs Not Appearing in Grafana

1. **Check Promtail targets**:
   ```bash
   curl http://localhost:9080/targets
   ```

2. **Verify log file paths** in Promtail config match volume mounts

3. **Check Loki ingestion**:
   ```bash
   curl "http://localhost:3100/loki/api/v1/query?query={service=\"api-gateway\"}&limit=1"
   ```

4. **Validate JSON structure** - ensure logs are valid JSON

#### High Log Volume

1. **Adjust log levels** in production:
   ```bash
   LOGGING_LEVEL=warn  # Reduce verbosity
   ```

2. **Implement log sampling** for high-traffic endpoints

3. **Configure Loki retention** policies in `loki-config.yml`

#### Performance Issues

1. **Monitor Promtail metrics** for backpressure
2. **Check disk I/O** on log volumes
3. **Adjust Loki resource limits** if needed

## Development Workflow

### Local Development

1. **Start logging stack**:
   ```bash
   make dev
   ```

2. **View logs in Grafana**:
   - Open http://localhost:3001
   - Navigate to "Service Boilerplate - Logs" dashboard

3. **Debug with structured logs**:
   ```bash
   # Query specific request
   {request_id="your-request-id"}
   ```

### Testing Log Configuration

1. **Generate test logs**:
   ```bash
   curl http://localhost:8080/health
   ```

2. **Verify in Grafana**:
   - Check "Log Volume by Service" panel
   - Search for recent entries

3. **Validate JSON parsing**:
   ```bash
   # Check if logs are valid JSON
   jq . docker/volumes/api-gateway/logs/api-gateway.log | head -5
   ```

## Security Considerations

### Production Deployment

1. **Enable authentication** in Loki:
   ```yaml
   auth_enabled: true
   ```

2. **Configure TLS** for all components

3. **Restrict Grafana access** with proper authentication

4. **Implement log retention** policies

5. **Monitor log volume** and costs

### Log Data Protection

1. **Sanitize sensitive data** before logging
2. **Use appropriate log levels** for sensitive operations
3. **Implement log encryption** at rest if required
4. **Configure access controls** for log viewing

## Scaling and Performance

### Loki Scaling

For high-volume deployments:

```yaml
# Increase retention
table_manager:
  retention_deletes_enabled: true
  retention_period: 30d

# Add caching
query_range:
  results_cache:
    cache:
      embedded_cache:
        enabled: true
        max_size_mb: 512
```

### Promtail Scaling

For multiple nodes:

```yaml
# Use service discovery
scrape_configs:
  - job_name: services
    file_sd_configs:
      - files:
          - /etc/promtail/targets/*.yaml
```

### Storage Considerations

- **Loki uses filesystem storage** by default
- **Monitor disk usage** regularly
- **Implement backup strategies** for log data
- **Consider object storage** (S3, GCS) for production

## Integration with CI/CD

### Automated Testing

Include log validation in CI:

```bash
# Test log structure
./scripts/validate-logs.sh

# Check Promtail configuration
docker-compose config promtail
```

### Deployment Checklist

- [ ] Loki configuration updated for environment
- [ ] Promtail targets configured for new services
- [ ] Grafana dashboards provisioned
- [ ] Log volumes created with proper permissions
- [ ] Service logging configuration verified
- [ ] Log retention policies set
- [ ] Monitoring alerts configured

## Best Practices

### Log Design

1. **Use structured logging** with consistent field names
2. **Include correlation IDs** for request tracing
3. **Log at appropriate levels** (debug/info/warn/error)
4. **Avoid logging sensitive data** (passwords, tokens, PII)
5. **Use meaningful log messages** that provide context

### Operations

1. **Monitor log volume** and set up alerts
2. **Implement log rotation** and retention policies
3. **Regularly review and optimize** queries
4. **Document custom dashboards** and their purpose
5. **Backup critical logs** for compliance

### Development

1. **Test logging in development** before production
2. **Use debug logging** for troubleshooting
3. **Document log fields** for new services
4. **Follow established patterns** for consistency
5. **Review logs during code review** process

## Troubleshooting Guide

### Common Issues

#### "No logs appearing in Grafana"

**Symptoms**: Dashboard shows no data despite services running

**Solutions**:
1. Check Promtail targets: `curl http://localhost:9080/targets`
2. Verify log file paths match volume mounts
3. Ensure services are writing to correct directories
4. Check Loki ingestion: `curl "http://localhost:3100/loki/api/v1/query?query={service=\"api-gateway\"}"`

#### "Invalid JSON in logs"

**Symptoms**: Promtail parsing errors in logs

**Solutions**:
1. Validate log JSON: `jq . /path/to/logfile`
2. Check for ANSI escape codes in files
3. Ensure `LOGGING_STRIP_ANSI_FROM_FILES=true`
4. Verify service logging configuration

#### "High memory usage"

**Symptoms**: Loki or Promtail consuming excessive memory

**Solutions**:
1. Reduce query time ranges in Grafana
2. Implement log sampling for high-volume services
3. Adjust Loki cache settings
4. Monitor Promtail metrics for backpressure

#### "Log files not rotating"

**Symptoms**: Single large log file despite rotation settings

**Solutions**:
1. Check lumberjack configuration in service code
2. Verify file permissions for rotation
3. Ensure sufficient disk space
4. Restart service to trigger rotation

### Debug Commands

```bash
# Check Promtail status
curl http://localhost:9080/targets | jq .

# Query Loki directly
curl "http://localhost:3100/loki/api/v1/query_range?query={service=\"api-gateway\"}&start=$(date +%s - 3600)&end=$(date +%s)&limit=10"

# Validate log format
tail -1 docker/volumes/api-gateway/logs/api-gateway.log | jq .

# Check volume mounts
docker inspect service-boilerplate-api-gateway | jq '.[0].Mounts[] | select(.Destination == "/app/logs")'

# Monitor log volume
du -sh docker/volumes/*/logs/
```

## Future Enhancements

### Planned Features

- **Log alerting** based on patterns and thresholds
- **Log archiving** to long-term storage
- **Advanced parsing** with custom pipelines
- **Multi-tenant logging** for different environments
- **Log analytics** with machine learning insights
- **Integration with external tools** (ELK, Splunk)

### Configuration Extensions

- **Dynamic service discovery** for Promtail targets
- **Environment-specific dashboards** auto-provisioning
- **Custom log parsers** for complex formats
- **Log encryption** and compliance features
- **Advanced retention policies** with tiered storage

---

## Quick Reference

### Ports
- **Grafana**: 3001
- **Loki**: 3100
- **Promtail**: 9080

### Volumes
- **Logs**: `docker/volumes/{service}/logs/`
- **Loki Data**: `docker/volumes/loki/data/`
- **Grafana Data**: `docker/volumes/grafana/data/`

### Key Files
- **Loki Config**: `docker/loki-config.yml`
- **Promtail Config**: `docker/promtail-config.yml`
- **Grafana Dashboard**: `docker/grafana/dashboards/service-logs.json`

### Essential Commands
```bash
# Start logging stack
make dev

# View logs in Grafana
open http://localhost:3001

# Check Promtail status
curl http://localhost:9080/targets

# Query logs directly
{level="error", service="api-gateway"}
```

This centralized logging setup provides a robust, scalable foundation for observability in microservices architectures.