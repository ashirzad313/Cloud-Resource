# Cloud Resource Management Platform

A portfolio-focused Go backend project that simulates cloud VM resource management. This is a learning-by-running project: one manageable step at a time, with a single cumulative README and manual verification before merging each milestone. No real VMs are provisioned.

## Project status and roadmap

**Milestone 1 complete:** Go HTTP server, routing and JSON health endpoint.

**Milestone 2 complete:** Dockerized Go API, PostgreSQL with Docker Compose, database connectivity and `/ready` manually verified.

**Milestone 3 in progress, Step 1:** Create the first versioned SQL migration for our `virtual_machines` table; manual database verification pending.

- [x] **1. Go API foundation (2–3 h):** entry point, HTTP routing, JSON health endpoint.
- [x] **2. Docker and PostgreSQL (4–6 h):** multi-stage Dockerfile, Compose networking, PostgreSQL connection and persistent volume; manually verified.
- [ ] **3. Database design (6–8 h):** schema, constraints, migrations and repository layer. *(Step 1: first VM table migration prepared; not yet applied locally.)*
- [ ] **4. VM REST API (8–10 h):** create, list, get and delete VM records.
- [ ] **5. VM lifecycle (8–12 h):** state transitions, worker and event history.
- [ ] **6. Authentication (8–12 h):** account signup, password hashing, tokens, ownership.
- [ ] **7. API improvements (4–6 h):** validation, pagination, filtering and errors.
- [ ] **8. Quality and verification (8–10 h):** manual integration and failure tests; revisit automated tests for regression prevention.
- [ ] **9. Delivery and CI/CD (6–9 h):** production Docker build, GitHub Actions, API documentation.
- [ ] **10. Polish (4–6 h):** logging, graceful shutdown, security and CV/interview-ready explanation.

**Estimated total: 58–82 learning hours.** Work in a milestone branch, merge into `main` only after local verification, then delete the branch.

## Technology stack

- Go 1.27+ with standard-library HTTP routing
- `github.com/jackc/pgx/v5/pgxpool` for PostgreSQL connectivity and connection pooling
- Docker multi-stage Go build; Docker Compose for two services
- PostgreSQL 17 (development database)
- Planned: schemas, VM lifecycle, authentication, GitHub Actions CI/CD

## Architecture today

```text
Mac / terminal
  |
  | curl localhost:8080
  v
Docker Desktop (Linux containers)
  |
  +-- api service (Go, port 8080)
  |     GET /health -> reports whether Go HTTP server is responding
  |     GET /ready  -> pings PostgreSQL; 503 when unavailable
  |         |
  |         | DATABASE_URL uses host "db", port 5432
  |         v
  +-- db service (PostgreSQL 17, internal port 5432)
          |
          +-- postgres_data (named Docker volume)
```

Docker Compose creates a private project network. The API can resolve `db` by service name; `localhost` inside the API would refer to the API container, not PostgreSQL. **PostgreSQL port 5432 is intentionally not published on the Mac**, so use `docker compose exec db` for local database inspection.

The database was initially empty in Milestone 2. In **Milestone 3**, we add a `virtual_machines` table using a versioned SQL migration. The `/ready` endpoint checks database connectivity, not schema correctness.

## Milestone 1: run without Docker

With Go 1.27 or newer installed:

```bash
go run ./cmd/api
```

Without a `DATABASE_URL`, the server can still run and `/health` returns HTTP 200, but `/ready` returns HTTP 503 because there is no configured database.

To use another port: `PORT=9090 go run ./cmd/api`.

## Milestone 2, Step 1: run just the API in Docker

```bash
docker build -t cloud-resource:dev .
docker run --rm --name cloud-api -p 127.0.0.1:8080:8080 cloud-resource:dev
```

In another terminal, `curl -i http://localhost:8080/health` should return HTTP 200 with `{"status":"ok"}`.

Stop the container using `Ctrl+C`. The `--rm` flag removes the stopped container. This mode does **not** connect PostgreSQL.

## Milestone 2, Step 2: API and PostgreSQL with Docker Compose

Prerequisites: Docker Desktop running, Docker Compose v2 (`docker compose`) available. No local PostgreSQL installation is required.

### 1. Configure local development credentials

```bash
cp .env.example .env
```

`.env.example` contains an **example local development password**. `.env` is excluded from Git and never committed. For now, keep the password alphanumeric because Compose inserts it into a connection URL. This is for local development, not a production secret-management design.

### 2. Start both services

First stop any manually running Go server or `docker run` container using host port 8080. Then:

```bash
docker compose up --build -d
docker compose ps
```

Compose starts `db`, waits for its PostgreSQL health check to pass, then starts `api`. The API connects using `DATABASE_URL=postgres://cloud_resource:...@db:5432/cloud_resource?sslmode=disable` (development only).

### 3. Verify API and database connectivity

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:8080/ready
docker compose logs api
```

Expected:

| Request | Status | Body |
| --- | --- | --- |
| `GET /health` | HTTP 200 | `{"status":"ok"}` |
| `GET /ready` | HTTP 200 | `{"status":"ready"}` |
| `GET /missing` | HTTP 404 | Default Go not-found response |

The API logs should include `connected to PostgreSQL`.

Ask PostgreSQL which database/user it is connected to:

```bash
docker compose exec db psql -U cloud_resource -d cloud_resource -c "SELECT current_database(), current_user;"
```

Expected: `cloud_resource` for both columns.

### 4. Optional: watch readiness react to a database outage

```bash
docker compose stop db
curl -i http://localhost:8080/health
curl -i http://localhost:8080/ready
docker compose start db
```

While PostgreSQL is stopped, the Go server can still answer `/health` with 200, but `/ready` should return HTTP 503 with `{"status":"not_ready"}`. After PostgreSQL has started and accepted connections again, `/ready` should return 200. This distinction is useful to orchestrators and monitoring systems.

### 5. Stop without deleting database data

```bash
docker compose down
```

This removes the Compose containers and network but keeps the named `postgres_data` volume, where PostgreSQL stores its files. **Do not use `docker compose down -v` unless you intentionally want to delete that database volume and its data.** We'll explore SQL persistence and migrations in the next milestone.

### Troubleshooting

- **Missing password variable:** run `cp .env.example .env` in the repository root.
- **Port 8080 already allocated:** stop your previous `go run` or `docker run` instance; or change the host-side port in `compose.yaml` to `127.0.0.1:8081:8080` and curl port 8081.
- **API exits at startup:** inspect `docker compose logs api` and `docker compose ps`; it fails fast when it cannot authenticate or reach PostgreSQL.
- **Database not healthy:** inspect `docker compose logs db`. On an old retained volume, changing `POSTGRES_PASSWORD` in `.env` does **not** reset an existing PostgreSQL user's password.
- **Docker daemon unavailable:** start Docker Desktop and retry `docker version`.

## Milestone 3, Step 1: create our first database table

A **migration** is a versioned SQL change to the database structure. The new files are:

- `migrations/000001_create_virtual_machines.up.sql` — creates the table and data constraints.
- `migrations/000001_create_virtual_machines.down.sql` — a destructive rollback that drops the table and **all VM data**. Do not run this casually.

The table has columns:

| Column | Type | Purpose |
| --- | --- | --- |
| `id` | `UUID` | Unique identifier, generated by PostgreSQL |
| `name` | `VARCHAR(100)` | Human-readable VM name; cannot be blank |
| `cpu_cores` | `INTEGER` | 1 to 64 cores |
| `memory_mb` | `INTEGER` | 512 to 262144 MB, in steps of 512 |
| `region` | `VARCHAR(50)` | Cloud region, cannot be blank |
| `status` | `VARCHAR(20)` | `stopped` by default, or starting/running/stopping/error |
| `created_at` | `TIMESTAMPTZ` | Creation time (UTC-capable timestamp) |
| `updated_at` | `TIMESTAMPTZ` | Update time (our Go code will maintain this later) |

The `CHECK` constraints enforce basic validity even if a different application writes to the database. We intentionally postpone user ownership and lifecycle-event tables until those features are introduced; later migrations will extend this schema.

### 1. Get the branch and start PostgreSQL

Run from the repository root:

```bash
git fetch origin
git switch --track origin/milestone/03-database-design
docker compose up -d
```

If you already have the local branch, use `git switch milestone/03-database-design` instead. Your existing `.env` file from Milestone 2 stays local and is reused. If it is missing, copy `.env.example` to `.env` before starting Compose.

### 2. Apply the migration exactly once

```bash
docker compose exec -T db psql -X -v ON_ERROR_STOP=1 -U cloud_resource -d cloud_resource < migrations/000001_create_virtual_machines.up.sql
```

Expected psql output includes `BEGIN`, `CREATE TABLE`, `COMMIT`. The `-T` flag disables a pseudo-terminal so shell redirection of the SQL file works. `ON_ERROR_STOP=1` makes psql return a failure on SQL errors.

**Important:** This initial exercise uses manual migrations and has no migration tracking yet. Running the same `CREATE TABLE` again will report `relation "virtual_machines" already exists`. Do not use the down migration just to make it rerunnable; we will add proper migration version tracking later.

### 3. Inspect the schema

```bash
docker compose exec db psql -U cloud_resource -d cloud_resource -c '\d virtual_machines'
```

Look for the eight columns, a UUID primary key, and the named `CHECK` constraints.

### 4. Insert valid data

```bash
docker compose exec db psql -U cloud_resource -d cloud_resource -c "INSERT INTO virtual_machines (name, cpu_cores, memory_mb, region) VALUES ('web-server-01', 2, 4096, 'eu-central-1') RETURNING id, name, status;"
```

Expected: one row containing a generated UUID, `web-server-01`, and default status `stopped`.

Confirm it was saved:

```bash
docker compose exec db psql -U cloud_resource -d cloud_resource -c "SELECT name, cpu_cores, memory_mb, region, status FROM virtual_machines;"
```

### 5. Try an invalid record

```bash
docker compose exec db psql -U cloud_resource -d cloud_resource -c "INSERT INTO virtual_machines (name, cpu_cores, memory_mb, region) VALUES ('invalid-vm', 0, 4096, 'eu-central-1');"
```

**Expected failure:** `violates check constraint "vm_cpu_cores_valid"`. PostgreSQL refuses to store an impossible CPU allocation.

### 6. Confirm the Docker volume retains the data

```bash
docker compose down
docker compose up -d
docker compose exec db psql -U cloud_resource -d cloud_resource -c "SELECT name, status FROM virtual_machines;"
```

The inserted VM should still exist. The data is held in the named `postgres_data` volume and is not removed by `docker compose down`.

**Do not use `docker compose down -v`** for this experiment, because that removes the database volume.

### Next within Milestone 3

After you verify the schema and understand the constraints, we'll introduce database migrations with tracking and the first Go repository operation. The VM REST API itself remains Milestone 4.

## Current repository structure

```text
cmd/api/main.go              # Server startup and database connection
internal/httpapi/router.go   # Health and readiness HTTP handlers
go.mod / go.sum              # Go dependencies and verified module checksums
Dockerfile                   # Multi-stage build and non-root runtime
.dockerignore                # Excludes unnecessary build context
compose.yaml                 # API, PostgreSQL, health check and named volume
.env.example                 # Example local development credentials
.gitignore                   # Excludes .env and local files from Git
README.md                    # All documentation and progress tracking
migrations/000001_*.sql      # Versioned VM table up/down migrations
```

## Future architecture

As we add cloud resource features, the architecture will evolve into:

```text
REST handlers -> application services -> repository layer -> PostgreSQL
```

A background worker inside the same Go application will eventually perform simulated VM lifecycle transitions. We do **not** plan Kubernetes, a microservice fleet, billing or real VM provisioning in the initial scope.
