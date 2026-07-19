# User Management Service

A high-performance, production-grade User Management Microservice built with **Go 1.25+**.

This project demonstrates **Advanced Backend Engineering** concepts including Clean Architecture, Distributed Caching, Event-Driven Messaging, and Resiliency Patterns — with a React frontend on top.

## Live Demo

- **App**: [usersvc.vercel.app](https://usersvc.vercel.app) — React frontend (Vercel)
- **API**: [usersvc.onrender.com](https://usersvc.onrender.com/health) — Go API + worker (Render), backed by Neon Postgres, Upstash Redis, and CloudAMQP

> The API sleeps after 15 idle minutes on the free tier — the first request may take ~30–60s to wake it.

---

## Architecture & Tech Stack

| Component | Technology | Role |
| :--- | :--- | :--- |
| **Language** | **Go (Golang)** | Core Logic (Standard Lib + optimized drivers) |
| **Database** | **PostgreSQL** | Primary Data Store (Source of Truth) |
| **Driver** | **pgx** | High-performance PostgreSQL driver |
| **Cache** | **Redis** | Distributed Cache (Cache-Aside Pattern) |
| **Messaging** | **RabbitMQ** | Asynchronous Event Bus (AMQP) |
| **Resiliency** | **Singleflight** | Cache Stampede Protection |
| **Frontend** | **React + Vite** | Admin UI (served by nginx) |
| **Deployment** | **Docker Compose** (local) / **Render + Vercel** (cloud) | Container Orchestration |

---

## Key Features

### 1. Clean Architecture
Separation of concerns is mostly strict:
-   **Transport**: HTTP Handlers & Middleware.
-   **Service**: Business Logic (Hashing, Caching, Event Publishing) — depends only on interfaces.
-   **Repository**: Raw SQL Queries (`database/sql`).
-   **Domain**: Pure Go structs.

### 2. Performance Optimizations
-   **Redis Cache-Aside**: Reads check Redis first (Sub-millisecond latency). Falls back to DB only on miss.
-   **Singleflight**: Protects the database from "Cache Stampedes". If 10,000 requests hit a cold cache simultaneously, only **ONE** DB query is fired. The result is shared with all 10,000 waiters.

### 3. Event-Driven (Async)
-   **Critical Path**: DB Creation is synchronous (User gets 201 Created immediately).
-   **Background Path**: RabbitMQ Event is published. A separate **Worker Process** consumes this event to handle slow tasks (e.g., Welcome Emails, Analytics) without blocking the API. Malformed messages are Nacked (no requeue) instead of silently dropped.
-   On single-process hosts, `EMBED_WORKER=true` runs the consumer inside the API binary.

### 4. Graceful Shutdown
-   Both the API (`http.Server.Shutdown` with a timeout context) and the worker intercept `SIGTERM`/`SIGINT`.
-   In-flight requests/tasks finish before exit — no dropped jobs during deployments.

### 5. Typed Errors
-   Sentinel errors (`ErrUserNotFound`, `ErrDuplicateEmail`) matched with `errors.Is` — no string matching.
-   Duplicate emails are detected via the Postgres driver's error type (SQLSTATE 23505), surfaced as **409 Conflict**.

### 6. Tested
-   Unit tests with hand-written fakes (no mocking frameworks), run with `go test ./... -race`.
-   Includes a concurrency test asserting that N concurrent cache misses collapse into **exactly one** repository call.

---

## Benchmark

`GET /users/{id}` (cached path), 100 concurrent connections for 10s — measured on a 2-core i3-1115G4 laptop running the **entire stack + the load generator** simultaneously:

```
Statistics        Avg      Stdev        Max
  Reqs/sec      4202.09    1875.94    8938.19
  Latency       23.75ms    13.72ms   283.11ms
  Latency Distribution
     50%    20.57ms
     75%    28.64ms
     90%    38.17ms
     95%    47.16ms
     99%    73.13ms
  HTTP codes:  2xx - 42140, 4xx - 0, 5xx - 0
  Throughput:  1.39MB/s
```

> Honest note: this is the cache-read path — it shows the ceiling of the system. Hammering writes or unique ids each time would settle lower, because then bcrypt, marshalling, and DB R/W get involved. Zero errors across 42K requests is the singleflight + cache-aside design doing its job.

Reproduce it yourself:

```bash
docker run --rm alpine/bombardier -c 100 -d 10s -l http://host.docker.internal:8080/users/{USER_ID}
```

---

## Getting Started

### Prerequisites
-   Docker & Docker Compose

### Run the Stack

```bash
# Starts API, Worker, Frontend, Postgres, Redis, RabbitMQ
docker-compose up --build
```

-   API: `localhost:8080`
-   Frontend: `localhost:3000`
-   RabbitMQ management UI: `localhost:15672` (guest/guest)

---

## API Usage

| Method | Route | Result |
| :--- | :--- | :--- |
| `POST` | `/users` | `201` + user, `409` on duplicate email |
| `GET` | `/users` | `200` + all users |
| `GET` | `/users/{id}` | `200` + user, `404` if unknown |
| `DELETE` | `/users/{id}` | `204`, `404` if unknown |
| `GET` | `/health` | `200` `{"status":"ok"}` |

### 1. Create User

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email": "engineer@example.com", "password": "securePass123"}'
```

**Response**: `201 Created` with User ID.

### 2. Get User (Test Caching)

```bash
#1st call (db hit -> cache write will happen)
curl http://localhost:8080/users/{USER_ID}

#2nd call (redis hit - scary fast)
curl http://localhost:8080/users/{USER_ID}
```

### 3. Delete User

```bash
curl -X DELETE http://localhost:8080/users/{USER_ID}
```

**Effect**: Removes user from **both** PostgreSQL and Redis.

---

## Tests

```bash
go test ./... -race
```

Covers: cached reads bypass the repository, cache population on miss, singleflight collapsing N concurrent misses into one DB call, and handler status codes (404 unknown id, 409 duplicate email).

---

## Simulation Testing

### Verify Graceful Shutdown
1.  Navigate to `cmd/worker/main.go`.
2.  Raise the `time.Sleep` in the consumer loop (e.g. to 10s) to simulate a slow task.
3.  Trigger a user creation.
4.  Immediately run `docker stop <worker_container_id>`.
5.  **Observation**: The container will NOT stop immediately. It waits for the in-flight task to finish, then logs "worker stopped gracefully".

---

## Deployment

-   **[render.yaml](render.yaml)** — Render Blueprint for the API (Docker runtime, `/health` health check, `EMBED_WORKER=true` so the worker rides inside the API process on the free tier).
-   **[frontend/vercel.json](frontend/vercel.json)** — Vercel config: Vite build + a rewrite that proxies `/api/*` to the Render API (no CORS needed).
-   Backing services: Neon (Postgres), Upstash (Redis), CloudAMQP (RabbitMQ) — all free tiers.
-   Push to `main` → Render rebuilds the API and Vercel rebuilds the frontend automatically.

---

## Directory Structure

```
.
├── cmd/
│   ├── api/            # REST API Main entrypoint (+ optional embedded worker)
│   └── worker/         # Background Worker Main entrypoint
├── internal/
│   ├── config/         # Env loading
│   ├── domain/         # Struct Definitions
│   ├── infrastructure/ # Redis & RabbitMQ Clients
│   ├── repository/     # SQL Logic
│   ├── service/        # Business Logic (The Glue)
│   └── transport/      # HTTP Handlers
├── frontend/           # React + Vite admin UI (nginx-served in Docker)
├── migrations/         # SQL Migration files (Embedded)
├── docker-compose.yml  # Infrastructure definition
├── render.yaml         # Render deployment blueprint
└── Dockerfile          # Multi-stage build
```
