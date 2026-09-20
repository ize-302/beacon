// Load-tests Beacon's REST surface directly, bypassing the Go simulator so
// results reflect API + Postgres capacity rather than the simulator's own
// route-planning CPU cost.
//
// Scenarios run concurrently:
//   - gps_batch_write : POST /gps-points/batch at a fixed batch rate
//                       (mirrors the simulator's real write shape)
//   - vehicle_list     : GET /vehicles  (joins each vehicle's last coordinate)
//   - vehicle_history  : GET /vehicles/{id}/history
//   - vehicle_churn    : POST /vehicles then DELETE it shortly after
//
// setup() fetches the live vehicle list once before the run and every
// scenario samples real IDs from it — gps_points has a foreign key on
// vehicle_id, and the batch insert is one atomic statement, so a single
// point referencing an ID that doesn't exist fails the *entire* batch, not
// just that point. Assuming a dense 1..N id range (this script's original
// approach) breaks the moment any vehicle has ever been deleted, and on a
// live, churning fleet that's the common case, not the edge case.
//
// Usage:
//   k6 run loadtest/k6/api_load.js
//   BASE_URL=http://127.0.0.1:8081 BATCH_SIZE=300 \
//     BATCH_RATE=20 DURATION=3m k6 run loadtest/k6/api_load.js
//
// Env vars (all optional):
//   BASE_URL      API root, default http://127.0.0.1:8081
//   BATCH_SIZE    points per gps-points/batch request, default 200
//   BATCH_RATE    batches/sec for the write scenario, default 10
//   READ_RATE     requests/sec for each read scenario, default 20
//   CHURN_VUS     concurrent VUs creating/deleting vehicles, default 5
//   DURATION      duration of the steady-state phase, default 2m
//   P95_WRITE_MS  pass/fail p95 latency threshold for the write scenario
//                 (ms), default 500
//   P95_READ_MS   pass/fail p95 latency threshold for both read scenarios
//                 (ms), default 300
//   MAX_FAIL_RATE pass/fail failure-rate threshold for writes and vehicle
//                 creation, default 0.01
//
// The threshold defaults were tuned against localhost latency. Testing a
// live remote deployment adds real network RTT on top of that, so raise
// these to match the environment rather than leaving the localhost numbers
// in place — otherwise a healthy deployment can false-fail on latency alone.

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8081';
const BATCH_SIZE = parseInt(__ENV.BATCH_SIZE || '200', 10);
const BATCH_RATE = parseInt(__ENV.BATCH_RATE || '10', 10);
const READ_RATE = parseInt(__ENV.READ_RATE || '20', 10);
const CHURN_VUS = parseInt(__ENV.CHURN_VUS || '5', 10);
const DURATION = __ENV.DURATION || '2m';
const P95_WRITE_MS = parseInt(__ENV.P95_WRITE_MS || '500', 10);
const P95_READ_MS = parseInt(__ENV.P95_READ_MS || '300', 10);
const MAX_FAIL_RATE = parseFloat(__ENV.MAX_FAIL_RATE || '0.01');

// Roughly Lagos's bounding box, matching the map the simulator drives on.
// Points don't need to be routable here — this scenario tests the write
// path, not the road graph.
const LAT_MIN = 6.39, LAT_MAX = 6.70;
const LON_MIN = 3.15, LON_MAX = 3.55;

const batchWriteFailRate = new Rate('gps_batch_write_failures');
const vehicleCreateFailRate = new Rate('vehicle_create_failures');
const vehicleReadFailRate = new Rate('vehicle_read_failures');
const historyReadFailRate = new Rate('vehicle_history_failures');
const insertedPerBatch = new Trend('gps_batch_inserted');

export const options = {
  scenarios: {
    gps_batch_write: {
      executor: 'constant-arrival-rate',
      exec: 'gpsBatchWrite',
      rate: BATCH_RATE,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: Math.max(10, BATCH_RATE * 2),
      maxVUs: Math.max(50, BATCH_RATE * 10),
    },
    vehicle_list: {
      executor: 'constant-arrival-rate',
      exec: 'vehicleListRead',
      rate: READ_RATE,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: Math.max(5, READ_RATE),
      maxVUs: Math.max(20, READ_RATE * 5),
    },
    vehicle_history: {
      executor: 'constant-arrival-rate',
      exec: 'vehicleHistoryRead',
      rate: READ_RATE,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: Math.max(5, READ_RATE),
      maxVUs: Math.max(20, READ_RATE * 5),
    },
    vehicle_churn: {
      executor: 'constant-vus',
      exec: 'vehicleChurn',
      vus: CHURN_VUS,
      duration: DURATION,
    },
  },
  thresholds: {
    'http_req_duration{name:gps_batch_write}': [`p(95)<${P95_WRITE_MS}`],
    'http_req_duration{name:vehicle_list}': [`p(95)<${P95_READ_MS}`],
    'http_req_duration{name:vehicle_history}': [`p(95)<${P95_READ_MS}`],
    gps_batch_write_failures: [`rate<${MAX_FAIL_RATE}`],
    vehicle_create_failures: [`rate<${MAX_FAIL_RATE}`],
  },
};

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function randomLat() {
  return LAT_MIN + Math.random() * (LAT_MAX - LAT_MIN);
}

function randomLon() {
  return LON_MIN + Math.random() * (LON_MAX - LON_MIN);
}

// Runs once, in a single VU, before any scenario starts. Its return value is
// handed to every scenario function as `data`.
export function setup() {
  const res = http.get(`${BASE_URL}/api/v1/vehicles`);
  if (res.status !== 200) {
    throw new Error(`setup: GET /vehicles returned ${res.status}, cannot build a valid vehicle id pool`);
  }
  const vehicleIds = (res.json('data') || [])
    .map((v) => v.id)
    .filter((id) => id != null);
  if (vehicleIds.length === 0) {
    throw new Error('setup: no vehicles exist yet — create some (or let the simulator run a moment) before load-testing gps_batch_write/vehicle_history');
  }
  return { vehicleIds };
}

function randomVehicleId(data) {
  return data.vehicleIds[randomInt(0, data.vehicleIds.length - 1)];
}

export function gpsBatchWrite(data) {
  const now = Date.now();
  const points = [];
  for (let i = 0; i < BATCH_SIZE; i++) {
    points.push({
      vehicle_id: randomVehicleId(data),
      bearing: Math.random() * 360,
      latitude: randomLat(),
      longitude: randomLon(),
      // spread timestamps slightly so a batch isn't perfectly simultaneous,
      // matching what the real sender accumulates over its flush interval.
      timestamp: now - randomInt(0, 200),
    });
  }

  const res = http.post(
    `${BASE_URL}/api/v1/gps-points/batch`,
    JSON.stringify({ points }),
    { headers: { 'Content-Type': 'application/json' }, tags: { name: 'gps_batch_write' } }
  );

  const ok = check(res, { 'gps batch 201': (r) => r.status === 201 });
  batchWriteFailRate.add(!ok);
  if (ok) {
    insertedPerBatch.add(res.json('data.inserted'));
  }
}

export function vehicleListRead() {
  const res = http.get(`${BASE_URL}/api/v1/vehicles`, { tags: { name: 'vehicle_list' } });
  const ok = check(res, { 'vehicle list 200': (r) => r.status === 200 });
  vehicleReadFailRate.add(!ok);
}

export function vehicleHistoryRead(data) {
  const id = randomVehicleId(data);
  const res = http.get(`${BASE_URL}/api/v1/vehicles/${id}/history`, {
    tags: { name: 'vehicle_history' },
  });
  const ok = check(res, { 'vehicle history not 5xx': (r) => r.status < 500 });
  historyReadFailRate.add(!ok);
}

export function vehicleChurn() {
  const plate = `LT-${__VU}-${__ITER}-${Date.now()}`;
  const createRes = http.post(
    `${BASE_URL}/api/v1/vehicles`,
    JSON.stringify({
      plate_number: plate,
    }),
    { headers: { 'Content-Type': 'application/json' }, tags: { name: 'vehicle_create' } }
  );

  const created = check(createRes, { 'vehicle create 201': (r) => r.status === 201 });
  vehicleCreateFailRate.add(!created);
  if (!created) {
    sleep(1);
    return;
  }

  const id = createRes.json('data.id');
  sleep(randomInt(1, 3));

  const delRes = http.del(`${BASE_URL}/api/v1/vehicles/${id}`, null, {
    tags: { name: 'vehicle_delete' },
  });
  check(delRes, { 'vehicle delete 204': (r) => r.status === 204 });
}
