# Beacon

![Screenshot](screenshot.png)

Real-time vehicle tracking system. Simulates GPS-equipped vehicles moving across Lagos road network and streams their positions to a live map dashboard.

## How it works

```mermaid
flowchart LR
    Sim["Simulator\n(BFS + haversine)"]
    API["Backend API"]
    DB[("PostgreSQL")]
    WS["WebSocket Hub"]
    Dash["Dashboard\n(Mapbox)"]

    Sim -->|"POST /gps-points"| API
    API --> DB
    API --> WS
    WS -->|"lat, lng, bearing, timestamp"| Dash
    Dash -->|"GET /gps (initial load)"| API
```

1. **Backend** seeds the database and exposes a REST API for vehicles, GPS devices, and GPS points.
2. **Simulator** fetches all registered GPS devices, builds a road graph from Lagos OSM data, and moves each vehicle independently along BFS-computed paths. Sends the vehicle's new position, bearing, and timestamp to the API. New GPS devices are detected and picked up automatically using Server-Sent Events.
3. **API** saves each GPS point and broadcasts it over WebSocket.
4. **Dashboard** renders vehicle markers on a Mapbox map. Markers appear on first REST load (if a last coordinate exists) or on first WebSocket ping. Each marker rotates to face its direction of travel and holds the correct bearing as the map is rotated. Clicking a marker loads its GPS history and draws the route on the map; the route grows in real time as the vehicle moves.

## Stack

| Layer     | Tech                                                                          |
| --------- | ----------------------------------------------------------------------------- |
| Backend   | Go, `net/http`, `go-chi/chi`, `gorilla/websocket`, PostgreSQL, Huma           |
| Simulator | Go, BFS pathfinding, `paulmach/osm` (OSM PBF parsing)                        |
| Dashboard | SolidJS, Vite, Mapbox GL JS, TanStack Query, Axios, openapi-generator client |

## Getting started

### Prerequisites

- Docker + Docker Compose — for the containerised backend
- Go 1.26+ — only if running the backend outside Docker
- Node.js 18+ — the dashboard always runs on the host
- Mapbox access token
- The Lagos OSM PBF file at `backend/cmd/simulator/map_data/lagos.osm.pbf`

### Environment files

Copy `backend/.env.example` to `backend/.env`:

```env
DB_HOST=beacon_postgres_db
DB_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=beacon

PORT=8080
API_BASE_URL=http://127.0.0.1:8080
```

`DB_HOST`/`DB_PORT` above are written for **Docker**, where the backend reaches Postgres by
service name on the container network. To run the backend on the host instead, point them at the
published port — `DB_HOST=localhost` and `DB_PORT=5433` (see the ports table below).

Compose also reads this file to provision the database, so `POSTGRES_USER`, `POSTGRES_PASSWORD`,
and `POSTGRES_DB` must be set or `docker compose` refuses to start. In deployment, a single
`DATABASE_URL` overrides all five database variables.

Create a `.env` file in `dashboard/`:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
VITE_MAPBOX_ACCESS_TOKEN=your_mapbox_token
```

### Run the backend with Docker (recommended)

From `backend/`:

```bash
docker compose up --build
```

This starts three containers — Postgres, the API, and the simulator — each as a single process.
The API waits for the database to pass its health check, and the simulator waits for the API to
pass its own before it starts posting positions. Source is bind-mounted, so `air` live-reloads
both Go services on save.

```bash
docker compose up beacon_backend      # api + database only, no simulator
docker compose logs -f beacon_simulator
docker compose down                   # add -v to also drop the database volume
```

The dashboard is not containerised — run it on the host alongside the stack:

```bash
cd dashboard && npm install && npm run dev
```

### Run everything on the host

Set `DB_HOST=localhost` and `DB_PORT=5433` in `backend/.env` first, and make sure a Postgres is
listening there (`docker compose up beacon_db` will do).

```bash
# API, with live reload
cd backend && make api

# Simulator (separate terminal)
cd backend && make sim

# Dashboard (separate terminal)
cd dashboard && npm install && npm run dev
```

`make dev` runs the API and simulator together in one terminal.

Without live reload:

```bash
cd backend && go run ./cmd/api
cd backend && go run ./cmd/simulator
```

The simulator resolves the OSM file relative to the working directory, so run it from `backend/`.

### Ports

| Service   | In Docker                | From the host    |
| --------- | ------------------------ | ---------------- |
| API       | `beacon_backend:8080`    | `localhost:8080` |
| Postgres  | `beacon_postgres_db:5432`| `localhost:5433` |
| Dashboard | not containerised        | `localhost:5173` |

Postgres is published on **5433** to avoid clashing with a local Postgres on 5432.

## API

Full interactive docs available at [`http://localhost:8080/swagger/`](http://localhost:8080/swagger/) when the backend is running.

OpenAPI JSON schema is served at `/openapi.json` and is used to generate the typed dashboard client via `openapi-generator`.

WebSocket endpoint `/ws` streams `{ gps_id, latitude, longitude, bearing, timestamp }` for every position update.

SSE endpoint `/api/v1/gps-devices/events` streams newly registered GPS devices so the simulator picks them up automatically without polling.
