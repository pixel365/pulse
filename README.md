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

This project is in a very early stage of development.

What exists today:

- configuration loading and validation
- typed check configuration models
- basic application/checker scaffolding
- an initial HTTP checker implementation

What does not exist yet in a finished form:

- stable runtime behavior
- persistent state storage
- finalized status aggregation rules
- public/internal status page generation
- stable external interfaces

## Important

Everything may change.

This includes:

- configuration format
- internal architecture
- runtime behavior
- package layout
- public interfaces

At this stage, the repository should be treated as an evolving prototype rather than a stable production-ready system.
