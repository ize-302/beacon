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
// Usage:
//   k6 run loadtest/k6/api_load.js
//   BASE_URL=http://127.0.0.1:8081 FLEET_SIZE=2000 BATCH_SIZE=300 \
//     BATCH_RATE=20 DURATION=3m k6 run loadtest/k6/api_load.js
//
// Env vars (all optional):
//   BASE_URL    API root, default http://127.0.0.1:8081
//   FLEET_SIZE  vehicle id range assumed to already exist, default 500
//   BATCH_SIZE  points per gps-points/batch request, default 200
//   BATCH_RATE  batches/sec for the write scenario, default 10
//   READ_RATE   requests/sec for each read scenario, default 20
//   CHURN_VUS   concurrent VUs creating/deleting vehicles, default 5
//   DURATION    duration of the steady-state phase, default 2m

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8081';
const FLEET_SIZE = parseInt(__ENV.FLEET_SIZE || '500', 10);
const BATCH_SIZE = parseInt(__ENV.BATCH_SIZE || '200', 10);
const BATCH_RATE = parseInt(__ENV.BATCH_RATE || '10', 10);
const READ_RATE = parseInt(__ENV.READ_RATE || '20', 10);
const CHURN_VUS = parseInt(__ENV.CHURN_VUS || '5', 10);
const DURATION = __ENV.DURATION || '2m';

// Roughly Lagos's bounding box, matching the map the simulator drives on.
// Points don't need to be routable here — this scenario tests the write
// path, not the road graph.
const LAT_MIN = 6.39, LAT_MAX = 6.70;
const LON_MIN = 3.15, LON_MAX = 3.55;
const VEHICLE_TYPES = ['car', 'bus', 'truck', 'van'];

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
    'http_req_duration{name:gps_batch_write}': ['p(95)<500'],
    'http_req_duration{name:vehicle_list}': ['p(95)<300'],
    'http_req_duration{name:vehicle_history}': ['p(95)<300'],
    gps_batch_write_failures: ['rate<0.01'],
    vehicle_create_failures: ['rate<0.01'],
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

export function gpsBatchWrite() {
  const now = Date.now();
  const points = [];
  for (let i = 0; i < BATCH_SIZE; i++) {
    points.push({
      vehicle_id: randomInt(1, FLEET_SIZE),
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

export function vehicleHistoryRead() {
  const id = randomInt(1, FLEET_SIZE);
  const res = http.get(`${BASE_URL}/api/v1/vehicles/${id}/history`, {
    tags: { name: 'vehicle_history' },
  });
  // A gap between FLEET_SIZE and actual seeded rows is expected; only count
  // real server failures.
  const ok = check(res, { 'vehicle history not 5xx': (r) => r.status < 500 });
  historyReadFailRate.add(!ok);
}

export function vehicleChurn() {
  const plate = `LT-${__VU}-${__ITER}-${Date.now()}`;
  const createRes = http.post(
    `${BASE_URL}/api/v1/vehicles`,
    JSON.stringify({
      plate_number: plate,
      vehicle_type: VEHICLE_TYPES[randomInt(0, VEHICLE_TYPES.length - 1)],
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
