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
- PostgreSQL-backed raw check execution storage
- persisted current check state
- persisted check state event history
- service-level state aggregation on read path
- migration CLI via `cmd/pulse-migrate`
- early REST API via `cmd/pulse-api`

What does not exist yet in a finished form:

- persisted service-level aggregation
- complete query/read API for current and historical state
- public/internal status page generation
- production-ready logging and observability
- stable external interfaces

## Running

The application expects configuration to be provided through `CONFIG_DIR`.
PostgreSQL is also required for execution and state storage.

Expected layout:

- `services.yaml`
- `checks/*.yaml`

Example configuration files are available in:

- `examples/services.yaml`
- `examples/checks/api-checks.yaml`

Validate configuration with:

```bash
API_HTTP_AUTH_TOKEN=dev-token API_HTTP_CUSTOM_HEADER=dev-value CONFIG_DIR=./examples go run ./cmd/pulse-validate
```

Before starting the app, apply migrations:

```bash
go run ./cmd/pulse-migrate up
```

Run with:

```bash
API_HTTP_AUTH_TOKEN=dev-token API_HTTP_CUSTOM_HEADER=dev-value CONFIG_DIR=./examples go run ./cmd/pulse
```

Run API with:

```bash
API_LISTEN_ADDR=:8080 INTERNAL_API_ENABLED=true CONFIG_DIR=./examples go run ./cmd/pulse-api
```

Check `spec` values may reference environment variables with `${VAR_NAME}` syntax:

```yaml
headers:
  Authorization: "Bearer ${API_HTTP_AUTH_TOKEN}"
```

Referenced variables must be present when the configuration is loaded. Values are resolved at check execution time and are not written back into the loaded config snapshot.

The API process has separate internal and public route groups.
Both are disabled by default and must be enabled explicitly:

- `INTERNAL_API_ENABLED=true` enables monitoring/admin-oriented endpoints
- `PUBLIC_API_ENABLED=true` enables public status endpoints

Implemented internal API endpoints:

- `GET /internal/v1/services` returns services with aggregated current status
- `GET /internal/v1/services/{serviceId}/checks/state`
- `GET /internal/v1/services/{serviceId}/checks/{checkId}/executions`
- `GET /internal/v1/services/{serviceId}/checks/{checkId}/timeline`
- `GET /internal/v1/services/{serviceId}/checks/{checkId}/buckets`

Placeholder public API endpoints:

- `GET /public/v1/status`

## Notes

A few implementation details are intentionally narrow at this stage:

- gRPC support currently targets only the standard health check API: `grpc.health.v1.Health/Check`
- raw execution history, current check state, and check state event history are stored in PostgreSQL
- API configuration is read from the latest valid hot-reloaded config snapshot
- internal architecture is still evolving
- timeline and bucket endpoints respect per-check `allowed_buckets`

## Important

Everything may change.

This includes:

- configuration format
- internal architecture
- runtime behavior
- package layout
- public interfaces

At this stage, the repository should be treated as an evolving prototype rather than a stable production-ready system.
