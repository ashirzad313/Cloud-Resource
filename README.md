# Cloud Resource Management Platform

A portfolio-focused backend project built incrementally with Go, REST APIs, PostgreSQL and Docker. It **simulates** a cloud provider's VM management API; it does not provision real virtual machines.

We learn each component by discussing the design, running the application, and verifying behavior manually before starting the next milestone. This README is the **single source of project documentation** and grows with each merged milestone.

## Project status

**Milestone 1 complete — Go HTTP API foundation.** The server and health endpoint have been run and verified manually.

### Roadmap

- [x] **1. Go API foundation (2–3 h):** entry point, HTTP routing, JSON health endpoint, manual verification.
- [ ] **2. Docker and PostgreSQL (4–6 h):** Dockerfiles, Compose, containers and networking.
- [ ] **3. Database design (6–8 h):** tables, constraints, SQL migrations and repository layer.
- [ ] **4. VM REST API (8–10 h):** create, list, get and delete VM records.
- [ ] **5. VM lifecycle (8–12 h):** start/stop commands, valid state transitions, worker and event history.
- [ ] **6. Authentication (8–12 h):** signup, password hashing, tokens and resource ownership.
- [ ] **7. API improvements (4–6 h):** validation, pagination, filtering and consistent errors.
- [ ] **8. Quality and verification (8–10 h):** manual API and integration scenarios, failure cases; revisit automated regression tests as complexity grows.
- [ ] **9. Delivery and CI/CD (6–9 h):** production Docker build, GitHub Actions and API documentation.
- [ ] **10. Final polish (4–6 h):** observability, graceful shutdown, security review and interview-ready README.

Estimated total: **58–82 learning hours**. Each milestone is developed in a separate branch, merged into `main` after verification, and then the working branch is deleted.

## Current tech stack

- Go 1.27+, using the standard-library `net/http` and `encoding/json` packages
- PostgreSQL — planned
- Docker / Docker Compose — planned
- GitHub Actions CI/CD — planned

## Architecture

For now, `cmd/api/main.go` starts the HTTP server and `internal/httpapi/router.go` registers routes and handlers.

Target architecture as the project grows:

```text
Client (curl / Postman)
          |
       HTTP API
          |
       Services (business rules, VM state machine)
          |
       Repositories (SQL access)
          |
       PostgreSQL
```

A background worker inside the same Go application will eventually process simulated VM transitions. We will use a modular monolith rather than microservices or Kubernetes.

## Run locally

Prerequisites: Git and Go 1.27 or newer.

```bash
go run ./cmd/api
```

The API listens on `http://localhost:8080` by default. To use a different port:

```bash
PORT=9090 go run ./cmd/api
```

### Manually verify Milestone 1

From a **second terminal**:

```bash
curl -i http://localhost:8080/health
curl -i -X POST http://localhost:8080/health
curl -i http://localhost:8080/missing
```

Expected outcomes:

| Request | HTTP status | Notes |
| --- | --- | --- |
| `GET /health` | 200 | JSON body: `{"status":"ok"}` |
| `POST /health` | 405 | Method Not Allowed |
| `GET /missing` | 404 | Not Found |

We currently use manual endpoint checks, by project preference, rather than committing automated test files at this stage.

## Current repository structure

```text
cmd/api/main.go              # API entry point and server setup
internal/httpapi/router.go   # HTTP routing and handlers
go.mod                       # Go module
README.md                    # Documentation and milestone checklist
.gitignore                   # Local file exclusions
```

## Planned functionality

Authenticated users will manage their own simulated VMs: create, list, inspect, start, stop and delete instances, with defined resource sizes and regions, valid state transitions and lifecycle event history.

Out of scope initially: real VM provisioning, Kubernetes, distributed microservices, billing and a frontend dashboard.
