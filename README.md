# pulse

`pulse` is an early-stage active monitoring system for checking the availability of services and infrastructure resources.

The project is focused on config-driven health checks for things like:

- HTTP APIs
- TCP endpoints
- gRPC health checks
- DNS records
- TLS certificates

In the future, the project is planned to grow toward:

- additional check types for databases, queues, and other infrastructure components
- inbound health/status signals from external systems
- internal and public status pages generated from collected state

## Current Status

This project is still in an early stage of development.

What exists today:

- configuration loading and validation
- typed check configuration models
- periodic check execution with retries and jitter
- HTTP checks
- TCP checks
- DNS checks
- TLS certificate checks
- gRPC health checks via `grpc.health.v1.Health/Check`

What does not exist yet in a finished form:

- persistent state storage
- finalized status aggregation rules
- public/internal status page generation
- production-ready logging and observability
- stable external interfaces

## Running

The application expects configuration to be provided through `CONFIG_DIR`.

Expected layout:

- `services.yaml`
- `checks/*.yaml`

Example configuration files are available in:

- `examples/services.yaml`
- `examples/checks/api-checks.yaml`

Run with:

```bash
CONFIG_DIR=./examples go run ./cmd/pulse
```

## Notes

A few implementation details are intentionally narrow at this stage:

- gRPC support currently targets only the standard health check API: `grpc.health.v1.Health/Check`
- result writing is still backed by a temporary fake writer
- internal architecture is still evolving

## Important

Everything may change.

This includes:

- configuration format
- internal architecture
- runtime behavior
- package layout
- public interfaces

At this stage, the repository should be treated as an evolving prototype rather than a stable production-ready system.
