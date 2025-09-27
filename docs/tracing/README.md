# Distributed Tracing Documentation

This directory contains comprehensive documentation for the distributed tracing implementation in the service boilerplate.

## 📚 Documentation Overview

| Document | Audience | Purpose |
|----------|----------|---------|
| [Overview & Architecture](overview.md) | All developers | High-level understanding of tracing system |
| [Tools & Libraries](tools.md) | Architects, DevOps | Technical stack and dependencies |
| [Implementation Details](implementation.md) | Backend developers | Code-level implementation details |
| [Service Template Integration](template.md) | Service creators | How tracing is integrated into new services |
| [Developer Guide](developer-guide.md) | Application developers | How to instrument new endpoints and operations |
| [Configuration](configuration.md) | DevOps, Platform engineers | Environment setup and configuration |
| [Monitoring & Troubleshooting](monitoring.md) | SREs, DevOps | Observability and debugging |
| [Best Practices](best-practices.md) | All developers | Guidelines and recommendations |

## 🚀 Quick Start

1. **New Service**: Tracing is automatically enabled when creating services with `create-service.sh`
2. **Existing Services**: Tracing is already implemented in API Gateway, Auth Service, and User Service
3. **View Traces**: Access Jaeger UI at `http://localhost:16686` in development
4. **Configuration**: Modify tracing settings in each service's `config.yaml`

## 🎯 Key Features

- **Automatic Instrumentation**: HTTP endpoints are automatically traced
- **Distributed Propagation**: Traces flow across service boundaries
- **Jaeger Integration**: Industry-standard tracing backend
- **Performance Optimized**: Configurable sampling rates
- **Developer Friendly**: Simple APIs for custom instrumentation

## 📋 Prerequisites

- Go 1.23+
- Docker & Docker Compose
- Understanding of microservices architecture

## 🔧 Architecture at a Glance

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │───▶│  Auth Service  │───▶│  User Service   │
│                 │    │                 │    │                 │
│ • HTTP Tracing  │    │ • HTTP Tracing  │    │ • HTTP Tracing  │
│ • Header Inject │    │ • Client Prop.  │    │ • DB Tracing    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌─────────────────────┐
                    │      Jaeger UI      │
                    │    (localhost:16686)│
                    └─────────────────────┘
```

## 📖 Getting Started

1. **Read the [Overview](overview.md)** to understand the architecture
2. **Check [Configuration](configuration.md)** for your environment
3. **Follow the [Developer Guide](developer-guide.md)** for custom instrumentation
4. **Use [Monitoring](monitoring.md)** for troubleshooting

## 🤝 Contributing

When adding new tracing instrumentation:

1. Follow the patterns in existing services
2. Update this documentation if needed
3. Test trace propagation across services
4. Ensure performance impact is minimal

## 📞 Support

- **Jaeger UI**: `http://localhost:16686` (development)
- **Service Logs**: Check individual service logs for tracing errors
- **Configuration**: Verify `tracing.enabled: true` in service configs

---

*This documentation covers OpenTelemetry-based distributed tracing with Jaeger backend.*</content>
</xai:function_call">README.md