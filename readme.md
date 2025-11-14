# Event Ingestion & Reporting Service

## Overview
This project implements a backend system to capture website analytics events. The system is designed to handle a **high volume of ingestion requests** very quickly using asynchronous processing and provides a **reporting API** to retrieve aggregated data.

The system consists of three main components:

1. **Ingestion API** – Receives events and pushes them to a queue.
2. **Worker (Processor)** – Processes queued events and writes them to the database.
3. **Reporting API** – Reads from the database to provide summary statistics.

---

## Architecture Decision

### Asynchronous Event Processing
- Events are ingested via the `/event` POST endpoint.
- Instead of writing directly to the database (which is slow under heavy load), events are pushed into a **Redis queue** immediately.
- A separate **worker service** continuously pulls events from Redis and writes them to Postgres.
- This design ensures that the ingestion endpoint is extremely fast and non-blocking, allowing high throughput (thousands of requests per second) without waiting for database writes.

**Flow Diagram:**
```
[Client] --> POST /event --> [Redis Queue] --> [Worker] --> [Postgres DB]
```

- Redis acts as an in-memory queue to decouple ingestion from processing.
- Postgres stores the finalized events for reporting purposes.

---

## Database Schema

**Table:** `events`

| Column      | Type        | Description                     |
|-------------|-------------|---------------------------------|
| id          | BIGSERIAL   | Primary key                     |
| site_id     | TEXT        | ID of the website               |
| event_type  | TEXT        | Type of event (e.g., page_view) |
| path        | TEXT        | Page path where event occurred  |
| user_id     | TEXT        | User identifier                 |
| timestamp   | TIMESTAMPTZ | Event timestamp                 |

**Indexes:**
- `idx_events_site_timestamp` on `(site_id, timestamp)` for fast date-based queries.
- `idx_events_path` on `(path)` for aggregating popular paths.

---

## Setup Instructions

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Redis & Postgres (run via Docker Compose)

### Steps

1. Clone the repository:
```bash
git clone https://github.com/shivamsahu-tech/event-ingestions.git
cd event-ingestion
```

2. Start Redis and Postgres using Docker Compose:
```bash
docker-compose up -d
```

3. Run database migrations:
```bash
./scripts/run_migrations.sh
```

4. Start the services:

```bash
./scripts/start.sh
```



5. Verify services:
   - Ingestion API: `http://localhost:8080/event`
   - Reporting API: `http://localhost:8080/stats`

---

## API Usage

### 1. POST /event

**Description:** Capture an event.

**Request:**
```bash
curl -X POST http://localhost:8080/event \
-H "Content-Type: application/json" \
-d '{
  "site_id": "test123",
  "event_type": "page_view",
  "path": "/home",
  "user_id": "user1"
}'
```

**Response:**

- HTTP 204 No Content (immediate acknowledgment)

### 2. GET /stats

**Description:** Retrieve aggregated stats for a site and date.

**Request:**
```bash
curl "http://localhost:8080/stats?site_id=test123&date=2025-11-14"
```

**Sample Response:**
```json
{
  "site_id": "test123",
  "date": "2025-11-14",
  "total_views": 1005,
  "unique_users": 4,
  "top_paths": [
    { "path": "/home", "views": 1000 },
    { "path": "/pricing", "views": 5 }
  ]
}
```

---

## Performance Testing

Using `hey`, the ingestion API was benchmarked:
```bash
hey -n 5000 -c 100 -m POST -H "Content-Type: application/json" \
-d '{"site_id":"test123","event_type":"page_view","path":"/home","user_id":"userX"}' \
http://localhost:8080/event
```

**Results (sample on 8GB RAM, Ryzen 5 5500U):**
- Requests/sec: ~9585
- Average latency: 9.9 ms
- Status codes: 100% 204

The high throughput demonstrates that the API efficiently handles high-volume ingestion using the asynchronous queue pattern.

---



## Project Structure
```
event-ingestion/
├── cmd/
│   ├── api/main.go       # Ingestion + Reporting API
│   └── worker/main.go    # Background worker
├── internal/
│   ├── db/               # Postgres connection
│   ├── models/           # Event model
│   └── queue/            # Redis queue utilities
├── migrations/           # SQL migrations
├── scripts/              # Run & migration scripts
├── docker-compose.yml    # Redis + Postgres setup
├── go.mod
├── go.sum
└── README.md
```

---
