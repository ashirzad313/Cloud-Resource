# Cloud Resource Management Platform

A portfolio-focused Go backend project that simulates cloud VM resource management. This is a learning-by-running project: one manageable step at a time, with a single cumulative README, direct commits to `main`, and manual verification at each milestone. No real VMs are provisioned.

## Project status and roadmap

**Milestone 1 complete:** Go HTTP server, routing and JSON health endpoint.

**Milestone 2 complete:** Dockerized Go API, PostgreSQL with Docker Compose, database connectivity and `/ready` manually verified.

**Milestone 3 complete:** VM schema, repository reads/writes, service validation, GET/POST endpoints and automatic migration tracking manually verified.

**Milestone 4 in progress:** individual VM GET and DELETE endpoints integrated into `main`; manual verification pending.

- [x] **1. Go API foundation (2–3 h):** entry point, HTTP routing, JSON health endpoint.
- [x] **2. Docker and PostgreSQL (4–6 h):** multi-stage Dockerfile, Compose networking, PostgreSQL connection and persistent volume; manually verified.
- [x] **3. Database design (6–8 h):** schema, constraints, migrations and repository layer.
- [ ] **4. VM REST API (8–10 h):** create, list, get and delete VM records. *(Create/list verified in Milestone 3; get/delete and status handling ready to verify.)*
- [ ] **5. VM lifecycle (8–12 h):** state transitions, worker and event history.
- [ ] **6. Authentication (8–12 h):** account signup, password hashing, tokens, ownership.
- [ ] **7. API improvements (4–6 h):** validation, pagination, filtering and errors.
- [ ] **8. Quality and verification (8–10 h):** manual integration and failure tests; revisit automated tests for regression prevention.
- [ ] **9. Delivery and CI/CD (6–9 h):** production Docker build, GitHub Actions, API documentation.
- [ ] **10. Polish (4–6 h):** logging, graceful shutdown, security and CV/interview-ready explanation.

**Estimated total: 58–82 learning hours.** For this solo learning project, work directly on `main` with descriptive commits and manual verification at each step. Feature branches are optional for risky experiments; do not mark a milestone complete before verification.

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
  |     GET /vms    -> lists VM records
  |     POST /vms   -> validates and creates a VM
  |     GET /vms/{id} -> retrieves an individual VM
  |     DELETE /vms/{id} -> deletes a stopped VM
  |         |
  |         | DATABASE_URL uses host "db", port 5432
  |         v
  +-- db service (PostgreSQL 17, internal port 5432)
          |
          +-- postgres_data (named Docker volume)
```

Docker Compose creates a private project network. The API can resolve `db` by service name; `localhost` inside the API would refer to the API container, not PostgreSQL. **PostgreSQL port 5432 is intentionally not published on the Mac**, so use `docker compose exec db` for local database inspection.

At API startup, embedded SQL migrations create or upgrade the database schema before the HTTP server starts. The `/ready` endpoint checks database connectivity, not continuous schema correctness.

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

## Milestone 3, Step 1: create our first database table (historical exercise)

**Do not execute the manual CREATE TABLE exercise on a new installation now.** Step 4 below applies migrations automatically. These commands document how we learned the initial schema.

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

### 1. Update the repository and start PostgreSQL

Run from the repository root:

```bash
git switch main
git pull --ff-only origin main
docker compose up -d
```

Your existing `.env` file from Milestone 2 stays local and is reused. If it is missing, copy `.env.example` to `.env` before starting Compose.

### 2. Apply the migration exactly once

```bash
docker compose exec -T db psql -X -v ON_ERROR_STOP=1 --single-transaction -U cloud_resource -d cloud_resource -f - < migrations/000001_create_virtual_machines.up.sql
```

Expected psql output includes `CREATE TABLE`. The `-T` flag disables a pseudo-terminal so shell redirection of the SQL file works. `ON_ERROR_STOP=1` makes psql return a failure on SQL errors.

**Historical note:** This exercise predates migration tracking. Running the same `CREATE TABLE` again reports that the relation already exists. Do not execute the down migration to make it rerunnable.

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

## Milestone 3, Step 2: Go repository reads (verified)

The VM domain model and PostgreSQL repository map SQL rows to Go structs. We manually verified `List(ctx)` using a temporary `vmcheck` executable. The temporary CLI and its Docker build instructions have now been removed; the real API replaces them.

## Milestone 3, Step 3: create and list VMs through REST

The API now wires a repository and a service together at startup:

```text
HTTP handler -> VM service -> repository interface -> PostgreSQL repository -> PostgreSQL
```

- The HTTP handler decodes JSON, rejects malformed or oversized requests, and selects HTTP status codes.
- The service validates names, CPU, memory and region using the same limits as our schema.
- The repository executes parameterized SQL (`$1`, `$2`, etc.) and returns PostgreSQL-generated IDs, status and timestamps.
- The model has JSON field tags for a consistent API response; the request type only accepts user-editable fields.

Authentication is not implemented yet. These endpoints are for local development only, bound to the Mac's loopback address by Compose. Do not expose this API publicly. Listing currently returns all records; ownership and pagination are planned.

### 1. Get the latest code and rebuild

```bash
git switch main
git pull --ff-only origin main
docker compose up --build -d
```

Keep the existing `.env` and PostgreSQL volume. Do not reapply the first migration if your table already exists. For fresh databases, Step 4 automatically creates the table.

### 2. List existing VMs

```bash
curl -i http://localhost:8080/vms
```

Expected: HTTP 200 with a JSON array containing the record created during Step 1 (or an empty array if no records exist).

### 3. Create a new VM

```bash
curl -i -X POST http://localhost:8080/vms \
  -H 'Content-Type: application/json' \
  -d '{"name":"database-server-01","cpu_cores":4,"memory_mb":8192,"region":"eu-central-1"}'
```

Expected: HTTP 201 Created, with the saved VM including a UUID, status `stopped`, and creation/update timestamps. The API simulates resources; no actual VM is provisioned. Each successful POST inserts a new record.

### 4. Verify service validation

```bash
curl -i -X POST http://localhost:8080/vms \
  -H 'Content-Type: application/json' \
  -d '{"name":"invalid-vm","cpu_cores":0,"memory_mb":8192,"region":"eu-central-1"}'
```

Expected: HTTP 400 Bad Request with a message that CPU cores must be between 1 and 64. The invalid row must not be inserted.

### 5. Compare API results with PostgreSQL

```bash
curl -i http://localhost:8080/vms
docker compose exec db psql -U cloud_resource -d cloud_resource -c "SELECT name, cpu_cores, memory_mb, region, status FROM virtual_machines;"
```

Both should show the newly created VM and any previous valid records.

If the API returns 500, inspect `docker compose logs api`. A missing-table error indicates a schema problem; check the startup migration logs. If the API was started without a configured database, VM routes return 503.

## Milestone 3, Step 4: automatic migration tracking

`migrations/migrate.go` embeds the forward SQL files using Go's `//go:embed` directive. The runner sorts the zero-padded filenames, obtains a PostgreSQL transaction-level advisory lock, creates a `schema_migrations` history table if needed, then executes and records pending migrations in one transaction. If any step fails, the transaction rolls back and the API does not start.

Only `*.up.sql` files are run. The runner owns the transaction, so forward migrations must not contain their own BEGIN/COMMIT or nontransactional SQL. Treat applied migration files as immutable and add new numbered files for future schema changes. This is a small forward-only runner for this project; it does not implement checksums or automatic rollback/down migrations.

### Existing development database: baseline once

If you manually applied migration 000001 in Step 1, you need to explicitly tell the runner that it is already applied. This is called **baselining**. Do not use these commands on a fresh database; the API will create its schema automatically.

```bash
git switch main
git pull --ff-only origin main
docker compose up -d --wait db
docker compose exec db psql -U cloud_resource -d cloud_resource -c '\d virtual_machines'
```

Confirm the existing table matches the eight columns and constraints in Step 1 before recording the baseline. The following command does not create or modify VM records. It refuses to baseline if the table is absent; it does not automatically compare the full schema.

```bash
docker compose exec -T db psql -X -v ON_ERROR_STOP=1 -U cloud_resource -d cloud_resource <<'SQL'
BEGIN;
SELECT pg_advisory_xact_lock(74632001);
DO $baseline$
BEGIN
    IF to_regclass('public.virtual_machines') IS NULL THEN
        RAISE EXCEPTION 'VM table is missing; do not baseline a fresh database';
    END IF;
END;
$baseline$;
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO schema_migrations (version)
VALUES ('000001_create_virtual_machines.up.sql')
ON CONFLICT (version) DO NOTHING;
COMMIT;
SQL
```

Then rebuild the API:

```bash
docker compose up --build -d
docker compose logs api
curl -i http://localhost:8080/vms
docker compose exec db psql -U cloud_resource -d cloud_resource -c "SELECT * FROM schema_migrations;"
```

Expected: API logs include `database migrations up to date`, existing VMs remain, and the history contains exactly one entry for migration 000001 at this milestone. Restart the API (`docker compose restart api`) and confirm the history is unchanged. Do not delete the volume.

### Fresh installation

With a new database, there is no baseline step. Copy `.env.example` to `.env` if needed, then run `docker compose up --build -d`. Expected startup logs include `applied migration: 000001_create_virtual_machines.up.sql`; `GET /vms` initially returns an empty array.

Optional separate fresh-database verification: stop the original API to free port 8080, then use a distinct Compose project name that has not been used before:

```bash
docker compose stop api
docker compose -p cloud-resource-migration-check up --build -d
docker compose -p cloud-resource-migration-check logs api
curl -i http://localhost:8080/vms
docker compose -p cloud-resource-migration-check down
docker compose up -d
```

This uses a separate named volume, leaving your original VM data intact. The check project's volume is retained after down. If startup fails, inspect API logs before making any changes to the database.

Milestone 3 verified by the project owner. Milestone 4 finishes the remaining VM REST operations.

## Milestone 4: individual VM REST resources

The existing collection routes are joined by individual resource routes. The service validates UUID format, the repository maps missing rows to a domain error, and HTTP handlers translate errors into status codes. No new dependency or database migration is required.

| Method and path | Success | Expected errors |
| --- | --- | --- |
| GET /vms | 200, JSON array | 503 if database not configured; 500 for unexpected storage errors |
| POST /vms | 201, JSON object and Location header | 400 invalid input, 413 oversized body |
| GET /vms/{id} | 200, JSON object | 400 malformed UUID, 404 missing VM |
| DELETE /vms/{id} | 204, empty body | 400 malformed UUID, 404 missing VM, 409 VM not stopped |

All routes remain local-development-only until authentication and ownership checks are implemented. Deletion permanently removes the simulated VM record. The repository obtains a row lock with SELECT ... FOR UPDATE and checks the status inside the same transaction as the DELETE. This prevents concurrent updates of that row from invalidating the status check. Only stopped VMs may be deleted.

### Manual verification

Update `main` and rebuild:

```bash
git switch main
git pull --ff-only origin main
docker compose up --build -d
```

Keep the existing database volume.

Create a **disposable** VM and copy the UUID from its response (or its Location header):

```bash
curl -i -X POST http://localhost:8080/vms \
  -H 'Content-Type: application/json' \
  -d '{"name":"delete-check","cpu_cores":2,"memory_mb":4096,"region":"eu-central-1"}'
```

Replace the example value below with that UUID.

```bash
VM_ID='PASTE-UUID-HERE'
curl -i "http://localhost:8080/vms/$VM_ID"
curl -i http://localhost:8080/vms/not-a-uuid
```

Expected: 200 for the disposable VM and 400 for the malformed ID.

To verify the deletion guard before the lifecycle API exists, temporarily set **only the disposable VM** to running using psql. This is a local manual test, not a real VM start operation:

```bash
docker compose exec -T db psql -X -v ON_ERROR_STOP=1 -U cloud_resource -d cloud_resource -v vm_id="$VM_ID" <<'SQL'
UPDATE virtual_machines SET status = 'running', updated_at = NOW()
WHERE id = :'vm_id'::uuid AND name = 'delete-check';
SQL
curl -i -X DELETE "http://localhost:8080/vms/$VM_ID"
curl -i "http://localhost:8080/vms/$VM_ID"
```

Expected: UPDATE 1, DELETE returns 409, GET returns 200 and status running. Return the disposable VM to stopped, then delete it:

```bash
docker compose exec -T db psql -X -v ON_ERROR_STOP=1 -U cloud_resource -d cloud_resource -v vm_id="$VM_ID" <<'SQL'
UPDATE virtual_machines SET status = 'stopped', updated_at = NOW()
WHERE id = :'vm_id'::uuid AND name = 'delete-check';
SQL
curl -i -X DELETE "http://localhost:8080/vms/$VM_ID"
curl -i "http://localhost:8080/vms/$VM_ID"
curl -i -X DELETE "http://localhost:8080/vms/$VM_ID"
curl -i http://localhost:8080/vms
```

Expected: UPDATE 1, DELETE 204 with no body, subsequent GET and DELETE 404, and list 200 with the original records intact. If UPDATE does not report 1, verify you copied the UUID of the disposable VM before continuing. Repeat GET /health and GET /ready to check previous behavior.

Milestone 5 will introduce lifecycle actions and asynchronous state changes.

## Current repository structure

```text
cmd/api/main.go              # Server startup and database connection
internal/vm/vm.go            # VM domain model
internal/vm/service.go       # Business rules and repository interface
internal/storage/postgres/vm_repository.go # SQL query and row mapping
internal/httpapi/router.go   # HTTP route registration and health/readiness
internal/httpapi/vm_handler.go # JSON request and response handling
go.mod / go.sum              # Go dependencies and verified module checksums
Dockerfile                   # Multi-stage build and non-root runtime
.dockerignore                # Excludes unnecessary build context
compose.yaml                 # API, PostgreSQL, health check and named volume
.env.example                 # Example local development credentials
.gitignore                   # Excludes .env and local files from Git
README.md                    # All documentation and progress tracking
migrations/000001_*.sql      # Versioned VM table up/down migrations
migrations/migrate.go        # Embedded migration runner and version tracking
```

## Future architecture

The VM API now follows this architecture:

```text
REST handlers -> application services -> repository layer -> PostgreSQL
```

A background worker inside the same Go application will eventually perform simulated VM lifecycle transitions. We do **not** plan Kubernetes, a microservice fleet, billing or real VM provisioning in the initial scope.
