# 🔄 JWT Key Rotation

This document provides comprehensive guidance on JWT key rotation in the service-boilerplate project, including automatic rotation, manual operations, monitoring, and troubleshooting.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [Automatic Rotation](#automatic-rotation)
- [Manual Rotation](#manual-rotation)
- [Monitoring & Health Checks](#monitoring--health-checks)
- [Database Schema](#database-schema)
- [Operational Procedures](#operational-procedures)
- [Troubleshooting](#troubleshooting)
- [Security Considerations](#security-considerations)

## Overview

JWT key rotation is a critical security practice that ensures cryptographic keys are regularly updated to minimize the impact of potential key compromise. The service-boilerplate implements comprehensive key rotation with:

- **Automatic time-based rotation** (default: 30 days)
- **Manual rotation capabilities** for immediate security response
- **Health monitoring** and alerting
- **Audit logging** for compliance
- **Zero-downtime rotation** with key overlap

## Architecture

### Key Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ KeyRotation     │    │   JWTUtils       │    │   Database      │
│ Manager         │────│   Service        │────│   (jwt_keys)    │
│                 │    │                  │    │                 │
│ • Background    │    │ • Key Generation │    │ • Key Storage   │
│   monitoring    │    │ • Rotation Logic │    │ • Validation    │
│ • Rotation      │    │                  │    │ • Metadata      │
│   triggers      │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Rotation Flow

1. **Monitoring**: Background goroutine checks rotation status every hour
2. **Trigger**: Rotation initiated when time interval exceeded or manual request
3. **Generation**: New RSA key pair generated with unique key ID
4. **Storage**: New key stored in database with metadata
5. **Transition**: New key becomes active, old key marked for expiration

## Current Implementation Status ✅

### Recently Implemented (January 2026)

**Thread-Safe JWT Utils:**
- ✅ Added `sync.RWMutex` for concurrent key access protection
- ✅ Implemented thread-safe token generation with mutex locking
- ✅ Enhanced `RefreshActiveKey()` method for database synchronization

**Periodic Key Refresh:**
- ✅ Added `StartKeyRefresher()` for automatic 5-minute database checks
- ✅ Implemented background goroutine for continuous key synchronization
- ✅ Service startup integration with automatic refresher activation

**Database Synchronization:**
- ✅ Enhanced `loadActiveKey()` to fetch current active key from database
- ✅ Added key change detection and in-memory update logic
- ✅ Implemented proper error handling and logging for refresh failures

**Zero-Downtime Rotation:**
- ✅ Key overlap support (60 minutes) during rotation transition
- ✅ Automatic key activation with old key expiration
- ✅ Comprehensive audit logging for rotation events

### Key Issues Resolved

**Before Fix:**
- ❌ Race conditions during concurrent token signing
- ❌ No automatic key synchronization with database
- ❌ Manual service restarts required for key changes
- ❌ Integration test failures due to stale key usage

**After Fix:**
- ✅ Thread-safe concurrent key access
- ✅ Automatic detection of database key changes within 5 minutes
- ✅ Seamless key rotation without service restart
- ✅ Consistent key usage across token issuance and validation
- ✅ Integration test reliability restored

## Configuration

### Environment Variables

```bash
# JWT Configuration
JWT_ROTATION_ENABLED=true
JWT_ROTATION_INTERVAL_DAYS=30
JWT_ROTATION_OVERLAP_MINUTES=60
JWT_KEY_REFRESH_INTERVAL_MINUTES=5
JWT_KEY_MONITORING_ENABLED=true
```

### Service Integration

**Auth Service Configuration:**
```yaml
services:
  auth-service:
    environment:
      - JWT_ROTATION_ENABLED=${JWT_ROTATION_ENABLED:-true}
      - JWT_KEY_REFRESH_INTERVAL_MINUTES=${JWT_KEY_REFRESH_INTERVAL_MINUTES:-5}
```

**API Gateway Integration:**
```yaml
services:
  api-gateway:
    environment:
      - AUTH_SERVICE_URL=http://auth-service:8083
      - JWT_KEY_CACHE_TTL=${JWT_KEY_CACHE_TTL:-300}  # 5 minutes for rapid refresh
      - JWT_FORCE_REFRESH=${JWT_FORCE_REFRESH:-false}
```

## Operational Procedures

### Manual Rotation

```bash
# Force immediate key rotation
curl -X POST http://auth-service:8083/api/v1/admin/rotate-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "security_incident", "force": true}'
```

### Key Status Check

```bash
# Check current active JWT key
curl -X GET http://auth-service:8083/api/v1/admin/jwt-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json"
```

## Monitoring & Health Checks

### JWT Health Endpoints

**Auth Service:**
- `GET /admin/jwt-keys` - Current key status and history
- `GET /admin/jwt-keys/{key_id}` - Specific key details
- `POST /admin/rotate-keys` - Manual rotation trigger
- `POST /webhooks/key-rotated` - Immediate key update notifications

**API Gateway:**
- `GET /health` - Includes JWT key validation test
- `POST /webhooks/key-rotated` - Receives auth service notifications

### Health Check Metrics

```bash
# Validate JWT system health
curl -s http://api-gateway:8080/health | jq '.jwt'
{
  "status": "healthy",
  "active_key_id": "jwt-key-abc123",
  "key_freshness_minutes": 2,
  "last_rotation": "2025-01-15T10:30:00Z",
  "refresh_errors_24h": 0
}
```

## Security Considerations

### Key Security Standards

- **RSA-2048** minimum key size for security
- **PKCS#8** format for private key storage
- **Key rotation** every 30 days or on security incidents
- **Immediate revocation** capability for compromised keys
- **Audit logging** for all key operations and access

### Access Control

```go
// Key management should require admin privileges
if !userHasRole("jwt_admin") {
    return errors.New("insufficient permissions for key management")
}
```

## Future Enhancement Plan 🚀

### Phase 1: Enhanced Key Distribution (Q2 2026)

**Real-time Key Synchronization:**
- [ ] Implement Redis pub/sub for immediate key rotation notifications
- [ ] Add webhook system for services not using Redis
- [ ] Implement service discovery integration for automatic key source updates

**Improved Caching Strategy:**
- [ ] Add key versioning to support incremental updates
- [ ] Implement smart cache invalidation based on rotation events
- [ ] Add key health metrics and monitoring dashboards

### Phase 2: Advanced Security Features (Q3 2026)

**Multi-Tenant Support:**
- [ ] Per-tenant JWT key isolation and rotation
- [ ] Tenant-specific key policies and compliance rules
- [ ] Cross-tenant token validation with dedicated key stores

**Enhanced Audit Trail:**
- [ ] Immutable audit logs with tamper-evident storage
- [ ] Compliance reporting for GDPR/CCPA requirements
- [ ] Automated security scanning and vulnerability detection

### Phase 3: Performance & Scalability (Q4 2026)

**High-Performance JWT:**
- [ ] Hardware security module (HSM) integration for key protection
- [ ] Elliptic curve cryptography (ECDSA) support for better performance
- [ ] JWT token caching and pooling for high-throughput scenarios
- [ ] Load testing and performance benchmarking

## Implementation Guidelines

### Key Rotation Enhancement Tasks

**When implementing new features:**

1. **Backward Compatibility**: Ensure existing services continue working
2. **Graceful Migration**: Support old and new key formats during transition
3. **Monitoring**: Add comprehensive metrics for all key operations
4. **Testing**: Include JWT rotation scenarios in integration test suite
5. **Security**: Follow NIST and OWASP guidelines for key management
6. **Documentation**: Update this document and API documentation

### Code Examples

**Advanced Key Refresh:**
```go
// Enhanced key refresher with multiple triggers
func (j *JWTUtils) StartAdvancedKeyRefresher(ctx context.Context, config AdvancedRefreshConfig) {
    go func() {
        select {
            case <-ctx.Done():
                return
            case <-config.RotationEvents.Chan():
                j.handleRotationEvent(event)
            case <-config.HealthCheck.Ticker.C:
                j.performHealthCheck()
            case <-time.After(config.RefreshInterval):
                if err := j.RefreshActiveKey(ctx); err != nil {
                    j.logger.WithError(err).Error("Key refresh failed")
                }
        }
    }()
}
```

**Event-Driven Updates:**
```go
// Handle key rotation events from Redis/webhooks
func (j *JWTUtils) handleRotationEvent(event KeyRotationEvent) error {
    switch event.EventType {
    case "key_rotated":
        return j.updateKeyFromEvent(event.NewKeyID, event.PublicPEM)
    case "key_revoked":
        return j.revokeKeyImmediate(event.KeyID)
    default:
        return fmt.Errorf("unknown rotation event type: %s", event.EventType)
    }
}
```

## Troubleshooting

### Common Issues

**Stale Keys:**
```bash
# Force cache invalidation in API Gateway
curl -X POST http://api-gateway:8080/admin/cache/invalidate \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"cache_type": "jwt_keys"}'
```

**Key Sync Failures:**
```bash
# Check key synchronization logs
docker logs auth-service | grep "JWT key refresh\|synchronization failed"

# Force immediate refresh
curl -X POST http://auth-service:8083/api/v1/admin/force-refresh \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

**Service Discovery Issues:**
```bash
# Validate service connectivity
curl -f http://auth-service:8083/health || echo "Auth service unreachable"
curl -f http://user-service:8081/health || echo "User service unreachable"
```

**Last Updated**: January 30, 2026
**Next Review**: March 31, 2026
**Responsible**: Infrastructure & Security Team
**Status**: ✅ **Current implementation stable** - All issues resolved

📋 **For detailed implementation plan, see: [JWT Enhancement Plan](jwt-enhancement-plan.md)**
6. **Cleanup**: Old keys expired after overlap period (default: 60 minutes)
7. **Audit**: All operations logged with actor identification

## Key Distribution and Caching

### API Gateway Key Retrieval

The API Gateway implements dynamic JWT public key distribution with intelligent caching:

- **Key Source**: Public keys retrieved from auth-service `/public-key` endpoint
- **Cache TTL**: Keys cached for 1 hour to minimize network overhead
- **Background Refresh**: Automatic refresh every 30 minutes to ensure key freshness
- **Retry Logic**: Exponential backoff with up to 10 retry attempts for resilient fetching
- **Health Checks**: Auth-service health verification before key retrieval attempts
- **Fallback Support**: Environment variable `JWT_PUBLIC_KEY` when auth-service unavailable

### Cache Management Flow

1. **Initial Load**: On startup, attempt to fetch key from auth-service
2. **Cache Hit**: Use cached key if within TTL (1 hour)
3. **Background Refresh**: Periodic refresh every 30 minutes regardless of TTL
4. **Failure Handling**: Use expired cached key if refresh fails (better than no key)
5. **Fallback**: Use environment variable if all retrieval attempts fail

### Cache Configuration

```go
type jwtKeyCache struct {
    key       interface{}  // Cached RSA public key
    fetchedAt time.Time    // Timestamp of last successful fetch
    ttl       time.Duration // Cache TTL (default: 1 hour)
}
```

### Monitoring Cache Status

The API Gateway status endpoint includes key cache information:

```bash
curl http://localhost:8080/status
```

**Response includes:**
- Key cache status (healthy/unhealthy)
- Last key fetch timestamp
- Next scheduled refresh time
- Cache hit/miss statistics

## Configuration

### Rotation Configuration

The key rotation system is configured in the auth-service with the following parameters:

```go
rotationConfig := services.RotationConfig{
    Enabled:        true,           // Enable/disable automatic rotation
    Type:           "time",         // Rotation trigger: "time", "usage", "manual"
    IntervalDays:   30,             // Days between automatic rotations
    MaxTokens:      100000,         // Max tokens per key (for usage-based rotation)
    OverlapMinutes: 60,             // Minutes to keep old key active
    CheckInterval:  1 * time.Hour,  // How often to check for rotation
}
```

### Environment Variables

Add to your environment configuration:

```bash
# Key Rotation Settings
JWT_ROTATION_ENABLED=true
JWT_ROTATION_TYPE=time
JWT_ROTATION_INTERVAL_DAYS=30
JWT_ROTATION_MAX_TOKENS=100000
JWT_ROTATION_OVERLAP_MINUTES=60
JWT_ROTATION_CHECK_INTERVAL=1h
```

### Configuration Validation

The system validates configuration on startup:

- `IntervalDays` must be between 1-365 days
- `OverlapMinutes` must be between 1-1440 minutes (24 hours)
- `CheckInterval` must be reasonable (not too frequent)

## Automatic Rotation

### Time-Based Rotation

The default rotation strategy rotates keys based on time intervals:

```go
// Check every hour if 30 days have passed since last rotation
if timeSinceLastRotation >= 30 * 24 * time.Hour {
    performRotation("time_interval_exceeded_30_days")
}
```

### Usage-Based Rotation (Future)

Planned feature for rotation based on token issuance count:

```go
// Rotate when key has issued too many tokens
if tokensIssuedByCurrentKey >= maxTokensPerKey {
    performRotation("usage_limit_exceeded")
}
```

### Background Monitoring

The `KeyRotationManager` runs as a background service:

```go
func (krm *KeyRotationManager) Start(ctx context.Context) {
    go krm.rotationLoop(ctx)
}

func (krm *KeyRotationManager) rotationLoop(ctx context.Context) {
    ticker := time.NewTicker(krm.config.CheckInterval)
    for {
        select {
        case <-ticker.C:
            krm.checkAndRotate(ctx)
        case <-ctx.Done():
            return
        }
    }
}
```

## Manual Rotation

### API Endpoint

Manual rotation requires admin privileges:

```bash
# Manual key rotation (admin only)
curl -X POST http://localhost:8083/api/v1/admin/rotate-keys \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json"
```

**Response:**

```json
{
  "message": "JWT keys rotated successfully"
}
```

### Admin Requirements

Only users with `admin` role can perform manual rotation:

```go
// Check admin role in handler
userRoles := middleware.GetAuthenticatedUserRoles(c)
isAdmin := false
for _, role := range userRoles {
    if role == "admin" {
        isAdmin = true
        break
    }
}
```

### Audit Logging

All rotation operations are audited:

```json
{
  "time": "2025-01-15T10:30:00Z",
  "level": "warn",
  "event_type": "admin_rotate_keys",
  "user_id": "admin-user-123",
  "entity_id": "",
  "service": "auth-service",
  "trace_id": "abc123",
  "span_id": "def456",
  "result": "success",
  "ip_address": "192.168.1.100",
  "user_agent": "curl/7.68.0"
}
```

## Monitoring & Health Checks

### Health Check Endpoint

The `/health` endpoint includes rotation status:

```bash
curl http://localhost:8083/health
```

**Response:**

```json
{
  "status": "healthy",
  "service": {
    "name": "auth-service",
    "version": "1.0.0",
    "environment": "production",
    "uptime": "2h30m15s"
  },
  "rotation": {
    "status": "healthy",
    "type": "time",
    "enabled": true,
    "days_since_last": 15.5,
    "next_rotation": "2025-02-14T10:30:00Z",
    "response_time": "5ms"
  }
}
```

### Status Endpoint

Detailed status at `/status`:

```bash
curl http://localhost:8083/status
```

**Response:**

```json
{
  "status": "healthy",
  "rotation": {
    "status": "healthy",
    "type": "time",
    "enabled": true,
    "days_since_last": 15.5,
    "next_rotation": "2025-02-14T10:30:00Z",
    "response_time": "5ms"
  },
  "jwt_keys": {
    "status": "healthy",
    "key_id": "key_20250101_abc123",
    "response_time": "2ms"
  }
}
```

### Monitoring Metrics

Key metrics to monitor:

- **Rotation Status**: Days since last rotation
- **Next Rotation Due**: When automatic rotation will occur
- **Key Health**: Current active key ID and accessibility
- **Key Cache Status**: API Gateway key cache health and refresh status
- **Key Retrieval Attempts**: Success/failure rates of key fetching from auth-service
- **Cache TTL Status**: Time until next key refresh
- **Rotation Failures**: Failed rotation attempts
- **Audit Events**: Rotation operation logs

### Alerting

Set up alerts for:

- Rotation failures
- Keys not rotated for >35 days (past default interval)
- Health check failures
- Manual rotation operations (security events)

## Database Schema

### JWT Keys Table

```sql
CREATE TABLE auth_service.jwt_keys (
    id BIGSERIAL PRIMARY KEY,
    key_id VARCHAR(255) UNIQUE NOT NULL,
    private_key_pem TEXT NOT NULL,
    public_key_pem TEXT NOT NULL,
    algorithm VARCHAR(50) DEFAULT 'RS256',
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    rotation_reason TEXT,
    rotated_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB
);
```

### Key Rotation Config Table

```sql
CREATE TABLE auth_service.key_rotation_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rotation_type VARCHAR(50) NOT NULL DEFAULT 'time',
    interval_days INTEGER DEFAULT 30,
    max_tokens INTEGER DEFAULT 100000,
    overlap_minutes INTEGER DEFAULT 60,
    enabled BOOLEAN DEFAULT true,
    last_rotation_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Indexes

```sql
CREATE INDEX idx_jwt_keys_rotated_at ON auth_service.jwt_keys(rotated_at);
CREATE INDEX idx_key_rotation_config_enabled ON auth_service.key_rotation_config(enabled);
```

## Operational Procedures

### Routine Monitoring

1. **Daily Health Checks**:

   ```bash
   # Check rotation status
   curl http://localhost:8083/health | jq '.rotation'
   ```

2. **Weekly Review**:

   - Verify rotation is enabled
   - Check days since last rotation
   - Review audit logs for rotation events

3. **Monthly Audit**:
   - Confirm automatic rotations are occurring
   - Verify key history and cleanup
   - Check for any rotation failures

### Emergency Rotation

For security incidents requiring immediate key rotation:

1. **Assess Impact**: Determine if immediate rotation is needed
2. **Notify Stakeholders**: Alert team of planned rotation
3. **Perform Rotation**:

   ```bash
   curl -X POST http://localhost:8083/api/v1/admin/rotate-keys \
     -H "Authorization: Bearer $ADMIN_TOKEN"
   ```

4. **Verify Success**: Check health endpoint and logs
5. **Monitor Impact**: Watch for authentication failures
6. **Communicate**: Update team on rotation completion

### Configuration Changes

To modify rotation settings:

1. **Update Configuration**: Modify rotation config in `main.go`
2. **Restart Service**: Apply configuration changes
3. **Verify Settings**: Check status endpoint
4. **Monitor Behavior**: Ensure rotations occur as expected

### Backup and Recovery

- **Key Backup**: JWT keys are stored in database (include in backups)
- **Configuration Backup**: Rotation settings are in code/config
- **Recovery**: Keys are regenerated if lost (tokens become invalid)

## Troubleshooting

### Common Issues

#### 1. Rotation Not Occurring

**Symptoms:**

- Days since rotation keeps increasing
- No rotation log entries

**Causes & Solutions:**

- **Rotation Disabled**: Check `enabled: true` in config
- **Service Restart**: Rotation manager stops on service restart
- **Database Issues**: Check database connectivity
- **Time Zone Issues**: Ensure server time is correct

**Debug:**

```bash
# Check rotation status
curl http://localhost:8083/status | jq '.rotation'

# Check logs for rotation attempts
docker logs auth-service | grep rotation
```

#### 2. Manual Rotation Fails

**Symptoms:**

- 403 Forbidden on rotation endpoint
- 500 Internal Server Error

**Causes & Solutions:**

- **Not Admin**: User must have `admin` role
- **Token Expired**: Use valid admin JWT token
- **Database Error**: Check database connectivity
- **Key Generation Failure**: Check entropy/randomness

**Debug:**

```bash
# Verify admin token
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8083/api/v1/auth/me

# Check service logs
docker logs auth-service | grep "rotate-keys"
```

#### 3. Health Check Shows Unhealthy

**Symptoms:**

- Rotation status: "unhealthy"
- Error messages in health response

**Causes & Solutions:**

- **Database Connection**: Check PostgreSQL connectivity
- **Key Access Issues**: Verify key storage and retrieval
- **Configuration Errors**: Validate rotation config
- **Permission Issues**: Check database user permissions

**Debug:**

```bash
# Detailed status
curl http://localhost:8083/status

# Database connectivity
docker exec auth-service pg_isready -h postgres -U postgres
```

#### 4. Tokens Become Invalid After Rotation

**Symptoms:**

- Users getting 401 errors after rotation
- Refresh tokens failing

**Causes & Solutions:**

- **Overlap Period**: Old keys should remain valid during overlap
- **Client Caching**: Clients may cache public keys
- **Gateway Issues**: API gateway may need key refresh

**Debug:**

```bash
# Check active keys
docker exec -it auth-service psql -U postgres -d service_db \
  -c "SELECT key_id, is_active, rotated_at FROM auth_service.jwt_keys ORDER BY created_at DESC LIMIT 5;"

# Check gateway logs
docker logs api-gateway | grep "jwt\|auth"
```

#### 5. Key Retrieval Failures

**Symptoms:**

- API Gateway unable to fetch JWT public keys from auth-service
- Authentication failures despite valid tokens
- Gateway logs show "Failed to fetch public key" errors
- Fallback to environment variable `JWT_PUBLIC_KEY`

**Causes & Solutions:**

- **Auth-service Unavailable**: Service not running or network issues
- **Health Check Failures**: Auth-service health endpoint returning errors
- **Key Generation Issues**: No active keys in auth-service database
- **Network Connectivity**: Docker network or firewall blocking communication
- **TLS/SSL Issues**: Certificate validation problems in production

**Debug:**

```bash
# Check auth-service health
curl http://localhost:8083/health

# Test public key endpoint directly
curl http://localhost:8083/public-key

# Check auth-service logs for key retrieval attempts
docker logs auth-service | grep "public-key\|JWT"

# Verify API Gateway can reach auth-service
docker exec api-gateway curl -v http://auth-service:8083/health

# Check for active keys in database
docker exec auth-service psql -U postgres -d service_db \
  -c "SELECT key_id, is_active, created_at FROM auth_service.jwt_keys WHERE is_active = true;"

# Check API Gateway status for cache information
curl http://localhost:8080/status
```

### Log Analysis

#### Rotation Events

```bash
# Find rotation events
docker logs auth-service | grep "key rotation"

# Manual rotations
docker logs auth-service | grep "admin_rotate_keys"
```

#### Audit Logs

```bash
# Rotation audit events
docker logs auth-service | jq 'select(.event_type == "admin_rotate_keys")'

# Failed operations
docker logs auth-service | jq 'select(.result == "failure")'
```

### Performance Issues

- **Frequent Checks**: Reduce `CheckInterval` if causing load
- **Database Load**: Monitor query performance on rotation tables
- **Memory Usage**: Key storage is minimal impact
- **Network**: Rotation doesn't affect normal operations

## Security Considerations

### Key Security

- **Private Key Protection**: Keys encrypted at rest in database
- **Access Control**: Only auth-service can access private keys
- **Audit Trail**: All key operations are logged
- **Secure Generation**: Cryptographically secure random generation

### Rotation Security

- **Regular Rotation**: Prevents long-term key compromise
- **Overlap Period**: Ensures no service disruption
- **Admin Control**: Manual rotation requires admin privileges
- **Audit Compliance**: Full logging for security reviews

### Operational Security

- **Monitoring**: Alert on rotation failures
- **Access Control**: Restrict manual rotation to admins
- **Incident Response**: Immediate rotation capability
- **Backup Security**: Secure key backup procedures

### Compliance

- **PCI DSS**: Regular key rotation requirements
- **SOX**: Audit trail requirements
- **GDPR**: Data protection through key rotation
- **ISO 27001**: Information security management

## Best Practices

### Configuration

- Use 30-day intervals for balance of security and usability
- Enable automatic rotation in production
- Set reasonable overlap periods (30-60 minutes)
- Monitor rotation health regularly

### Operations

- Test rotation in staging before production
- Have manual rotation procedures documented
- Monitor for rotation failures
- Keep audit logs for compliance periods

### Security

- Rotate keys immediately on suspected compromise
- Use admin roles judiciously for rotation access
- Monitor for unusual rotation patterns
- Regular security audits of key management

### Maintenance

- Regular review of rotation settings
- Update rotation intervals based on threat model
- Archive old keys securely
- Test recovery procedures

## Related Documentation

- [Security Architecture](security-architecture.md) - Overall security model
- [Authentication API Examples](auth-api-examples.md) - Auth flow examples
- [Troubleshooting Auth & Logging](troubleshooting-auth-logging.md) - Issue resolution
- [Database Migrations](migrations/) - Schema management
- [Logging System](logging-system.md) - Audit logging details

