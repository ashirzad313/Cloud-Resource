# Development roadmap

## Goal and scope

Build a portfolio-quality, well-tested cloud resource management **simulation** with a single Go API, PostgreSQL, Docker Compose, automated tests and GitHub Actions. This is not a cloud provider or a real hypervisor: VM states are simulated in persisted application data.

## Architecture (target)

HTTP handlers -> application services -> repository interfaces -> PostgreSQL

A background worker inside the same application will eventually process queued VM lifecycle transitions. Keep one deployable API application rather than introducing microservices or Kubernetes.

## Planned functional behavior

- User registration and authentication; users can access only their own resources.
- Create/list/get/delete VM instances with configured name, CPU, memory and region.
- Start and stop VMs via valid state transitions. Track events and errors.
- Basic validation, pagination and predictable API responses.
- Unit/integration tests; containerized local development; API docs and CI checks.

## Milestones and estimated learning time

- [ ] **1. Go API foundation (2–3 h):** establish layout, `/health` endpoint and initial HTTP tests. *Implementation prepared; verify locally and merge the PR to complete.*
- [ ] **2. Docker and PostgreSQL (4–6 h):** container basics, Compose services, environment configuration.
- [ ] **3. Database design (6–8 h):** schemas, constraints, migrations, repository layer.
- [ ] **4. VM REST API (8–10 h):** create/list/get/delete VM resources.
- [ ] **5. VM lifecycle (8–12 h):** state machine, start/stop, background processing, event history.
- [ ] **6. Auth and security (8–12 h):** accounts, password hashing, token authentication, ownership.
- [ ] **7. API improvements (4–6 h):** validation, pagination, filtering, error responses.
- [ ] **8. Automated testing (8–10 h):** unit, integration and failure-path testing.
- [ ] **9. Delivery and CI/CD (6–9 h):** multi-stage Docker build, GitHub Actions, API documentation.
- [ ] **10. Polish (4–6 h):** logging, graceful shutdown, security checks, README, interview narrative.

**Estimated total: 58–82 hours.** Estimates include learning, local verification and discussion. We advance only after the current milestone is understood and verified.

## Scope exclusions

No real VM provisioning, Kubernetes, microservice decomposition, payment/billing system or frontend dashboard in the initial project.
