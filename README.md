# Cloud Resource Management Platform

An educational, portfolio-focused Go backend that will simulate managing cloud virtual machines through a REST API. We are building the project incrementally, testing and explaining each milestone before adding more features.

> **Current status:** Milestone 1 — Go API foundation. VM management, authentication, PostgreSQL, Docker and CI/CD are planned but **not implemented yet**.

## Technology choices

- Go (standard-library `net/http` for the initial API)
- PostgreSQL (upcoming milestone)
- Docker Compose (upcoming milestone)
- REST/JSON; Go automated tests; GitHub Actions (upcoming milestones)

## Requirements

- Go 1.27 or newer (see https://go.dev/dl/)

## Run locally

```bash
go run ./cmd/api
```

The API listens on `http://localhost:8080` by default. To choose a different port:

```bash
PORT=9090 go run ./cmd/api
```

In a second terminal:

```bash
curl -i http://localhost:8080/health
```

Expected HTTP 200 response with JSON body:

```json
{"status":"ok"}
```

## Test

```bash
go test ./...
```

Tests cover the health endpoint, method restrictions, and unknown routes.

## Structure

```text
cmd/api/main.go                 # API process entry point
internal/httpapi/router.go      # HTTP routes and handlers
internal/httpapi/router_test.go # Handler tests
docs/ROADMAP.md                 # Scope and milestone checklist
```

## Scope

The eventual API will let authenticated users create, list, inspect, start, stop and delete **simulated** VMs, with validated lifecycle transitions and event history. It will not provision real cloud compute.

See [the development roadmap](docs/ROADMAP.md) for planned functionality and progress.
