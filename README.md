# Cloud Resource Management Platform

A portfolio-focused backend project built incrementally with Go, REST APIs, PostgreSQL and Docker. It **simulates** a cloud provider's VM management API; it does not provision real virtual machines.

We learn each component by discussing the design, running the application and verifying behavior manually before starting the next step. This README is the **single source of project documentation** and grows with each milestone.

## Project status

- **Milestone 1: complete.** Go HTTP server and `GET /health` have been manually verified.
- **Milestone 2: in progress, Step 1.** The Dockerfile and `.dockerignore` are prepared on the milestone branch. **Build and runtime verification on your machine are still pending.**
- **Milestone 2, Step 2 (not started):** PostgreSQL, Docker Compose and API-to-database connectivity.

### Roadmap

- [x] **1. Go API foundation (2–3 h):** entry point, HTTP routing, JSON health endpoint, manual verification.
- [ ] **2. Docker and PostgreSQL (4–6 h):** Dockerfile, Compose, containers and networking. *(Step 1: Dockerize the Go API — pending verification.)*
- [ ] **3. Database design (6–8 h):** tables, constraints, SQL migrations and repository layer.
- [ ] **4. VM REST API (8–10 h):** create, list, get and delete VM records.
- [ ] **5. VM lifecycle (8–12 h):** start/stop commands, valid state transitions, worker and event history.
- [ ] **6. Authentication (8–12 h):** signup, password hashing, tokens and resource ownership.
- [ ] **7. API improvements (4–6 h):** validation, pagination, filtering and consistent errors.
- [ ] **8. Quality and verification (8–10 h):** manual API and integration scenarios, failure cases; revisit automated regression tests as complexity grows.
- [ ] **9. Delivery and CI/CD (6–9 h):** production Docker build, GitHub Actions and API documentation.
- [ ] **10. Final polish (4–6 h):** observability, graceful shutdown, security review and interview-ready README.

Estimated total: **58–82 learning hours**. Each milestone is developed in a separate branch and merged into `main` only after verification; the working branch is then deleted.

## Current tech stack

- Go 1.27+, standard-library `net/http` and `encoding/json`
- Docker multi-stage build (Milestone 2, Step 1)
- PostgreSQL — planned
- Docker Compose — planned
- GitHub Actions CI/CD — planned

## Architecture

Currently, `cmd/api/main.go` starts the HTTP server and `internal/httpapi/router.go` registers routes and handlers.

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

We will use a modular monolith with an internal background worker rather than microservices or Kubernetes.

## Run locally without Docker (Milestone 1)

Prerequisite: Go 1.27 or newer.

```bash
go run ./cmd/api
```

The API listens on `http://localhost:8080`. To change the port:

```bash
PORT=9090 go run ./cmd/api
```

### Manual endpoint checks

From a second terminal:

```bash
curl -i http://localhost:8080/health
curl -i -X POST http://localhost:8080/health
curl -i http://localhost:8080/missing
```

| Request | HTTP status | Notes |
| --- | --- | --- |
| `GET /health` | 200 | JSON body: `{"status":"ok"}` |
| `POST /health` | 405 | Method Not Allowed |
| `GET /missing` | 404 | Not Found |

We currently perform manual checks by project preference, rather than committing automated test files.

## Run with Docker (Milestone 2, Step 1)

Prerequisite: Docker Desktop or Docker Engine with a running Docker daemon.

### 1. Build the Docker image

```bash
docker build -t cloud-resource:dev .
```

This builds a Go binary in the first stage and copies only the binary into a minimal, non-root runtime image in the second stage. The `scratch` runtime image contains no shell or Go compiler.

### 2. Run the container

Stop any other process using port 8080, then run:

```bash
docker run --rm --name cloud-api -p 127.0.0.1:8080:8080 cloud-resource:dev
```

This runs in the foreground. Keep the terminal open. The port mapping makes port 8080 inside the container reachable only on your computer's loopback address. `EXPOSE 8080` in the Dockerfile is just documentation; `-p` actually publishes the port.

### 3. Verify from a second terminal

```bash
curl -i http://localhost:8080/health
docker ps
```

Expected: HTTP 200 and `{"status":"ok"}`; `docker ps` lists `cloud-api` as running.

Optionally repeat the POST and missing-route checks above; their results should be unchanged.

### 4. Stop and clean up

Press `Ctrl+C` in the terminal running `docker run`. The `--rm` flag removes the stopped container automatically. The image remains available for reuse.

If `Ctrl+C` does not stop the container, run `docker stop cloud-api` in a second terminal.

### Troubleshooting

- **Cannot connect to the Docker daemon:** start Docker Desktop / the Docker service and retry `docker version`.
- **Port already in use:** stop the locally running Go server or map a different host port with `-p 127.0.0.1:8081:8080`, then request `localhost:8081`.
- **Container exits unexpectedly:** run `docker run --name cloud-api ...` without `--rm` to preserve it for `docker logs cloud-api`, then `docker rm cloud-api` after inspection.
- **Docker Hub image download fails:** check internet connectivity and that Docker can pull `golang:1.27-alpine`.

## Current repository structure

```text
cmd/api/main.go              # API entry point and server setup
internal/httpapi/router.go   # HTTP routing and handlers
go.mod                       # Go module
Dockerfile                   # Multi-stage Docker image definition
.dockerignore                # Exclusions from Docker build context
README.md                    # Documentation and milestone checklist
.gitignore                   # Git exclusions
```

## Planned functionality

Authenticated users will manage their own simulated VMs: create, list, inspect, start, stop and delete instances, with defined resource sizes and regions, valid state transitions and lifecycle event history.

Out of scope initially: real VM provisioning, Kubernetes, distributed microservices, billing and a frontend dashboard.
